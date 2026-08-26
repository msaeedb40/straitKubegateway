// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package ir

import (
	"net/netip"
	"testing"

	sgtypes "github.com/straitkubegateway/straitkubegateway/pkg/types"
)

// ============================================================================
// NetworkState
// ============================================================================

func TestNewNetworkState(t *testing.T) {
	ns := NewNetworkState()
	if ns == nil {
		t.Fatal("expected non-nil NetworkState")
	}
	if ns.Endpoints == nil {
		t.Error("Endpoints map should be initialized")
	}
	if ns.Services == nil {
		t.Error("Services map should be initialized")
	}
	if ns.Generation != 0 {
		t.Errorf("expected Generation 0, got %d", ns.Generation)
	}
	if len(ns.Policies) != 0 {
		t.Errorf("expected empty Policies, got %d", len(ns.Policies))
	}
	if len(ns.Routes) != 0 {
		t.Errorf("expected empty Routes, got %d", len(ns.Routes))
	}
	if len(ns.NATRules) != 0 {
		t.Errorf("expected empty NATRules, got %d", len(ns.NATRules))
	}
	if len(ns.TunnelPeers) != 0 {
		t.Errorf("expected empty TunnelPeers, got %d", len(ns.TunnelPeers))
	}
	if len(ns.TransitSegments) != 0 {
		t.Errorf("expected empty TransitSegments, got %d", len(ns.TransitSegments))
	}
}

func TestNetworkStateMutation(t *testing.T) {
	ns := NewNetworkState()

	// Add an endpoint
	ep := &Endpoint{
		ID:       1,
		Identity: sgtypes.Identity(256),
		IP:       netip.MustParseAddr("10.0.0.1"),
	}
	ns.Endpoints[ep.ID] = ep
	ns.Generation++

	if len(ns.Endpoints) != 1 {
		t.Errorf("expected 1 endpoint, got %d", len(ns.Endpoints))
	}
	if ns.Generation != 1 {
		t.Errorf("expected Generation 1, got %d", ns.Generation)
	}
}

// ============================================================================
// Endpoint IR
// ============================================================================

func TestEndpointFields(t *testing.T) {
	ep := &Endpoint{
		ID:               42,
		Identity:         sgtypes.Identity(500),
		IP:               netip.MustParseAddr("10.244.0.5"),
		IPv6:             netip.MustParseAddr("fd00::5"),
		NodeIP:           netip.MustParseAddr("192.168.1.10"),
		Namespace:        "production",
		PodName:          "nginx-abc123",
		Labels:           map[string]string{"app": "nginx", "env": "prod"},
		NetNSPath:        "/var/run/netns/abc123",
		IfIndex:          7,
		ContainerIfIndex: 3,
		SegmentID:        sgtypes.SegmentID(100),
		Generation:       10,
	}

	if ep.ID != 42 {
		t.Errorf("expected ID 42, got %d", ep.ID)
	}
	if ep.Identity != sgtypes.Identity(500) {
		t.Errorf("expected identity 500, got %s", ep.Identity)
	}
	if ep.IP.String() != "10.244.0.5" {
		t.Errorf("expected IP 10.244.0.5, got %s", ep.IP)
	}
	if ep.IPv6.String() != "fd00::5" {
		t.Errorf("expected IPv6 fd00::5, got %s", ep.IPv6)
	}
	if ep.Namespace != "production" {
		t.Errorf("expected namespace production, got %s", ep.Namespace)
	}
	if ep.PodName != "nginx-abc123" {
		t.Errorf("expected pod name nginx-abc123, got %s", ep.PodName)
	}
	if ep.SegmentID != sgtypes.SegmentID(100) {
		t.Errorf("expected segment 100, got %s", ep.SegmentID)
	}
	if ep.Labels["app"] != "nginx" {
		t.Errorf("expected label app=nginx, got %s", ep.Labels["app"])
	}
}

// ============================================================================
// Service IR
// ============================================================================

