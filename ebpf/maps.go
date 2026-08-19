// Package ebpf provides Go bindings, map managers, and NetKit loader
// interfaces for the straitKubegateway dataplane.
package ebpf

import (
	"encoding/binary"
	"fmt"
	"net/netip"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/straitKubegateway/straitKubegateway/pkg/bpf"
)

// EndpointMapValue matches struct endpoint_info in bpf/maps/maps.h.
type EndpointMapValue struct {
	IfIndex   uint32
	Identity  uint32
	SegmentID uint32
	MAC       [6]byte
	Pad       uint16
	RxBytes   uint64
	TxBytes   uint64
}

// EndpointMapManager provides type-safe access to the BPF endpoints_map.
type EndpointMapManager struct {
	mu      sync.RWMutex
	bpfMap  *ebpf.Map
	pinPath string
}

// NewEndpointMapManager creates or loads the pinned endpoint map.
func NewEndpointMapManager(m *ebpf.Map) *EndpointMapManager {
	return &EndpointMapManager{
		bpfMap:  m,
		pinPath: bpf.MapPinPath("endpoints_map"),
	}
}

// AddEndpoint inserts or updates an endpoint in the BPF map.
func (m *EndpointMapManager) AddEndpoint(ip netip.Addr, val EndpointMapValue) error {
	if m.bpfMap == nil {
		return nil // No-op in mock/fallback mode
	}

	ip4 := ip.As4()
	key := binary.LittleEndian.Uint32(ip4[:])

	m.mu.Lock()
	defer m.mu.Unlock()

	return m.bpfMap.Put(key, val)
}

// DeleteEndpoint removes an endpoint from the BPF map.
func (m *EndpointMapManager) DeleteEndpoint(ip netip.Addr) error {
	if m.bpfMap == nil {
		return nil
	}

	ip4 := ip.As4()
	key := binary.LittleEndian.Uint32(ip4[:])

	m.mu.Lock()
	defer m.mu.Unlock()

	return m.bpfMap.Delete(key)
}

// LookupEndpoint reads an endpoint value by IP.
func (m *EndpointMapManager) LookupEndpoint(ip netip.Addr) (*EndpointMapValue, error) {
	if m.bpfMap == nil {
		return nil, fmt.Errorf("bpf map is nil")
	}

	ip4 := ip.As4()
	key := binary.LittleEndian.Uint32(ip4[:])

	var val EndpointMapValue
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err := m.bpfMap.Lookup(key, &val); err != nil {
		return nil, err
	}
	return &val, nil
}
