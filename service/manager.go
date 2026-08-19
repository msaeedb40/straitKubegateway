package service

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// Manager coordinates all Service LB definitions, backend pools, and BPF maps.
type Manager struct {
	mu          sync.RWMutex
	services    map[string]*Service // "namespace/name:port" -> Service
	backendPool *BackendPool
	nextSvcID   uint32
}

// NewManager creates a new Service LB manager.
func NewManager() *Manager {
	return &Manager{
		services:    make(map[string]*Service),
		backendPool: NewBackendPool(),
		nextSvcID:   1,
	}
}

// UpsertService creates or updates a Service with its backends.
func (m *Manager) UpsertService(ns, name string, vip netip.Addr, port uint16, proto types.Protocol, algo Algorithm, affinity bool, affinityTimeout uint32, endpoints []BackendEndpoint) *Service {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s/%s:%d/%s", ns, name, port, proto)

	svc, exists := m.services[key]
	if !exists {
		svc = &Service{
			ID:        m.nextSvcID,
			Namespace: ns,
			Name:      name,
			VIP:       vip,
			Port:      port,
			Protocol:  proto,
		}
		m.nextSvcID++
		m.services[key] = svc
	}

	svc.Algorithm = algo
	svc.SessionAffinity = affinity
	svc.AffinityTimeout = affinityTimeout

	// Register backends
	backends := make([]*Backend, 0, len(endpoints))
	for _, ep := range endpoints {
		b := m.backendPool.RegisterBackend(ep.IP, ep.Port, ep.Weight)
		backends = append(backends, b)
	}
	svc.Backends = backends

	return svc
}

// DeleteService removes a Service from the load balancer.
func (m *Manager) DeleteService(ns, name string, port uint16, proto types.Protocol) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := fmt.Sprintf("%s/%s:%d/%s", ns, name, port, proto)
	delete(m.services, key)
}

// GetService retrieves a service by namespace and name.
func (m *Manager) GetService(ns, name string, port uint16, proto types.Protocol) (*Service, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s/%s:%d/%s", ns, name, port, proto)
	svc, exists := m.services[key]
	return svc, exists
}

// ListServices returns all currently configured Services.
func (m *Manager) ListServices() []*Service {
	m.mu.RLock()
	defer m.mu.RUnlock()

	res := make([]*Service, 0, len(m.services))
	for _, s := range m.services {
		res = append(res, s)
	}
	return res
}

// BackendEndpoint is an input descriptor for registering a backend.
type BackendEndpoint struct {
	IP     netip.Addr
	Port   uint16
	Weight uint32
}
