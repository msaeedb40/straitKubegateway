// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package types

import (
	"net/netip"
	"testing"
)

func TestIdentity(t *testing.T) {
	if !IdentityWorld.IsReserved() {
		t.Errorf("IdentityWorld should be reserved")
	}
	if !IdentityHost.IsReserved() {
		t.Errorf("IdentityHost should be reserved")
	}
	if IdentityMin.IsReserved() {
		t.Errorf("IdentityMin should not be reserved")
	}

	if got := IdentityWorld.String(); got != "world" {
		t.Errorf("got %s, want world", got)
	}
	if got := Identity(500).String(); got != "500" {
		t.Errorf("got %s, want 500", got)
	}
}

func TestSegmentID(t *testing.T) {
	if !SegmentBackbone.IsBackbone() {
		t.Errorf("SegmentBackbone should be backbone (0)")
	}
	if SegmentID(10).IsBackbone() {
		t.Errorf("SegmentID(10) should not be backbone")
	}
	if got := SegmentBackbone.String(); got != "0" {
		t.Errorf("got %s, want 0", got)
	}
}

func TestCIDR(t *testing.T) {
	cidr, err := ParseCIDR("10.244.0.0/16")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cidr.IsIPv4() {
		t.Errorf("expected IPv4")
	}
	if cidr.IsIPv6() {
		t.Errorf("expected not IPv6")
	}

	ipInside := netip.MustParseAddr("10.244.1.5")
	ipOutside := netip.MustParseAddr("10.245.1.5")

	if !cidr.Contains(ipInside) {
		t.Errorf("expected %s to be inside %s", ipInside, cidr)
	}
	if cidr.Contains(ipOutside) {
		t.Errorf("expected %s to be outside %s", ipOutside, cidr)
	}

	other := MustParseCIDR("10.244.1.0/24")
	if !cidr.Overlaps(other) {
		t.Errorf("expected %s to overlap with %s", cidr, other)
	}
}

func TestMAC(t *testing.T) {
	macStr := "00:1a:2b:3c:4d:5e"
	mac, err := ParseMAC(macStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mac.String() != macStr {
		t.Errorf("got %s, want %s", mac.String(), macStr)
	}
	if mac.IsZero() {
		t.Errorf("expected non-zero MAC")
	}
	if !(MAC{}).IsZero() {
		t.Errorf("expected zero MAC")
	}
}

func TestSplitCIDR(t *testing.T) {
	cidrs, err := SplitCIDR("10.0.0.0/8, 172.16.0.0/12 , 192.168.0.0/16")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cidrs) != 3 {
		t.Fatalf("got %d CIDRs, want 3", len(cidrs))
	}
}
