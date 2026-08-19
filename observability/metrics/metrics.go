// Package metrics provides Prometheus metric collectors for straitKubegateway.
// It enforces low-cardinality metric label designs as required by the architecture.
package metrics

import (
	"sync"
)

// Registry holds the Prometheus metric collectors for straitKubegateway.
type Registry struct {
	mu sync.RWMutex
	// Metric values tracked in-memory
	packetDropsTotal    uint64
	packetsForwardTotal uint64
	activeEndpoints     int64
	activeGateways      int64
	activeTunnels       int64
}

var (
	defaultRegistry *Registry
	once            sync.Once
)

// DefaultRegistry returns the singleton metrics registry.
func DefaultRegistry() *Registry {
	once.Do(func() {
		defaultRegistry = &Registry{}
	})
	return defaultRegistry
}

// IncPacketDrops increments the packet drop counter.
func (r *Registry) IncPacketDrops() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.packetDropsTotal++
}

// IncPacketsForwarded increments the packet forwarding counter.
func (r *Registry) IncPacketsForwarded() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.packetsForwardTotal++
}

// SetActiveEndpoints updates the active endpoints gauge.
func (r *Registry) SetActiveEndpoints(count int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.activeEndpoints = count
}

// SetActiveGateways updates the active gateways gauge.
func (r *Registry) SetActiveGateways(count int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.activeGateways = count
}

// SetActiveTunnels updates the active tunnels gauge.
func (r *Registry) SetActiveTunnels(count int64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.activeTunnels = count
}

// Snapshot returns a copy of current metric values.
func (r *Registry) Snapshot() (drops, forwarded uint64, endpoints, gateways, tunnels int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.packetDropsTotal, r.packetsForwardTotal, r.activeEndpoints, r.activeGateways, r.activeTunnels
}
