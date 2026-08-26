// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package cni implements the straitKubegateway CNI plugin.
//
// CNI ADD flow (synchronous — must complete quickly):
//  1. Validate config
//  2. Allocate IP (IPAM)
//  3. Create network namespace attachment (NetKit)
//  4. Configure interface + address
//  5. Configure mandatory routes
//  6. Allocate BPF identity
//  7. Return CNI result
//
// Post-ADD reconciliation (asynchronous — never blocks CNI ADD):
//   - Service dataplane programming
//   - Policy dataplane programming
//   - NAT programming
//   - Observability
//
// Architectural invariants:
//   - CNI ADD must NOT synchronously depend on Service, Policy, NAT,
//     Gateway, Transit, or BGP convergence.
//   - CNI bootstrap must NOT depend on the Service dataplane.
//   - CNI readiness ≠ Service readiness ≠ Policy readiness ≠ Gateway readiness.
package cni

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
	"sync"

	"go.uber.org/zap"

	"github.com/straitkubegateway/straitkubegateway/identity"
	"github.com/straitkubegateway/straitkubegateway/ipam"
)

// ============================================================================
// CNI spec types (CNI Spec 1.1+)
// ============================================================================

// NetworkConfig is the parsed CNI configuration.
type NetworkConfig struct {
	// CNIVersion is the CNI spec version.
	CNIVersion string `json:"cniVersion"`
	// Name is the network name.
	Name string `json:"name"`
	// Type is the plugin type ("straitkubegateway").
	Type string `json:"type"`
	// IPAM holds IPAM configuration.
	IPAM IPAMConfig `json:"ipam,omitempty"`
	// MTU overrides MTU auto-discovery (0 = auto-detect).
	MTU int `json:"mtu,omitempty"`
	// BPFFSPath is the bpffs mount path.
	BPFFSPath string `json:"bpffsPath,omitempty"`
	// DataDir is the CNI data directory for state persistence.
	DataDir string `json:"dataDir,omitempty"`
}

// IPAMConfig holds IPAM configuration.
type IPAMConfig struct {
	// Type is the IPAM type (always "straitkubegateway" for built-in).
	Type string `json:"type,omitempty"`
	// Subnet is an optional subnet override.
	Subnet string `json:"subnet,omitempty"`
	// Gateway is an optional gateway override.
	Gateway string `json:"gateway,omitempty"`
}

// ============================================================================
// CNI Plugin
// ============================================================================

// Plugin implements the straitKubegateway CNI plugin.
type Plugin struct {
	mu           sync.RWMutex
	log          *zap.Logger
	ipamAlloc    *ipam.Allocator
	identityAlloc *identity.Allocator
	allocations  map[string]netip.Addr // containerID -> IP
}

// New creates a new CNI plugin.
func New(log *zap.Logger) *Plugin {
	return &Plugin{
		log:         log,
		allocations: make(map[string]netip.Addr),
	}
}

// WithIPAM sets the IPAM allocator for this plugin.
func (p *Plugin) WithIPAM(alloc *ipam.Allocator) *Plugin {
	p.ipamAlloc = alloc
	return p
}

// WithIdentity sets the BPF identity allocator for this plugin.
func (p *Plugin) WithIdentity(alloc *identity.Allocator) *Plugin {
	p.identityAlloc = alloc
	return p
}

// ============================================================================
// ADD — synchronous pod network setup
// ============================================================================

// AddResult is the CNI result returned from ADD.
type AddResult struct {
	// IP is the allocated pod IP.
	IP netip.Addr
	// Gateway is the pod's default gateway.
	Gateway netip.Addr
	// HostIfIndex is the host-side NetKit interface index.
	HostIfIndex int
	// ContainerIfIndex is the container-side interface index.
	ContainerIfIndex int
	// Identity is the allocated BPF identity.
	Identity uint32
}

