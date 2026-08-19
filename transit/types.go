// Package transit provides Transit Gateway multi-cluster peering,
// 32-bit segment isolation (0..4294967295), route propagation, and encrypted overlay tunnels.
package transit

import (
	"net/netip"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// AttachmentType defines the type of entity attached to a Transit Gateway.
type AttachmentType string

const (
	AttachmentTypeVPC     AttachmentType = "VPC"
	AttachmentTypeCluster AttachmentType = "Cluster"
	AttachmentTypeSubnet  AttachmentType = "Subnet"
	AttachmentTypePeering AttachmentType = "Peering"
)

// Attachment represents a connection to a Transit Gateway.
type Attachment struct {
	ID         string
	Name       string
	Type       AttachmentType
	SegmentID  types.SegmentID // 0 .. 4294967295
	ClusterID  string
	Subnets    []netip.Prefix
	Propagated bool
}

// TransitRoute represents a route inside a Transit Gateway route table.
type TransitRoute struct {
	Destination  netip.Prefix
	NextHop      netip.Addr
	AttachmentID string
	SegmentID    types.SegmentID
	Metric       uint32
}

// PeeringLink represents an encrypted peering connection between two Transit Gateways.
type PeeringLink struct {
	ID             string
	LocalCluster   string
	RemoteCluster  string
	RemoteEndpoint netip.AddrPort
	Encryption     string // WireGuard, IPsec
	Segments       []types.SegmentID
	Healthy        bool
}
