// Package identity provides Kubernetes-native workload identity management,
// mapping label selectors and namespaces to deterministic 32-bit security identities.
package identity

import (
	"fmt"
	"sync"

	pkgident "github.com/straitKubegateway/straitKubegateway/pkg/identity"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// Manager coordinates local and cluster-wide identity allocation for pods and nodes.
type Manager struct {
	mu        sync.RWMutex
	allocator pkgident.Allocator
}

// NewManager creates a new identity manager wrapping the core allocator.
func NewManager(alloc pkgident.Allocator) *Manager {
	if alloc == nil {
		alloc = pkgident.NewLocalAllocator()
	}
	return &Manager{
		allocator: alloc,
	}
}

// Allocate resolves or creates an identity for the given label set.
func (m *Manager) Allocate(labels map[string]string) (types.Identity, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.allocator.Allocate(labels)
}

// Release releases a reference to the given identity.
func (m *Manager) Release(id types.Identity) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.allocator.Release(id)
}

// Lookup resolves the label set associated with an identity ID.
func (m *Manager) Lookup(id types.Identity) (map[string]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	labels, ok := m.allocator.LookupByID(id)
	if !ok {
		return nil, fmt.Errorf("identity %d not found", id)
	}
	return labels, nil
}
