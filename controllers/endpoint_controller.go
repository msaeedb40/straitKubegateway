package controllers

import (
	"context"
	"fmt"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// EndpointReconciler reconciles endpoint registrations across the cluster.
type EndpointReconciler struct {
	mu        sync.RWMutex
	endpoints map[uint64]*types.Endpoint
	logger    *logging.Logger
}

// NewEndpointReconciler creates a new Endpoint controller.
func NewEndpointReconciler() *EndpointReconciler {
	return &EndpointReconciler{
		endpoints: make(map[uint64]*types.Endpoint),
		logger:    logging.DefaultLogger(),
	}
}

// RegisterEndpoint adds or updates an endpoint record.
func (r *EndpointReconciler) RegisterEndpoint(ctx context.Context, ep *types.Endpoint) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.endpoints[ep.ID] = ep
	r.logger.Debug(fmt.Sprintf("registered endpoint id=%d pod=%s/%s ip=%s", ep.ID, ep.Namespace, ep.PodName, ep.IPv4), &types.Metadata{
		Component: "endpoint-controller",
		Namespace: ep.Namespace,
		PodName:   ep.PodName,
	})
}

// UnregisterEndpoint removes an endpoint record.
func (r *EndpointReconciler) UnregisterEndpoint(ctx context.Context, id uint64) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.endpoints, id)
}
