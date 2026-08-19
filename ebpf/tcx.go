package ebpf

import (
	"fmt"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// TCXManager manages TCX ingress and egress program attachment.
type TCXManager struct {
	mu           sync.RWMutex
	ingressLinks map[int]link.Link // ifindex -> attached TCX link
	egressLinks  map[int]link.Link // ifindex -> attached TCX link
}

// NewTCXManager creates a new TCX manager instance.
func NewTCXManager() *TCXManager {
	return &TCXManager{
		ingressLinks: make(map[int]link.Link),
		egressLinks:  make(map[int]link.Link),
	}
}

// AttachTCXIngress attaches a TCX program to network interface ingress.
func (m *TCXManager) AttachTCXIngress(ifindex int, prog *ebpf.Program) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := link.AttachTCX(link.TCXOptions{
		Program:   prog,
		Attach:    ebpf.AttachTCXIngress,
		Interface: ifindex,
	})
	if err != nil {
		return fmt.Errorf("attach TCX ingress to ifindex %d: %w", ifindex, err)
	}

	m.ingressLinks[ifindex] = l
	return nil
}

// AttachTCXEgress attaches a TCX program to network interface egress.
func (m *TCXManager) AttachTCXEgress(ifindex int, prog *ebpf.Program) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := link.AttachTCX(link.TCXOptions{
		Program:   prog,
		Attach:    ebpf.AttachTCXEgress,
		Interface: ifindex,
	})
	if err != nil {
		return fmt.Errorf("attach TCX egress to ifindex %d: %w", ifindex, err)
	}

	m.egressLinks[ifindex] = l
	return nil
}

// DetachTCX detaches TCX programs from the given interface.
func (m *TCXManager) DetachTCX(ifindex int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if l, ok := m.ingressLinks[ifindex]; ok {
		_ = l.Close()
		delete(m.ingressLinks, ifindex)
	}
	if l, ok := m.egressLinks[ifindex]; ok {
		_ = l.Close()
		delete(m.egressLinks, ifindex)
	}
	return nil
}

// Close releases all attached TCX links.
func (m *TCXManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for ifindex, l := range m.ingressLinks {
		_ = l.Close()
		delete(m.ingressLinks, ifindex)
	}
	for ifindex, l := range m.egressLinks {
		_ = l.Close()
		delete(m.egressLinks, ifindex)
	}
	return nil
}
