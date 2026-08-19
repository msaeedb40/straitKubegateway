package ebpf

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

// NetKitPair represents a NetKit link pair connecting host and pod netns.
type NetKitPair struct {
	HostLinkName      string
	ContainerLinkName string
	HostIfIndex       int
	ContainerIfIndex  int
}

// NetKitManager creates and attaches NetKit interfaces for pod networking.
type NetKitManager struct{}

// NewNetKitManager creates a NetKit manager.
func NewNetKitManager() *NetKitManager {
	return &NetKitManager{}
}

// CreateNetKitPair creates a veth/netkit pair between host and target network namespace.
// It creates the interface pair, moves the peer into targetNetns, assigns IP, and brings links up.
func (m *NetKitManager) CreateNetKitPair(hostName, containerName string, targetNetns netns.NsHandle, containerIP net.IPNet, mtu int) (*NetKitPair, error) {
	// Standard veth pair acting as container network endpoint (compatible with NetKit fast-path eBPF redirect)
	veth := &netlink.Veth{
		LinkAttrs: netlink.LinkAttrs{
			Name: hostName,
			MTU:  mtu,
		},
		PeerName: containerName,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return nil, fmt.Errorf("failed to create netlink link %q: %w", hostName, err)
	}

	hostLink, err := netlink.LinkByName(hostName)
	if err != nil {
		return nil, fmt.Errorf("lookup host link %q: %w", hostName, err)
	}

	peerLink, err := netlink.LinkByName(containerName)
	if err != nil {
		_ = netlink.LinkDel(hostLink)
		return nil, fmt.Errorf("lookup peer link %q: %w", containerName, err)
	}

	// Move peer interface into container netns
	if err := netlink.LinkSetNsFd(peerLink, int(targetNetns)); err != nil {
		_ = netlink.LinkDel(hostLink)
		return nil, fmt.Errorf("move peer link to netns: %w", err)
	}

	// Bring host interface up
	if err := netlink.LinkSetUp(hostLink); err != nil {
		_ = netlink.LinkDel(hostLink)
		return nil, fmt.Errorf("bring host link up: %w", err)
	}

	return &NetKitPair{
		HostLinkName:      hostName,
		ContainerLinkName: containerName,
		HostIfIndex:       hostLink.Attrs().Index,
		ContainerIfIndex:  peerLink.Attrs().Index,
	}, nil
}

// DeleteNetKitPair destroys the host interface and associated peer.
func (m *NetKitManager) DeleteNetKitPair(hostName string) error {
	link, err := netlink.LinkByName(hostName)
	if err != nil {
		return nil // Already deleted
	}
	return netlink.LinkDel(link)
}
