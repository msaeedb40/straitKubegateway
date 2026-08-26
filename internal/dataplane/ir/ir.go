// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package ir defines the Dataplane Intermediate Representation (IR).
//
// The IR is the ONLY boundary between the straitKubegateway control plane
// and the kernel dataplane. Controllers produce IR; the Compiler translates
// IR into BPF/netlink state. Nothing else touches BPF maps directly.
package ir

import (
	"net/netip"

	sgtypes "github.com/straitkubegateway/straitkubegateway/pkg/types"
)

// ============================================================================
// Generation / revision tracking
// Invariant: every state transition must be generation/revision based.
// ============================================================================

// Generation is a monotonically increasing version used for optimistic
// concurrency control in IR state transitions.
type Generation uint64

// ============================================================================
// NetworkState — the complete desired dataplane state
// ============================================================================

// NetworkState is the top-level desired dataplane state produced by all
// controllers and consumed by the Compiler.
type NetworkState struct {
	// Generation is incremented on every mutation.
	Generation Generation

	// Endpoints is the map of all endpoints keyed by endpoint ID.
	Endpoints map[uint64]*Endpoint

	// Services is the map of all Kubernetes services.
	Services map[ServiceKey]*Service

	// Policies is the ordered list of active network policies.
	Policies []*Policy

	// Routes is the set of routes to program.
	Routes []*Route

	// NATRules is the set of NAT rules to program.
	NATRules []*NATRule

	// TunnelPeers is the set of tunnel peers for overlay networking.
	TunnelPeers []*TunnelPeer

	// TransitSegments is the set of active transit segments.
	TransitSegments []*TransitSegment
}

// NewNetworkState creates an empty NetworkState.
func NewNetworkState() *NetworkState {
	return &NetworkState{
		Endpoints: make(map[uint64]*Endpoint),
		Services:  make(map[ServiceKey]*Service),
	}
}

// ============================================================================
// Endpoint IR
// ============================================================================

// Endpoint represents a single network endpoint (pod, host, or remote).
type Endpoint struct {
	// ID is the unique numeric identifier for this endpoint.
	ID uint64

	// Identity is the BPF identity for policy matching.
	Identity sgtypes.Identity

	// IP is the primary IPv4 address.
	IP netip.Addr

	// IPv6 is the primary IPv6 address (dual-stack).
	IPv6 netip.Addr

	// NodeIP is the IP of the host node.
	NodeIP netip.Addr

	// Namespace is the Kubernetes namespace.
	Namespace string

	// PodName is the Kubernetes pod name.
	PodName string

	// Labels are the pod labels used for policy selection.
	Labels map[string]string

	// NetNSPath is the path to the pod's network namespace.
	NetNSPath string

	// IfIndex is the host-side interface index for this endpoint.
	IfIndex int

	// ContainerIfIndex is the container-side interface index.
	ContainerIfIndex int

	// SegmentID is the transit segment this endpoint belongs to.
	SegmentID sgtypes.SegmentID

	// Generation is the revision of this endpoint entry.
	Generation Generation
}

// ============================================================================
// Service IR
// ============================================================================

// ServiceKey uniquely identifies a Kubernetes service.
type ServiceKey struct {
	Namespace string
	Name      string
}

// ServiceType enumerates Kubernetes service types.
type ServiceType string

const (
	ServiceTypeClusterIP    ServiceType = "ClusterIP"
	ServiceTypeNodePort     ServiceType = "NodePort"
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"
	ServiceTypeExternalName ServiceType = "ExternalName"
)

// Service is the IR representation of a Kubernetes Service.
// Invariant: Kubernetes API state is compiled into BPF maps so that
// packets never reach a userspace proxy.
type Service struct {
	Key       ServiceKey
	Type      ServiceType
	ClusterIP netip.Addr
	Ports     []ServicePort
	Backends  []Backend
	Algorithm sgtypes.LBAlgorithm
	// SessionAffinity enables session affinity for this service.
	SessionAffinity bool
	// DSR enables Direct Server Return.
	DSR bool
	Generation Generation
}

// ServicePort maps a frontend port to a backend port.
type ServicePort struct {
	Protocol sgtypes.Protocol
	Port     uint16
	NodePort uint16
	TargetPort uint16
}

