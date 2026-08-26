// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package identity implements numeric identity allocation and caching
// for the straitKubegateway eBPF policy engine.
package identity

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	sgtypes "github.com/straitkubegateway/straitkubegateway/pkg/types"
)

// ============================================================================
// Allocator
// ============================================================================

// Allocator manages the allocation and release of numeric network identities.
// Thread-safe; safe for concurrent use from multiple goroutines.
type Allocator struct {
	mu       sync.Mutex
	next     atomic.Uint32
	free     []sgtypes.Identity
	used     map[sgtypes.Identity]struct{}
	byLabels map[string]sgtypes.Identity
}

// NewAllocator returns a new identity allocator starting at IdentityMin.
func NewAllocator() *Allocator {
	a := &Allocator{
		used:     make(map[sgtypes.Identity]struct{}),
		byLabels: make(map[string]sgtypes.Identity),
	}
	a.next.Store(uint32(sgtypes.IdentityMin))
	return a
}

// Allocate allocates a new identity for the given label set.
// If an identity already exists for these labels, it is returned.
func (a *Allocator) Allocate(_ context.Context, labels string) (sgtypes.Identity, error) {
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
	delete(a.used, id)
	a.free = append(a.free, id)
}

// Get returns the identity for a given label set, or IdentityUnknown.
func (a *Allocator) Get(labels string) sgtypes.Identity {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.byLabels[labels]
}

// Count returns the number of currently allocated user identities.
func (a *Allocator) Count() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.used)
}

func (a *Allocator) allocateLocked() (sgtypes.Identity, error) {
	// Reuse freed identities first
	if len(a.free) > 0 {
		id := a.free[0]
		a.free = a.free[1:]
		a.used[id] = struct{}{}
		return id, nil
	}
	// Allocate from the counter
	raw := a.next.Add(1) - 1
	id := sgtypes.Identity(raw)
	if id > sgtypes.IdentityMax {
		return 0, fmt.Errorf("identity space exhausted (max %d)", sgtypes.IdentityMax)
	}
	a.used[id] = struct{}{}
	return id, nil
}

// ============================================================================
// Label helpers
// ============================================================================

// Labels represents a sorted, stable key→value label set used for
// identity calculation.
type Labels map[string]string

// Key returns a stable canonical string key for the label set.
// The key is used as the identity cache key.
func (l Labels) Key() string {
	if len(l) == 0 {
		return ""
	}
	// Sort keys for determinism
	keys := make([]string, 0, len(l))
	for k := range l {
		keys = append(keys, k)
	}
	sortStrings(keys)
	var b []byte
	for i, k := range keys {
		if i > 0 {
			b = append(b, ',')
		}
		b = append(b, k...)
		b = append(b, '=')
		b = append(b, l[k]...)
	}
	return string(b)
}

// sortStrings sorts a string slice in place (insertion sort — good for small slices).
func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		key := s[i]
		j := i - 1
		for j >= 0 && s[j] > key {
			s[j+1] = s[j]
			j--
		}
		s[j+1] = key
	}
}

// ============================================================================
// Store
// ============================================================================

// Store persists identity→labels mappings for use across restarts.
type Store struct {
	mu      sync.RWMutex
	byID    map[sgtypes.Identity]Labels
	byLabel map[string]sgtypes.Identity
}

// NewStore creates a new identity store.
func NewStore() *Store {
	return &Store{
		byID:    make(map[sgtypes.Identity]Labels),
		byLabel: make(map[string]sgtypes.Identity),
	}
}

// Set stores an identity→labels mapping.
func (s *Store) Set(id sgtypes.Identity, labels Labels) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.byID[id] = labels
	s.byLabel[labels.Key()] = id
}

// GetByID returns the labels for a given identity.
func (s *Store) GetByID(id sgtypes.Identity) (Labels, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	l, ok := s.byID[id]
	return l, ok
}

// GetByLabels returns the identity for a given label set.
func (s *Store) GetByLabels(labels Labels) (sgtypes.Identity, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byLabel[labels.Key()]
	return id, ok
}

// Delete removes an identity from the store.
func (s *Store) Delete(id sgtypes.Identity) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if labels, ok := s.byID[id]; ok {
		delete(s.byLabel, labels.Key())
	}
	delete(s.byID, id)
}

// Count returns the number of stored identities.
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.byID)
}
