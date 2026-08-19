package nat_test

import (
	"net/netip"
	"testing"
	"time"

	"github.com/straitKubegateway/straitKubegateway/nat"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

func TestConntrackTable(t *testing.T) {
	ct := nat.NewConntrackTable(100*time.Millisecond, 50*time.Millisecond)

	fwd := nat.Tuple{
		SrcIP:    netip.MustParseAddr("10.244.1.5"),
		DstIP:    netip.MustParseAddr("1.1.1.1"),
		SrcPort:  45678,
		DstPort:  443,
		Protocol: types.ProtocolTCP,
	}
	rev := nat.Tuple{
		SrcIP:    netip.MustParseAddr("1.1.1.1"),
		DstIP:    netip.MustParseAddr("192.168.1.100"),
		SrcPort:  443,
		DstPort:  32800,
		Protocol: types.ProtocolTCP,
	}

	entry := ct.Track(fwd, rev, nat.StateEstablished, 1)
	if entry.State != nat.StateEstablished {
		t.Errorf("expected StateEstablished, got %v", entry.State)
	}

	if ct.ActiveConnections() != 1 {
		t.Errorf("expected 1 active connection, got %d", ct.ActiveConnections())
	}

	// Lookup forward
	fwdEntry, found := ct.LookupForward(fwd)
	if !found || fwdEntry != entry {
		t.Errorf("forward lookup failed")
	}

	// Lookup reverse
	revEntry, found := ct.LookupReverse(rev)
	if !found || revEntry != entry {
		t.Errorf("reverse lookup failed")
	}

	// Test eviction
	time.Sleep(150 * time.Millisecond)
	evicted := ct.EvictExpired()
	if evicted != 1 {
		t.Errorf("expected 1 evicted connection, got %d", evicted)
	}
	if ct.ActiveConnections() != 0 {
		t.Errorf("expected 0 active connections after eviction, got %d", ct.ActiveConnections())
	}
}

func TestSNATEngine(t *testing.T) {
	engine := nat.NewSNATEngine(32000, 32005)

	rule := nat.SNATRule{
		SourceCIDR: netip.MustParsePrefix("10.244.0.0/16"),
		EgressIP:   netip.MustParseAddr("192.168.1.50"),
	}
	engine.AddRule(rule)

	srcIP := netip.MustParseAddr("10.244.1.20")
	egressIP, port1, err := engine.Translate(srcIP, 12345, types.ProtocolTCP)
	if err != nil {
		t.Fatalf("SNAT translate failed: %v", err)
	}
	if egressIP.String() != "192.168.1.50" {
		t.Errorf("expected egress IP 192.168.1.50, got %s", egressIP)
	}
	if port1 < 32000 || port1 > 32005 {
		t.Errorf("port %d out of expected range [32000, 32005]", port1)
	}

	// Allocate next port
	_, port2, err := engine.Translate(srcIP, 12346, types.ProtocolTCP)
	if err != nil {
		t.Fatalf("SNAT translate 2 failed: %v", err)
	}
	if port1 == port2 {
		t.Errorf("expected distinct allocated ports, got %d and %d", port1, port2)
	}

	// Release port1
	engine.ReleasePort(egressIP, port1, types.ProtocolTCP)
}

func TestDNATEngine(t *testing.T) {
	dnat := nat.NewDNATEngine()

	rule := nat.DNATRule{
		VIP:        netip.MustParseAddr("198.51.100.1"),
		Port:       80,
		Protocol:   types.ProtocolTCP,
		TargetIP:   netip.MustParseAddr("10.244.2.10"),
		TargetPort: 8080,
	}
	dnat.AddRule(rule)

	targetIP, targetPort, err := dnat.Translate(netip.MustParseAddr("198.51.100.1"), 80, types.ProtocolTCP)
	if err != nil {
		t.Fatalf("DNAT translate failed: %v", err)
	}
	if targetIP.String() != "10.244.2.10" || targetPort != 8080 {
		t.Errorf("expected 10.244.2.10:8080, got %s:%d", targetIP, targetPort)
	}
}

func TestMasqueradeEngine(t *testing.T) {
	masq := nat.NewMasqueradeEngine()

	rule := nat.MasqueradeRule{
		PodCIDR: netip.MustParsePrefix("10.244.0.0/16"),
		ExcludedCIDRs: []netip.Prefix{
			netip.MustParsePrefix("10.96.0.0/12"),  // Service CIDR
			netip.MustParsePrefix("10.244.0.0/16"), // Internal Pod CIDR
		},
		EgressNodeIP: netip.MustParseAddr("192.168.1.10"),
	}
	masq.AddRule(rule)

	srcPod := netip.MustParseAddr("10.244.1.15")

	// 1. External internet traffic -> Should masquerade
	shouldMasq, nodeIP := masq.ShouldMasquerade(srcPod, netip.MustParseAddr("8.8.8.8"))
	if !shouldMasq || nodeIP.String() != "192.168.1.10" {
		t.Errorf("expected masquerade to 192.168.1.10, got masq=%v ip=%s", shouldMasq, nodeIP)
	}

	// 2. Intra-cluster service traffic -> Should NOT masquerade
	shouldMasqSvc, _ := masq.ShouldMasquerade(srcPod, netip.MustParseAddr("10.96.0.1"))
	if shouldMasqSvc {
		t.Errorf("expected intra-cluster service traffic NOT to be masqueraded")
	}
}

func TestNAT64Engine(t *testing.T) {
	nat64, err := nat.NewNAT64Engine(nat.DefaultNAT64Prefix, "192.0.2.0/24")
	if err != nil {
		t.Fatalf("create NAT64 engine failed: %v", err)
	}

	ipv4 := netip.MustParseAddr("198.51.100.1")
	synthIPv6, err := nat64.SynthesizeIPv6(ipv4)
	if err != nil {
		t.Fatalf("synthesize IPv6 failed: %v", err)
	}

	if !nat64.IsNAT64Address(synthIPv6) {
		t.Errorf("expected synthesized address %s to be recognized as NAT64", synthIPv6)
	}

	extractedIPv4, err := nat64.ExtractIPv4(synthIPv6)
	if err != nil {
		t.Fatalf("extract IPv4 failed: %v", err)
	}
	if extractedIPv4 != ipv4 {
		t.Errorf("expected extracted IPv4 %s, got %s", ipv4, extractedIPv4)
	}
}
