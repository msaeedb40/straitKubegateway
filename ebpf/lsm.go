package ebpf

import (
	"fmt"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// LSMManager manages Linux Security Module (LSM) BPF programs.
type LSMManager struct {
	mu    sync.RWMutex
	links []link.Link
}

// NewLSMManager creates a new LSM BPF manager.
func NewLSMManager() *LSMManager {
	return &LSMManager{
		links: make([]link.Link, 0),
	}
}

// AttachLSM attaches an LSM BPF program to a kernel security hook.
func (m *LSMManager) AttachLSM(prog *ebpf.Program) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	l, err := link.AttachLSM(link.LSMOptions{
		Program: prog,
	})
	if err != nil {
		return fmt.Errorf("attach LSM program: %w", err)
	}

	m.links = append(m.links, l)
	return nil
}

// Close detaches all attached LSM programs.
func (m *LSMManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, l := range m.links {
		_ = l.Close()
	}
	m.links = nil
	return nil
}
