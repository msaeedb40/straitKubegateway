// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"testing"
)

func TestIdentityAllocationAndDimensions(t *testing.T) {
	ctx := context.Background()
	alloc := NewAllocator()

	dim1 := Dimensions{
		Namespace: "production",
		PodLabels: map[string]string{"app": "checkout", "version": "v2"},
		Cluster:   "cluster-primary",
		Segment:   "10",
		Gateway:   "main-gw",
		HTTPRoute: "checkout-route",
	}

	key1 := BuildIdentityKey(dim1)
	id1, err := alloc.Allocate(ctx, key1)
	if err != nil {
		t.Fatalf("allocate failed: %v", err)
	}

	if id1 < IdentityMin {
		t.Fatalf("expected id >= %d, got %d", IdentityMin, id1)
	}

	// Idempotent allocation should return same ID
	id1Again, err := alloc.Allocate(ctx, key1)
	if err != nil {
		t.Fatalf("re-allocate failed: %v", err)
	}
	if id1Again != id1 {
		t.Fatalf("expected identical id %d, got %d", id1, id1Again)
	}

	// Lookup by ID
	recoveredLabels := alloc.LookupLabels(id1)
	if recoveredLabels != key1 {
		t.Fatalf("expected %s, got %s", key1, recoveredLabels)
	}

	// Allocate different dimension
	dim2 := Dimensions{
		Namespace: "production",
		PodLabels: map[string]string{"app": "payment"},
		Cluster:   "cluster-primary",
		Segment:   "10",
	}
	key2 := BuildIdentityKey(dim2)
	id2, err := alloc.Allocate(ctx, key2)
	if err != nil {
		t.Fatalf("allocate id2 failed: %v", err)
	}
	if id2 == id1 {
		t.Fatalf("expected different IDs for different dimensions, got both %d", id1)
	}

	if alloc.Count() != 2 {
		t.Fatalf("expected 2 allocated, got %d", alloc.Count())
	}

	// Release id1
	alloc.Release(ctx, key1)
	if alloc.Count() != 1 {
		t.Fatalf("expected 1 allocated after release, got %d", alloc.Count())
	}
}
