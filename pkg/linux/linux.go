// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package linux provides Linux kernel interaction helpers including
// sysctl, capabilities, kernel version parsing, and netlink utilities.
package linux

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

// ============================================================================
// Kernel version
// ============================================================================

// KernelVersion represents a parsed Linux kernel version.
type KernelVersion struct {
	Major int
	Minor int
	Patch int
}

// String returns the dotted kernel version string.
func (k KernelVersion) String() string {
	return fmt.Sprintf("%d.%d.%d", k.Major, k.Minor, k.Patch)
}

// IsAtLeast reports whether this kernel version is >= the given version.
func (k KernelVersion) IsAtLeast(major, minor, patch int) bool {
	if k.Major != major {
		return k.Major > major
	}
	if k.Minor != minor {
		return k.Minor > minor
	}
	return k.Patch >= patch
}

// GetKernelVersion returns the running kernel version.
func GetKernelVersion() (KernelVersion, error) {
	var uts unix.Utsname
	if err := unix.Uname(&uts); err != nil {
		return KernelVersion{}, fmt.Errorf("uname: %w", err)
	}
	// Convert int8 array to string
	var buf []byte
	for _, c := range uts.Release {
		if c == 0 {
			break
		}
		buf = append(buf, byte(c))
	}
	return parseKernelVersion(string(buf))
}

func parseKernelVersion(s string) (KernelVersion, error) {
	// Strip anything after the first non-version character (e.g. "-generic")
	dashIdx := strings.IndexAny(s, "-+~")
	if dashIdx > 0 {
		s = s[:dashIdx]
	}
	parts := strings.SplitN(s, ".", 3)
	if len(parts) < 2 {
		return KernelVersion{}, fmt.Errorf("cannot parse kernel version %q", s)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return KernelVersion{}, fmt.Errorf("major: %w", err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return KernelVersion{}, fmt.Errorf("minor: %w", err)
	}
	patch := 0
	if len(parts) == 3 {
		patch, _ = strconv.Atoi(parts[2])
	}
	return KernelVersion{Major: major, Minor: minor, Patch: patch}, nil
}

// MinimumKernelVersion is the minimum supported kernel (6.7).
var MinimumKernelVersion = KernelVersion{Major: 6, Minor: 7, Patch: 0}

// ProductionKernelVersion is the production baseline kernel (6.12 LTS).
var ProductionKernelVersion = KernelVersion{Major: 6, Minor: 12, Patch: 0}

// CheckKernelVersion verifies the running kernel meets the minimum requirement.
func CheckKernelVersion() error {
	kv, err := GetKernelVersion()
	if err != nil {
		return fmt.Errorf("kernel version check: %w", err)
	}
	if !kv.IsAtLeast(MinimumKernelVersion.Major, MinimumKernelVersion.Minor, MinimumKernelVersion.Patch) {
		return fmt.Errorf("kernel %s is below minimum required %s", kv, MinimumKernelVersion)
	}
	return nil
}

// ============================================================================
// Sysctl
// ============================================================================

// GetSysctl reads a sysctl value from /proc/sys.
func GetSysctl(name string) (string, error) {
	path := "/proc/sys/" + strings.ReplaceAll(name, ".", "/")
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("sysctl %q: %w", name, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// SetSysctl writes a value to a sysctl parameter.
func SetSysctl(name, value string) error {
	path := "/proc/sys/" + strings.ReplaceAll(name, ".", "/")
	if err := os.WriteFile(path, []byte(value+"\n"), 0644); err != nil {
		return fmt.Errorf("sysctl %q = %q: %w", name, value, err)
	}
	return nil
}

// SetIPv4Forwarding enables IPv4 packet forwarding.
func SetIPv4Forwarding() error {
	return SetSysctl("net.ipv4.ip_forward", "1")
}

// SetIPv6Forwarding enables IPv6 packet forwarding.
func SetIPv6Forwarding() error {
	return SetSysctl("net.ipv6.conf.all.forwarding", "1")
}

// SetRPFilter disables reverse path filter for an interface.
func SetRPFilter(iface string) error {
	return SetSysctl(fmt.Sprintf("net.ipv4.conf.%s.rp_filter", iface), "0")
}

// ============================================================================
// Capabilities
// ============================================================================

// RequiredCapabilities lists the Linux capabilities needed by straitd.
// All other capabilities are dropped.
var RequiredCapabilities = []string{
	"CAP_NET_ADMIN",
	"CAP_SYS_ADMIN",
	"CAP_NET_RAW",
	"CAP_PERFMON",
	"CAP_BPF",
	"CAP_SYS_RESOURCE",
	"CAP_SYS_PTRACE",
}

// CheckCapability checks if the current process has a given capability.
// Uses PR_GET_SECUREBITS / prctl as a lightweight check.
func CheckCapability(cap uint) bool {
	// Use capget syscall
	type capHeader struct {
		version uint32
		pid     int32
	}
	type capData struct {
		effective   uint32
		permitted   uint32
		inheritable uint32
	}
	hdr := capHeader{version: 0x20080522} // LINUX_CAPABILITY_VERSION_3
	var data [2]capData
	_, _, errno := unix.Syscall(unix.SYS_CAPGET,
		uintptr(unsafe.Pointer(&hdr)),
		uintptr(unsafe.Pointer(&data[0])),
		0,
	)
	if errno != 0 {
		return false
	}
	if cap < 32 {
		return data[0].effective&(1<<cap) != 0
	}
	return data[1].effective&(1<<(cap-32)) != 0
}

// ============================================================================
// Namespace helpers
// ============================================================================

// NetNSPath returns the network namespace path for a given PID.
func NetNSPath(pid int) string {
	return fmt.Sprintf("/proc/%d/ns/net", pid)
}

// CurrentNetNSPath returns the network namespace path for the current process.
func CurrentNetNSPath() string {
	return "/proc/self/ns/net"
}

// OpenNetNS opens a network namespace by path and returns its file descriptor.
func OpenNetNS(path string) (int, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC, 0)
	if err != nil {
		return -1, fmt.Errorf("open netns %q: %w", path, err)
	}
	return fd, nil
}

// ============================================================================
// BPF filesystem
// ============================================================================

// BPFFSDefaultPath is the default bpffs mount point.
const BPFFSDefaultPath = "/sys/fs/bpf"

// IsBPFFSMounted checks whether bpffs is mounted at the given path.
func IsBPFFSMounted(path string) (bool, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return false, fmt.Errorf("statfs %q: %w", path, err)
	}
	// BPF_FS_MAGIC = 0xcafe4a11
	const BPFFSMagic = 0xcafe4a11
	return stat.Type == BPFFSMagic, nil
}

// MountBPFFS mounts bpffs at the given path if not already mounted.
func MountBPFFS(path string) error {
	mounted, err := IsBPFFSMounted(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if mounted {
		return nil
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("mkdir %q: %w", path, err)
	}
	if err := unix.Mount("bpffs", path, "bpf", 0, ""); err != nil {
		return fmt.Errorf("mount bpffs at %q: %w", path, err)
	}
	return nil
}
