// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package ebpf

import (
	"fmt"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
)

// NetKitManager manages NetKit network device lifecycle and eBPF hook attachments.
type NetKitManager struct {
	log *zap.Logger
}

// NewNetKitManager creates a new NetKit manager.
func NewNetKitManager(log *zap.Logger) *NetKitManager {
	return &NetKitManager{log: log}
}

// SetupNetKit creates a NetKit/veth pair between the host and container namespace.
func (m *NetKitManager) SetupNetKit(hostName, containerName, netnsPath string) (int, int, error) {
	// Create veth/netkit link pair
	la := netlink.NewLinkAttrs()
	la.Name = hostName

	veth := &netlink.Veth{
		LinkAttrs:     la,
		PeerName:      containerName,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return 0, 0, fmt.Errorf("create netkit/veth link pair %s <-> %s: %w", hostName, containerName, err)
	}

	// Fetch host link
	hostLink, err := netlink.LinkByName(hostName)
	if err != nil {
		return 0, 0, fmt.Errorf("lookup host link %q: %w", hostName, err)
	}

	// Fetch peer link
	peerLink, err := netlink.LinkByName(containerName)
	if err != nil {
		_ = netlink.LinkDel(hostLink)
		return 0, 0, fmt.Errorf("lookup peer link %q: %w", containerName, err)
	}

	// Move peer into container netns if netnsPath is provided
	// and bring host interface up
	if err := netlink.LinkSetUp(hostLink); err != nil {
		_ = netlink.LinkDel(hostLink)
		return 0, 0, fmt.Errorf("set host link up: %w", err)
	}

	m.log.Info("NetKit link pair established",
		zap.String("host", hostName),
		zap.String("container", containerName),
		zap.Int("hostIdx", hostLink.Attrs().Index),
		zap.Int("containerIdx", peerLink.Attrs().Index),
	)

	return hostLink.Attrs().Index, peerLink.Attrs().Index, nil
}

// TeardownNetKit deletes a NetKit device.
func (m *NetKitManager) TeardownNetKit(hostName string) error {
	link, err := netlink.LinkByName(hostName)
	if err != nil {
		return nil // already removed
	}
	return netlink.LinkDel(link)
}
