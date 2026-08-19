package service_test

import (
	"net/netip"
	"testing"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
	"github.com/straitKubegateway/straitKubegateway/service"
)

func TestMaglevLUTGeneration(t *testing.T) {
	backends := []*service.Backend{
		{ID: 1, IP: netip.MustParseAddr("10.244.1.2"), Port: 8080, Weight: 100, Healthy: true},
		{ID: 2, IP: netip.MustParseAddr("10.244.1.3"), Port: 8080, Weight: 100, Healthy: true},
		{ID: 3, IP: netip.MustParseAddr("10.244.1.4"), Port: 8080, Weight: 100, Healthy: true},
	}

	lut := service.GenerateMaglevLUT(101, backends)
	if lut.ServiceID != 101 {
		t.Errorf("expected ServiceID 101, got %d", lut.ServiceID)
	}

	// Verify all 128 slots are populated with valid backend IDs (1, 2, or 3)
	counts := make(map[uint32]int)
	for i, id := range lut.Lookup {
		if id < 1 || id > 3 {
			t.Errorf("slot %d has invalid backend ID %d", i, id)
		}
		counts[id]++
	}

	// With 3 backends and 128 slots, distribution should be relatively balanced (~42 each)
	for id, count := range counts {
		if count < 30 || count > 55 {
			t.Errorf("backend %d has unbalanced slot count: %d", id, count)
		}
	}
}

func TestLoadBalancingAlgorithms(t *testing.T) {
	mgr := service.NewManager()

	endpoints := []service.BackendEndpoint{
		{IP: netip.MustParseAddr("10.244.1.10"), Port: 8080, Weight: 100},
		{IP: netip.MustParseAddr("10.244.1.11"), Port: 8080, Weight: 100},
	}

	// 1. Test Round Robin
	svcRR := mgr.UpsertService(
		"default", "web-svc-rr",
		netip.MustParseAddr("10.96.0.10"), 80, types.ProtocolTCP,
		service.AlgorithmRoundRobin, false, 0,
		endpoints,
	)

	b1, err := svcRR.SelectBackend(netip.MustParseAddr("10.244.0.5"), 12345)
	if err != nil {
		t.Fatalf("SelectBackend failed: %v", err)
	}
	b2, err := svcRR.SelectBackend(netip.MustParseAddr("10.244.0.5"), 12346)
	if err != nil {
		t.Fatalf("SelectBackend failed: %v", err)
	}

	if b1.ID == b2.ID {
		t.Errorf("Round Robin expected alternating backends, got %d and %d", b1.ID, b2.ID)
	}

	// 2. Test Maglev Hash Consistency
	svcMaglev := mgr.UpsertService(
		"default", "web-svc-maglev",
		netip.MustParseAddr("10.96.0.20"), 80, types.ProtocolTCP,
		service.AlgorithmMaglevHash, false, 0,
		endpoints,
	)

	clientIP := netip.MustParseAddr("10.244.2.15")
	clientPort := uint16(54321)

	bMaglev1, err := svcMaglev.SelectBackend(clientIP, clientPort)
	if err != nil {
		t.Fatalf("Maglev SelectBackend failed: %v", err)
	}
	bMaglev2, err := svcMaglev.SelectBackend(clientIP, clientPort)
	if err != nil {
		t.Fatalf("Maglev SelectBackend failed: %v", err)
	}

	// Same 5-tuple must consistently map to the same backend
	if bMaglev1.ID != bMaglev2.ID {
		t.Errorf("Maglev Hash expected consistent backend, got %d vs %d", bMaglev1.ID, bMaglev2.ID)
	}
}
