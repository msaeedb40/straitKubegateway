// Package identity provides an allocator for BPF network identities.
// Each endpoint (pod) receives a unique identity used in the eBPF
// dataplane for policy decisions and routing.
package identity

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	sgtypes "github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// Allocator manages the allocation and release of BPF identities.
type Allocator interface {
	// Allocate assigns a new identity for the given labels.
	// If an identity already exists for the exact label set, it is reused.
	Allocate(labels map[string]string) (sgtypes.Identity, error)

	// Release decrements the reference count for the identity and frees
	// it when no endpoints reference it.
	Release(id sgtypes.Identity) error

	// Lookup returns the identity for the given labels, if one exists.
	Lookup(labels map[string]string) (sgtypes.Identity, bool)

	// LookupByID returns the labels associated with the given identity.
	LookupByID(id sgtypes.Identity) (map[string]string, bool)
}

// localAllocator is an in-memory identity allocator.
// In production this will be backed by a CRD-based distributed allocator.
type localAllocator struct {
	mu sync.RWMutex

	// next is the next identity to allocate.
	next sgtypes.Identity

	// identities maps identity -> labels.
	identities map[sgtypes.Identity]map[string]string

	// labelIndex maps label hash -> identity for deduplication.
	labelIndex map[string]sgtypes.Identity

	// refCounts tracks how many endpoints reference each identity.
	refCounts map[sgtypes.Identity]int
}

// NewLocalAllocator creates a new in-memory identity allocator.
// User identities start above IdentityReservedMax.
func NewLocalAllocator() Allocator {
	return &localAllocator{
		next:       sgtypes.IdentityReservedMax + 1,
		identities: make(map[sgtypes.Identity]map[string]string),
		labelIndex: make(map[string]sgtypes.Identity),
		refCounts:  make(map[sgtypes.Identity]int),
	}
}

func (a *localAllocator) Allocate(labels map[string]string) (sgtypes.Identity, error) {
	key := hashLabels(labels)

	a.mu.Lock()
	defer a.mu.Unlock()

	// Reuse existing identity for the same label set.
	if id, ok := a.labelIndex[key]; ok {
		a.refCounts[id]++
		return id, nil
	}

	// Allocate a new identity.
	id := a.next
	a.next++

	// Copy labels for immutability.
	stored := make(map[string]string, len(labels))
	for k, v := range labels {
		stored[k] = v
	}

	a.identities[id] = stored
	a.labelIndex[key] = id
	a.refCounts[id] = 1

	return id, nil
}

func (a *localAllocator) Release(id sgtypes.Identity) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	count, ok := a.refCounts[id]
	if !ok {
		return fmt.Errorf("identity %d not found", id)
	}

	count--
	if count <= 0 {
		labels := a.identities[id]
		key := hashLabels(labels)
		delete(a.identities, id)
		delete(a.labelIndex, key)
		delete(a.refCounts, id)
	} else {
		a.refCounts[id] = count
	}

	return nil
}

func (a *localAllocator) Lookup(labels map[string]string) (sgtypes.Identity, bool) {
	key := hashLabels(labels)

	a.mu.RLock()
	defer a.mu.RUnlock()

	id, ok := a.labelIndex[key]
	return id, ok
}

func (a *localAllocator) LookupByID(id sgtypes.Identity) (map[string]string, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	labels, ok := a.identities[id]
	if !ok {
		return nil, false
	}

	// Return a copy.
	result := make(map[string]string, len(labels))
	for k, v := range labels {
		result[k] = v
	}
	return result, true
}

// hashLabels creates a deterministic string key from a label map.
func hashLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString("=")
		sb.WriteString(labels[k])
		sb.WriteString(";")
	}
	return sb.String()
}
