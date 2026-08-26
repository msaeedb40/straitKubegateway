// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package platform implements Linux platform abstractions:
// cgroup v2 resource management, systemd integration, kernel capabilities,
// and process supervision for straitKubegateway.
package platform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

// ============================================================================
// cgroup v2
// ============================================================================

// CgroupV2Root is the default cgroup v2 mount point.
const CgroupV2Root = "/sys/fs/cgroup"

// CgroupManager manages cgroup v2 resource control for straitd.
type CgroupManager struct {
	log  *zap.Logger
	root string
	name string
}

// NewCgroupManager creates a new cgroup v2 manager.
func NewCgroupManager(name string, log *zap.Logger) *CgroupManager {
	return &CgroupManager{
		log:  log,
		root: CgroupV2Root,
		name: name,
	}
}

// Path returns the cgroup directory path.
func (m *CgroupManager) Path() string {
	return filepath.Join(m.root, m.name)
}

// Create creates the cgroup directory.
func (m *CgroupManager) Create() error {
	if err := os.MkdirAll(m.Path(), 0755); err != nil {
		return fmt.Errorf("create cgroup %q: %w", m.Path(), err)
	}
	m.log.Debug("cgroup created", zap.String("path", m.Path()))
	return nil
}

// SetCPUWeight sets the cpu.weight value (1-10000).
func (m *CgroupManager) SetCPUWeight(weight uint64) error {
	return m.writeControl("cpu.weight", strconv.FormatUint(weight, 10))
}

// SetMemoryMax sets the memory.max limit in bytes.
func (m *CgroupManager) SetMemoryMax(bytes uint64) error {
	return m.writeControl("memory.max", strconv.FormatUint(bytes, 10))
}

// SetMemoryHigh sets the memory.high soft limit in bytes.
func (m *CgroupManager) SetMemoryHigh(bytes uint64) error {
	return m.writeControl("memory.high", strconv.FormatUint(bytes, 10))
}

// SetIOMax sets io.max limits. Format: "MAJOR:MINOR rbps=N wbps=N riops=N wiops=N".
func (m *CgroupManager) SetIOMax(devMajorMinor string, rbps, wbps uint64) error {
	v := fmt.Sprintf("%s rbps=%d wbps=%d", devMajorMinor, rbps, wbps)
	return m.writeControl("io.max", v)
}

// AddPID adds a PID to the cgroup.
func (m *CgroupManager) AddPID(pid int) error {
	return m.writeControl("cgroup.procs", strconv.Itoa(pid))
}

// AddCurrentPID adds the current process to the cgroup.
func (m *CgroupManager) AddCurrentPID() error {
	return m.AddPID(os.Getpid())
}

// Delete removes the cgroup directory.
func (m *CgroupManager) Delete() error {
	return os.Remove(m.Path())
}

func (m *CgroupManager) writeControl(file, value string) error {
	path := filepath.Join(m.Path(), file)
	if err := os.WriteFile(path, []byte(value+"\n"), 0644); err != nil {
		return fmt.Errorf("cgroup %s/%s = %q: %w", m.name, file, value, err)
	}
	return nil
}

// IsCgroupV2 reports whether the system is using cgroup v2.
func IsCgroupV2() bool {
	var stat unix.Statfs_t
	if err := unix.Statfs(CgroupV2Root, &stat); err != nil {
		return false
	}
	// CGROUP2_SUPER_MAGIC = 0x63677270
	return stat.Type == 0x63677270
}

// ============================================================================
// systemd integration
// ============================================================================

// SystemdNotifier sends systemd service manager notifications.
type SystemdNotifier struct {
	socketPath string
	log        *zap.Logger
}

// NewSystemdNotifier creates a notifier using the NOTIFY_SOCKET env var.
func NewSystemdNotifier(log *zap.Logger) *SystemdNotifier {
	sock := os.Getenv("NOTIFY_SOCKET")
	return &SystemdNotifier{socketPath: sock, log: log}
}

// Ready sends the READY=1 notification to systemd.
func (n *SystemdNotifier) Ready() error {
	return n.send("READY=1")
}

// Stopping sends the STOPPING=1 notification to systemd.
func (n *SystemdNotifier) Stopping() error {
	return n.send("STOPPING=1")
}

