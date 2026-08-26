// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"testing"

	sgtypes "github.com/straitkubegateway/straitkubegateway/pkg/types"
)

func TestAllocator(t *testing.T) {
	alloc := NewAllocator()
	ctx := context.Background()

	// Allocate identity for pod-a
	id1, err := alloc.Allocate(ctx, "app=nginx,env=prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id1 < sgtypes.IdentityMin {
		t.Errorf("expected identity >= %d, got %d", sgtypes.IdentityMin, id1)
	}

	// Idempotent allocation for identical labels
	id2, err := alloc.Allocate(ctx, "app=nginx,env=prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id1 != id2 {
		t.Errorf("expected identical identity for same labels, got %d and %d", id1, id2)
	}

	// Allocate different identity for different labels
	id3, err := alloc.Allocate(ctx, "app=redis,env=prod")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id3 == id1 {
		t.Errorf("expected distinct identities, both got %d", id3)
	}

	// Release and reuse
	alloc.Release(ctx, "app=nginx,env=prod")
	if alloc.Get("app=nginx,env=prod") != sgtypes.IdentityUnknown {
		t.Errorf("expected identity to be released")
	}

	id4, err := alloc.Allocate(ctx, "app=frontend,env=staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id4 != id1 {
		t.Errorf("expected freed identity %d to be reused, got %d", id1, id4)
	}
}

func TestLabelsKey(t *testing.T) {
	l1 := Labels{"env": "prod", "app": "nginx"}
	l2 := Labels{"app": "nginx", "env": "prod"}

	if l1.Key() != l2.Key() {
		t.Errorf("expected deterministic label keys, got %s and %s", l1.Key(), l2.Key())
	}
	if l1.Key() != "app=nginx,env=prod" {
		t.Errorf("got %s, want app=nginx,env=prod", l1.Key())
	}
}

func TestStore(t *testing.T) {
	store := NewStore()
	labels := Labels{"app": "api"}
	id := sgtypes.Identity(300)

	store.Set(id, labels)

	if got, ok := store.GetByID(id); !ok || got.Key() != labels.Key() {
		t.Errorf("expected labels %v, got %v", labels, got)
	}
	if gotID, ok := store.GetByLabels(labels); !ok || gotID != id {
		t.Errorf("expected ID %d, got %d", id, gotID)
	}
	if store.Count() != 1 {
		t.Errorf("expected 1 stored item, got %d", store.Count())
	}

	store.Delete(id)
	if store.Count() != 0 {
		t.Errorf("expected 0 stored items after delete, got %d", store.Count())
	}
}
