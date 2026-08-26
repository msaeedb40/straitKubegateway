// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package ebpf provides eBPF program loading, NetKit attachment,
// and bpffs pin management.
package ebpf

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/straitkubegateway/straitkubegateway/pkg/bpf"
	"github.com/straitkubegateway/straitkubegateway/pkg/linux"
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

// LoadMap creates a versioned eBPF map.
func (l *Loader) LoadMap(spec bpf.MapSpec) (*bpf.Map, error) {
	if spec.PinPath == "" {
		spec.PinPath = filepath.Join(l.bpffsPath, "straitkubegateway", spec.Name)
	}
	m, err := bpf.CreateMap(spec)
	if err != nil {
		return nil, err
	}
	l.maps[spec.Name] = m
	l.log.Info("BPF map created and pinned",
		zap.String("name", spec.Name),
		zap.String("pinPath", spec.PinPath),
	)
	return m, nil
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
