package ebpf

import (
	"fmt"
	"os"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// SockOpsManager manages socket-level acceleration programs attached to cgroups.
type SockOpsManager struct {
	mu     sync.RWMutex
	cgroup link.Link
}

// NewSockOpsManager creates a new sockops manager.
func NewSockOpsManager() *SockOpsManager {
	return &SockOpsManager{}
}

// AttachCgroup attaches a sockops program to a cgroup v2 path.
func (m *SockOpsManager) AttachCgroup(cgroupPath string, prog *ebpf.Program) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	f, err := os.Open(cgroupPath)
	if err != nil {
		return fmt.Errorf("open cgroup path %q: %w", cgroupPath, err)
	}
	defer f.Close()

	l, err := link.AttachCgroup(link.CgroupOptions{
		Path:    cgroupPath,
		Attach:  ebpf.AttachCGroupSockOps,
		Program: prog,
	})
	if err != nil {
		return fmt.Errorf("attach sockops to cgroup %q: %w", cgroupPath, err)
	}

	m.cgroup = l
	return nil
}

// Close detaches the sockops cgroup program.
func (m *SockOpsManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.cgroup != nil {
		err := m.cgroup.Close()
		m.cgroup = nil
		return err
	}
	return nil
}
