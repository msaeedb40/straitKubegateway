// Package routing provides IP routing table management and FIB interaction
// for straitKubegateway node networking.
package routing

import (
	"fmt"
	"net/netip"
	"sync"
)

// RouteEntry represents a destination prefix with next hop and interface index.
type RouteEntry struct {
	Dst       netip.Prefix
	NextHop   netip.Addr
	IfIndex   int
	SegmentID uint32
	Metric    int
}

// Table holds in-memory routing table state for the node.
type Table struct {
	mu     sync.RWMutex
	routes map[string]RouteEntry // Prefix -> RouteEntry
}

// NewTable creates an in-memory routing table.
func NewTable() *Table {
	return &Table{
		routes: make(map[string]RouteEntry),
	}
}

// AddRoute adds or updates a route.
func (t *Table) AddRoute(r RouteEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.routes[r.Dst.String()] = r
}

// DeleteRoute removes a route by prefix.
func (t *Table) DeleteRoute(dst netip.Prefix) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.routes, dst.String())
}

// Lookup finds the longest prefix match for an IP address.
func (t *Table) Lookup(ip netip.Addr) (*RouteEntry, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var bestMatch *RouteEntry
	var maxPrefixLen = -1

	for _, r := range t.routes {
		if r.Dst.Contains(ip) {
			if r.Dst.Bits() > maxPrefixLen {
				maxPrefixLen = r.Dst.Bits()
				match := r
				bestMatch = &match
			}
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no route to host %s", ip)
	}
	return bestMatch, nil
}

// ListRoutes returns a copy of all current routes.
func (t *Table) ListRoutes() []RouteEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()

	res := make([]RouteEntry, 0, len(t.routes))
	for _, r := range t.routes {
		res = append(res, r)
	}
	return res
}