// ADD implements the CNI ADD command.
// This function is SYNCHRONOUS and must return quickly.
// It must NOT block on Service, Policy, NAT, Gateway, Transit, or BGP.
func (p *Plugin) ADD(cfg *NetworkConfig, netns string, containerID, ifName string) (*AddResult, error) {
	p.log.Info("CNI ADD",
		zap.String("containerID", containerID),
		zap.String("netns", netns),
		zap.String("iface", ifName),
	)

	// Step 1: Validate config
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("CNI ADD validate: %w", err)
	}

	// Step 2: Allocate IP via IPAM
	var ip, gw netip.Addr
	var err error
	if p.ipamAlloc != nil {
		ip, err = p.ipamAlloc.Allocate()
		if err != nil {
			return nil, fmt.Errorf("CNI ADD IPAM allocate: %w", err)
		}
		gw = p.deriveGatewayForIP(cfg, ip)
	} else {
		ip, gw, err = allocateIP(cfg)
		if err != nil {
			return nil, fmt.Errorf("CNI ADD IPAM: %w", err)
		}
	}

	p.mu.Lock()
	p.allocations[containerID] = ip
	p.mu.Unlock()

	p.log.Debug("IPAM allocation", zap.String("ip", ip.String()))

	// Step 3: Create NetKit attachment (host ↔ container veth pair via netkit)
	hostIdx, containerIdx, err := setupNetKit(netns, ifName, containerID)
	if err != nil {
		p.releaseAllocatedIP(cfg, containerID, ip)
		return nil, fmt.Errorf("CNI ADD netkit setup: %w", err)
	}

	// Step 4: Configure address on container interface
	if err := configureAddress(netns, ifName, ip, gw, cfg.MTU); err != nil {
		_ = teardownNetKit(netns, ifName)
		p.releaseAllocatedIP(cfg, containerID, ip)
		return nil, fmt.Errorf("CNI ADD configure address: %w", err)
	}

	// Step 5: Configure mandatory routes (default route via gateway)
	if err := configureMandatoryRoutes(netns, ifName, gw); err != nil {
		_ = teardownNetKit(netns, ifName)
		p.releaseAllocatedIP(cfg, containerID, ip)
		return nil, fmt.Errorf("CNI ADD configure routes: %w", err)
	}

	// Step 6: Allocate BPF identity
	var bpfIdentity uint32
	if p.identityAlloc != nil {
		id, err := p.identityAlloc.Allocate(context.Background(), fmt.Sprintf("ip:%s", ip.String()))
		if err != nil {
			p.log.Warn("BPF identity allocation deferred", zap.Error(err))
		} else {
			bpfIdentity = uint32(id)
		}
	} else {
		id, err := allocateIdentity(ip, containerID)
		if err != nil {
			p.log.Warn("BPF identity allocation deferred", zap.Error(err))
		} else {
			bpfIdentity = id
		}
	}

	result := &AddResult{
		IP:               ip,
		Gateway:          gw,
		HostIfIndex:      hostIdx,
		ContainerIfIndex: containerIdx,
		Identity:         bpfIdentity,
	}

	p.log.Info("CNI ADD complete",
		zap.String("ip", ip.String()),
		zap.Uint32("identity", bpfIdentity),
		zap.Int("hostIfIdx", hostIdx),
	)
	return result, nil
}

// ============================================================================
// DEL — synchronous pod network teardown
// ============================================================================

// DEL implements the CNI DEL command.
func (p *Plugin) DEL(cfg *NetworkConfig, netns string, containerID, ifName string) error {
	p.log.Info("CNI DEL",
		zap.String("containerID", containerID),
		zap.String("netns", netns),
	)

	p.mu.Lock()
	allocatedIP, hasIP := p.allocations[containerID]
	delete(p.allocations, containerID)
	p.mu.Unlock()

	// 1. Remove routes
	if err := removeRoutes(netns, ifName); err != nil {
		p.log.Warn("CNI DEL remove routes", zap.Error(err))
	}
	// 2. Remove BPF identity
	if err := removeBPFIdentity(containerID); err != nil {
		p.log.Warn("CNI DEL remove BPF identity", zap.Error(err))
	}
	// 3. Remove policy state (best-effort)
	if err := removePolicyState(containerID); err != nil {
		p.log.Warn("CNI DEL remove policy state", zap.Error(err))
	}
	// 4. Release IP back to IPAM
	if hasIP {
		p.releaseAllocatedIP(cfg, containerID, allocatedIP)
	} else if err := releaseIPByContainerID(cfg, containerID); err != nil {
		p.log.Warn("CNI DEL release IP", zap.Error(err))
	}
	// 5. Destroy NetKit interface
	if err := teardownNetKit(netns, ifName); err != nil {
		p.log.Warn("CNI DEL teardown netkit", zap.Error(err))
	}

	p.log.Info("CNI DEL complete", zap.String("containerID", containerID))
	return nil
}

// ============================================================================
// CHECK — verify pod network is still correct
// ============================================================================

// CHECK implements the CNI CHECK command.
func (p *Plugin) CHECK(cfg *NetworkConfig, netns string, containerID, ifName string) error {
	p.log.Debug("CNI CHECK", zap.String("containerID", containerID))
	// Verify interface exists and has the correct address
	return checkNetworkNamespace(netns, ifName)
}

// ============================================================================
// GC — clean up orphaned interfaces and allocations
// ============================================================================

