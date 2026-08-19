// Package kernel provides low-level Linux kernel capability, sysctl, and namespace helpers.
package kernel

import (
	"fmt"
	"os"
	"strings"

	"github.com/straitKubegateway/straitKubegateway/pkg/linux"
)

// Environment holds the detected Linux kernel features and runtime capabilities.
type Environment struct {
	KernelVersion   string
	IsKernel67Plus  bool
	IsKernel612Plus bool
	BPFFSMounted    bool
	CgroupV2        bool
	MissingCaps     []string
}

// DetectEnvironment inspects the host kernel and validates prerequisites.
func DetectEnvironment() (*Environment, error) {
	ver, err := linux.KernelVersion()
	if err != nil {
		return nil, fmt.Errorf("detect kernel version: %w", err)
	}

	env := &Environment{
		KernelVersion: ver,
	}

	if err := linux.CheckMinKernelVersion(6, 7); err == nil {
		env.IsKernel67Plus = true
	}
	if err := linux.CheckMinKernelVersion(6, 12); err == nil {
		env.IsKernel612Plus = true
	}

	mounted, _ := linux.IsBPFFilesystemMounted("/sys/fs/bpf")
	env.BPFFSMounted = mounted

	data, err := os.ReadFile("/proc/mounts")
	if err == nil {
		env.CgroupV2 = strings.Contains(string(data), "cgroup2")
	}

	return env, nil
}
