// Package netutil provides network utility functions for IP address manipulation,
// netlink operations, and CIDR arithmetic used throughout straitKubegateway.
package netutil

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// InterfaceByIndex returns the network interface with the given index.
func InterfaceByIndex(index int) (*net.Interface, error) {
	return net.InterfaceByIndex(index)
}

// InterfaceByName returns the network interface with the given name.
func InterfaceByName(name string) (*net.Interface, error) {
	return net.InterfaceByName(name)
}

// GetLinkByName returns the netlink.Link for the named interface.
func GetLinkByName(name string) (netlink.Link, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, fmt.Errorf("link %q not found: %w", name, err)
	}
	return link, nil
}

// GetLinkByIndex returns the netlink.Link for the given interface index.
func GetLinkByIndex(index int) (netlink.Link, error) {
	link, err := netlink.LinkByIndex(index)
	if err != nil {
		return nil, fmt.Errorf("link index %d not found: %w", index, err)
	}
	return link, nil
}

// SetLinkUp brings a network interface up.
func SetLinkUp(link netlink.Link) error {
	return netlink.LinkSetUp(link)
}

// SetLinkDown brings a network interface down.
func SetLinkDown(link netlink.Link) error {
	return netlink.LinkSetDown(link)
}

// SetLinkNsFd moves a link to the network namespace represented by the file descriptor.
func SetLinkNsFd(link netlink.Link, fd int) error {
	return netlink.LinkSetNsFd(link, fd)
}

// AddRoute adds a route to the kernel routing table.
func AddRoute(route *netlink.Route) error {
	return netlink.RouteAdd(route)
}

// DeleteRoute removes a route from the kernel routing table.
func DeleteRoute(route *netlink.Route) error {
	return netlink.RouteDel(route)
}

// ReplaceRoute adds or replaces a route in the kernel routing table.
func ReplaceRoute(route *netlink.Route) error {
	return netlink.RouteReplace(route)
}

// ListRoutes lists routes matching the given filter.
func ListRoutes(link netlink.Link, family int) ([]netlink.Route, error) {
	return netlink.RouteList(link, family)
}

// AddAddress adds an IP address to a network interface.
func AddAddress(link netlink.Link, addr *netlink.Addr) error {
	return netlink.AddrAdd(link, addr)
}

// DeleteAddress removes an IP address from a network interface.
func DeleteAddress(link netlink.Link, addr *netlink.Addr) error {
	return netlink.AddrDel(link, addr)
}

// ListAddresses lists addresses on a network interface.
func ListAddresses(link netlink.Link, family int) ([]netlink.Addr, error) {
	return netlink.AddrList(link, family)
}

// GetCurrentNetNS returns the current network namespace.
func GetCurrentNetNS() (netns.NsHandle, error) {
	return netns.Get()
}

// GetNetNSByPath opens a network namespace by its filesystem path.
func GetNetNSByPath(path string) (netns.NsHandle, error) {
	return netns.GetFromPath(path)
}

// SetNetNS sets the current goroutine's network namespace.
func SetNetNS(ns netns.NsHandle) error {
	return netns.Set(ns)
}

// GetMTU returns the MTU of the named interface.
func GetMTU(name string) (int, error) {
	link, err := GetLinkByName(name)
	if err != nil {
		return 0, err
	}
	return link.Attrs().MTU, nil
}

// DetectMTU auto-detects the best MTU for the given interface, accounting
// for encapsulation overhead.
func DetectMTU(name string, encapOverhead int) (int, error) {
	mtu, err := GetMTU(name)
	if err != nil {
		return 0, err
	}
	return mtu - encapOverhead, nil
}

// FirstIPInPrefix returns the first usable IP in a prefix (skips network address).
func FirstIPInPrefix(prefix netip.Prefix) netip.Addr {
	addr := prefix.Addr()
	return addr.Next()
}

// PrefixContains checks whether a prefix contains a given address.
func PrefixContains(prefix netip.Prefix, addr netip.Addr) bool {
	return prefix.Contains(addr)
}

// IsIPv4 reports whether the address is IPv4.
func IsIPv4(addr netip.Addr) bool {
	return addr.Is4()
}

// IsIPv6 reports whether the address is IPv6.
func IsIPv6(addr netip.Addr) bool {
	return addr.Is6()
}

// SetSysctl writes a value to a sysctl path.
func SetSysctl(path, value string) error {
	sysctlPath := filepath.Join("/proc/sys", strings.ReplaceAll(path, ".", "/"))
	return os.WriteFile(sysctlPath, []byte(value), 0644)
}
