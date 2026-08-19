package controllers

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
	"github.com/straitKubegateway/straitKubegateway/service"
)

// EndpointSliceDescriptor represents a Kubernetes EndpointSlice resource.
type EndpointSliceDescriptor struct {
	Namespace   string
	ServiceName string
	Endpoints   []EndpointDescriptor
	Port        uint16
	Protocol    types.Protocol
}

// EndpointDescriptor represents a single backend endpoint within an EndpointSlice.
type EndpointDescriptor struct {
	IP      string
	Port    uint16
	Ready   bool
	Serving bool
	Weight  uint32
}

// EndpointSliceController reconciles Kubernetes EndpointSlices and maintains Service backend pools.
type EndpointSliceController struct {
	mu        sync.RWMutex
	lbManager *service.Manager
	logger    *logging.Logger
}

// NewEndpointSliceController creates a new EndpointSlice controller.
func NewEndpointSliceController(lbMgr *service.Manager) *EndpointSliceController {
	return &EndpointSliceController{
		lbManager: lbMgr,
		logger:    logging.DefaultLogger(),
	}
}

// ReconcileEndpointSlice updates the backends associated with a Service.
func (c *EndpointSliceController) ReconcileEndpointSlice(ctx context.Context, slice EndpointSliceDescriptor) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	svc, exists := c.lbManager.GetService(slice.Namespace, slice.ServiceName, slice.Port, slice.Protocol)
	if !exists {
		return fmt.Errorf("service %s/%s:%d/%s not found", slice.Namespace, slice.ServiceName, slice.Port, slice.Protocol)
	}

	backends := make([]service.BackendEndpoint, 0, len(slice.Endpoints))
	for _, ep := range slice.Endpoints {
		if !ep.Ready {
			continue
		}
		ip, err := netip.ParseAddr(ep.IP)
		if err != nil {
			continue
		}

		weight := ep.Weight
		if weight == 0 {
			weight = 100
		}

		port := ep.Port
		if port == 0 {
			port = slice.Port
		}

		backends = append(backends, service.BackendEndpoint{
			IP:     ip,
			Port:   port,
			Weight: weight,
		})
	}

	// Re-upsert service with new backends
	c.lbManager.UpsertService(
		svc.Namespace,
		svc.Name,
		svc.VIP,
		svc.Port,
		svc.Protocol,
		svc.Algorithm,
		svc.SessionAffinity,
		svc.AffinityTimeout,
		backends,
	)

	c.logger.Debug(fmt.Sprintf("reconciled EndpointSlice for %s/%s: %d active backends",
		slice.Namespace, slice.ServiceName, len(backends)), &types.Metadata{
		Component:   "endpointslice-controller",
		Namespace:   slice.Namespace,
		ServiceName: slice.ServiceName,
	})

	return nil
}
