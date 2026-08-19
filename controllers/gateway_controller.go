package controllers

import (
	"context"
	"fmt"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/api/v1alpha1"
	"github.com/straitKubegateway/straitKubegateway/gateway"
	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// GatewayController reconciles Gateway resources and manages active GatewayInstances.
type GatewayController struct {
	mu         sync.RWMutex
	gatewayMgr *gateway.Manager
	logger     *logging.Logger
}

// NewGatewayController creates a new Gateway controller.
func NewGatewayController(gwMgr *gateway.Manager) *GatewayController {
	return &GatewayController{
		gatewayMgr: gwMgr,
		logger:     logging.DefaultLogger(),
	}
}

// ReconcileGateway processes a Gateway resource and updates listener bindings.
func (c *GatewayController) ReconcileGateway(ctx context.Context, gw *v1alpha1.Gateway) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	listeners := make([]gateway.Listener, 0, len(gw.Spec.Listeners))
	for _, l := range gw.Spec.Listeners {
		listeners = append(listeners, gateway.Listener{
			Name:     l.Name,
			Port:     uint16(l.Port),
			Protocol: l.Protocol,
		})
	}

	instance := gateway.GatewayInstance{
		ID:        string(gw.UID),
		Namespace: gw.Namespace,
		Name:      gw.Name,
		SegmentID: types.SegmentID(gw.Spec.SegmentID),
		Listeners: listeners,
	}

	c.gatewayMgr.UpsertGateway(instance)
	c.logger.Info(fmt.Sprintf("reconciled Gateway %s/%s with %d listeners (mode=%s, segment=%d)",
		gw.Namespace, gw.Name, len(listeners), gw.Spec.Mode, gw.Spec.SegmentID), &types.Metadata{
		Component: "gateway-controller",
		Namespace: gw.Namespace,
		GatewayID: string(gw.UID),
	})

	return nil
}

// DeleteGateway removes a Gateway instance.
func (c *GatewayController) DeleteGateway(ctx context.Context, ns, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.gatewayMgr.DeleteGateway(ns, name)
	c.logger.Info(fmt.Sprintf("deleted Gateway %s/%s", ns, name), &types.Metadata{
		Component: "gateway-controller",
		Namespace: ns,
	})
}
