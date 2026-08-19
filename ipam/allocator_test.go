package ipam_test

import (
	"testing"

	"github.com/straitKubegateway/straitKubegateway/ipam"
)

func TestIPAMAllocator(t *testing.T) {
	alloc, err := ipam.NewAllocator("10.244.1.0/24")
	if err != nil {
		t.Fatalf("unexpected error creating allocator: %v", err)
	}

	gw := alloc.Gateway()
	if gw.String() != "10.244.1.1" {
		t.Errorf("expected gateway 10.244.1.1, got %s", gw)
	}

	ip1, err := alloc.Allocate("container-1")
	if err != nil {
		t.Fatalf("allocate failed: %v", err)
	}
	if ip1.String() != "10.244.1.2" {
		t.Errorf("expected 10.244.1.2, got %s", ip1)
	}

	ip2, err := alloc.Allocate("container-2")
	if err != nil {
		t.Fatalf("allocate failed: %v", err)
	}
	if ip2.String() != "10.244.1.3" {
		t.Errorf("expected 10.244.1.3, got %s", ip2)
	}

	if alloc.AllocatedCount() != 2 {
		t.Errorf("expected count 2, got %d", alloc.AllocatedCount())
	}

	releasedIP, err := alloc.Release("container-1")
	if err != nil {
		t.Fatalf("release failed: %v", err)
	}
	if releasedIP != ip1 {
		t.Errorf("expected released IP %s, got %s", ip1, releasedIP)
	}

	if alloc.AllocatedCount() != 1 {
		t.Errorf("expected count 1, got %d", alloc.AllocatedCount())
	}
}