func TestServiceIR(t *testing.T) {
	svc := &Service{
		Key:       ServiceKey{Namespace: "default", Name: "orders"},
		Type:      ServiceTypeClusterIP,
		ClusterIP: netip.MustParseAddr("10.96.0.10"),
		Ports: []ServicePort{
			{Protocol: sgtypes.ProtocolTCP, Port: 80, TargetPort: 8080},
			{Protocol: sgtypes.ProtocolTCP, Port: 443, TargetPort: 8443, NodePort: 30443},
		},
		Backends: []Backend{
			{ID: 1, IP: netip.MustParseAddr("10.244.0.5"), Port: 8080, Weight: 100, Healthy: true},
			{ID: 2, IP: netip.MustParseAddr("10.244.0.6"), Port: 8080, Weight: 50, Healthy: false},
		},
		Algorithm:       sgtypes.LBAlgorithmMaglev,
		SessionAffinity: true,
		DSR:             false,
		Generation:      5,
	}

	if svc.Key.Name != "orders" {
		t.Errorf("expected service name orders, got %s", svc.Key.Name)
	}
	if svc.Type != ServiceTypeClusterIP {
		t.Errorf("expected ClusterIP type, got %s", svc.Type)
	}
	if len(svc.Ports) != 2 {
		t.Errorf("expected 2 ports, got %d", len(svc.Ports))
	}
	if len(svc.Backends) != 2 {
		t.Errorf("expected 2 backends, got %d", len(svc.Backends))
	}
	if !svc.Backends[0].Healthy {
		t.Error("expected backend 0 to be healthy")
	}
	if svc.Backends[1].Healthy {
		t.Error("expected backend 1 to be unhealthy")
	}
	if svc.Algorithm != sgtypes.LBAlgorithmMaglev {
		t.Errorf("expected maglev algorithm, got %s", svc.Algorithm)
	}
	if !svc.SessionAffinity {
		t.Error("expected session affinity enabled")
	}
}

func TestServiceTypes(t *testing.T) {
	types := []ServiceType{
		ServiceTypeClusterIP,
		ServiceTypeNodePort,
		ServiceTypeLoadBalancer,
		ServiceTypeExternalName,
	}
	expected := []string{"ClusterIP", "NodePort", "LoadBalancer", "ExternalName"}

	for i, st := range types {
		if string(st) != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], st)
		}
	}
}

// ============================================================================
// Policy IR
// ============================================================================

func TestPolicyIR(t *testing.T) {
	pol := &Policy{
		ID:       "default/web-policy/ingress/0",
		Priority: 50,
		Direction: PolicyDirectionIngress,
		SourceIdentities: []sgtypes.Identity{sgtypes.Identity(256), sgtypes.Identity(257)},
		DestIdentities:   []sgtypes.Identity{sgtypes.Identity(300)},
		Ports:            []sgtypes.Protocol{sgtypes.ProtocolTCP},
		Action:           sgtypes.PolicyActionAllow,
		Generation:       3,
	}

	if pol.Priority != 50 {
		t.Errorf("expected priority 50, got %d", pol.Priority)
	}
	if pol.Direction != PolicyDirectionIngress {
		t.Errorf("expected ingress direction, got %s", pol.Direction)
	}
	if len(pol.SourceIdentities) != 2 {
		t.Errorf("expected 2 source identities, got %d", len(pol.SourceIdentities))
	}
	if pol.Action != sgtypes.PolicyActionAllow {
		t.Errorf("expected Allow action, got %s", pol.Action)
	}
}

func TestPolicyDirections(t *testing.T) {
	if PolicyDirectionIngress != "Ingress" {
		t.Errorf("expected Ingress, got %s", PolicyDirectionIngress)
	}
	if PolicyDirectionEgress != "Egress" {
		t.Errorf("expected Egress, got %s", PolicyDirectionEgress)
	}
}

// ============================================================================
// Route IR
// ============================================================================

func TestRouteIR(t *testing.T) {
	route := &Route{
		Destination: netip.MustParsePrefix("10.20.0.0/16"),
		NextHop:     netip.MustParseAddr("172.30.0.5"),
		Dev:         "sg-vxlan0",
		Type:        RouteTypeVXLAN,
		Metric:      100,
		TableID:     42,
		Generation:  1,
	}

	if route.Destination.String() != "10.20.0.0/16" {
		t.Errorf("expected dest 10.20.0.0/16, got %s", route.Destination)
	}
	if route.Type != RouteTypeVXLAN {
		t.Errorf("expected vxlan type, got %s", route.Type)
	}
	if route.Metric != 100 {
		t.Errorf("expected metric 100, got %d", route.Metric)
	}
}