// Backend is a single service backend endpoint.
type Backend struct {
	ID       uint32
	IP       netip.Addr
	Port     uint16
	Weight   uint32
	NodeIP   netip.Addr
	Healthy  bool
}

// ============================================================================
// Policy IR
// ============================================================================

// PolicyDirection is the traffic direction a policy rule applies to.
type PolicyDirection string

const (
	PolicyDirectionIngress PolicyDirection = "Ingress"
	PolicyDirectionEgress  PolicyDirection = "Egress"
)

// Policy is the IR representation of a compiled network policy.
// The policy engine evaluates rules in priority order (lower = higher priority).
// Deterministic semantics:
//   - Default ingress: Deny
//   - Default egress: Allow
//   - On equal priority: Deny overrides Allow
type Policy struct {
	// ID is the unique policy identifier (used in observability metadata).
	ID string
	// Priority determines evaluation order (0=highest, 255=lowest).
	Priority uint8
	// Direction is Ingress or Egress.
	Direction PolicyDirection
	// SourceIdentities are the BPF identities allowed/denied as source.
	SourceIdentities []sgtypes.Identity
	// DestIdentities are the BPF identities allowed/denied as destination.
	DestIdentities []sgtypes.Identity
	// Ports are the port/protocol matchers for this rule.
	Ports []sgtypes.Protocol
	// Action is the compiled rule action.
	Action sgtypes.PolicyAction // defined in pkg/types
	Generation Generation
}

// ============================================================================
// Route IR
// ============================================================================

// RouteType identifies how this route is programmed.
type RouteType string

const (
	RouteTypeDirect  RouteType = "direct"
	RouteTypeVXLAN   RouteType = "vxlan"
	RouteTypeGeneve  RouteType = "geneve"
	RouteTypeGRE     RouteType = "gre"
	RouteTypeWireGuard RouteType = "wireguard"
)

// Route is the IR representation of a routing table entry.
type Route struct {
	// Destination is the destination prefix.
	Destination netip.Prefix
	// NextHop is the next-hop IP address.
	NextHop netip.Addr
	// Dev is the output interface name.
	Dev string
	// Type is the route encapsulation type.
	Type RouteType
	// Metric is the route metric.
	Metric uint32
	// TableID is the Linux routing table ID.
	TableID uint32
	Generation Generation
}

// ============================================================================
// NAT IR
// ============================================================================

// NATType specifies the NAT direction.
type NATType string

const (
	NATTypeSNAT NATType = "SNAT"
	NATTypeDNAT NATType = "DNAT"
	NATTypeMasq NATType = "Masquerade"
	NATTypeNAT64 NATType = "NAT64"
)

// NATRule is the IR representation of a NAT rule.
type NATRule struct {
	Type       NATType
	Match      netip.Prefix
	RewriteIP  netip.Addr
	RewritePort uint16
	OutDev     string
	Generation Generation
}

// ============================================================================
// Tunnel IR
// ============================================================================

// TunnelPeer represents a remote node reachable via an overlay tunnel.
type TunnelPeer struct {
	// NodeIP is the remote node's IP.
	NodeIP netip.Addr
	// TunnelIP is the remote tunnel endpoint IP.
	TunnelIP netip.Addr
	// PodCIDR is the remote pod CIDR served by this node.
	PodCIDR netip.Prefix
	// Mode is the tunnel encapsulation mode.
	Mode sgtypes.TunnelMode
	// Port is the tunnel UDP port.
	Port uint16
	Generation Generation
}

// ============================================================================
// Transit Segment IR
// ============================================================================

// TransitSegment represents an active transit segment in the IR.
type TransitSegment struct {
	// ID is the 32-bit segment identifier (0 = backbone).
	ID sgtypes.SegmentID
	// Clusters are the cluster IDs participating in this segment.
	Clusters []sgtypes.ClusterID
	// Routes are the inter-segment routes.
	Routes []TransitRoute
	Generation Generation
}

// TransitRoute is a route within a transit segment.
type TransitRoute struct {
	CIDR    netip.Prefix
	NextHop string // attachment name or IP
}
