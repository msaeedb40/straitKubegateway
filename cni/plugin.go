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

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"

	"github.com/straitkubegateway/straitkubegateway/identity"
	"github.com/straitkubegateway/straitkubegateway/ipam"
	"github.com/straitkubegateway/straitkubegateway/platform"
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
	mu            sync.RWMutex
	log           *zap.Logger
	ipamAlloc     *ipam.Allocator
	identityAlloc *identity.Allocator
	allocations   map[string]netip.Addr // containerID -> IP
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
		return nil, fmt.Errorf("CNI ADD: IPAM allocator not configured — " +
			"call WithIPAM() before ADD")
	}

	p.mu.Lock()
	p.allocations[containerID] = ip
	p.mu.Unlock()

	p.log.Debug("IPAM allocation", zap.String("ip", ip.String()), zap.String("gw", gw.String()))

	// hostDev is the host-side veth name, derived from containerID prefix
	hostDev := "sg-" + containerID[:min(len(containerID), 10)]

	// Step 3: Create NetKit/veth attachment (host ↔ container)
	hostIdx, containerIdx, err := setupNetKit(netns, hostDev, ifName)
	if err != nil {
		p.releaseAllocatedIP(containerID, ip)
		return nil, fmt.Errorf("CNI ADD netkit setup: %w", err)
	}

	// Step 4: Configure address on container interface
	if err := configureAddress(netns, ifName, ip, gw, cfg.MTU); err != nil {
		_ = teardownNetKit(hostDev)
		p.releaseAllocatedIP(containerID, ip)
		return nil, fmt.Errorf("CNI ADD configure address: %w", err)
	}

	// Step 5: Configure mandatory routes (default route via gateway)
	if err := configureMandatoryRoutes(netns, ifName, gw); err != nil {
		_ = teardownNetKit(hostDev)
		p.releaseAllocatedIP(containerID, ip)
		return nil, fmt.Errorf("CNI ADD configure routes: %w", err)
	}

	// Step 6: Allocate BPF identity
	// Best-effort: failure does NOT block pod creation (identity may be
	// deferred until the dataplane reconciles asynchronously).
	var bpfIdentity uint32
	if p.identityAlloc != nil {
		labelKey := identity.BuildIdentityKey(identity.Dimensions{
			Namespace: extractNamespace(cfg),
			PodLabels: map[string]string{"ip": ip.String()},
		})
		id, idErr := p.identityAlloc.Allocate(context.Background(), labelKey)
		if idErr != nil {
			p.log.Warn("BPF identity allocation deferred", zap.Error(idErr))
		} else {
			bpfIdentity = uint32(id)
		}
	} else {
		p.log.Warn("identity allocator not configured — BPF identity deferred")
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
		zap.String("gw", gw.String()),
		zap.Uint32("identity", bpfIdentity),
		zap.Int("hostIfIdx", hostIdx),
		zap.String("hostDev", hostDev),
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

	hostDev := "sg-" + containerID[:min(len(containerID), 10)]

	// 1. Remove routes inside the container netns
	if err := removeRoutes(netns, ifName); err != nil {
		p.log.Warn("CNI DEL remove routes", zap.Error(err))
	}
	// 2. Remove BPF identity
	if p.identityAlloc != nil {
		labelKey := identity.BuildIdentityKey(identity.Dimensions{
			PodLabels: map[string]string{"ip": allocatedIP.String()},
		})
		p.identityAlloc.Release(context.Background(), labelKey)
	}
	// 3. Remove policy state (best-effort, non-blocking)
	// Policy cleanup is handled asynchronously by the dataplane reconciler
	// 4. Release IP back to IPAM
	if hasIP && p.ipamAlloc != nil {
		p.ipamAlloc.Release(allocatedIP)
	}
	// 5. Destroy host-side NetKit interface (peer is destroyed automatically)
	if err := teardownNetKit(hostDev); err != nil {
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
	// Verify interface exists and is UP inside the container's network namespace
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
			if p.identityAlloc != nil {
				labelKey := identity.BuildIdentityKey(identity.Dimensions{
					PodLabels: map[string]string{"ip": ip.String()},
				})
				p.identityAlloc.Release(ctx, labelKey)
			}
			if p.ipamAlloc != nil {
				p.ipamAlloc.Release(ip)
			}
			hostDev := "sg-" + cid[:min(len(cid), 10)]
			_ = teardownNetKit(hostDev)
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

func (p *Plugin) releaseAllocatedIP(containerID string, ip netip.Addr) {
	p.mu.Lock()
	delete(p.allocations, containerID)
	p.mu.Unlock()
	if p.ipamAlloc != nil {
		p.ipamAlloc.Release(ip)
	}
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
// Internal helpers
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

// setupNetKit creates a veth pair: host-side stays in host ns, peer is moved
// into the container netns by ebpf.NetKitManager.SetupNetKit.
func setupNetKit(netnsPath, hostDev, containerDev string) (hostIdx, containerIdx int, err error) {
	la := netlink.NewLinkAttrs()
	la.Name = hostDev

	veth := &netlink.Veth{
		LinkAttrs: la,
		PeerName:  containerDev,
	}
	if err := netlink.LinkAdd(veth); err != nil {
		return 0, 0, fmt.Errorf("create veth %s<->%s: %w", hostDev, containerDev, err)
	}

	hostLink, err := netlink.LinkByName(hostDev)
	if err != nil {
		return 0, 0, fmt.Errorf("lookup host link %q: %w", hostDev, err)
	}
	peerLink, err := netlink.LinkByName(containerDev)
	if err != nil {
		_ = netlink.LinkDel(hostLink)
		return 0, 0, fmt.Errorf("lookup peer link %q: %w", containerDev, err)
	}

	// Move peer into container netns
	if netnsPath != "" {
		nsFD, err := unix.Open(netnsPath, unix.O_RDONLY|unix.O_CLOEXEC, 0)
		if err != nil {
			_ = netlink.LinkDel(hostLink)
			return 0, 0, fmt.Errorf("open container netns %q: %w", netnsPath, err)
		}
		defer unix.Close(nsFD)
		if err := netlink.LinkSetNsFd(peerLink, nsFD); err != nil {
			_ = netlink.LinkDel(hostLink)
			return 0, 0, fmt.Errorf("move peer %q into netns %q: %w", containerDev, netnsPath, err)
		}
	}

	if err := netlink.LinkSetUp(hostLink); err != nil {
		_ = netlink.LinkDel(hostLink)
		return 0, 0, fmt.Errorf("bring up host link %q: %w", hostDev, err)
	}

	return hostLink.Attrs().Index, peerLink.Attrs().Index, nil
}

func teardownNetKit(hostDev string) error {
	link, err := netlink.LinkByName(hostDev)
	if err != nil {
		return nil // already gone
	}
	return netlink.LinkDel(link)
}

// configureAddress enters the container netns and assigns the IP address and MTU.
func configureAddress(netnsPath, ifName string, ip, gw netip.Addr, mtu int) error {
	ns, err := platform.OpenNetNSByPath(netnsPath)
	if err != nil {
		return fmt.Errorf("open container netns %q: %w", netnsPath, err)
	}
	defer ns.Close()

	return ns.Do(func() error {
		link, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("link %q not found in container netns: %w", ifName, err)
		}

		bits := 32
		if ip.Is6() {
			bits = 128
		}
		addr := &netlink.Addr{
			IPNet: &net.IPNet{
				IP:   ip.AsSlice(),
				Mask: fullMask(bits),
			},
		}
		if err := netlink.AddrAdd(link, addr); err != nil {
			return fmt.Errorf("assign IP %s to %q: %w", ip, ifName, err)
		}

		if mtu > 0 {
			if err := netlink.LinkSetMTU(link, mtu); err != nil {
				return fmt.Errorf("set MTU %d on %q: %w", mtu, ifName, err)
			}
		}

		return netlink.LinkSetUp(link)
	})
}

// configureMandatoryRoutes adds the default route via gw inside the container netns.
func configureMandatoryRoutes(netnsPath, ifName string, gw netip.Addr) error {
	ns, err := platform.OpenNetNSByPath(netnsPath)
	if err != nil {
		return fmt.Errorf("open container netns %q: %w", netnsPath, err)
	}
	defer ns.Close()

	return ns.Do(func() error {
		link, err := netlink.LinkByName(ifName)
		if err != nil {
			return fmt.Errorf("link %q not found in container netns: %w", ifName, err)
		}
		gwSlice := gw.AsSlice()
		route := &netlink.Route{
			LinkIndex: link.Attrs().Index,
			Gw:        gwSlice,
		}
		if err := netlink.RouteAdd(route); err != nil {
			return fmt.Errorf("add default route via %s on %q: %w", gw, ifName, err)
		}
		return nil
	})
}

// removeRoutes flushes routes on ifName inside the container netns.
func removeRoutes(netnsPath, ifName string) error {
	ns, err := platform.OpenNetNSByPath(netnsPath)
	if err != nil {
		// Netns may already be gone on DEL — best-effort
		return nil
	}
	defer ns.Close()

	return ns.Do(func() error {
		link, err := netlink.LinkByName(ifName)
		if err != nil {
			return nil // interface already gone
		}
		routes, err := netlink.RouteList(link, netlink.FAMILY_ALL)
		if err != nil {
			return nil
		}
		for i := range routes {
			_ = netlink.RouteDel(&routes[i])
		}
		return nil
	})
}

// checkNetworkNamespace verifies that ifName exists and is UP inside the
// container's network namespace (not the host netns).
func checkNetworkNamespace(netnsPath, ifName string) error {
	ns, err := platform.OpenNetNSByPath(netnsPath)
	if err != nil {
		return fmt.Errorf("open container netns %q for CHECK: %w", netnsPath, err)
	}
	defer ns.Close()

	var checkErr error
	_ = ns.Do(func() error {
		link, err := netlink.LinkByName(ifName)
		if err != nil {
			checkErr = fmt.Errorf("interface %q not found in container netns: %w", ifName, err)
			return nil
		}
		if link.Attrs().Flags&net.FlagUp == 0 {
			checkErr = fmt.Errorf("interface %q is down in container netns", ifName)
		}
		return nil
	})
	return checkErr
}

// fullMask returns a full-coverage IP mask for the given bit length.
func fullMask(bits int) []byte {
	if bits == 32 {
		return []byte{0xff, 0xff, 0xff, 0xff}
	}
	mask := make([]byte, 16)
	for i := range mask {
		mask[i] = 0xff
	}
	return mask
}

// extractNamespace extracts the pod namespace from the CNI config if present.
func extractNamespace(cfg *NetworkConfig) string {
	if cfg == nil {
		return ""
	}
	return cfg.Name // placeholder; real impl reads from CNI_ARGS env
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