func TestRouteTypes(t *testing.T) {
	types := []RouteType{
		RouteTypeDirect,
		RouteTypeVXLAN,
		RouteTypeGeneve,
		RouteTypeGRE,
		RouteTypeWireGuard,
	}
	expected := []string{"direct", "vxlan", "geneve", "gre", "wireguard"}

	for i, rt := range types {
		if string(rt) != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], rt)
		}
	}
}

// ============================================================================
// NAT IR
// ============================================================================

func TestNATRuleIR(t *testing.T) {
	rule := &NATRule{
		Type:        NATTypeSNAT,
		Match:       netip.MustParsePrefix("10.244.0.0/16"),
		RewriteIP:   netip.MustParseAddr("192.168.1.1"),
		RewritePort: 0,
		OutDev:      "eth0",
		Generation:  2,
	}

	if rule.Type != NATTypeSNAT {
		t.Errorf("expected SNAT, got %s", rule.Type)
	}
	if rule.Match.String() != "10.244.0.0/16" {
		t.Errorf("expected match 10.244.0.0/16, got %s", rule.Match)
	}
}

func TestNATTypes(t *testing.T) {
	types := []NATType{NATTypeSNAT, NATTypeDNAT, NATTypeMasq, NATTypeNAT64}
	expected := []string{"SNAT", "DNAT", "Masquerade", "NAT64"}

	for i, nt := range types {
		if string(nt) != expected[i] {
			t.Errorf("expected %s, got %s", expected[i], nt)
		}
	}
}

// ============================================================================
// Tunnel IR
// ============================================================================

func TestTunnelPeerIR(t *testing.T) {
	tp := &TunnelPeer{
		NodeIP:   netip.MustParseAddr("192.168.1.20"),
		TunnelIP: netip.MustParseAddr("172.30.0.20"),
		PodCIDR:  netip.MustParsePrefix("10.244.1.0/24"),
		Mode:     sgtypes.TunnelModeVXLAN,
		Port:     4789,
	}

	if tp.Mode != sgtypes.TunnelModeVXLAN {
		t.Errorf("expected vxlan mode, got %s", tp.Mode)
	}
	if tp.Port != 4789 {
		t.Errorf("expected port 4789, got %d", tp.Port)
	}
}

// ============================================================================
// Transit Segment IR
// ============================================================================

func TestTransitSegmentIR(t *testing.T) {
	seg := &TransitSegment{
		ID:       sgtypes.SegmentID(100),
		Clusters: []sgtypes.ClusterID{"cluster-a", "cluster-b"},
		Routes: []TransitRoute{
			{
				CIDR:    netip.MustParsePrefix("10.20.0.0/16"),
				NextHop: "gateway-cluster-b",
			},
		},
		Generation: 7,
	}

	if seg.ID != sgtypes.SegmentID(100) {
		t.Errorf("expected segment 100, got %s", seg.ID)
	}
	if len(seg.Clusters) != 2 {
		t.Errorf("expected 2 clusters, got %d", len(seg.Clusters))
	}
	if len(seg.Routes) != 1 {
		t.Errorf("expected 1 route, got %d", len(seg.Routes))
	}
	if seg.Routes[0].NextHop != "gateway-cluster-b" {
		t.Errorf("expected nexthop gateway-cluster-b, got %s", seg.Routes[0].NextHop)
	}
}

func TestTransitSegmentBackbone(t *testing.T) {
	seg := &TransitSegment{
		ID: sgtypes.SegmentBackbone,
	}
	if !seg.ID.IsBackbone() {
		t.Error("expected backbone segment")
	}
}

// ============================================================================
// Generation
// ============================================================================

func TestGenerationMonotonicity(t *testing.T) {
	var g Generation = 0
	for i := 0; i < 10; i++ {
		prev := g
		g++
		if g <= prev {
			t.Errorf("generation should be monotonically increasing: prev=%d, current=%d", prev, g)
		}
	}
}

func TestGenerationOverflow(t *testing.T) {
	var g Generation = ^Generation(0) - 1 // near max uint64
	g++
	if g != ^Generation(0) {
		t.Errorf("expected max generation, got %d", g)
	}
}
