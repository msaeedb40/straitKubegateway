// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package types defines core shared primitive types used across all
// straitKubegateway packages.
package types

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"strings"
)

// ============================================================================
// Identity
// ============================================================================

// Identity is a numeric network identity used in eBPF policy enforcement.
// Identities are allocated per workload endpoint.
type Identity uint32

const (
	// IdentityUnknown is the zero/unassigned identity.
	IdentityUnknown Identity = 0
	// IdentityWorld represents traffic to/from outside the cluster.
	IdentityWorld Identity = 1
	// IdentityHost represents traffic to/from the host network namespace.
	IdentityHost Identity = 2
	// IdentityInit represents workloads that have not yet received an identity.
	IdentityInit Identity = 3
	// IdentityRemoteNode represents traffic from remote cluster nodes.
	IdentityRemoteNode Identity = 4
	// IdentityMin is the minimum value for user-allocated identities.
	IdentityMin Identity = 256
	// IdentityMax is the maximum value for user-allocated identities.
	IdentityMax Identity = 0xFFFFFFFF
)

// IsReserved returns true if this identity is a reserved/system identity.
func (id Identity) IsReserved() bool {
	return id < IdentityMin
}

// String returns the string representation of the identity.
func (id Identity) String() string {
	switch id {
	case IdentityUnknown:
		return "unknown"
	case IdentityWorld:
		return "world"
	case IdentityHost:
		return "host"
	case IdentityInit:
		return "init"
	case IdentityRemoteNode:
		return "remote-node"
	default:
		return fmt.Sprintf("%d", uint32(id))
	}
}

// ============================================================================
// SegmentID
// ============================================================================

// SegmentID is a 32-bit transit segment identifier.
// Segment 0 is the backbone segment; all segments are isolated by default.
type SegmentID uint32

const (
	// SegmentBackbone is the default backbone segment (0).
	SegmentBackbone SegmentID = 0
	// SegmentMax is the maximum segment ID (2^32 - 1).
	SegmentMax SegmentID = 4294967295
)

// IsBackbone returns true if this is the backbone segment.
func (s SegmentID) IsBackbone() bool { return s == SegmentBackbone }

// String returns the string representation of the segment ID.
func (s SegmentID) String() string { return fmt.Sprintf("%d", uint32(s)) }

// ============================================================================
// ClusterID
// ============================================================================

// ClusterID uniquely identifies a Kubernetes cluster in a federation.
type ClusterID string

// NodeID uniquely identifies a Kubernetes node within a cluster.
type NodeID string

// ============================================================================
// CIDR
// ============================================================================

// CIDR wraps netip.Prefix providing convenience methods used throughout
// the straitKubegateway dataplane.
type CIDR struct {
	netip.Prefix
}

// ParseCIDR parses a CIDR string into a CIDR value.
func ParseCIDR(s string) (CIDR, error) {
	p, err := netip.ParsePrefix(s)
	if err != nil {
		return CIDR{}, fmt.Errorf("invalid CIDR %q: %w", s, err)
	}
	return CIDR{p.Masked()}, nil
}

// MustParseCIDR parses a CIDR string or panics.
func MustParseCIDR(s string) CIDR {
	c, err := ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return c
}

// Contains reports whether ip is within the CIDR.
func (c CIDR) Contains(ip netip.Addr) bool { return c.Prefix.Contains(ip) }

// Overlaps reports whether two CIDRs overlap.
func (c CIDR) Overlaps(other CIDR) bool { return c.Prefix.Overlaps(other.Prefix) }

// IsIPv4 reports whether this CIDR is IPv4.
func (c CIDR) IsIPv4() bool { return c.Addr().Is4() }

// IsIPv6 reports whether this CIDR is IPv6.
func (c CIDR) IsIPv6() bool { return c.Addr().Is6() }

// ToNetIPNet converts to a *net.IPNet.
func (c CIDR) ToNetIPNet() *net.IPNet {
	prefix := c.Prefix
	ip := prefix.Addr().AsSlice()
	bits := prefix.Bits()
	totalBits := prefix.Addr().BitLen()
	mask := net.CIDRMask(bits, totalBits)
	return &net.IPNet{IP: ip, Mask: mask}
}

// ============================================================================
// IP
// ============================================================================

// IP wraps netip.Addr providing helpers for the straitKubegateway dataplane.
type IP struct {
	netip.Addr
}

// ParseIP parses an IP address string.
func ParseIP(s string) (IP, error) {
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return IP{}, fmt.Errorf("invalid IP %q: %w", s, err)
	}
	return IP{addr}, nil
}

// MustParseIP parses an IP or panics.
func MustParseIP(s string) IP {
	ip, err := ParseIP(s)
	if err != nil {
		panic(err)
	}
	return ip
}

// To4Bytes returns the 4-byte IPv4 representation, or an error if not IPv4.
func (ip IP) To4Bytes() ([4]byte, error) {
	a := ip.Addr
	if !a.Is4() {
		return [4]byte{}, fmt.Errorf("IP %s is not IPv4", a)
	}
	return a.As4(), nil
}

// To16Bytes returns the 16-byte representation.
func (ip IP) To16Bytes() [16]byte { return ip.Addr.As16() }

// ToUint32 converts an IPv4 address to a uint32 (network byte order).
func (ip IP) ToUint32() (uint32, error) {
	b, err := ip.To4Bytes()
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b[:]), nil
}

// ============================================================================
// MAC
// ============================================================================

// MAC represents an Ethernet MAC address.
type MAC [6]byte

