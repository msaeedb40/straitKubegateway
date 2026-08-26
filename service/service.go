// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package service implements the straitKubegateway eBPF service load balancer.
//
// Architectural invariant: Kubernetes Service state is compiled into eBPF maps
// so that packets never reach kube-proxy or a userspace proxy.
// All Service/EndpointSlice state is discovered dynamically from the API server.
package service

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"go.uber.org/zap"

	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/ir"
	sgtypes "github.com/straitkubegateway/straitkubegateway/pkg/types"
)

// ============================================================================
// Service Manager
// ============================================================================

// Manager manages Kubernetes service state and compiles it into IR.
type Manager struct {
	mu     sync.RWMutex
	client client.Client
	log    *zap.Logger
	// services maps namespace/name → Service IR
	services map[ir.ServiceKey]*ir.Service
}

// NewManager creates a new service manager.
func NewManager(c client.Client, log *zap.Logger) *Manager {
	return &Manager{
		client:   c,
		log:      log,
		services: make(map[ir.ServiceKey]*ir.Service),
	}
}

// ============================================================================
// Reconcile — called by the Service controller
// ============================================================================

// Reconcile fetches the current service and its endpoint slices, then
// compiles the result into the IR Service representation.
func (m *Manager) Reconcile(ctx context.Context, key types.NamespacedName) (*ir.Service, error) {
	var svc corev1.Service
	if err := m.client.Get(ctx, key, &svc); err != nil {
		return nil, client.IgnoreNotFound(err)
	}

	// Fetch endpoint slices for this service
	var esList discoveryv1.EndpointSliceList
	if err := m.client.List(ctx, &esList,
		client.InNamespace(key.Namespace),
		client.MatchingLabels{discoveryv1.LabelServiceName: key.Name},
	); err != nil {
		return nil, fmt.Errorf("list endpoint slices for %s: %w", key, err)
	}

	irSvc, err := compileService(&svc, esList.Items)
	if err != nil {
		return nil, fmt.Errorf("compile service %s: %w", key, err)
	}

	m.mu.Lock()
	m.services[ir.ServiceKey{Namespace: key.Namespace, Name: key.Name}] = irSvc
	m.mu.Unlock()

	m.log.Info("service reconciled",
		zap.String("namespace", key.Namespace),
		zap.String("name", key.Name),
		zap.String("clusterIP", irSvc.ClusterIP.String()),
		zap.Int("backends", len(irSvc.Backends)),
		zap.String("type", string(irSvc.Type)),
	)
	return irSvc, nil
}

// Delete removes a service from the manager.
func (m *Manager) Delete(key types.NamespacedName) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.services, ir.ServiceKey{Namespace: key.Namespace, Name: key.Name})
}

// GetAll returns all currently known services.
func (m *Manager) GetAll() map[ir.ServiceKey]*ir.Service {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[ir.ServiceKey]*ir.Service, len(m.services))
	for k, v := range m.services {
		out[k] = v
	}
	return out
}

// ============================================================================
// Service compiler: Kubernetes → IR
// ============================================================================

func compileService(svc *corev1.Service, slices []discoveryv1.EndpointSlice) (*ir.Service, error) {
	clusterIP, err := parseOptionalIP(svc.Spec.ClusterIP)
	if err != nil {
		return nil, fmt.Errorf("invalid ClusterIP %q: %w", svc.Spec.ClusterIP, err)
	}

	ports := make([]ir.ServicePort, 0, len(svc.Spec.Ports))
	for _, p := range svc.Spec.Ports {
		ports = append(ports, ir.ServicePort{
			Protocol:   mapProtocol(p.Protocol),
			Port:       uint16(p.Port),
			NodePort:   uint16(p.NodePort),
			TargetPort: uint16(p.TargetPort.IntVal),
		})
	}

	backends, err := compileBackends(slices)
	if err != nil {
		return nil, err
	}

	return &ir.Service{
		Key: ir.ServiceKey{
			Namespace: svc.Namespace,
			Name:      svc.Name,
		},
		Type:      mapServiceType(svc.Spec.Type),
		ClusterIP: clusterIP,
		Ports:     ports,
		Backends:  backends,
		Algorithm: sgtypes.LBAlgorithmMaglev, // default
		SessionAffinity: svc.Spec.SessionAffinity == corev1.ServiceAffinityClientIP,
	}, nil
}

func compileBackends(slices []discoveryv1.EndpointSlice) ([]ir.Backend, error) {
	var backends []ir.Backend
	var id uint32
	for _, slice := range slices {
		for _, ep := range slice.Endpoints {
			if ep.Conditions.Ready != nil && !*ep.Conditions.Ready {
				continue
			}
			for _, addr := range ep.Addresses {
				ip, err := netip.ParseAddr(addr)
				if err != nil {
					continue
				}
				id++
				var port uint16
				if len(slice.Ports) > 0 && slice.Ports[0].Port != nil {
					port = uint16(*slice.Ports[0].Port)
				}
				backends = append(backends, ir.Backend{
					ID:      id,
					IP:      ip,
					Port:    port,
					Weight:  1,
					Healthy: true,
				})
			}
		}
	}
	return backends, nil
}

func mapProtocol(p corev1.Protocol) sgtypes.Protocol {
	switch p {
	case corev1.ProtocolTCP:
		return sgtypes.ProtocolTCP
	case corev1.ProtocolUDP:
		return sgtypes.ProtocolUDP
	default:
		return sgtypes.ProtocolTCP
	}
}

func mapServiceType(t corev1.ServiceType) ir.ServiceType {
	switch t {
	case corev1.ServiceTypeNodePort:
		return ir.ServiceTypeNodePort
	case corev1.ServiceTypeLoadBalancer:
		return ir.ServiceTypeLoadBalancer
	case corev1.ServiceTypeExternalName:
		return ir.ServiceTypeExternalName
	default:
		return ir.ServiceTypeClusterIP
	}
}

func parseOptionalIP(s string) (netip.Addr, error) {
	if s == "" || s == "None" {
		return netip.Addr{}, nil
	}
	return netip.ParseAddr(s)
}

// ============================================================================
// Maglev hash (LB algorithm)
// ============================================================================

// MaglevTableSize is the Maglev consistent hash table size.
// Must be prime. 127 is used for 128-slot tables.
const MaglevTableSize = 127

// MaglevLookup performs a Maglev hash lookup for a 5-tuple flow.
// Returns the backend index.
func MaglevLookup(table []uint32, srcIP, dstIP netip.Addr, srcPort, dstPort uint16, proto uint8) int {
	if len(table) == 0 {
		return 0
	}
	h := hashFlow(srcIP, dstIP, srcPort, dstPort, proto)
	return int(table[h%uint32(len(table))])
}

// hashFlow produces a 32-bit hash of the 5-tuple using FNV-1a.
func hashFlow(srcIP, dstIP netip.Addr, srcPort, dstPort uint16, proto uint8) uint32 {
	const offset32 = 2166136261
	const prime32 = 16777619
	h := uint32(offset32)
	fnv := func(b []byte) {
		for _, c := range b {
			h ^= uint32(c)
			h *= prime32
		}
	}
	fnv(srcIP.AsSlice())
	fnv(dstIP.AsSlice())
	fnv([]byte{byte(srcPort >> 8), byte(srcPort)})
	fnv([]byte{byte(dstPort >> 8), byte(dstPort)})
	fnv([]byte{proto})
	return h
}
