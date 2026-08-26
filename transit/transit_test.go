// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package transit

import (
	"context"
	"fmt"
	"net/netip"
	"testing"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apiv1 "github.com/straitkubegateway/straitkubegateway/api/v1alpha1"
	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/ir"
	sgtypes "github.com/straitkubegateway/straitkubegateway/pkg/types"
)

// ============================================================================
// Test helpers
// ============================================================================

func newTestManager(objs ...runtime.Object) *Manager {
	scheme := runtime.NewScheme()
	_ = apiv1.AddToScheme(scheme)

	builder := fake.NewClientBuilder().WithScheme(scheme)
	for _, obj := range objs {
		builder = builder.WithRuntimeObjects(obj)
	}
	c := builder.Build()
	return NewManager(c, zap.NewNop())
}

func makeTransitGateway(name string, segmentID uint32, topology string, peers []apiv1.TransitPeer) *apiv1.TransitGateway {
	return &apiv1.TransitGateway{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiv1.TransitGatewaySpec{
			SegmentID: segmentID,
			Topology:  topology,
			Peers:     peers,
		},
	}
}

func makeAttachment(name string, segA, segB uint32, routes []apiv1.TransitSegmentRoute) *apiv1.TransitSegmentAttachment {
	return &apiv1.TransitSegmentAttachment{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: apiv1.TransitSegmentAttachmentSpec{
			SegmentID:     segA,
			PeerSegmentID: segB,
			Routes:        routes,
		},
	}
}

// ============================================================================
// Unit Tests
// ============================================================================

func TestNewManager(t *testing.T) {
	mgr := newTestManager()
	if mgr == nil {
		t.Fatal("expected non-nil manager")
	}
	if mgr.segments == nil {
		t.Error("segments map should be initialized")
	}
	if mgr.gateways == nil {
		t.Error("gateways map should be initialized")
	}
	if len(mgr.segments) != 0 {
		t.Errorf("expected 0 segments, got %d", len(mgr.segments))
	}
	if len(mgr.gateways) != 0 {
		t.Errorf("expected 0 gateways, got %d", len(mgr.gateways))
	}
}

func TestReconcileGateway(t *testing.T) {
	tg := makeTransitGateway("gw-alpha", 100, "hub-spoke", []apiv1.TransitPeer{
		{
			ClusterName: "cluster-b",
			Endpoint:    "172.30.0.5",
			PodCIDRs:    []string{"10.20.0.0/16"},
		},
		{
			ClusterName: "cluster-c",
			Endpoint:    "172.30.0.6",
			PodCIDRs:    []string{"10.30.0.0/16", "10.31.0.0/16"},
		},
	})

	mgr := newTestManager(tg)
	ctx := context.Background()

	key := types.NamespacedName{Name: "gw-alpha"}
	if err := mgr.ReconcileGateway(ctx, key); err != nil {
		t.Fatalf("ReconcileGateway failed: %v", err)
	}

	// Verify segment was created
	seg, ok := mgr.segments[sgtypes.SegmentID(100)]
	if !ok {
		t.Fatal("expected segment 100 to exist")
	}

	// Verify routes from peers
	if len(seg.Routes) != 3 {
		t.Errorf("expected 3 routes, got %d", len(seg.Routes))
	}

	// Verify cluster IDs
	if len(seg.Clusters) != 2 {
		t.Errorf("expected 2 clusters, got %d", len(seg.Clusters))
	}
	foundB, foundC := false, false
	for _, cid := range seg.Clusters {
		if cid == sgtypes.ClusterID("cluster-b") {
			foundB = true
		}
		if cid == sgtypes.ClusterID("cluster-c") {
			foundC = true
		}
	}
	if !foundB || !foundC {
		t.Errorf("expected clusters cluster-b and cluster-c, got %v", seg.Clusters)
	}

	// Verify gateway map entry
	if _, ok := mgr.gateways[key]; !ok {
		t.Error("expected gateway map entry")
	}
}

func TestReconcileGatewayNotFound(t *testing.T) {
	mgr := newTestManager()
	ctx := context.Background()

	key := types.NamespacedName{Name: "nonexistent"}
	if err := mgr.ReconcileGateway(ctx, key); err != nil {
		t.Fatalf("expected nil error for not-found, got: %v", err)
	}
}

func TestReconcileGatewayInvalidPeerCIDR(t *testing.T) {
	tg := makeTransitGateway("gw-bad-cidr", 200, "mesh", []apiv1.TransitPeer{
		{
			ClusterName: "cluster-x",
			Endpoint:    "1.2.3.4",
			PodCIDRs:    []string{"not-a-cidr"},
		},
	})

	mgr := newTestManager(tg)
	ctx := context.Background()

	key := types.NamespacedName{Name: "gw-bad-cidr"}
	if err := mgr.ReconcileGateway(ctx, key); err == nil {
		t.Fatal("expected error for invalid peer CIDR")
	}
}

