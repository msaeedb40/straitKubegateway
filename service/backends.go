package service

import (
	"fmt"
	"net/netip"
	"sync"
)

// BackendPool tracks active backend endpoints across services.
type BackendPool struct {
	mu       sync.RWMutex
	backends map[uint32]*Backend
	nextID   uint32
}

// NewBackendPool initializes an empty backend pool.
func NewBackendPool() *BackendPool {
	return &BackendPool{
		backends: make(map[uint32]*Backend),
		nextID:   1,
	}
}

// RegisterBackend adds a backend to the pool and returns its assigned ID.
func (p *BackendPool) RegisterBackend(ip netip.Addr, port uint16, weight uint32) *Backend {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check if already registered
	for _, b := range p.backends {
		if b.IP == ip && b.Port == port {
			b.Weight = weight
			b.Healthy = true
			return b
		}
	}

	id := p.nextID
	p.nextID++

	b := &Backend{
		ID:      id,
		IP:      ip,
		Port:    port,
		Weight:  weight,
		Healthy: true,
	}
	p.backends[id] = b
	return b
}

// UnregisterBackend removes a backend by ID.
func (p *BackendPool) UnregisterBackend(id uint32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.backends, id)
}

// SetHealth updates the health status of a backend.
func (p *BackendPool) SetHealth(id uint32, healthy bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	b, exists := p.backends[id]
	if !exists {
		return fmt.Errorf("backend %d not found", id)
	}
	b.Healthy = healthy
	return nil
}

// GetBackend retrieves a backend by ID.
func (p *BackendPool) GetBackend(id uint32) (*Backend, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	b, exists := p.backends[id]
	return b, exists
}
