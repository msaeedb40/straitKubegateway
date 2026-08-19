// Package dataplane coordinates CNI lifecycle, NetKit links, BPF maps,
// IPAM allocations, and node routing into a unified Linux kernel dataplane.
package dataplane

import (
	"fmt"
	"net"
	"net/netip"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/ebpf"
	"github.com/straitKubegateway/straitKubegateway/ipam"
	"github.com/straitKubegateway/straitKubegateway/pkg/identity"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
	"github.com/straitKubegateway/straitKubegateway/routing"
	"github.com/vishvananda/netns"
)

// PodNetworkRequest encapsulates parameters for CNI ADD or DEL.
type PodNetworkRequest struct {
	ContainerID string
	NetnsPath   string
	IfName      string
	Namespace   string
	PodName     string
	SegmentID   types.SegmentID
	Labels      map[string]string
	MTU         int
}

// PodNetworkResult holds the allocated network configuration for a pod.
type PodNetworkResult struct {
	IP        netip.Addr
	Gateway   netip.Addr
	PrefixLen int
	Identity  types.Identity
	HostVeth  string
	IfIndex   int
}

// Manager orchestrates all dataplane subsystems.
type Manager struct {
	mu           sync.RWMutex
	ipam         *ipam.Allocator
	identity     identity.Allocator
	netkit       *ebpf.NetKitManager
	epMapManager *ebpf.EndpointMapManager
	routingTable *routing.Table
	endpoints    map[string]*types.Endpoint // ContainerID -> Endpoint
}

// Config configures the Dataplane Manager.
type Config struct {
	PodCIDR string
}

// NewManager initializes the dataplane manager and its underlying components.
func NewManager(cfg Config) (*Manager, error) {
	ipamAlloc, err := ipam.NewAllocator(cfg.PodCIDR)
	if err != nil {
		return nil, fmt.Errorf("init ipam allocator: %w", err)
	}

	bpfLoader, _ := ebpf.NewLoader()
	var bpfCollection *ebpf.Collection
	if bpfLoader != nil {
		bpfCollection, _ = bpfLoader.LoadMaps()
	}

	var epMapManager *ebpf.EndpointMapManager
	if bpfCollection != nil && bpfCollection.EndpointsMap != nil {
		epMapManager = ebpf.NewEndpointMapManager(bpfCollection.EndpointsMap)
	}

	return &Manager{
		ipam:         ipamAlloc,
		identity:     identity.NewLocalAllocator(),
		netkit:       ebpf.NewNetKitManager(),
		epMapManager: epMapManager,
		routingTable: routing.NewTable(),
		endpoints:    make(map[string]*types.Endpoint),
	}, nil
}

// AddPodNetwork configures networking for a new pod container (CNI ADD flow).
// Flow: IPAM allocation -> NetKit interface -> Identity allocation -> BPF map update -> READY.
func (m *Manager) AddPodNetwork(req PodNetworkRequest) (*PodNetworkResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 1. IPAM allocation
	ip, err := m.ipam.Allocate(req.ContainerID)
	if err != nil {
		return nil, fmt.Errorf("ipam allocation failed: %w", err)
	}

	// 2. Identity allocation
	id, err := m.identity.Allocate(req.Labels)
	if err != nil {
		_, _ = m.ipam.Release(req.ContainerID)
		return nil, fmt.Errorf("identity allocation failed: %w", err)
	}

	// 3. NetKit device creation
	targetNs, err := netns.GetFromPath(req.NetnsPath)
	if err != nil {
		_ = m.identity.Release(id)
		_, _ = m.ipam.Release(req.ContainerID)
		return nil, fmt.Errorf("open container netns %q: %w", req.NetnsPath, err)
	}
	defer targetNs.Close()

	hostVeth := fmt.Sprintf("sg-%s", req.ContainerID[:min(8, len(req.ContainerID))])
	ipNet := net.IPNet{
		IP:   net.ParseIP(ip.String()),
		Mask: net.CIDRMask(m.ipam.Prefix().Bits(), 32),
	}

	mtu := req.MTU
	if mtu <= 0 {
		mtu = 1500
	}

	pair, err := m.netkit.CreateNetKitPair(hostVeth, req.IfName, targetNs, ipNet, mtu)
	if err != nil {
		_ = m.identity.Release(id)
		_, _ = m.ipam.Release(req.ContainerID)
		return nil, fmt.Errorf("create netkit link pair: %w", err)
	}

	// 4. Update BPF Endpoint Map
	if m.epMapManager != nil {
		_ = m.epMapManager.AddEndpoint(ip, ebpf.EndpointMapValue{
			IfIndex:   uint32(pair.HostIfIndex),
			Identity:  id.Uint32(),
			SegmentID: req.SegmentID.Uint32(),
		})
	}

	ep := &types.Endpoint{
		ID:          uint64(pair.HostIfIndex),
		Identity:    id,
		IPv4:        ip,
		Namespace:   req.Namespace,
		PodName:     req.PodName,
		ContainerID: req.ContainerID,
		IfIndex:     pair.HostIfIndex,
		SegmentID:   req.SegmentID,
		Labels:      req.Labels,
		State:       types.EndpointStateReady,
	}
	m.endpoints[req.ContainerID] = ep

	return &PodNetworkResult{
		IP:        ip,
		Gateway:   m.ipam.Gateway(),
		PrefixLen: m.ipam.Prefix().Bits(),
		Identity:  id,
		HostVeth:  hostVeth,
		IfIndex:   pair.HostIfIndex,
	}, nil
}

// DelPodNetwork removes networking for a deleted pod container (CNI DEL flow).
// Flow: remove BPF map entry -> release Identity -> release IPAM -> delete NetKit link.
func (m *Manager) DelPodNetwork(containerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	ep, exists := m.endpoints[containerID]
	if !exists {
		return nil // Idempotent DEL
	}

	// 1. Remove from BPF endpoint map
	if m.epMapManager != nil {
		_ = m.epMapManager.DeleteEndpoint(ep.IPv4)
	}

	// 2. Release identity
	_ = m.identity.Release(ep.Identity)

	// 3. Release IP
	_, _ = m.ipam.Release(containerID)

	// 4. Destroy NetKit link
	hostVeth := fmt.Sprintf("sg-%s", containerID[:min(8, len(containerID))])
	_ = m.netkit.DeleteNetKitPair(hostVeth)

	delete(m.endpoints, containerID)
	return nil
}

// GetEndpoint returns an active endpoint by container ID.
func (m *Manager) GetEndpoint(containerID string) (*types.Endpoint, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ep, ok := m.endpoints[containerID]
	return ep, ok
}

// ListEndpoints returns a snapshot of all active endpoints.
func (m *Manager) ListEndpoints() []*types.Endpoint {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make([]*types.Endpoint, 0, len(m.endpoints))
	for _, ep := range m.endpoints {
		res = append(res, ep)
	}
	return res
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
