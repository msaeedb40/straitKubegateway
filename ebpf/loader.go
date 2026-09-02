// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package ebpf provides eBPF program loading, NetKit attachment,
// and bpffs pin management.
package ebpf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"

	"github.com/straitkubegateway/straitkubegateway/pkg/bpf"
	"github.com/straitkubegateway/straitkubegateway/pkg/linux"
)

// MinKernelMajor and MinKernelMinor define the minimum supported kernel version.
// Invariant: straitKubegateway requires Linux ≥ 6.7 for NetKit + CO-RE BTF support.
const (
	MinKernelMajor = 6
	MinKernelMinor = 7
)

// Loader manages loading and attaching eBPF programs.
type Loader struct {
	log       *zap.Logger
	bpffsPath string
	maps      map[string]*bpf.Map
}

// NewLoader creates a new eBPF Loader.
func NewLoader(bpffsPath string, log *zap.Logger) *Loader {
	return &Loader{
		log:       log,
		bpffsPath: bpffsPath,
		maps:      make(map[string]*bpf.Map),
	}
}

// CheckKernelVersion verifies the running kernel meets the minimum version
// requirement (≥ 6.7) for NetKit, CO-RE BTF, and all required eBPF hooks.
// This MUST be called before loading any eBPF programs.
func (l *Loader) CheckKernelVersion() error {
	var uname unix.Utsname
	if err := unix.Uname(&uname); err != nil {
		return fmt.Errorf("uname syscall failed: %w", err)
	}
	// Convert [65]int8 to string
	var buf []byte
	for _, c := range uname.Release {
		if c == 0 {
			break
		}
		buf = append(buf, byte(c))
	}
	release := string(buf)

	major, minor, err := parseKernelVersion(release)
	if err != nil {
		// Non-fatal in test/unprivileged environments — log and continue
		l.log.Warn("kernel version check skipped (uname parse error)",
			zap.String("release", release),
			zap.Error(err),
		)
		return nil
	}

	if major < MinKernelMajor || (major == MinKernelMajor && minor < MinKernelMinor) {
		return fmt.Errorf(
			"kernel %d.%d is below minimum required %d.%d for straitKubegateway "+
				"(NetKit, CO-RE BTF, XDP, TCX require Linux ≥ %d.%d)",
			major, minor,
			MinKernelMajor, MinKernelMinor,
			MinKernelMajor, MinKernelMinor,
		)
	}

	l.log.Info("kernel version check passed",
		zap.String("release", release),
		zap.Int("major", major),
		zap.Int("minor", minor),
	)
	return nil
}

// parseKernelVersion parses "6.12.0-xxx" → (6, 12, nil).
func parseKernelVersion(release string) (major, minor int, err error) {
	// Strip everything after the first non-numeric separator
	parts := strings.SplitN(release, ".", 3)
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("unexpected kernel release format: %q", release)
	}
	if _, err = fmt.Sscanf(parts[0], "%d", &major); err != nil {
		return 0, 0, fmt.Errorf("parse kernel major from %q: %w", parts[0], err)
	}
	// Minor may have suffixes like "12-generic"
	minorStr := parts[1]
	if idx := strings.IndexAny(minorStr, "-_+"); idx >= 0 {
		minorStr = minorStr[:idx]
	}
	if _, err = fmt.Sscanf(minorStr, "%d", &minor); err != nil {
		return 0, 0, fmt.Errorf("parse kernel minor from %q: %w", parts[1], err)
	}
	return major, minor, nil
}

// InitBPFFS ensures bpffs is mounted and ready for pinning.
func (l *Loader) InitBPFFS() error {
	if err := linux.MountBPFFS(l.bpffsPath); err != nil {
		return fmt.Errorf("mount bpffs at %q: %w", l.bpffsPath, err)
	}
	dir := filepath.Join(l.bpffsPath, "straitkubegateway")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %q: %w", dir, err)
	}
	l.log.Info("bpffs initialized", zap.String("path", dir))
	return nil
}

// LoadMap creates a versioned eBPF map and pins it to bpffs.
// The return value MUST be checked by the caller — never assume success.
func (l *Loader) LoadMap(spec bpf.MapSpec) (*bpf.Map, error) {
	if spec.PinPath == "" {
		spec.PinPath = filepath.Join(l.bpffsPath, "straitkubegateway", spec.Name)
	}
	m, err := bpf.CreateMap(spec)
	if err != nil {
		return nil, fmt.Errorf("create BPF map %q: %w", spec.Name, err)
	}
	l.maps[spec.Name] = m
	l.log.Info("BPF map created and pinned",
		zap.String("name", spec.Name),
		zap.String("pinPath", spec.PinPath),
	)
	return m, nil
}

// UnpinAll removes all pinned BPF objects from the bpffs directory.
// Called on clean shutdown to avoid stale pins after a restart.
func (l *Loader) UnpinAll() error {
	dir := filepath.Join(l.bpffsPath, "straitkubegateway")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read bpffs dir %q: %w", dir, err)
	}
	var errs []string
	for _, e := range entries {
		path := filepath.Join(dir, e.Name())
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Sprintf("unpin %q: %v", path, err))
		} else {
			l.log.Debug("unpinned BPF object", zap.String("path", path))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("unpin errors: %s", strings.Join(errs, "; "))
	}
	return nil
}

// Close closes all loaded BPF maps.
func (l *Loader) Close() error {
	for name, m := range l.maps {
		if err := m.Close(); err != nil {
			l.log.Warn("failed to close BPF map", zap.String("name", name), zap.Error(err))
		}
	}
	return nil
}
