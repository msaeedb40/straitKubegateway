package transit

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// TransitGatewayEngine manages segment isolation, route tables, and inter-segment policies.
type TransitGatewayEngine struct {
	mu          sync.RWMutex
	id          string
	name        string
	asn         uint32
	attachments map[string]*Attachment             // ID -> Attachment
	routes      map[types.SegmentID][]TransitRoute // SegmentID -> Routes
	peerings    map[string]*PeeringLink
}

// NewTransitGatewayEngine creates a new Transit Gateway engine.
func NewTransitGatewayEngine(id, name string, asn uint32) *TransitGatewayEngine {
	if asn == 0 {
		asn = 64512
	}
	return &TransitGatewayEngine{
		id:          id,
		name:        name,
		asn:         asn,
		attachments: make(map[string]*Attachment),
		routes:      make(map[types.SegmentID][]TransitRoute),
		peerings:    make(map[string]*PeeringLink),
	}
}

// Attach registers an attachment to the Transit Gateway and auto-propagates routes.
func (tg *TransitGatewayEngine) Attach(att Attachment) {
	tg.mu.Lock()
	defer tg.mu.Unlock()

	attCopy := att
	tg.attachments[att.ID] = &attCopy

	// Auto-propagate attachment subnets into its segment route table
	for _, subnet := range att.Subnets {
		r := TransitRoute{
			Destination:  subnet,
			AttachmentID: att.ID,
			SegmentID:    att.SegmentID,
			Metric:       100,
		}
		tg.routes[att.SegmentID] = append(tg.routes[att.SegmentID], r)
	}
}

// Detach removes an attachment and its routes.
func (tg *TransitGatewayEngine) Detach(attachmentID string) {
	tg.mu.Lock()
	defer tg.mu.Unlock()

	att, exists := tg.attachments[attachmentID]
	if !exists {
		return
	}

	segID := att.SegmentID
	delete(tg.attachments, attachmentID)

	// Filter out routes belonging to detached attachment
	var remaining []TransitRoute
	for _, r := range tg.routes[segID] {
		if r.AttachmentID != attachmentID {
			remaining = append(remaining, r)
		}
	}
	tg.routes[segID] = remaining
}

// AddRoute manually adds a route into a specific segment.
func (tg *TransitGatewayEngine) AddRoute(route TransitRoute) {
	tg.mu.Lock()
	defer tg.mu.Unlock()
	tg.routes[route.SegmentID] = append(tg.routes[route.SegmentID], route)
}

// LookupRoute resolves the destination within the caller's segment.
// Segments are strictly isolated unless explicitly routed.
func (tg *TransitGatewayEngine) LookupRoute(srcSegment types.SegmentID, dstIP netip.Addr) (*TransitRoute, error) {
	tg.mu.RLock()
	defer tg.mu.RUnlock()

	routes, exists := tg.routes[srcSegment]
	if !exists {
		return nil, fmt.Errorf("no route table found for segment %d", srcSegment)
	}

	var bestMatch *TransitRoute
	var maxPrefixLen int

	for i := range routes {
		r := &routes[i]
		if r.Destination.Contains(dstIP) {
			if r.Destination.Bits() > maxPrefixLen {
				maxPrefixLen = r.Destination.Bits()
				matchCopy := *r
				bestMatch = &matchCopy
			}
		}
	}

	if bestMatch == nil {
		return nil, fmt.Errorf("no route to %s in segment %d", dstIP, srcSegment)
	}

	return bestMatch, nil
}
