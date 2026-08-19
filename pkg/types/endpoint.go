package types

import "net/netip"

// Endpoint represents a network endpoint in the straitKubegateway dataplane.
// An endpoint is typically a pod's network interface, identified by its
// IP address and BPF identity.
type Endpoint struct {
	// ID is the unique identifier for this endpoint.
	ID uint64

	// Identity is the BPF identity assigned to this endpoint.
	Identity Identity

	// IPv4 is the endpoint's IPv4 address, if any.
	IPv4 netip.Addr

	// IPv6 is the endpoint's IPv6 address, if any.
	IPv6 netip.Addr

	// NodeID identifies the node hosting this endpoint.
	NodeID string

	// Namespace is the Kubernetes namespace.
	Namespace string

	// PodName is the Kubernetes pod name.
	PodName string

	// ContainerID is the container runtime ID.
	ContainerID string

	// NetNsCookie is the network namespace cookie for this endpoint.
	NetNsCookie uint64

	// IfIndex is the host-side interface index (NetKit host endpoint).
	IfIndex int

	// SegmentID is the network segment this endpoint belongs to.
	SegmentID SegmentID

	// Labels are the Kubernetes labels on the endpoint's pod.
	Labels map[string]string

	// State is the current operational state of the endpoint.
	State EndpointState
}

// EndpointState represents the lifecycle state of an endpoint.
type EndpointState uint8

const (
	// EndpointStateCreating indicates the endpoint is being set up.
	EndpointStateCreating EndpointState = iota

	// EndpointStateReady indicates the endpoint is fully operational.
	EndpointStateReady

	// EndpointStateDisconnecting indicates the endpoint is being torn down.
	EndpointStateDisconnecting

	// EndpointStateDisconnected indicates the endpoint has been removed.
	EndpointStateDisconnected
)

// String returns a human-readable representation of the endpoint state.
func (s EndpointState) String() string {
	switch s {
	case EndpointStateCreating:
		return "creating"
	case EndpointStateReady:
		return "ready"
	case EndpointStateDisconnecting:
		return "disconnecting"
	case EndpointStateDisconnected:
		return "disconnected"
	default:
		return "unknown"
	}
}

// HasIPv4 reports whether the endpoint has an IPv4 address.
func (e *Endpoint) HasIPv4() bool {
	return e.IPv4.IsValid()
}

// HasIPv6 reports whether the endpoint has an IPv6 address.
func (e *Endpoint) HasIPv6() bool {
	return e.IPv6.IsValid()
}
