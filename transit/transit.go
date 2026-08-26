// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package transit implements the straitKubegateway multi-cluster transit gateway.
//
// Supported topologies:
//   - Hub-and-spoke (cluster A as hub, B/C/D as spokes)
//   - Mesh (all clusters connected to all others)
//   - Peer-to-peer (two clusters directly connected)
//   - Gateway-to-gateway (cross-gateway routing)
//
// Segment model:
//   - Segment IDs are 32-bit unsigned integers (0–4294967295)
//   - Segment 0 is the backbone segment (default)
//   - All segments are isolated by default
//   - Inter-segment communication is via TransitSegmentAttachment CRDs
package transit

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/straitkubegateway/straitkubegateway/api/v1alpha1"
	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/ir"
	sgtypes "github.com/straitkubegateway/straitkubegateway/pkg/types"
)

// ============================================================================
// Transit Manager
// ============================================================================

// Manager manages multi-cluster transit gateway state.
type Manager struct {
	mu       sync.RWMutex
	client   client.Client
	log      *zap.Logger
	segments map[sgtypes.SegmentID]*ir.TransitSegment
	gateways map[types.NamespacedName]*ir.TransitSegment
}

// NewManager creates a new transit manager.
func NewManager(c client.Client, log *zap.Logger) *Manager {
	return &Manager{
		client:   c,
		log:      log,
		segments: make(map[sgtypes.SegmentID]*ir.TransitSegment),
		gateways: make(map[types.NamespacedName]*ir.TransitSegment),
	}
}

// ============================================================================
// TransitGateway reconciliation
// ============================================================================

// ReconcileGateway reconciles a TransitGateway CRD into transit IR state.
func (m *Manager) ReconcileGateway(ctx context.Context, key types.NamespacedName) error {
	var tg apiv1.TransitGateway
	if err := m.client.Get(ctx, key, &tg); err != nil {
		return client.IgnoreNotFound(err)
	}

	seg, err := m.compileGateway(&tg)
	if err != nil {
		return fmt.Errorf("compile transit gateway %s: %w", key, err)
	}

	m.mu.Lock()
	m.segments[seg.ID] = seg
	m.gateways[key] = seg
	m.mu.Unlock()

	m.log.Info("transit gateway reconciled",
		zap.String("name", key.Name),
		zap.String("segmentID", seg.ID.String()),
		zap.String("topology", tg.Spec.Topology),
		zap.Int("peers", len(tg.Spec.Peers)),
	)
	return nil
}

// DeleteGateway removes a transit gateway.
func (m *Manager) DeleteGateway(key types.NamespacedName) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if seg, ok := m.gateways[key]; ok {
		delete(m.segments, seg.ID)
		delete(m.gateways, key)
	}
}

// ============================================================================
// TransitSegmentAttachment reconciliation
// ============================================================================

// ReconcileAttachment reconciles a TransitSegmentAttachment CRD.
func (m *Manager) ReconcileAttachment(ctx context.Context, key types.NamespacedName) error {
	var att apiv1.TransitSegmentAttachment
	if err := m.client.Get(ctx, key, &att); err != nil {
		return client.IgnoreNotFound(err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure both segments exist
	segA := m.getOrCreateSegment(sgtypes.SegmentID(att.Spec.SegmentID))
	segB := m.getOrCreateSegment(sgtypes.SegmentID(att.Spec.PeerSegmentID))

	// Program inter-segment routes from the attachment spec
	for _, route := range att.Spec.Routes {
		prefix, err := netip.ParsePrefix(route.CIDR)
		if err != nil {
			m.log.Warn("invalid route CIDR in attachment",
				zap.String("cidr", route.CIDR),
				zap.Error(err),
			)
			continue
		}
		tr := ir.TransitRoute{CIDR: prefix, NextHop: route.NextHop}
		segA.Routes = appendUniqueRoute(segA.Routes, tr)
		segB.Routes = appendUniqueRoute(segB.Routes, tr)
	}

	m.log.Info("transit attachment reconciled",
		zap.String("name", key.Name),
		zap.Uint32("segmentA", att.Spec.SegmentID),
		zap.Uint32("segmentB", att.Spec.PeerSegmentID),
		zap.Int("routes", len(att.Spec.Routes)),
	)
	return nil
}

// ============================================================================
// State access
// ============================================================================

// GetAllSegments returns the current transit segment IR for compilation.
func (m *Manager) GetAllSegments() []*ir.TransitSegment {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*ir.TransitSegment, 0, len(m.segments))
	for _, s := range m.segments {
		out = append(out, s)
	}
	return out
}

// ============================================================================
// Internal helpers
// ============================================================================

func (m *Manager) compileGateway(tg *apiv1.TransitGateway) (*ir.TransitSegment, error) {
	seg := m.getOrCreateSegment(sgtypes.SegmentID(tg.Spec.SegmentID))

	// Add peer cluster CIDRs as routes
	for _, peer := range tg.Spec.Peers {
		for _, cidrStr := range peer.PodCIDRs {
			prefix, err := netip.ParsePrefix(cidrStr)
			if err != nil {
				return nil, fmt.Errorf("invalid peer pod CIDR %q: %w", cidrStr, err)
			}
			seg.Routes = appendUniqueRoute(seg.Routes, ir.TransitRoute{
				CIDR:    prefix,
				NextHop: peer.Endpoint,
			})
		}
		// Add peer cluster ID
		seg.Clusters = appendUniqueCID(seg.Clusters, sgtypes.ClusterID(peer.ClusterName))
	}
	return seg, nil
}

func (m *Manager) getOrCreateSegment(id sgtypes.SegmentID) *ir.TransitSegment {
	if s, ok := m.segments[id]; ok {
		return s
	}
	s := &ir.TransitSegment{ID: id}
	m.segments[id] = s
	return s
}

func appendUniqueRoute(routes []ir.TransitRoute, r ir.TransitRoute) []ir.TransitRoute {
	for _, existing := range routes {
		if existing.CIDR == r.CIDR && existing.NextHop == r.NextHop {
			return routes
		}
	}
	return append(routes, r)
}

func appendUniqueCID(ids []sgtypes.ClusterID, id sgtypes.ClusterID) []sgtypes.ClusterID {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}

// ============================================================================
// Topology helpers
// ============================================================================

// TopologyHubSpoke describes a hub-and-spoke transit segment topology.
// Hub: a single gateway that all spokes connect to.
// Backbone segment (0) connects all segments by default.
//
// Example:
//
//	Cluster-A --- Gateway1 --- Cluster-C  (segment=1)
//	               |
//	           Cluster-B  (backbone segment=0)
type TopologyHubSpoke struct {
	HubCluster   sgtypes.ClusterID
	SpokeCluster []sgtypes.ClusterID
	SegmentID    sgtypes.SegmentID
}

// TopologyMesh describes a full-mesh transit topology where all clusters
// are connected to each other.
type TopologyMesh struct {
	Clusters  []sgtypes.ClusterID
	SegmentID sgtypes.SegmentID
}

// TopologyPeerToPeer describes a direct peer-to-peer transit connection.
type TopologyPeerToPeer struct {
	ClusterA  sgtypes.ClusterID
	ClusterB  sgtypes.ClusterID
	SegmentID sgtypes.SegmentID
}
