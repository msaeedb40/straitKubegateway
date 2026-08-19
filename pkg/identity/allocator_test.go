package identity_test

import (
	"testing"

	"github.com/straitKubegateway/straitKubegateway/pkg/identity"
)

func TestLocalAllocator(t *testing.T) {
	alloc := identity.NewLocalAllocator()

	labelsA := map[string]string{"app": "frontend", "env": "prod"}
	labelsB := map[string]string{"app": "backend", "env": "prod"}

	idA1, err := alloc.Allocate(labelsA)
	if err != nil {
		t.Fatalf("unexpected error allocating idA: %v", err)
	}

	// Id should be above reserved range
	if idA1.IsReserved() {
		t.Errorf("allocated id %d is in reserved range", idA1)
	}

	// Duplicate allocation should reuse id
	idA2, err := alloc.Allocate(labelsA)
	if err != nil {
		t.Fatalf("unexpected error allocating duplicate idA: %v", err)
	}
	if idA1 != idA2 {
		t.Errorf("expected duplicate labels to get same ID, got %d vs %d", idA1, idA2)
	}

	// Different labels get distinct id
	idB, err := alloc.Allocate(labelsB)
	if err != nil {
		t.Fatalf("unexpected error allocating idB: %v", err)
	}
	if idA1 == idB {
		t.Errorf("expected different labels to get different IDs, got %d", idB)
	}

	// Lookup
	foundID, ok := alloc.Lookup(labelsA)
	if !ok || foundID != idA1 {
		t.Errorf("expected to lookup %d, got %d (ok=%v)", idA1, foundID, ok)
	}

	// Release ref 1
	if err := alloc.Release(idA1); err != nil {
		t.Fatalf("unexpected error releasing idA1: %v", err)
	}
	// Still referenced once, lookup should succeed
	if _, ok := alloc.Lookup(labelsA); !ok {
		t.Error("expected labelsA to still be allocated with 1 ref")
	}

	// Release ref 2
	if err := alloc.Release(idA2); err != nil {
		t.Fatalf("unexpected error releasing idA2: %v", err)
	}
	// Now freed
	if _, ok := alloc.Lookup(labelsA); ok {
		t.Error("expected labelsA to be released")
	}
}