func TestReconcileGatewayBackboneSegment(t *testing.T) {
	tg := makeTransitGateway("gw-backbone", 0, "hub-spoke", []apiv1.TransitPeer{
		{
			ClusterName: "spoke-1",
			Endpoint:    "10.0.0.1",
			PodCIDRs:    []string{"10.100.0.0/16"},
		},
	})

	mgr := newTestManager(tg)
	ctx := context.Background()

	key := types.NamespacedName{Name: "gw-backbone"}
	if err := mgr.ReconcileGateway(ctx, key); err != nil {
		t.Fatalf("ReconcileGateway backbone failed: %v", err)
	}

	seg, ok := mgr.segments[sgtypes.SegmentBackbone]
	if !ok {
		t.Fatal("expected backbone segment 0")
	}
	if !seg.ID.IsBackbone() {
		t.Error("expected segment ID to be backbone")
	}
}

func TestDeleteGateway(t *testing.T) {
	tg := makeTransitGateway("gw-delete", 300, "peer-to-peer", []apiv1.TransitPeer{
		{
			ClusterName: "cluster-d",
			Endpoint:    "172.16.0.1",
			PodCIDRs:    []string{"10.40.0.0/16"},
		},
	})

	mgr := newTestManager(tg)
	ctx := context.Background()

	key := types.NamespacedName{Name: "gw-delete"}
	if err := mgr.ReconcileGateway(ctx, key); err != nil {
		t.Fatalf("ReconcileGateway failed: %v", err)
	}

	// Verify it exists
	if _, ok := mgr.segments[sgtypes.SegmentID(300)]; !ok {
		t.Fatal("expected segment 300 before delete")
	}

	mgr.DeleteGateway(key)

	// Verify cleanup
	if _, ok := mgr.segments[sgtypes.SegmentID(300)]; ok {
		t.Error("expected segment 300 to be removed after delete")
	}
	if _, ok := mgr.gateways[key]; ok {
		t.Error("expected gateway entry to be removed after delete")
	}
}

func TestDeleteGatewayNonexistent(t *testing.T) {
	mgr := newTestManager()
	key := types.NamespacedName{Name: "nope"}
	// Should not panic
	mgr.DeleteGateway(key)
}

func TestReconcileAttachment(t *testing.T) {
	att := makeAttachment("att-100-200", 100, 200, []apiv1.TransitSegmentRoute{
		{CIDR: "10.50.0.0/16", NextHop: "172.30.0.10"},
		{CIDR: "10.60.0.0/16", NextHop: "172.30.0.11"},
	})

	mgr := newTestManager(att)
	ctx := context.Background()

	key := types.NamespacedName{Name: "att-100-200"}
	if err := mgr.ReconcileAttachment(ctx, key); err != nil {
		t.Fatalf("ReconcileAttachment failed: %v", err)
	}

	// Both segments should exist
	segA, okA := mgr.segments[sgtypes.SegmentID(100)]
	segB, okB := mgr.segments[sgtypes.SegmentID(200)]
	if !okA || !okB {
		t.Fatalf("expected both segments to exist, got segA=%v segB=%v", okA, okB)
	}

	// Both segments should have the routes
	if len(segA.Routes) != 2 {
		t.Errorf("expected 2 routes on segA, got %d", len(segA.Routes))
	}
	if len(segB.Routes) != 2 {
		t.Errorf("expected 2 routes on segB, got %d", len(segB.Routes))
	}
}

func TestReconcileAttachmentInvalidCIDR(t *testing.T) {
	att := makeAttachment("att-bad", 100, 200, []apiv1.TransitSegmentRoute{
		{CIDR: "bad-cidr", NextHop: "172.30.0.10"},
		{CIDR: "10.70.0.0/16", NextHop: "172.30.0.12"},
	})

	mgr := newTestManager(att)
	ctx := context.Background()

	key := types.NamespacedName{Name: "att-bad"}
	// Invalid CIDR should be skipped (logged), not return an error
	if err := mgr.ReconcileAttachment(ctx, key); err != nil {
		t.Fatalf("expected no error (invalid CIDRs are skipped), got: %v", err)
	}

	// The valid route should still be programmed
	seg := mgr.segments[sgtypes.SegmentID(100)]
	if len(seg.Routes) != 1 {
		t.Errorf("expected 1 valid route, got %d", len(seg.Routes))
	}
}

func TestReconcileAttachmentNotFound(t *testing.T) {
	mgr := newTestManager()
	ctx := context.Background()

	key := types.NamespacedName{Name: "nonexistent"}
	if err := mgr.ReconcileAttachment(ctx, key); err != nil {
		t.Fatalf("expected nil error for not-found attachment, got: %v", err)
	}
}

