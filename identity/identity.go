// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package identity implements numeric security identity allocation,
// multi-dimensional label hashing, and caching for the straitKubegateway
// eBPF policy engine.
package identity

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	sgtypes "github.com/straitkubegateway/straitkubegateway/pkg/types"
)

// Identity is the numeric BPF security identity.
type Identity = sgtypes.Identity

const (
	IdentityUnknown    = sgtypes.IdentityUnknown
	IdentityWorld      = sgtypes.IdentityWorld
	IdentityHost       = sgtypes.IdentityHost
	IdentityInit       = sgtypes.IdentityInit
	IdentityRemoteNode = sgtypes.IdentityRemoteNode
	IdentityMin        = sgtypes.IdentityMin
	IdentityMax        = sgtypes.IdentityMax
)

// ============================================================================
// Identity Allocator
// ============================================================================

// Allocator manages thread-safe allocation, caching, and release of numeric identities.
type Allocator struct {
	mu       sync.Mutex
	next     atomic.Uint32
	free     []Identity
	used     map[Identity]struct{}
	byLabels map[string]Identity
	byID     map[Identity]string
}

// NewAllocator returns a new identity allocator starting at IdentityMin.
func NewAllocator() *Allocator {
	a := &Allocator{
		used:     make(map[Identity]struct{}),
		byLabels: make(map[string]Identity),
		byID:     make(map[Identity]string),
	}
	a.next.Store(uint32(IdentityMin))
	return a
}

// Allocate allocates a new numeric identity for the given multi-dimensional label set.
// If an identity already exists for these labels, it is returned from cache.
func (a *Allocator) Allocate(_ context.Context, labels string) (Identity, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if id, ok := a.byLabels[labels]; ok {
		return id, nil
	}

	id, err := a.allocateLocked()
	if err != nil {
		return 0, err
	}
	a.byLabels[labels] = id
	a.byID[id] = labels
	return id, nil
}

// Release releases the identity previously allocated for labels.
func (a *Allocator) Release(_ context.Context, labels string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	id, ok := a.byLabels[labels]
	if !ok {
		return
	}
	delete(a.byLabels, labels)
	delete(a.byID, id)
	delete(a.used, id)
	a.free = append(a.free, id)
}

// Get returns the identity for a given label string, or IdentityUnknown.
func (a *Allocator) Get(labels string) Identity {
	a.mu.Lock()
	defer a.mu.Unlock()
	if id, ok := a.byLabels[labels]; ok {
		return id
	}
	return IdentityUnknown
}

// LookupLabels returns the label string associated with a given identity.
func (a *Allocator) LookupLabels(id Identity) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.byID[id]
}

// Count returns the number of currently allocated identities.
func (a *Allocator) Count() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.used)
}

func (a *Allocator) allocateLocked() (Identity, error) {
	if len(a.free) > 0 {
		id := a.free[0]
		a.free = a.free[1:]
		a.used[id] = struct{}{}
		return id, nil
	}
	raw := a.next.Add(1) - 1
	id := Identity(raw)
	if id > IdentityMax {
		return 0, fmt.Errorf("identity space exhausted (max %d)", IdentityMax)
	}
	a.used[id] = struct{}{}
	return id, nil
}

// ============================================================================
// Multi-Dimensional Label Dimension Hashing
// ============================================================================

// Dimensions represents all selector dimensions for a pod or endpoint.
type Dimensions struct {
	Namespace string
	PodLabels map[string]string
	Cluster   string
	Segment   string
	Gateway   string
	HTTPRoute string
	TCPRoute  string
	UDPRoute  string
	GRPCRoute string
	TLSRoute  string
}

// BuildIdentityKey serializes multi-dimensional labels deterministically.
func BuildIdentityKey(d Dimensions) string {
	var parts []string

	if d.Namespace != "" {
		parts = append(parts, "k8s:io.kubernetes.pod.namespace="+d.Namespace)
	}
	if d.Cluster != "" {
		parts = append(parts, "cluster="+d.Cluster)
	}
	if d.Segment != "" {
		parts = append(parts, "segment="+d.Segment)
	}
	if d.Gateway != "" {
		parts = append(parts, "gw="+d.Gateway)
	}
	if d.HTTPRoute != "" {
		parts = append(parts, "httproute="+d.HTTPRoute)
	}
	if d.TCPRoute != "" {
		parts = append(parts, "tcproute="+d.TCPRoute)
	}
	if d.UDPRoute != "" {
		parts = append(parts, "udproute="+d.UDPRoute)
	}
	if d.GRPCRoute != "" {
		parts = append(parts, "grpcroute="+d.GRPCRoute)
	}
	if d.TLSRoute != "" {
		parts = append(parts, "tlsroute="+d.TLSRoute)
	}

	// Sort pod labels for deterministic output
	var labelKeys []string
	for k := range d.PodLabels {
		labelKeys = append(labelKeys, k)
	}
	sort.Strings(labelKeys)
	for _, k := range labelKeys {
		parts = append(parts, "k8s:"+k+"="+d.PodLabels[k])
	}

	return strings.Join(parts, ";")
}
