// Package linux provides abstractions for Linux kernel interfaces used by
// straitKubegateway, including capabilities, sysctl, namespaces, and netlink.
package linux

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// KernelVersion returns the running kernel version string.
func KernelVersion() (string, error) {
	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		return "", fmt.Errorf("uname: %w", err)
	}
	return strings.TrimRight(string(uname.Release[:]), "\x00"), nil
}

// CheckMinKernelVersion verifies the running kernel meets the minimum version.
// Returns an error if the kernel is too old.
func CheckMinKernelVersion(minMajor, minMinor int) error {
	ver, err := KernelVersion()
	if err != nil {
		return err
	}
	var major, minor int
	if _, err := fmt.Sscanf(ver, "%d.%d", &major, &minor); err != nil {
		return fmt.Errorf("cannot parse kernel version %q: %w", ver, err)
	}
	if major < minMajor || (major == minMajor && minor < minMinor) {
		return fmt.Errorf("kernel %d.%d is below minimum %d.%d", major, minor, minMajor, minMinor)
	}
	return nil
}

// Capability represents a Linux capability.
type Capability int

const (
	CapNetAdmin    Capability = unix.CAP_NET_ADMIN
	CapSysAdmin    Capability = unix.CAP_SYS_ADMIN
	CapNetRaw      Capability = unix.CAP_NET_RAW
	CapPerfmon     Capability = 38 // CAP_PERFMON
	CapBPF         Capability = 39 // CAP_BPF
	CapSysResource Capability = unix.CAP_SYS_RESOURCE
	CapSysPtrace   Capability = unix.CAP_SYS_PTRACE
)

// RequiredCapabilities returns the capabilities needed by straitKubegateway.
func RequiredCapabilities() []Capability {
	return []Capability{
		CapNetAdmin,
		CapSysAdmin,
		CapNetRaw,
		CapPerfmon,
		CapBPF,
		CapSysResource,
		CapSysPtrace,
	}
}

// String returns the capability name.
func (c Capability) String() string {
	switch c {
	case CapNetAdmin:
		return "CAP_NET_ADMIN"
	case CapSysAdmin:
		return "CAP_SYS_ADMIN"
	case CapNetRaw:
		return "CAP_NET_RAW"
	case CapPerfmon:
		return "CAP_PERFMON"
	case CapBPF:
		return "CAP_BPF"
	case CapSysResource:
		return "CAP_SYS_RESOURCE"
	case CapSysPtrace:
		return "CAP_SYS_PTRACE"
	default:
		return fmt.Sprintf("CAP_%d", int(c))
	}
}

// ReadSysctl reads a sysctl value.
func ReadSysctl(name string) (string, error) {
	path := filepath.Join("/proc/sys", strings.ReplaceAll(name, ".", "/"))
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read sysctl %q: %w", name, err)
	}
	return strings.TrimSpace(string(data)), nil
}

// WriteSysctl writes a sysctl value.
func WriteSysctl(name, value string) error {
	path := filepath.Join("/proc/sys", strings.ReplaceAll(name, ".", "/"))
	if err := os.WriteFile(path, []byte(value), 0644); err != nil {
		return fmt.Errorf("write sysctl %q=%q: %w", name, value, err)
	}
	return nil
}

// EnableIPForwarding enables IPv4 and IPv6 forwarding.
func EnableIPForwarding() error {
	if err := WriteSysctl("net.ipv4.ip_forward", "1"); err != nil {
		return err
	}
	return WriteSysctl("net.ipv6.conf.all.forwarding", "1")
}

// DisableRPFilter disables reverse path filtering on the given interface.
func DisableRPFilter(iface string) error {
	return WriteSysctl(fmt.Sprintf("net.ipv4.conf.%s.rp_filter", iface), "0")
}

// IsBPFFilesystemMounted checks whether bpffs is mounted at the given path.
func IsBPFFilesystemMounted(path string) (bool, error) {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return false, fmt.Errorf("read /proc/mounts: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == path && fields[2] == "bpf" {
			return true, nil
		}
	}
	return false, nil
}

// MountBPFFilesystem mounts bpffs at the given path.
func MountBPFFilesystem(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("mkdir %q: %w", path, err)
	}
	return unix.Mount("bpffs", path, "bpf", 0, "")
}

// EnsureBPFFilesystem ensures bpffs is mounted, mounting it if necessary.
func EnsureBPFFilesystem(path string) error {
	mounted, err := IsBPFFilesystemMounted(path)
	if err != nil {
		return err
	}
	if mounted {
		return nil
	}
	return MountBPFFilesystem(path)
}

// NamespaceType represents a Linux namespace type.
type NamespaceType int

const (
	NamespaceNet NamespaceType = unix.CLONE_NEWNET
	NamespacePID NamespaceType = unix.CLONE_NEWPID
	NamespaceMnt NamespaceType = unix.CLONE_NEWNS
)

// GetNetNSCookie returns the network namespace cookie for the given path.
func GetNetNSCookie(nsPath string) (uint64, error) {
	f, err := os.Open(nsPath)
	if err != nil {
		return 0, fmt.Errorf("open netns %q: %w", nsPath, err)
	}
	defer f.Close()

	cookie, err := unix.GetsockoptUint64(int(f.Fd()), unix.SOL_SOCKET, unix.SO_NETNS_COOKIE)
	if err != nil {
		return 0, fmt.Errorf("get netns cookie: %w", err)
	}
	return cookie, nil
}