func TestGetAllSegments(t *testing.T) {
	tg1 := makeTransitGateway("gw-1", 100, "mesh", []apiv1.TransitPeer{
		{ClusterName: "c1", Endpoint: "1.1.1.1", PodCIDRs: []string{"10.0.0.0/16"}},
	})
	tg2 := makeTransitGateway("gw-2", 200, "mesh", []apiv1.TransitPeer{
		{ClusterName: "c2", Endpoint: "2.2.2.2", PodCIDRs: []string{"10.1.0.0/16"}},
	})

	mgr := newTestManager(tg1, tg2)
	ctx := context.Background()

	_ = mgr.ReconcileGateway(ctx, types.NamespacedName{Name: "gw-1"})
	_ = mgr.ReconcileGateway(ctx, types.NamespacedName{Name: "gw-2"})

	all := mgr.GetAllSegments()
	if len(all) != 2 {
		t.Fatalf("expected 2 segments, got %d", len(all))
	}

	foundIDs := map[sgtypes.SegmentID]bool{}
	for _, seg := range all {
		foundIDs[seg.ID] = true
	}
	if !foundIDs[sgtypes.SegmentID(100)] || !foundIDs[sgtypes.SegmentID(200)] {
		t.Errorf("expected segments 100 and 200, got %v", foundIDs)
	}
}

func TestGetAllSegmentsEmpty(t *testing.T) {
	mgr := newTestManager()
	all := mgr.GetAllSegments()
	if len(all) != 0 {
		t.Errorf("expected 0 segments, got %d", len(all))
	}
}

func TestGetOrCreateSegment(t *testing.T) {
	mgr := newTestManager()

	seg1 := mgr.getOrCreateSegment(sgtypes.SegmentID(42))
	if seg1 == nil {
		t.Fatal("expected non-nil segment")
	}
	if seg1.ID != sgtypes.SegmentID(42) {
		t.Errorf("expected segment ID 42, got %s", seg1.ID)
	}

	// Idempotent: same segment returned
	seg2 := mgr.getOrCreateSegment(sgtypes.SegmentID(42))
	if seg1 != seg2 {
		t.Error("expected same segment pointer on second call")
	}

	// Different ID: different segment
	seg3 := mgr.getOrCreateSegment(sgtypes.SegmentID(99))
	if seg3 == seg1 {
		t.Error("expected different segment for different ID")
	}
}

func TestAppendUniqueRoute(t *testing.T) {
	routes := []ir.TransitRoute{
		{CIDR: mustParsePrefix("10.0.0.0/16"), NextHop: "gw-1"},
	}

	// Same route should be deduped
	routes = appendUniqueRoute(routes, ir.TransitRoute{
		CIDR: mustParsePrefix("10.0.0.0/16"), NextHop: "gw-1",
	})
	if len(routes) != 1 {
		t.Errorf("expected dedup, got %d routes", len(routes))
	}

	// Different route should be appended
	routes = appendUniqueRoute(routes, ir.TransitRoute{
		CIDR: mustParsePrefix("10.1.0.0/16"), NextHop: "gw-2",
	})
	if len(routes) != 2 {
		t.Errorf("expected 2 routes, got %d", len(routes))
	}

	// Same CIDR, different nexthop: should be appended
	routes = appendUniqueRoute(routes, ir.TransitRoute{
		CIDR: mustParsePrefix("10.0.0.0/16"), NextHop: "gw-3",
	})
	if len(routes) != 3 {
		t.Errorf("expected 3 routes (same CIDR, diff nexthop), got %d", len(routes))
	}
}

func TestAppendUniqueCID(t *testing.T) {
	ids := []sgtypes.ClusterID{"cluster-a"}

	// Duplicate should be deduped
	ids = appendUniqueCID(ids, "cluster-a")
	if len(ids) != 1 {
		t.Errorf("expected dedup, got %d", len(ids))
	}

	// Unique should be appended
	ids = appendUniqueCID(ids, "cluster-b")
	if len(ids) != 2 {
		t.Errorf("expected 2, got %d", len(ids))
	}
}

