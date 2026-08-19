package nat

import (
	"fmt"
	"net/netip"
	"sync"
)

const (
	// DefaultNAT64Prefix is the RFC 6052 well-known prefix "64:ff9b::/96".
	DefaultNAT64Prefix = "64:ff9b::/96"
)

// NAT64Engine handles stateful translation between IPv6-only clients and IPv4 servers.
type NAT64Engine struct {
	mu          sync.RWMutex
	nat64Prefix netip.Prefix
	poolIPv4    netip.Prefix
	nextPort    uint16
}

// NewNAT64Engine creates a NAT64 translation engine.
func NewNAT64Engine(prefixStr, ipv4PoolStr string) (*NAT64Engine, error) {
	if prefixStr == "" {
		prefixStr = DefaultNAT64Prefix
	}
	prefix, err := netip.ParsePrefix(prefixStr)
	if err != nil {
		return nil, fmt.Errorf("invalid NAT64 prefix %q: %w", prefixStr, err)
	}

	if ipv4PoolStr == "" {
		ipv4PoolStr = "192.0.2.0/24" // RFC 5737 TEST-NET-1 or custom pool
	}
	pool4, err := netip.ParsePrefix(ipv4PoolStr)
	if err != nil {
		return nil, fmt.Errorf("invalid IPv4 pool %q: %w", ipv4PoolStr, err)
	}

	return &NAT64Engine{
		nat64Prefix: prefix,
		poolIPv4:    pool4,
		nextPort:    40000,
	}, nil
}

// Prefix returns the active NAT64 IPv6 prefix.
func (n *NAT64Engine) Prefix() netip.Prefix {
	return n.nat64Prefix
}

// IsNAT64Address checks if an IPv6 destination address belongs to the NAT64 prefix.
func (n *NAT64Engine) IsNAT64Address(addr netip.Addr) bool {
	if !addr.Is6() {
		return false
	}
	return n.nat64Prefix.Contains(addr)
}

// ExtractIPv4 extracts the embedded IPv4 address from the last 32 bits of a /96 NAT64 IPv6 address.
func (n *NAT64Engine) ExtractIPv4(addr6 netip.Addr) (netip.Addr, error) {
	if !n.IsNAT64Address(addr6) {
		return netip.Addr{}, fmt.Errorf("address %s does not match NAT64 prefix %s", addr6, n.nat64Prefix)
	}

	b := addr6.As16()
	// Extract bytes 12..15 (last 4 bytes for /96 prefix)
	ip4 := netip.AddrFrom4([4]byte{b[12], b[13], b[14], b[15]})
	return ip4, nil
}

// SynthesizeIPv6 creates a synthetic IPv6 address embedding an IPv4 address into the /96 prefix.
func (n *NAT64Engine) SynthesizeIPv6(addr4 netip.Addr) (netip.Addr, error) {
	if !addr4.Is4() {
		return netip.Addr{}, fmt.Errorf("expected IPv4 address, got %s", addr4)
	}

	prefixBytes := n.nat64Prefix.Addr().As16()
	ip4Bytes := addr4.As4()

	var result [16]byte
	copy(result[0:12], prefixBytes[0:12])
	copy(result[12:16], ip4Bytes[0:4])

	return netip.AddrFrom16(result), nil
}
