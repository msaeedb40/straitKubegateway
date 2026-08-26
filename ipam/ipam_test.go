// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package ipam

import (
	"net/netip"
	"testing"

	"go.uber.org/zap"
)

func TestDynamicIPAMAllocationAndRelease(t *testing.T) {
	log := zap.NewNop()
	cidr := netip.MustParsePrefix("10.244.1.0/28") // small pool for testing arbitrary prefix lengths
	alloc := NewAllocator([]netip.Prefix{cidr}, log)

	if alloc.Used() != 0 {
		t.Fatalf("expected 0 used, got %d", alloc.Used())
	}

	// Gateway should be dynamically derived as the 1st host IP
	expectedGw := netip.MustParseAddr("10.244.1.1")
	firstIP, err := alloc.Allocate()
	if err != nil {
		t.Fatalf("allocate failed: %v", err)
	}

	gw, err := alloc.GatewayForIP(firstIP)
	if err != nil {
		t.Fatalf("gateway derivation failed: %v", err)
	}
	if gw != expectedGw {
		t.Fatalf("expected gateway %s, got %s", expectedGw, gw)
	}

	if firstIP == gw {
		t.Fatalf("allocated IP must not equal gateway address")
	}

	// Allocate remaining addresses
	allocated := []netip.Addr{firstIP}
	for i := 0; i < 12; i++ {
		ip, err := alloc.Allocate()
		if err != nil {
			t.Fatalf("allocation %d failed: %v", i, err)
		}
		if !cidr.Contains(ip) {
			t.Fatalf("allocated IP %s outside CIDR %s", ip, cidr)
		}
		allocated = append(allocated, ip)
	}

	if alloc.Used() != len(allocated) {
		t.Fatalf("expected %d used, got %d", len(allocated), alloc.Used())
	}

	// Release an address and re-allocate it
	releasedIP := allocated[0]
	alloc.Release(releasedIP)
	if alloc.Used() != len(allocated)-1 {
		t.Fatalf("expected %d used after release, got %d", len(allocated)-1, alloc.Used())
	}

	reallocatedIP, err := alloc.Allocate()
	if err != nil {
		t.Fatalf("re-allocate failed: %v", err)
	}
	if reallocatedIP != releasedIP {
		t.Logf("reallocated IP: %s (released: %s)", reallocatedIP, releasedIP)
	}
}

func TestDynamicDualStackAllocation(t *testing.T) {
	log := zap.NewNop()
	v4CIDR := netip.MustParsePrefix("192.168.100.0/24")
	v6CIDR := netip.MustParsePrefix("fd00:10:244::/64")

	alloc := NewAllocator([]netip.Prefix{v4CIDR, v6CIDR}, log)

	dual, err := alloc.AllocateDualStack()
	if err != nil {
		t.Fatalf("dual-stack allocate failed: %v", err)
	}

	if !dual.IPv4.IsValid() || !dual.IPv4.Is4() {
		t.Fatalf("expected valid IPv4, got %v", dual.IPv4)
	}
	if !dual.IPv6.IsValid() || !dual.IPv6.Is6() {
		t.Fatalf("expected valid IPv6, got %v", dual.IPv6)
	}

	if !v4CIDR.Contains(dual.IPv4) {
		t.Fatalf("IPv4 %s not in CIDR %s", dual.IPv4, v4CIDR)
	}
	if !v6CIDR.Contains(dual.IPv6) {
		t.Fatalf("IPv6 %s not in CIDR %s", dual.IPv6, v6CIDR)
	}

	// Test dynamic CIDR addition on the fly
	newCIDR := netip.MustParsePrefix("172.16.50.0/26")
	alloc.AddCIDR(newCIDR)
	if len(alloc.cidrs) != 3 {
		t.Fatalf("expected 3 pools after AddCIDR, got %d", len(alloc.cidrs))
	}
}

func TestBroadcastAddressNotAllocated(t *testing.T) {
	log := zap.NewNop()
	// /29 has 8 addresses: .0 (base), .1 (gw), .2-.6 (5 usable hosts), .7 (broadcast)
	cidr := netip.MustParsePrefix("10.0.0.0/29")
	alloc := NewAllocator([]netip.Prefix{cidr}, log)

	broadcast := netip.MustParseAddr("10.0.0.7")
	base := netip.MustParseAddr("10.0.0.0")
	gw := netip.MustParseAddr("10.0.0.1")

	var allocated []netip.Addr
	for i := 0; i < 5; i++ {
		ip, err := alloc.Allocate()
		if err != nil {
			t.Fatalf("allocate %d failed: %v", i, err)
		}
		if ip == broadcast {
			t.Fatalf("allocated broadcast address %s", ip)
		}
		if ip == base {
			t.Fatalf("allocated base address %s", ip)
		}
		if ip == gw {
			t.Fatalf("allocated gateway address %s", ip)
		}
		allocated = append(allocated, ip)
	}

	// Next allocation should fail because pool is exhausted
	_, err := alloc.Allocate()
	if err == nil {
		t.Fatalf("expected pool to be exhausted, but allocation succeeded")
	}
}
