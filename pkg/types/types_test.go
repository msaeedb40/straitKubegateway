package types_test

import (
	"net/netip"
	"testing"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

func TestIdentity(t *testing.T) {
	if !types.IdentityHost.IsReserved() {
		t.Error("expected IdentityHost to be reserved")
	}
	if !types.IdentityWorld.IsReserved() {
		t.Error("expected IdentityWorld to be reserved")
	}
	if !types.IdentityKubeAPIServer.IsReserved() {
		t.Error("expected IdentityKubeAPIServer to be reserved")
	}
	if !types.IdentityCoreDNS.IsReserved() {
		t.Error("expected IdentityCoreDNS to be reserved")
	}
	if types.IdentityUnknown.IsValid() {
		t.Error("expected IdentityUnknown to be invalid")
	}

	userId := types.Identity(256)
	if userId.IsReserved() {
		t.Error("expected userId 256 to not be reserved")
	}
	if !userId.IsValid() {
		t.Error("expected userId 256 to be valid")
	}
}

func TestSegmentID(t *testing.T) {
	if !types.SegmentBackbone.IsBackbone() {
		t.Error("expected SegmentBackbone to report IsBackbone = true")
	}
	seg := types.SegmentID(42)
	if seg.IsBackbone() {
		t.Error("expected SegmentID(42) to report IsBackbone = false")
	}
	if seg.Uint32() != 42 {
		t.Errorf("expected 42, got %d", seg.Uint32())
	}
}

func TestCIDR(t *testing.T) {
	cidr, err := types.ParseCIDR("10.244.0.0/16")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cidr.IsIPv4() {
		t.Error("expected IPv4 CIDR")
	}
	if cidr.IsIPv6() {
		t.Error("expected non-IPv6 CIDR")
	}

	ipInside := netip.MustParseAddr("10.244.1.5")
	ipOutside := netip.MustParseAddr("10.245.1.5")

	if !cidr.Contains(ipInside) {
		t.Errorf("expected CIDR to contain %s", ipInside)
	}
	if cidr.Contains(ipOutside) {
		t.Errorf("expected CIDR not to contain %s", ipOutside)
	}
}

func TestPortRange(t *testing.T) {
	pr := types.PortRange{Start: 80, End: 443}
	if !pr.Contains(80) || !pr.Contains(443) || !pr.Contains(200) {
		t.Error("expected ports 80, 443, 200 to be in range")
	}
	if pr.Contains(79) || pr.Contains(444) {
		t.Error("expected ports 79, 444 to be outside range")
	}
}
