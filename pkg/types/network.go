package types

import (
	"fmt"
	"net/netip"
)

// AddressFamily represents an IP address family.
type AddressFamily uint8

const (
	// IPv4 represents the IPv4 address family.
	IPv4 AddressFamily = 4

	// IPv6 represents the IPv6 address family.
	IPv6 AddressFamily = 6
)

// String returns the human-readable address family name.
func (af AddressFamily) String() string {
	switch af {
	case IPv4:
		return "IPv4"
	case IPv6:
		return "IPv6"
	default:
		return fmt.Sprintf("AF(%d)", af)
	}
}

// Protocol represents a network protocol.
type Protocol uint8

const (
	ProtocolTCP  Protocol = 6
	ProtocolUDP  Protocol = 17
	ProtocolICMP Protocol = 1
)

// String returns the protocol name.
func (p Protocol) String() string {
	switch p {
	case ProtocolTCP:
		return "TCP"
	case ProtocolUDP:
		return "UDP"
	case ProtocolICMP:
		return "ICMP"
	default:
		return fmt.Sprintf("proto(%d)", p)
	}
}

// Direction represents the direction of network traffic.
type Direction uint8

const (
	// DirectionIngress represents incoming traffic.
	DirectionIngress Direction = iota

	// DirectionEgress represents outgoing traffic.
	DirectionEgress
)

// String returns the direction name.
func (d Direction) String() string {
	switch d {
	case DirectionIngress:
		return "ingress"
	case DirectionEgress:
		return "egress"
	default:
		return "unknown"
	}
}

// L4Addr represents a Layer 4 address (IP + port + protocol).
type L4Addr struct {
	Addr     netip.Addr
	Port     uint16
	Protocol Protocol
}

// String returns a human-readable representation.
func (a L4Addr) String() string {
	return fmt.Sprintf("%s:%d/%s", a.Addr, a.Port, a.Protocol)
}

// CIDR wraps netip.Prefix and provides utility methods.
type CIDR struct {
	netip.Prefix
}

// NewCIDR creates a CIDR from a netip.Prefix.
func NewCIDR(p netip.Prefix) CIDR {
	return CIDR{Prefix: p}
}

// ParseCIDR parses a CIDR string (e.g. "10.0.0.0/8").
func ParseCIDR(s string) (CIDR, error) {
	p, err := netip.ParsePrefix(s)
	if err != nil {
		return CIDR{}, fmt.Errorf("invalid CIDR %q: %w", s, err)
	}
	return NewCIDR(p), nil
}

// MustParseCIDR is like ParseCIDR but panics on error.
// Use only in tests and initialization.
func MustParseCIDR(s string) CIDR {
	c, err := ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return c
}

// Contains reports whether the CIDR contains the given address.
func (c CIDR) Contains(addr netip.Addr) bool {
	return c.Prefix.Contains(addr)
}

// IsIPv4 reports whether this is an IPv4 CIDR.
func (c CIDR) IsIPv4() bool {
	return c.Addr().Is4()
}

// IsIPv6 reports whether this is an IPv6 CIDR.
func (c CIDR) IsIPv6() bool {
	return c.Addr().Is6()
}

// IPRange represents a contiguous range of IP addresses.
type IPRange struct {
	Start netip.Addr
	End   netip.Addr
}

// PortRange represents a range of ports.
type PortRange struct {
	Start uint16
	End   uint16
}

// Contains reports whether the port range contains the given port.
func (pr PortRange) Contains(port uint16) bool {
	return port >= pr.Start && port <= pr.End
}