func TestTopologyTypes(t *testing.T) {
	hs := TopologyHubSpoke{
		HubCluster:   "hub",
		SpokeCluster: []sgtypes.ClusterID{"spoke-1", "spoke-2"},
		SegmentID:    sgtypes.SegmentID(10),
	}
	if hs.HubCluster != "hub" {
		t.Errorf("unexpected hub: %s", hs.HubCluster)
	}
	if len(hs.SpokeCluster) != 2 {
		t.Errorf("expected 2 spokes, got %d", len(hs.SpokeCluster))
	}

	mesh := TopologyMesh{
		Clusters:  []sgtypes.ClusterID{"a", "b", "c"},
		SegmentID: sgtypes.SegmentID(20),
	}
	if len(mesh.Clusters) != 3 {
		t.Errorf("expected 3 mesh clusters, got %d", len(mesh.Clusters))
	}

	p2p := TopologyPeerToPeer{
		ClusterA:  "alpha",
		ClusterB:  "beta",
		SegmentID: sgtypes.SegmentID(30),
	}
	if p2p.ClusterA != "alpha" || p2p.ClusterB != "beta" {
		t.Errorf("unexpected p2p clusters: %s, %s", p2p.ClusterA, p2p.ClusterB)
	}
}

func TestReconcileMultipleGatewaysSameSegment(t *testing.T) {
	tg1 := makeTransitGateway("gw-same-1", 500, "mesh", []apiv1.TransitPeer{
		{ClusterName: "c1", Endpoint: "1.1.1.1", PodCIDRs: []string{"10.0.0.0/16"}},
	})
	tg2 := makeTransitGateway("gw-same-2", 500, "mesh", []apiv1.TransitPeer{
		{ClusterName: "c2", Endpoint: "2.2.2.2", PodCIDRs: []string{"10.1.0.0/16"}},
	})

	mgr := newTestManager(tg1, tg2)
	ctx := context.Background()

	_ = mgr.ReconcileGateway(ctx, types.NamespacedName{Name: "gw-same-1"})
	_ = mgr.ReconcileGateway(ctx, types.NamespacedName{Name: "gw-same-2"})

	// Both gateways share segment 500, routes should be accumulated
	seg := mgr.segments[sgtypes.SegmentID(500)]
	if seg == nil {
		t.Fatal("expected segment 500")
	}
	if len(seg.Routes) != 2 {
		t.Errorf("expected 2 routes on shared segment, got %d", len(seg.Routes))
	}
	if len(seg.Clusters) != 2 {
		t.Errorf("expected 2 clusters on shared segment, got %d", len(seg.Clusters))
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkReconcileGateway(b *testing.B) {
	for _, peerCount := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("peers=%d", peerCount), func(b *testing.B) {
			peers := make([]apiv1.TransitPeer, peerCount)
			for i := range peers {
				peers[i] = apiv1.TransitPeer{
					ClusterName: fmt.Sprintf("cluster-%d", i),
					Endpoint:    fmt.Sprintf("10.%d.%d.1", i/256, i%256),
					PodCIDRs:    []string{fmt.Sprintf("10.%d.0.0/16", i%256)},
				}
			}
			tg := makeTransitGateway("bench-gw", 100, "mesh", peers)
			mgr := newTestManager(tg)
			ctx := context.Background()
			key := types.NamespacedName{Name: "bench-gw"}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Reset segment state for each iteration
				mgr.segments = make(map[sgtypes.SegmentID]*ir.TransitSegment)
				mgr.gateways = make(map[types.NamespacedName]*ir.TransitSegment)
				_ = mgr.ReconcileGateway(ctx, key)
			}
		})
	}
}

func BenchmarkGetAllSegments(b *testing.B) {
	for _, count := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("segments=%d", count), func(b *testing.B) {
			mgr := newTestManager()
			for i := 0; i < count; i++ {
				mgr.segments[sgtypes.SegmentID(i)] = &ir.TransitSegment{
					ID: sgtypes.SegmentID(i),
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = mgr.GetAllSegments()
			}
		})
	}
}

func BenchmarkAppendUniqueRoute(b *testing.B) {
	base := make([]ir.TransitRoute, 100)
	for i := range base {
		base[i] = ir.TransitRoute{
			CIDR:    mustParsePrefix(fmt.Sprintf("10.%d.0.0/16", i)),
			NextHop: fmt.Sprintf("gw-%d", i),
		}
	}
	newRoute := ir.TransitRoute{
		CIDR:    mustParsePrefix("192.168.0.0/16"),
		NextHop: "gw-new",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		routes := make([]ir.TransitRoute, len(base))
		copy(routes, base)
		_ = appendUniqueRoute(routes, newRoute)
	}
}

func BenchmarkAppendUniqueCID(b *testing.B) {
	base := make([]sgtypes.ClusterID, 100)
	for i := range base {
		base[i] = sgtypes.ClusterID(fmt.Sprintf("cluster-%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ids := make([]sgtypes.ClusterID, len(base))
		copy(ids, base)
		_ = appendUniqueCID(ids, "cluster-new")
	}
}

// ============================================================================
// Helpers
// ============================================================================

func mustParsePrefix(s string) netip.Prefix {
	p, err := netip.ParsePrefix(s)
	if err != nil {
		panic(fmt.Sprintf("bad prefix %q: %v", s, err))
	}
	return p
}
