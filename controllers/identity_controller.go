package controllers

import (
	"context"
	"fmt"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/identity"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// IdentityDescriptor describes a workload label set for identity allocation.
type IdentityDescriptor struct {
	Namespace string
	PodName   string
	Labels    map[string]string
}

// IdentityController reconciles workload labels into deterministic 32-bit BPF identities.
type IdentityController struct {
	mu        sync.RWMutex
	allocator identity.Allocator
	logger    *logging.Logger
	podIdents map[string]types.Identity // "namespace/podName" -> Identity
}

// NewIdentityController creates a new Identity reconciler.
func NewIdentityController(alloc identity.Allocator) *IdentityController {
	if alloc == nil {
		alloc = identity.NewLocalAllocator()
	}
	return &IdentityController{
		allocator: alloc,
		logger:    logging.DefaultLogger(),
		podIdents: make(map[string]types.Identity),
	}
}

// ReconcilePodIdentity assigns or resolves a 32-bit security identity for a pod workload.
func (c *IdentityController) ReconcilePodIdentity(ctx context.Context, desc IdentityDescriptor) (types.Identity, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s/%s", desc.Namespace, desc.PodName)

	id, err := c.allocator.Allocate(desc.Labels)
	if err != nil {
		return 0, fmt.Errorf("allocate identity for %s: %w", key, err)
	}

	c.podIdents[key] = id

	c.logger.Info(fmt.Sprintf("assigned Identity %d to pod %s", id, key), &types.Metadata{
		Component: "identity-controller",
		Namespace: desc.Namespace,
		PodName:   desc.PodName,
	})

	return id, nil
}

// ReleasePodIdentity releases the allocated identity for a terminated pod.
func (c *IdentityController) ReleasePodIdentity(ctx context.Context, ns, podName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := fmt.Sprintf("%s/%s", ns, podName)
	id, exists := c.podIdents[key]
	if !exists {
		return nil
	}

	delete(c.podIdents, key)
	return c.allocator.Release(id)
}
