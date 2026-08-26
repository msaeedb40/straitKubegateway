// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package net provides network utilities used across straitKubegateway.
package net

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// ============================================================================
// MTU
// ============================================================================

// GetMTU returns the MTU of a network interface by name.
func GetMTU(ifaceName string) (int, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return 0, fmt.Errorf("interface %q not found: %w", ifaceName, err)
	}
	return iface.MTU, nil
}

// GetDefaultMTU returns the MTU of the interface that has the default route.
// Falls back to 1500 if detection fails.
func GetDefaultMTU() int {
	iface, err := defaultRouteInterface()
	if err != nil {
		return 1500
	}
	mtu, err := GetMTU(iface)
	if err != nil {
		return 1500
	}
	return mtu
}

// OverlayMTU computes the effective MTU for an overlay network given the
// host MTU and the encapsulation overhead in bytes.
func OverlayMTU(hostMTU, overhead int) int {
	result := hostMTU - overhead
	if result < 576 {
		result = 576
	}
	return result
}

// VXLANOverhead is the byte overhead added by VXLAN encapsulation.
const VXLANOverhead = 50

// GeneveOverhead is the byte overhead added by Geneve encapsulation.
const GeneveOverhead = 50

// GREOverhead is the byte overhead added by GRE encapsulation.
const GREOverhead = 20

// WireGuardOverhead is the byte overhead added by WireGuard encapsulation.
const WireGuardOverhead = 80

// defaultRouteInterface returns the interface name of the default route.
func defaultRouteInterface() (string, error) {
	data, err := os.ReadFile("/proc/net/route")
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines[1:] {
		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}
		dest := fields[1]
		flags, _ := strconv.ParseInt(fields[3], 16, 64)
		// RTF_UP | RTF_GATEWAY
		if dest == "00000000" && flags&0x3 == 0x3 {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("default route not found")
}

// ============================================================================
// Prefix utilities
// ============================================================================

// Contains reports whether the prefix contains the given IP address.
func Contains(prefix netip.Prefix, addr netip.Addr) bool {
	return prefix.Contains(addr)
}

// Overlaps reports whether two prefixes overlap.
func Overlaps(a, b netip.Prefix) bool {
	return a.Overlaps(b)
}

// IsIPv4 reports whether a prefix is IPv4.
func IsIPv4(p netip.Prefix) bool { return p.Addr().Is4() }

// IsIPv6 reports whether a prefix is IPv6.
func IsIPv6(p netip.Prefix) bool { return p.Addr().Is6() }

// PrefixToNetIPNet converts netip.Prefix to *net.IPNet.
func PrefixToNetIPNet(p netip.Prefix) *net.IPNet {
	ip := p.Addr().AsSlice()
	bits := p.Bits()
	total := p.Addr().BitLen()
	return &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, total)}
}

// NetIPNetToPrefix converts *net.IPNet to netip.Prefix.
func NetIPNetToPrefix(n *net.IPNet) (netip.Prefix, error) {
	addr, ok := netip.AddrFromSlice(n.IP)
	if !ok {
		return netip.Prefix{}, fmt.Errorf("invalid IP in net.IPNet")
	}
	ones, _ := n.Mask.Size()
	return netip.PrefixFrom(addr.Unmap(), ones), nil
}

// ============================================================================
// Interface helpers
// ============================================================================

// InterfaceExists reports whether a network interface with the given name exists.
func InterfaceExists(name string) bool {
	_, err := net.InterfaceByName(name)
	return err == nil
}

// GetInterfaceAddrs returns the IP addresses assigned to an interface.
func GetInterfaceAddrs(name string) ([]netip.Prefix, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("interface %q: %w", name, err)
	}
	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("addrs for %q: %w", name, err)
	}
	var prefixes []netip.Prefix
	for _, a := range addrs {
		var ipnet *net.IPNet
		switch v := a.(type) {
		case *net.IPNet:
			ipnet = v
		case *net.IPAddr:
			ipnet = &net.IPNet{IP: v.IP, Mask: net.CIDRMask(128, 128)}
		}
		if ipnet == nil {
			continue
		}
		p, err := NetIPNetToPrefix(ipnet)
		if err != nil {
			continue
		}
		prefixes = append(prefixes, p)
	}
	return prefixes, nil
}

// GetInterfaceMAC returns the hardware MAC address of an interface.
func GetInterfaceMAC(name string) (net.HardwareAddr, error) {
	iface, err := net.InterfaceByName(name)
	if err != nil {
		return nil, fmt.Errorf("interface %q: %w", name, err)
	}
	return iface.HardwareAddr, nil
}

// ============================================================================
// Socket helpers
// ============================================================================

// SetSocketReuseAddr sets SO_REUSEADDR on a socket file descriptor.
func SetSocketReuseAddr(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
}

// SetSocketReusePort sets SO_REUSEPORT on a socket file descriptor.
func SetSocketReusePort(fd int) error {
	return unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
}

// ============================================================================
// IP range helpers
// ============================================================================

// IPInRange reports whether ip falls within [start, end] inclusive.
func IPInRange(ip, start, end net.IP) bool {
	if ip.To4() != nil {
		ip = ip.To4()
		start = start.To4()
		end = end.To4()
	}
	for i := range ip {
		if i >= len(start) || i >= len(end) {
			break
		}
		if ip[i] < start[i] || ip[i] > end[i] {
			return false
		}
		if ip[i] > start[i] && ip[i] < end[i] {
			return true
		}
	}
	return true
}

// FirstIP returns the first usable host IP in a prefix.
func FirstIP(prefix netip.Prefix) netip.Addr {
	return prefix.Masked().Addr().Next()
}

// LastIP returns the last IP in a prefix (broadcast for IPv4).
func LastIP(prefix netip.Prefix) netip.Addr {
	a := prefix.Masked().Addr()
	bits := a.BitLen()
	prefixBits := prefix.Bits()
	hostBits := bits - prefixBits
	b := a.As16()
	for i := bits - 1; i >= bits-hostBits; i-- {
		byteIdx := (bits - 1 - i) / 8
		bitIdx := uint(i % 8)
		b[15-byteIdx] |= 1 << bitIdx
	}
	last, _ := netip.AddrFromSlice(b[:])
	return last.Unmap()
}
