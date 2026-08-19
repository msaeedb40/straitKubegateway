package ebpf

import (
	"fmt"
	"os"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// CgroupBPFManager manages cgroup-level socket and SKB enforcement programs.
type CgroupBPFManager struct {
	mu    sync.RWMutex
	links []link.Link
}

// NewCgroupBPFManager creates a new cgroup BPF manager.
func NewCgroupBPFManager() *CgroupBPFManager {
	return &CgroupBPFManager{
		links: make([]link.Link, 0),
	}
}

// Attach attaches a BPF program to a cgroup v2 directory path.
func (m *CgroupBPFManager) Attach(cgroupPath string, attachType ebpf.AttachType, prog *ebpf.Program) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	f, err := os.Open(cgroupPath)
	if err != nil {
		return fmt.Errorf("open cgroup path %q: %w", cgroupPath, err)
	}
	defer f.Close()

	l, err := link.AttachCgroup(link.CgroupOptions{
		Path:    cgroupPath,
		Attach:  attachType,
		Program: prog,
	})
	if err != nil {
		return fmt.Errorf("attach cgroup program (%v) to %q: %w", attachType, cgroupPath, err)
	}

	m.links = append(m.links, l)
	return nil
}

// Close detaches all managed cgroup programs.
func (m *CgroupBPFManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, l := range m.links {
		_ = l.Close()
	}
	m.links = nil
	return nil
}