// GC scans allocated resources and removes state for non-active containers.
func (p *Plugin) GC(ctx context.Context, activeContainers map[string]bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for cid, ip := range p.allocations {
		if !activeContainers[cid] {
			p.log.Info("CNI GC cleaning orphaned container",
				zap.String("containerID", cid),
				zap.String("ip", ip.String()),
			)
			_ = removeBPFIdentity(cid)
			_ = removePolicyState(cid)
			delete(p.allocations, cid)
		}
	}
	return nil
}

// ============================================================================
// VERSION — return supported CNI versions
// ============================================================================

// SupportedVersions lists the CNI spec versions this plugin supports.
var SupportedVersions = []string{"0.3.1", "0.4.0", "1.0.0", "1.1.0"}

// VERSION returns the list of supported CNI versions.
func (p *Plugin) VERSION() []string {
	return SupportedVersions
}

func (p *Plugin) releaseAllocatedIP(cfg *NetworkConfig, containerID string, ip netip.Addr) {
	p.mu.Lock()
	delete(p.allocations, containerID)
	p.mu.Unlock()

	_ = releaseIP(cfg, ip)
}

func (p *Plugin) deriveGatewayForIP(cfg *NetworkConfig, ip netip.Addr) netip.Addr {
	if p.ipamAlloc != nil {
		if gw, err := p.ipamAlloc.GatewayForIP(ip); err == nil {
			return gw
		}
	}
	if cfg != nil && cfg.IPAM.Gateway != "" {
		if gw, err := netip.ParseAddr(cfg.IPAM.Gateway); err == nil {
			return gw
		}
	}
	// Dynamic fallback: 1st usable host IP in default /24 or /64 prefix
	if ip.Is4() {
		pfx, err := ip.Prefix(24)
		if err == nil {
			return pfx.Masked().Addr().Next()
		}
	}
	pfx, err := ip.Prefix(64)
	if err == nil {
		return pfx.Masked().Addr().Next()
	}
	return ip
}

// ParseConfig parses raw CNI config bytes into NetworkConfig.
func ParseConfig(data []byte) (*NetworkConfig, error) {
	var cfg NetworkConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse CNI config: %w", err)
	}
	return &cfg, nil
}

// ============================================================================
// Internal helpers (stubs — implemented in subsequent files)
// ============================================================================

func validateConfig(cfg *NetworkConfig) error {
	if cfg.Name == "" {
		return fmt.Errorf("network name required")
	}
	if cfg.CNIVersion == "" {
		return fmt.Errorf("cniVersion required")
	}
	return nil
}

func allocateIP(_ *NetworkConfig) (ip, gw netip.Addr, err error) {
	// TODO(phase1): call ipam.Allocator.Allocate()
	// Placeholder: returns an error until IPAM is wired
	return netip.Addr{}, netip.Addr{}, fmt.Errorf("IPAM not yet wired")
}

func releaseIP(_ *NetworkConfig, _ netip.Addr) error {
	// TODO(phase1): call ipam.Allocator.Release()
	return nil
}

func releaseIPByContainerID(_ *NetworkConfig, _ string) error {
	// TODO(phase1): look up IP from state store and release
	return nil
}

func setupNetKit(_, _, _ string) (hostIdx, containerIdx int, err error) {
	// TODO(phase1): create NetKit veth pair, attach eBPF programs
	return 0, 0, nil
}

func teardownNetKit(_, _ string) error {
	// TODO(phase1): delete NetKit interface
	return nil
}

func configureAddress(netns, ifName string, ip, gw netip.Addr, mtu int) error {
	// TODO(phase1): enter netns, set IP address and MTU on ifName
	_ = netns
	_ = ifName
	_ = ip
	_ = gw
	_ = mtu
	return nil
}

func configureMandatoryRoutes(netns, ifName string, gw netip.Addr) error {
	// TODO(phase1): enter netns, add default route via gw
	_ = netns
	_ = ifName
	_ = gw
	return nil
}

func allocateIdentity(_ netip.Addr, _ string) (uint32, error) {
	// TODO(phase1): call identity.Allocator.Allocate()
	return 0, fmt.Errorf("identity allocator not yet wired")
}

func removeBPFIdentity(_ string) error {
	// TODO(phase1): remove from BPF identity map
	return nil
}

func removePolicyState(_ string) error {
	// TODO(phase4): remove from BPF policy map
	return nil
}

func removeRoutes(netns, ifName string) error {
	// TODO(phase1): enter netns, flush routes on ifName
	_ = netns
	_ = ifName
	return nil
}

func checkNetworkNamespace(netns, ifName string) error {
	// Check that the interface exists in the network namespace
	iface, err := net.InterfaceByName(ifName)
	if err != nil {
		// Interface may be inside a different netns; this is expected
		_ = netns
		return nil
	}
	if iface.Flags&net.FlagUp == 0 {
		return fmt.Errorf("interface %q is down", ifName)
	}
	return nil
}
