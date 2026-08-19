// Package resource provides cgroup v2 resource management (CPU, memory, I/O)
// for straitd node runtime.
package resource

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	CgroupV2MountPoint = "/sys/fs/cgroup"
	DefaultCgroupPath  = "/sys/fs/cgroup/straitkubegateway"
)

// Manager manages cgroup v2 resource limits and process containment.
type Manager struct {
	basePath string
}

// NewManager creates a cgroup v2 manager.
func NewManager(path string) *Manager {
	if path == "" {
		path = DefaultCgroupPath
	}
	return &Manager{basePath: path}
}

// IsCgroupV2Available checks whether cgroup v2 is mounted on the host.
func IsCgroupV2Available() bool {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return false
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == CgroupV2MountPoint && fields[2] == "cgroup2" {
			return true
		}
	}
	return false
}

// EnsureCgroup creates the cgroup directory if it does not exist.
func (m *Manager) EnsureCgroup() error {
	if err := os.MkdirAll(m.basePath, 0755); err != nil {
		return fmt.Errorf("create cgroup %q: %w", m.basePath, err)
	}
	return nil
}

// SetCPULimit sets the cpu.max quota and period in microseconds (e.g., quota=100000, period=100000 for 1 core).
func (m *Manager) SetCPULimit(quotaUs, periodUs int64) error {
	val := fmt.Sprintf("%d %d", quotaUs, periodUs)
	if quotaUs <= 0 {
		val = "max 100000"
	}
	return os.WriteFile(filepath.Join(m.basePath, "cpu.max"), []byte(val), 0644)
}

// SetMemoryLimit sets memory.max in bytes.
func (m *Manager) SetMemoryLimit(bytes int64) error {
	val := fmt.Sprintf("%d", bytes)
	if bytes <= 0 {
		val = "max"
	}
	return os.WriteFile(filepath.Join(m.basePath, "memory.max"), []byte(val), 0644)
}

// SetMemoryHigh sets memory.high throttle watermark in bytes.
func (m *Manager) SetMemoryHigh(bytes int64) error {
	val := fmt.Sprintf("%d", bytes)
	if bytes <= 0 {
		val = "max"
	}
	return os.WriteFile(filepath.Join(m.basePath, "memory.high"), []byte(val), 0644)
}

// AddProcess adds a PID to the cgroup.procs file.
func (m *Manager) AddProcess(pid int) error {
	return os.WriteFile(filepath.Join(m.basePath, "cgroup.procs"), []byte(fmt.Sprintf("%d\n", pid)), 0644)
}

// GetMemoryUsage reads current memory usage in bytes.
func (m *Manager) GetMemoryUsage() (int64, error) {
	data, err := os.ReadFile(filepath.Join(m.basePath, "memory.current"))
	if err != nil {
		return 0, err
	}
	var usage int64
	_, err = fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &usage)
	return usage, err
}
