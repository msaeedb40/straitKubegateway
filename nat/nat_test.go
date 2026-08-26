// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package nat

import (
	"net/netip"
	"testing"

	"go.uber.org/zap"
)

func TestNAT64Translation(t *testing.T) {
	// IPv4 address to translate
	ipv4 := netip.MustParseAddr("192.0.2.33")

	// Synthesize NAT64 IPv6 address: 64:ff9b::192.0.2.33
	ipv6, err := SynthesizeNAT64(ipv4)
	if err != nil {
		t.Fatalf("unexpected error synthesizing NAT64: %v", err)
	}

	expectedPrefix := netip.MustParsePrefix("64:ff9b::/96")
	if !expectedPrefix.Contains(ipv6) {
		t.Errorf("synthesized IPv6 %s should be in %s", ipv6, expectedPrefix)
	}

	// Translate back to IPv4
	recoveredIPv4, err := TranslateNAT64(ipv6)
	if err != nil {
		t.Fatalf("unexpected error translating NAT64: %v", err)
	}

	if recoveredIPv4 != ipv4 {
		t.Errorf("got %s, want %s", recoveredIPv4, ipv4)
	}
}

func TestNATManager(t *testing.T) {
	log := zap.NewNop()
	mgr := NewManager(log)

	src := netip.MustParsePrefix("10.244.0.0/16")
	rewriteIP := netip.MustParseAddr("192.168.1.100")

	mgr.AddSNATRule(src, rewriteIP, "eth0")
	mgr.AddMasqueradeRule(src, "eth0")

	rules := mgr.GetRules()
	if len(rules) != 2 {
		t.Fatalf("got %d rules, want 2", len(rules))
	}
}

func TestConntrackTable(t *testing.T) {
	ct := NewConntrackTable()

	entry := &ConntrackEntry{
		SrcIP:     netip.MustParseAddr("10.244.1.2"),
		DstIP:     netip.MustParseAddr("10.96.0.1"),
		SrcPort:   45000,
		DstPort:   443,
		Proto:     6, // TCP
		NATedIP:   netip.MustParseAddr("10.244.2.5"),
		NATedPort: 6443,
		State:     "ESTABLISHED",
	}

	ct.Insert(entry)

	got, ok := ct.Lookup(entry.SrcIP, entry.DstIP, entry.SrcPort, entry.DstPort, entry.Proto)
	if !ok || got.NATedPort != 6443 {
		t.Errorf("failed to lookup conntrack entry")
	}

	if ct.Count() != 1 {
		t.Errorf("got count %d, want 1", ct.Count())
	}

	ct.Delete(entry.SrcIP, entry.DstIP, entry.SrcPort, entry.DstPort, entry.Proto)
	if ct.Count() != 0 {
		t.Errorf("got count %d after delete, want 0", ct.Count())
	}
}