// ParseMAC parses a MAC address string.
func ParseMAC(s string) (MAC, error) {
	hw, err := net.ParseMAC(s)
	if err != nil {
		return MAC{}, fmt.Errorf("invalid MAC %q: %w", s, err)
	}
	if len(hw) != 6 {
		return MAC{}, fmt.Errorf("MAC %q is not 48-bit", s)
	}
	var m MAC
	copy(m[:], hw)
	return m, nil
}

// String returns the colon-separated hex representation.
func (m MAC) String() string {
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		m[0], m[1], m[2], m[3], m[4], m[5])
}

// IsZero reports whether this is the zero MAC address.
func (m MAC) IsZero() bool { return m == (MAC{}) }

// ============================================================================
// Endpoint
// ============================================================================

// Endpoint represents a network endpoint (pod or service backend).
type Endpoint struct {
	// ID is the numeric endpoint identifier.
	ID uint64
	// Identity is the BPF identity for policy enforcement.
	Identity Identity
	// IP is the endpoint's primary IP address.
	IP IP
	// Port is the endpoint's port (0 for pod endpoints).
	Port uint16
	// NodeID is the node hosting this endpoint.
	NodeID NodeID
	// ClusterID is the cluster this endpoint belongs to.
	ClusterID ClusterID
	// Namespace is the Kubernetes namespace.
	Namespace string
	// PodName is the Kubernetes pod name.
	PodName string
}

// String returns a human-readable description of the endpoint.
func (e *Endpoint) String() string {
	return fmt.Sprintf("endpoint{id=%d ip=%s identity=%s ns=%s pod=%s}",
		e.ID, e.IP, e.Identity, e.Namespace, e.PodName)
}

// ============================================================================
// ObservabilityMetadata
// ============================================================================

// ObservabilityMetadata is the common metadata attached to every major event,
// log entry, metric, trace, flow event, and policy decision.
// Invariant: every major object carries all of these fields.
type ObservabilityMetadata struct {
	ClusterID  ClusterID `json:"clusterID"`
	NodeID     NodeID    `json:"nodeID"`
	Namespace  string    `json:"namespace,omitempty"`
	Pod        string    `json:"pod,omitempty"`
	Service    string    `json:"service,omitempty"`
	Endpoint   string    `json:"endpoint,omitempty"`
	FlowID     string    `json:"flowID,omitempty"`
	TraceID    string    `json:"traceID,omitempty"`
	PolicyID   string    `json:"policyID,omitempty"`
	SegmentID  SegmentID `json:"segmentID"`
	GatewayID  string    `json:"gatewayID,omitempty"`
}

// ============================================================================
// TunnelMode
// ============================================================================

// TunnelMode specifies the overlay encapsulation protocol.
type TunnelMode string

const (
	TunnelModeVXLAN    TunnelMode = "vxlan"
	TunnelModeGeneve   TunnelMode = "geneve"
	TunnelModeGRE      TunnelMode = "gre"
	TunnelModeDisabled TunnelMode = "disabled"
)

// ============================================================================
// PolicyAction
// ============================================================================

// PolicyAction defines what to do when a policy rule matches.
type PolicyAction string

const (
	PolicyActionAllow  PolicyAction = "Allow"
	PolicyActionDeny   PolicyAction = "Deny"
	PolicyActionReject PolicyAction = "Reject"
)

// ============================================================================
// Protocol
// ============================================================================

// Protocol represents a network protocol for policy rules.
type Protocol string

const (
	ProtocolTCP  Protocol = "TCP"
	ProtocolUDP  Protocol = "UDP"
	ProtocolICMP Protocol = "ICMP"
)

// ============================================================================
// LBAlgorithm
// ============================================================================

// LBAlgorithm selects the load-balancing algorithm for service backends.
type LBAlgorithm string

const (
	LBAlgorithmMaglev             LBAlgorithm = "maglev"
	LBAlgorithmRoundRobin         LBAlgorithm = "round-robin"
	LBAlgorithmLeastConnections   LBAlgorithm = "least-connections"
	LBAlgorithmWeightedRoundRobin LBAlgorithm = "weighted-round-robin"
	LBAlgorithmIPHash             LBAlgorithm = "ip-hash"
	LBAlgorithmRandom             LBAlgorithm = "random"
	LBAlgorithmFailover           LBAlgorithm = "failover"
)

// ============================================================================
// EncryptionType
// ============================================================================

// EncryptionType specifies the encryption backend.
type EncryptionType string

const (
	EncryptionTypeWireGuard EncryptionType = "wireguard"
	EncryptionTypeIPsec     EncryptionType = "ipsec"
	EncryptionTypeDisabled  EncryptionType = "disabled"
)

// ============================================================================
// ReadinessCondition keys
// ============================================================================

// Readiness condition type names used in StraitNode status.
const (
	ConditionCNIReady     = "CNIReady"
	ConditionServiceReady = "ServiceReady"
	ConditionPolicyReady  = "PolicyReady"
	ConditionGatewayReady = "GatewayReady"
	ConditionTransitReady = "TransitReady"
	ConditionBGPReady     = "BGPReady"
)

// ============================================================================
// Helpers
// ============================================================================

// SplitCIDR splits a comma-separated list of CIDRs into individual CIDR values.
func SplitCIDR(s string) ([]CIDR, error) {
	parts := strings.Split(s, ",")
	cidrs := make([]CIDR, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		c, err := ParseCIDR(p)
		if err != nil {
			return nil, err
		}
		cidrs = append(cidrs, c)
	}
	return cidrs, nil
}
