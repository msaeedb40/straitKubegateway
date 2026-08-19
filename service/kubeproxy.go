// Package service provides full kube-proxy replacement functionality,
// handling ClusterIP, NodePort, LoadBalancer, ExternalIPs, Direct Server Return (DSR),
// HostPort, and Hairpin NAT via eBPF.
package service

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// ServiceType mirrors Kubernetes Service types.
type ServiceType string

const (
	ServiceTypeClusterIP    ServiceType = "ClusterIP"
	ServiceTypeNodePort     ServiceType = "NodePort"
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"
	ServiceTypeExternalName ServiceType = "ExternalName"
)

// KubeProxyService represents a complete Kubernetes Service in the dataplane.
type KubeProxyService struct {
	Namespace       string
	Name            string
	Type            ServiceType
	ClusterIP       netip.Addr
	NodePorts       map[uint16]uint16 // NodePort -> TargetPort
	LoadBalancerIPs []netip.Addr
	ExternalIPs     []netip.Addr
	Port            uint16
	Protocol        types.Protocol
	Algorithm       Algorithm
	SessionAffinity bool
	AffinityTimeout uint32
	DSR             bool // Direct Server Return
	HairpinMode     bool
	Backends        []*Backend
}

// KubeProxyManager manages the full suite of kube-proxy replacement operations.
type KubeProxyManager struct {
	mu           sync.RWMutex
	services     map[string]*KubeProxyService // "namespace/name:port/proto" -> Service
	nodeIPs      map[netip.Addr]bool
	nodePortMap  map[uint16]*KubeProxyService // NodePort -> Service
	lbIPMap      map[string]*KubeProxyService // "lbIP:port" -> Service
	backendPool  *BackendPool
	hairpinCIDRs []netip.Prefix
}

// NewKubeProxyManager creates a new kube-proxy replacement manager.
func NewKubeProxyManager(nodeIPs []netip.Addr, podCIDRs []netip.Prefix) *KubeProxyManager {
	nodeIPMap := make(map[netip.Addr]bool)
	for _, ip := range nodeIPs {
		nodeIPMap[ip] = true
	}

	return &KubeProxyManager{
		services:     make(map[string]*KubeProxyService),
		nodeIPs:      nodeIPMap,
		nodePortMap:  make(map[uint16]*KubeProxyService),
		lbIPMap:      make(map[string]*KubeProxyService),
		backendPool:  NewBackendPool(),
		hairpinCIDRs: podCIDRs,
	}
}

// UpsertKubeService adds or updates a complete Kubernetes Service definition.
func (m *KubeProxyManager) UpsertKubeService(svc KubeProxyService, endpoints []BackendEndpoint) *KubeProxyService {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s/%s:%d/%s", svc.Namespace, svc.Name, svc.Port, svc.Protocol)

	// Register backends
	backends := make([]*Backend, 0, len(endpoints))
	for _, ep := range endpoints {
		b := m.backendPool.RegisterBackend(ep.IP, ep.Port, ep.Weight)
		backends = append(backends, b)
	}
	svc.Backends = backends

	serviceCopy := svc
	m.services[key] = &serviceCopy

	// Index NodePorts
	for np := range svc.NodePorts {
		m.nodePortMap[np] = &serviceCopy
	}

	// Index LoadBalancer IPs
	for _, lbIP := range svc.LoadBalancerIPs {
		lbKey := fmt.Sprintf("%s:%d/%s", lbIP, svc.Port, svc.Protocol)
		m.lbIPMap[lbKey] = &serviceCopy
	}

	return &serviceCopy
}

// DeleteKubeService removes a Service from all routing tables.
func (m *KubeProxyManager) DeleteKubeService(ns, name string, port uint16, proto types.Protocol) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s/%s:%d/%s", ns, name, port, proto)
	svc, exists := m.services[key]
	if !exists {
		return
	}

	for np := range svc.NodePorts {
		delete(m.nodePortMap, np)
	}
	for _, lbIP := range svc.LoadBalancerIPs {
		lbKey := fmt.Sprintf("%s:%d/%s", lbIP, svc.Port, svc.Protocol)
		delete(m.lbIPMap, lbKey)
	}

	delete(m.services, key)
}

// LookupDestination translates ClusterIP, NodePort, or LoadBalancer IP to a backend endpoint.
func (m *KubeProxyManager) LookupDestination(clientIP, dstIP netip.Addr, dstPort uint16, proto types.Protocol) (*Backend, bool, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var targetSvc *KubeProxyService

	// 1. Check ClusterIP
	for _, s := range m.services {
		if s.ClusterIP == dstIP && s.Port == dstPort && s.Protocol == proto {
			targetSvc = s
			break
		}
	}

	// 2. Check NodePort (if dstIP is one of the node IPs)
	if targetSvc == nil && m.nodeIPs[dstIP] {
		if s, ok := m.nodePortMap[dstPort]; ok && s.Protocol == proto {
			targetSvc = s
		}
	}

	// 3. Check LoadBalancer IP
	if targetSvc == nil {
		lbKey := fmt.Sprintf("%s:%d/%s", dstIP, dstPort, proto)
		if s, ok := m.lbIPMap[lbKey]; ok {
			targetSvc = s
		}
	}

	if targetSvc == nil {
		return nil, false, false, nil // Not a service packet
	}

	// 4. Select Backend
	backend, err := targetSvc.selectBackend(clientIP)
	if err != nil {
		return nil, false, false, err
	}

	// 5. Check if Hairpin NAT is required (client IP is identical to selected backend IP)
	isHairpin := clientIP == backend.IP
	isDSR := targetSvc.DSR

	return backend, isDSR, isHairpin, nil
}

func (s *KubeProxyService) selectBackend(clientIP netip.Addr) (*Backend, error) {
	healthy := make([]*Backend, 0, len(s.Backends))
	for _, b := range s.Backends {
		if b.Healthy {
			healthy = append(healthy, b)
		}
	}

	if len(healthy) == 0 {
		return nil, fmt.Errorf("no healthy backends for service %s/%s", s.Namespace, s.Name)
	}

	// Default to Maglev LUT
	lut := GenerateMaglevLUT(1, healthy)
	slot := uint32(clientIP.As4()[3]) % MaglevTableSize
	targetID := lut.Lookup[slot]

	for _, b := range healthy {
		if b.ID == targetID {
			return b, nil
		}
	}
	return healthy[0], nil
}
