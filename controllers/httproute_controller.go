package controllers

import (
	"context"
	"fmt"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/gateway"
	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// HTTPRouteDescriptor represents an HTTPRoute resource for reconciliation.
type HTTPRouteDescriptor struct {
	Namespace   string
	Name        string
	GatewayNs   string
	GatewayName string
	Hostnames   []string
	Rules       []gateway.HTTPRouteRule
}

// HTTPRouteController reconciles Gateway API HTTPRoutes and binds them to Gateways.
type HTTPRouteController struct {
	mu         sync.RWMutex
	gatewayMgr *gateway.Manager
	logger     *logging.Logger
}

// NewHTTPRouteController creates a new HTTPRoute controller.
func NewHTTPRouteController(gwMgr *gateway.Manager) *HTTPRouteController {
	return &HTTPRouteController{
		gatewayMgr: gwMgr,
		logger:     logging.DefaultLogger(),
	}
}

// ReconcileHTTPRoute updates an HTTPRoute and binds it to its parent Gateway.
func (c *HTTPRouteController) ReconcileHTTPRoute(ctx context.Context, desc HTTPRouteDescriptor) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	route := gateway.Route{
		ID:        fmt.Sprintf("%s/%s", desc.Namespace, desc.Name),
		Namespace: desc.Namespace,
		Name:      desc.Name,
		Hostnames: desc.Hostnames,
		Rules:     desc.Rules,
		Protocol:  types.ProtocolTCP,
	}

	if err := c.gatewayMgr.UpsertRoute(desc.GatewayNs, desc.GatewayName, route); err != nil {
		return err
	}

	c.logger.Info(fmt.Sprintf("reconciled HTTPRoute %s/%s -> Gateway %s/%s (%d rules, %d hostnames)",
		desc.Namespace, desc.Name, desc.GatewayNs, desc.GatewayName, len(desc.Rules), len(desc.Hostnames)), &types.Metadata{
		Component: "httproute-controller",
		Namespace: desc.Namespace,
		RouteID:   route.ID,
	})

	return nil
}

// DeleteHTTPRoute removes an HTTPRoute from its parent Gateway.
func (c *HTTPRouteController) DeleteHTTPRoute(ctx context.Context, gwNs, gwName, routeNs, routeName string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.gatewayMgr.DeleteRoute(gwNs, gwName, routeNs, routeName)
	c.logger.Info(fmt.Sprintf("deleted HTTPRoute %s/%s from Gateway %s/%s", routeNs, routeName, gwNs, gwName), &types.Metadata{
		Component: "httproute-controller",
		Namespace: routeNs,
	})
}