// Status sends a STATUS=<msg> notification to systemd.
func (n *SystemdNotifier) Status(msg string) error {
	return n.send("STATUS=" + msg)
}

// Watchdog sends a WATCHDOG=1 keepalive to systemd.
func (n *SystemdNotifier) Watchdog() error {
	return n.send("WATCHDOG=1")
}

func (n *SystemdNotifier) send(msg string) error {
	if n.socketPath == "" {
		return nil // not running under systemd
	}
	conn, err := unix.Socket(unix.AF_UNIX, unix.SOCK_DGRAM, 0)
	if err != nil {
		return fmt.Errorf("systemd notify socket: %w", err)
	}
	defer unix.Close(conn)
	addr := &unix.SockaddrUnix{Name: n.socketPath}
	return unix.Sendto(conn, []byte(msg), 0, addr)
}

// ============================================================================
// Graceful shutdown
// ============================================================================

// GracefulStopper manages graceful shutdown for straitd.
type GracefulStopper struct {
	cancel context.CancelFunc
	log    *zap.Logger
}

// NewGracefulStopper wraps a context cancellation.
func NewGracefulStopper(cancel context.CancelFunc, log *zap.Logger) *GracefulStopper {
	return &GracefulStopper{cancel: cancel, log: log}
}

// Stop initiates a graceful shutdown.
func (g *GracefulStopper) Stop() {
	g.log.Info("initiating graceful shutdown")
	g.cancel()
}

// ============================================================================
// Kernel namespace helpers
// ============================================================================

// NetNS represents a Linux network namespace.
type NetNS struct {
	fd   int
	path string
}

// OpenNetNSByPath opens a network namespace from a path.
func OpenNetNSByPath(path string) (*NetNS, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("open netns %q: %w", path, err)
	}
	return &NetNS{fd: fd, path: path}, nil
}

// OpenNetNSByPID opens the network namespace of a process.
func OpenNetNSByPID(pid int) (*NetNS, error) {
	return OpenNetNSByPath(fmt.Sprintf("/proc/%d/ns/net", pid))
}

// FD returns the netns file descriptor.
func (n *NetNS) FD() int { return n.fd }

// Path returns the netns path.
func (n *NetNS) Path() string { return n.path }

// Close closes the netns file descriptor.
func (n *NetNS) Close() error {
	if n.fd < 0 {
		return nil
	}
	err := unix.Close(n.fd)
	n.fd = -1
	return err
}

// Do executes fn in the network namespace, then restores the original.
func (n *NetNS) Do(fn func() error) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Get the current netns
	origFd, err := unix.Open("/proc/self/ns/net", unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("open current netns: %w", err)
	}
	defer unix.Close(origFd)

	// Enter the target netns
	if err := unix.Setns(n.fd, unix.CLONE_NEWNET); err != nil {
		return fmt.Errorf("enter netns %q: %w", n.path, err)
	}

	// Run the function
	fnErr := fn()

	// Restore original netns
	if err := unix.Setns(origFd, unix.CLONE_NEWNET); err != nil {
		return fmt.Errorf("restore netns: %w", err)
	}
	return fnErr
}

// ============================================================================
// sysctl helpers
// ============================================================================

// NetworkSysctls sets the required network sysctl parameters for straitd.
func NetworkSysctls() error {
	settings := map[string]string{
		"net.ipv4.ip_forward":                "1",
		"net.ipv6.conf.all.forwarding":       "1",
		"net.ipv4.conf.all.rp_filter":        "0",
		"net.ipv4.conf.default.rp_filter":    "0",
		"net.bridge.bridge-nf-call-iptables": "0",
		"net.ipv4.conf.all.arp_announce":     "2",
		"net.ipv4.conf.all.arp_ignore":       "1",
		"net.core.bpf_jit_enable":            "1",
	}
	for key, val := range settings {
		path := "/proc/sys/" + strings.ReplaceAll(key, ".", "/")
		if err := os.WriteFile(path, []byte(val+"\n"), 0644); err != nil {
			// Non-fatal: some sysctls may not exist on all kernels
			continue
		}
	}
	return nil
}
