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

// TransitController reconciles TransitGateway, Segment, and TransitAttachment CRDs.
type TransitController struct {
	mu     sync.RWMutex
	engine *transit.TransitGatewayEngine
	logger *logging.Logger
}

// NewTransitController creates a new Transit controller.
func NewTransitController(engine *transit.TransitGatewayEngine) *TransitController {
	return &TransitController{
		engine: engine,
		logger: logging.DefaultLogger(),
	}
}

// ReconcileTransitGateway reconciles a TransitGateway CRD.
func (c *TransitController) ReconcileTransitGateway(ctx context.Context, tg *v1alpha1.TransitGateway) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.logger.Info(fmt.Sprintf("reconciled TransitGateway %s (Topology=%s, SegmentID=%d, TunnelType=%s)",
		tg.Name, tg.Spec.Topology, tg.Spec.SegmentID, tg.Spec.TunnelType), &types.Metadata{
		Component: "transit-controller",
		SegmentID: fmt.Sprintf("%d", tg.Spec.SegmentID),
	})

	return nil
}

// ReconcileAttachment reconciles a TransitAttachment CRD.
func (c *TransitController) ReconcileAttachment(ctx context.Context, att *v1alpha1.TransitAttachment) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var subnets []netip.Prefix
	for _, s := range att.Spec.PodCIDRs {
		p, err := netip.ParsePrefix(s)
		if err == nil {
			subnets = append(subnets, p)
		}
	}
	for _, s := range att.Spec.ServiceCIDRs {
		p, err := netip.ParsePrefix(s)
		if err == nil {
			subnets = append(subnets, p)
		}
	}

	attachment := transit.Attachment{
		ID:        string(att.UID),
		Name:      att.Name,
		Type:      transit.AttachmentTypeCluster,
		SegmentID: types.SegmentID(att.Spec.SegmentID),
		ClusterID: att.Spec.ClusterID,
		Subnets:   subnets,
	}

	c.engine.Attach(attachment)
	c.logger.Info(fmt.Sprintf("attached %s to TransitGateway %s (ClusterID=%s, Segment=%d, Subnets=%v)",
		att.Name, att.Spec.TransitGatewayRef, att.Spec.ClusterID, att.Spec.SegmentID, subnets), &types.Metadata{
		Component: "transit-controller",
		SegmentID: fmt.Sprintf("%d", att.Spec.SegmentID),
		ClusterID: att.Spec.ClusterID,
	})

	return nil
}
