// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package nat implements SNAT, DNAT, Masquerade, Conntrack, and NAT64
// for the straitKubegateway eBPF dataplane.
package nat

import (
	"fmt"
	"net/netip"
	"sync"

	"go.uber.org/zap"

	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/ir"
)

// ============================================================================
// NAT Manager
// ============================================================================

// Manager manages NAT rules and conntrack state for straitKubegateway.
type Manager struct {
	mu    sync.RWMutex
	rules []*ir.NATRule
	log   *zap.Logger
}

// NewManager creates a new NAT manager.
func NewManager(log *zap.Logger) *Manager {
	return &Manager{log: log}
}

// ============================================================================
// Phase 3A: SNAT / DNAT / Masquerade / Conntrack
// ============================================================================

// AddSNATRule adds a source NAT rule.
// Traffic matching src will have its source address rewritten to rewriteIP.
func (m *Manager) AddSNATRule(src netip.Prefix, rewriteIP netip.Addr, outDev string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rules = append(m.rules, &ir.NATRule{
		Type:      ir.NATTypeSNAT,
		Match:     src,
		RewriteIP: rewriteIP,
		OutDev:    outDev,
	})
	m.log.Debug("SNAT rule added",
		zap.String("src", src.String()),
		zap.String("rewriteIP", rewriteIP.String()),
		zap.String("outDev", outDev),
	)
}

// AddDNATRule adds a destination NAT rule.
// Traffic to dstIP:dstPort will be redirected to rewriteIP:rewritePort.
func (m *Manager) AddDNATRule(dst netip.Prefix, rewriteIP netip.Addr, rewritePort uint16) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rules = append(m.rules, &ir.NATRule{
		Type:        ir.NATTypeDNAT,
		Match:       dst,
		RewriteIP:   rewriteIP,
		RewritePort: rewritePort,
	})
	m.log.Debug("DNAT rule added",
		zap.String("dst", dst.String()),
		zap.String("rewriteIP", rewriteIP.String()),
		zap.Uint16("rewritePort", rewritePort),
	)
}

// AddMasqueradeRule adds a masquerade rule for the given source prefix.
// Used for pod-to-external traffic.
func (m *Manager) AddMasqueradeRule(src netip.Prefix, outDev string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rules = append(m.rules, &ir.NATRule{
		Type:   ir.NATTypeMasq,
		Match:  src,
		OutDev: outDev,
	})
	m.log.Debug("masquerade rule added",
		zap.String("src", src.String()),
		zap.String("outDev", outDev),
	)
}

// GetRules returns the current NAT rule set for compilation into IR.
func (m *Manager) GetRules() []*ir.NATRule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*ir.NATRule, len(m.rules))
	copy(out, m.rules)
	return out
}

// ============================================================================
// Phase 3B: NAT64
// ============================================================================

// NAT64Prefix is the well-known NAT64 prefix (RFC 6052, 64:ff9b::/96).
var NAT64Prefix = netip.MustParsePrefix("64:ff9b::/96")

// AddNAT64Rule adds a NAT64 translation rule for IPv6→IPv4 traffic.
func (m *Manager) AddNAT64Rule(ipv6Prefix netip.Prefix, ipv4Pool netip.Prefix) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rules = append(m.rules, &ir.NATRule{
		Type:  ir.NATTypeNAT64,
		Match: ipv6Prefix,
		// RewriteIP holds the first address of the IPv4 pool
		RewriteIP: ipv4Pool.Addr(),
	})
	m.log.Debug("NAT64 rule added",
		zap.String("ipv6Prefix", ipv6Prefix.String()),
		zap.String("ipv4Pool", ipv4Pool.String()),
	)
}

// TranslateNAT64 performs the IPv6→IPv4 address extraction from an NAT64 address.
// IPv4 address is embedded in the low 32 bits of the NAT64 address.
func TranslateNAT64(ipv6 netip.Addr) (netip.Addr, error) {
	if !ipv6.Is6() {
		return netip.Addr{}, fmt.Errorf("not an IPv6 address: %s", ipv6)
	}
	b := ipv6.As16()
	// For 64:ff9b::/96, the IPv4 address is in bytes 12–15
	ipv4 := netip.AddrFrom4([4]byte{b[12], b[13], b[14], b[15]})
	return ipv4, nil
}

// SynthesizeNAT64 synthesizes an NAT64 IPv6 address from an IPv4 address
// using the well-known prefix 64:ff9b::/96.
func SynthesizeNAT64(ipv4 netip.Addr) (netip.Addr, error) {
	if !ipv4.Is4() {
		return netip.Addr{}, fmt.Errorf("not an IPv4 address: %s", ipv4)
	}
	pfx := NAT64Prefix.Addr().As16()
	v4 := ipv4.As4()
	pfx[12] = v4[0]
	pfx[13] = v4[1]
	pfx[14] = v4[2]
	pfx[15] = v4[3]
	return netip.AddrFrom16(pfx), nil
}

// ============================================================================
// Conntrack
// ============================================================================

// ConntrackEntry represents a single connection tracking entry.
type ConntrackEntry struct {
	SrcIP   netip.Addr
	DstIP   netip.Addr
	SrcPort uint16
	DstPort uint16
	Proto   uint8

	// NATed is the post-NAT address/port
	NATedIP   netip.Addr
	NATedPort uint16

	// State is the conntrack state (NEW, ESTABLISHED, etc.)
	State string
}

// ConntrackTable is an in-memory conntrack table for testing and fallback.
// In production, conntrack state is maintained in BPF maps.
type ConntrackTable struct {
	mu      sync.RWMutex
	entries map[conntrackKey]*ConntrackEntry
}

type conntrackKey struct {
	srcIP, dstIP     netip.Addr
	srcPort, dstPort uint16
	proto            uint8
}

// NewConntrackTable creates a new conntrack table.
func NewConntrackTable() *ConntrackTable {
	return &ConntrackTable{
		entries: make(map[conntrackKey]*ConntrackEntry),
	}
}

// Insert inserts or updates a conntrack entry.
func (t *ConntrackTable) Insert(e *ConntrackEntry) {
	t.mu.Lock()
	defer t.mu.Unlock()
	k := conntrackKey{e.SrcIP, e.DstIP, e.SrcPort, e.DstPort, e.Proto}
	t.entries[k] = e
}

// Lookup finds a conntrack entry by 5-tuple.
func (t *ConntrackTable) Lookup(srcIP, dstIP netip.Addr, srcPort, dstPort uint16, proto uint8) (*ConntrackEntry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	k := conntrackKey{srcIP, dstIP, srcPort, dstPort, proto}
	e, ok := t.entries[k]
	return e, ok
}

// Delete removes a conntrack entry.
func (t *ConntrackTable) Delete(srcIP, dstIP netip.Addr, srcPort, dstPort uint16, proto uint8) {
	t.mu.Lock()
	defer t.mu.Unlock()
	k := conntrackKey{srcIP, dstIP, srcPort, dstPort, proto}
	delete(t.entries, k)
}

// Count returns the number of active conntrack entries.
func (t *ConntrackTable) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.entries)
}
