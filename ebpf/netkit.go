// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package ebpf

import (
	"fmt"
	"os"

	"github.com/vishvananda/netlink"
	"go.uber.org/zap"
	"golang.org/x/sys/unix"
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
//
// The host-side interface (hostName) remains in the host network namespace.
// The container-side interface (containerName) is moved into the network namespace
// at netnsPath, making it visible inside the container.
//
// Bug fix (v1): previously the peer was not moved into the container netns.
func (m *NetKitManager) SetupNetKit(hostName, containerName, netnsPath string) (int, int, error) {
	// Create veth/netkit link pair in the host namespace
	la := netlink.NewLinkAttrs()
	la.Name = hostName

	veth := &netlink.Veth{
		LinkAttrs: la,
		PeerName:  containerName,
	}

	if err := netlink.LinkAdd(veth); err != nil {
		return 0, 0, fmt.Errorf("create netkit/veth link pair %s <-> %s: %w", hostName, containerName, err)
	}

	// Fetch host-side link
	hostLink, err := netlink.LinkByName(hostName)
	if err != nil {
		return 0, 0, fmt.Errorf("lookup host link %q: %w", hostName, err)
	}

	// Fetch peer (container-side) link
	peerLink, err := netlink.LinkByName(containerName)
	if err != nil {
		_ = netlink.LinkDel(hostLink)
		return 0, 0, fmt.Errorf("lookup peer link %q: %w", containerName, err)
	}

	// Move peer into container network namespace if netnsPath is provided.
	// This is required so the interface appears as "eth0" (or similar) inside the pod.
	if netnsPath != "" {
		nsFD, err := openNetNS(netnsPath)
		if err != nil {
			_ = netlink.LinkDel(hostLink)
			return 0, 0, fmt.Errorf("open container netns %q: %w", netnsPath, err)
		}
		defer unix.Close(nsFD)

		if err := netlink.LinkSetNsFd(peerLink, nsFD); err != nil {
			_ = netlink.LinkDel(hostLink)
			return 0, 0, fmt.Errorf("move peer %q into netns %q: %w", containerName, netnsPath, err)
		}
	}

	// Bring host-side interface up
	if err := netlink.LinkSetUp(hostLink); err != nil {
		_ = netlink.LinkDel(hostLink)
		return 0, 0, fmt.Errorf("set host link %q up: %w", hostName, err)
	}

	hostIdx := hostLink.Attrs().Index
	peerIdx := peerLink.Attrs().Index

	m.log.Info("NetKit link pair established",
		zap.String("host", hostName),
		zap.String("container", containerName),
		zap.String("netns", netnsPath),
		zap.Int("hostIdx", hostIdx),
		zap.Int("containerIdx", peerIdx),
	)

	return hostIdx, peerIdx, nil
}

// TeardownNetKit deletes a NetKit/veth device by host-side interface name.
// Idempotent: if the link no longer exists, returns nil.
func (m *NetKitManager) TeardownNetKit(hostName string) error {
	link, err := netlink.LinkByName(hostName)
	if err != nil {
		// Link not found — already deleted or never created
		return nil
	}
	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("delete netkit link %q: %w", hostName, err)
	}
	m.log.Debug("NetKit link torn down", zap.String("host", hostName))
	return nil
}

// openNetNS opens a network namespace file descriptor from a path.
// The caller is responsible for closing the returned fd with unix.Close.
func openNetNS(path string) (int, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return -1, fmt.Errorf("open netns %q: %w", path, err)
	}
	fd := int(f.Fd())
	// Detach from the *os.File so that Close() on the file doesn't close the fd —
	// the caller manages the fd lifetime explicitly.
	// We duplicate so that os.File.Close() doesn't close our fd.
	newFD, err := unix.Dup(fd)
	f.Close()
	if err != nil {
		return -1, fmt.Errorf("dup netns fd: %w", err)
	}
	return newFD, nil
}
