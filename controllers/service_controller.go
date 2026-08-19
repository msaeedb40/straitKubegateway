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

// ServiceDescriptor defines a Kubernetes Service representation for reconciliation.
type ServiceDescriptor struct {
	Namespace       string
	Name            string
	ClusterIP       string
	Port            uint16
	Protocol        types.Protocol
	Algorithm       service.Algorithm
	SessionAffinity bool
	AffinityTimeout uint32
}

// ServiceController reconciles Kubernetes Services and compiles state into the Service LB manager.
type ServiceController struct {
	mu        sync.RWMutex
	lbManager *service.Manager
	logger    *logging.Logger
}

// NewServiceController creates a new Service reconciler.
func NewServiceController(lbMgr *service.Manager) *ServiceController {
	return &ServiceController{
		lbManager: lbMgr,
		logger:    logging.DefaultLogger(),
	}
}

// ReconcileService processes a Service definition and updates the dataplane.
func (c *ServiceController) ReconcileService(ctx context.Context, desc ServiceDescriptor, backends []service.BackendEndpoint) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	vip, err := netip.ParseAddr(desc.ClusterIP)
	if err != nil {
		return fmt.Errorf("invalid cluster IP %q: %w", desc.ClusterIP, err)
	}

	algo := desc.Algorithm
	if algo == 0 {
		algo = service.AlgorithmMaglevHash // Default to Maglev Hash
	}

	svc := c.lbManager.UpsertService(
		desc.Namespace,
		desc.Name,
		vip,
		desc.Port,
		desc.Protocol,
		algo,
		desc.SessionAffinity,
		desc.AffinityTimeout,
		backends,
	)

	c.logger.Info(fmt.Sprintf("reconciled Service %s/%s (VIP %s:%d/%s, algo=%d, backends=%d)",
		desc.Namespace, desc.Name, vip, desc.Port, desc.Protocol, algo, len(svc.Backends)), &types.Metadata{
		Component:   "service-controller",
		Namespace:   desc.Namespace,
		ServiceName: desc.Name,
	})

	return nil
}

// DeleteService removes a Service from the load balancer.
func (c *ServiceController) DeleteService(ctx context.Context, ns, name string, port uint16, proto types.Protocol) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lbManager.DeleteService(ns, name, port, proto)
	c.logger.Info(fmt.Sprintf("deleted Service %s/%s (%d/%s)", ns, name, port, proto), &types.Metadata{
		Component:   "service-controller",
		Namespace:   ns,
		ServiceName: name,
	})
}
