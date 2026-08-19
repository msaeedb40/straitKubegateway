package ebpf

import (
	"fmt"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// XDPManager manages XDP program lifecycle and interface attachment.
type XDPManager struct {
	mu    sync.RWMutex
	links map[int]link.Link // ifindex -> attached XDP link
}

// NewXDPManager creates a new XDP manager instance.
func NewXDPManager() *XDPManager {
	return &XDPManager{
		links: make(map[int]link.Link),
	}
}

// AttachXDP attaches an XDP program to the specified network interface.
func (m *XDPManager) AttachXDP(ifindex int, prog *ebpf.Program, flags link.XDPAttachFlags) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.links[ifindex]; exists {
		return fmt.Errorf("XDP program already attached to ifindex %d", ifindex)
	}

	l, err := link.AttachXDP(link.XDPOptions{
		Program:   prog,
		Interface: ifindex,
		Flags:     flags,
	})
	if err != nil {
		return fmt.Errorf("attach XDP to ifindex %d: %w", ifindex, err)
	}

	m.links[ifindex] = l
	return nil
}

// DetachXDP detaches the XDP program from the specified interface.
func (m *XDPManager) DetachXDP(ifindex int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, exists := m.links[ifindex]
	if !exists {
		return nil
	}

	if err := l.Close(); err != nil {
		return fmt.Errorf("detach XDP from ifindex %d: %w", ifindex, err)
	}

	delete(m.links, ifindex)
	return nil
}

// Close detaches all managed XDP links.
func (m *XDPManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for ifindex, l := range m.links {
		_ = l.Close()
		delete(m.links, ifindex)
	}
	return nil
}
