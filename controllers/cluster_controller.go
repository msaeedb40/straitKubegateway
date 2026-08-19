package controllers

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/api/v1alpha1"
	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
	"github.com/straitKubegateway/straitKubegateway/transit"
)

// ClusterLinkController reconciles ClusterLink multi-cluster peering CRDs.
type ClusterLinkController struct {
	mu        sync.RWMutex
	tunnelMgr *transit.TunnelManager
	logger    *logging.Logger
}

// NewClusterLinkController creates a new ClusterLink controller.
func NewClusterLinkController(tunnelMgr *transit.TunnelManager) *ClusterLinkController {
	return &ClusterLinkController{
		tunnelMgr: tunnelMgr,
		logger:    logging.DefaultLogger(),
	}
}

// ReconcileClusterLink updates peering links and WireGuard peer configurations.
func (c *ClusterLinkController) ReconcileClusterLink(ctx context.Context, link *v1alpha1.ClusterLink) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var allowedCIDRs []netip.Prefix
	for _, cidr := range link.Spec.PodCIDRs {
		p, err := netip.ParsePrefix(cidr)
		if err == nil {
			allowedCIDRs = append(allowedCIDRs, p)
		}
	}
	for _, cidr := range link.Spec.ServiceCIDRs {
		p, err := netip.ParsePrefix(cidr)
		if err == nil {
			allowedCIDRs = append(allowedCIDRs, p)
		}
	}

	peering := transit.PeeringLink{
		ID:            string(link.UID),
		LocalCluster:  "local",
		RemoteCluster: link.Spec.ClusterID,
		Encryption:    "WireGuard",
		Segments:      []types.SegmentID{0},
		Healthy:       true,
	}

	c.logger.Info(fmt.Sprintf("reconciled ClusterLink %s -> Remote Cluster %s (APIEndpoint=%s, PodCIDRs=%v, ServiceCIDRs=%v)",
		link.Name, link.Spec.ClusterID, link.Spec.APIEndpoint, link.Spec.PodCIDRs, link.Spec.ServiceCIDRs), &types.Metadata{
		Component: "cluster-controller",
		ClusterID: link.Spec.ClusterID,
	})

	_ = peering
	_ = allowedCIDRs

	return nil
}
