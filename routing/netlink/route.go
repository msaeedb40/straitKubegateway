package netlinkroute

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/vishvananda/netlink"
)

// SyncRoute installs or replaces a kernel route via netlink.
func SyncRoute(dst netip.Prefix, gw netip.Addr, linkIndex int) error {
	_, ipnet, err := net.ParseCIDR(dst.String())
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}

	route := &netlink.Route{
		LinkIndex: linkIndex,
		Dst:       ipnet,
	}

	if gw.IsValid() && !gw.IsUnspecified() {
		route.Gw = net.ParseIP(gw.String())
	}

	if err := netlink.RouteReplace(route); err != nil {
		return fmt.Errorf("failed to install netlink route to %s: %w", dst, err)
	}
	return nil
}

// RemoveRoute deletes a kernel route via netlink.
func RemoveRoute(dst netip.Prefix, linkIndex int) error {
	_, ipnet, err := net.ParseCIDR(dst.String())
	if err != nil {
		return fmt.Errorf("invalid CIDR: %w", err)
	}

	route := &netlink.Route{
		LinkIndex: linkIndex,
		Dst:       ipnet,
	}

	return netlink.RouteDel(route)
}
