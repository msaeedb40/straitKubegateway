// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package routing implements Linux FIB management, policy-based routing,
// tunnel setup, and netlink route programming for straitKubegateway.
package routing

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

// ============================================================================
// Manager
// ============================================================================

// Manager manages the Linux routing table for straitKubegateway.
type Manager struct {
	log       *zap.Logger
	tableID   int
	tunnelDev string
}

// NewManager creates a new routing Manager.
// tableID is the Linux policy routing table to use (0 = main table).
func NewManager(log *zap.Logger, tableID int) *Manager {
	return &Manager{
		log:     log,
		tableID: tableID,
	}
}

// ============================================================================
// Route programming
// ============================================================================

// AddRoute adds a unicast route to the configured routing table.
func (m *Manager) AddRoute(dst netip.Prefix, nexthop netip.Addr, dev string, metric uint32) error {
	link, err := netlink.LinkByName(dev)
	if err != nil {
		return fmt.Errorf("link %q: %w", dev, err)
	}

	ipNet := prefixToIPNet(dst)
	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       ipNet,
		Gw:        nexthop.AsSlice(),
		Priority:  int(metric),
		Table:     m.tableID,
	}
	if err := netlink.RouteAdd(route); err != nil {
		return fmt.Errorf("route add %s via %s dev %s: %w", dst, nexthop, dev, err)
	}
	m.log.Debug("route added",
		zap.String("dst", dst.String()),
		zap.String("nexthop", nexthop.String()),
		zap.String("dev", dev),
	)
	return nil
}

// DeleteRoute removes a route from the routing table.
func (m *Manager) DeleteRoute(dst netip.Prefix, dev string) error {
	link, err := netlink.LinkByName(dev)
	if err != nil {
		return fmt.Errorf("link %q: %w", dev, err)
	}
	ipNet := prefixToIPNet(dst)
	route := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       ipNet,
		Table:     m.tableID,
	}
	if err := netlink.RouteDel(route); err != nil {
		return fmt.Errorf("route del %s dev %s: %w", dst, dev, err)
	}
	m.log.Debug("route deleted", zap.String("dst", dst.String()))
	return nil
}

// EnsureRoute ensures the route exists, adding it if not present.
func (m *Manager) EnsureRoute(dst netip.Prefix, nexthop netip.Addr, dev string, metric uint32) error {
	existing, err := m.getRoute(dst, dev)
	if err == nil && existing != nil {
		return nil // already exists
	}
	return m.AddRoute(dst, nexthop, dev, metric)
}

// getRoute returns an existing route or nil.
func (m *Manager) getRoute(dst netip.Prefix, dev string) (*netlink.Route, error) {
	link, err := netlink.LinkByName(dev)
	if err != nil {
		return nil, err
	}
	filter := &netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       prefixToIPNet(dst),
		Table:     m.tableID,
	}
	routes, err := netlink.RouteListFiltered(netlink.FAMILY_ALL, filter, netlink.RT_FILTER_DST|netlink.RT_FILTER_OIF)
	if err != nil || len(routes) == 0 {
		return nil, fmt.Errorf("route not found")
	}
	return &routes[0], nil
}

// ============================================================================
// Tunnel management
// ============================================================================

// TunnelPeer is a remote node reachable via an overlay tunnel.
type TunnelPeer struct {
	NodeIP   netip.Addr
	PodCIDR  netip.Prefix
	TunnelIP netip.Addr
	Port     uint16
	Mode     string // "vxlan", "geneve", "gre"
}

// UpsertTunnelPeer programs the FDB entry and route for a remote tunnel peer.
func (m *Manager) UpsertTunnelPeer(peer TunnelPeer) error {
	m.log.Debug("upserting tunnel peer",
		zap.String("nodeIP", peer.NodeIP.String()),
		zap.String("mode", peer.Mode),
		zap.String("podCIDR", peer.PodCIDR.String()),
	)
	switch peer.Mode {
	case "vxlan":
		return m.upsertVXLANPeer(peer)
	case "geneve":
		return m.upsertGenevePeer(peer)
	case "gre":
		return m.upsertGREPeer(peer)
	default:
		return fmt.Errorf("unknown tunnel mode %q", peer.Mode)
	}
}

func (m *Manager) upsertVXLANPeer(peer TunnelPeer) error {
	// Program VXLAN FDB: associate peer MAC→VTEP IP in the VXLAN interface
	link, err := netlink.LinkByName("sg-vxlan0")
	if err != nil {
		return fmt.Errorf("sg-vxlan0 not found: %w", err)
	}
	// Zero MAC triggers flood-and-learn; use unicast for known peers
	entry := &netlink.Neigh{
		LinkIndex:    link.Attrs().Index,
		State:        netlink.NUD_PERMANENT,
		Family:       syscallAF_BRIDGE,
		Flags:        netlink.NTF_SELF,
		IP:           peer.TunnelIP.AsSlice(),
		HardwareAddr: net.HardwareAddr{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	if err := netlink.NeighSet(entry); err != nil {
		return fmt.Errorf("vxlan fdb set peer %s: %w", peer.NodeIP, err)
	}
	return nil
}

func (m *Manager) upsertGenevePeer(_ TunnelPeer) error {
	// TODO: Geneve peer programming via netlink
	return nil
}

func (m *Manager) upsertGREPeer(_ TunnelPeer) error {
	// TODO: GRE peer programming via netlink
	return nil
}

// DeleteTunnelPeer removes routing state for a remote tunnel peer.
func (m *Manager) DeleteTunnelPeer(peer TunnelPeer) error {
	m.log.Debug("deleting tunnel peer", zap.String("nodeIP", peer.NodeIP.String()))
	return m.DeleteRoute(peer.PodCIDR, m.tunnelDev)
}

// ============================================================================
// Bootstrap API path
// ============================================================================

// EnsureAPIServerRoute ensures the Kubernetes API server is reachable via
// the node routing table WITHOUT using the Service dataplane.
// Invariant: bootstrap API path must NOT depend on Service LB.
func (m *Manager) EnsureAPIServerRoute(apiServerAddr netip.Addr, dev string) error {
	apiServerPrefix := netip.PrefixFrom(apiServerAddr, apiServerAddr.BitLen())
	return m.EnsureRoute(apiServerPrefix, netip.Addr{}, dev, 0)
}

// ============================================================================
// Helpers
// ============================================================================

func prefixToIPNet(p netip.Prefix) *net.IPNet {
	ip := p.Addr().AsSlice()
	bits := p.Bits()
	total := p.Addr().BitLen()
	return &net.IPNet{
		IP:   ip,
		Mask: net.CIDRMask(bits, total),
	}
}

// syscallAF_BRIDGE is the AF_BRIDGE address family constant.
const syscallAF_BRIDGE = 7
