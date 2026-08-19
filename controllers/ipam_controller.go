// Package controllers provides Kubernetes reconcilers for straitKubegateway CRDs.
package controllers

import (
	"context"
	"fmt"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/api/v1alpha1"
	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// IPAMReconciler manages IPAMPool allocations.
type IPAMReconciler struct {
	mu     sync.Mutex
	pools  map[string]*v1alpha1.IPAMPool
	logger *logging.Logger
}

// NewIPAMReconciler creates a new IPAM controller.
func NewIPAMReconciler() *IPAMReconciler {
	return &IPAMReconciler{
		pools:  make(map[string]*v1alpha1.IPAMPool),
		logger: logging.DefaultLogger(),
	}
}

// Reconcile handles reconciliation of an IPAMPool resource.
func (r *IPAMReconciler) Reconcile(ctx context.Context, pool *v1alpha1.IPAMPool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.pools[pool.Name] = pool
	r.logger.Info(fmt.Sprintf("reconciled IPAMPool %q with %d CIDRs", pool.Name, len(pool.Spec.CIDRs)), &types.Metadata{
		Component: "ipam-controller",
	})
	return nil
}
