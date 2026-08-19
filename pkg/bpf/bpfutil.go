// Package bpf provides abstractions for BPF map operations, program loading,
// and pinning used by the straitKubegateway eBPF dataplane.
package bpf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cilium/ebpf"
)

const (
	// DefaultBPFRoot is the default bpffs mount point.
	DefaultBPFRoot = "/sys/fs/bpf"

	// PinBasePath is the base path where straitKubegateway pins BPF objects.
	PinBasePath = "/sys/fs/bpf/straitkubegateway"
)

// MapSpec defines the specification for a BPF map.
type MapSpec struct {
	Name       string
	Type       ebpf.MapType
	KeySize    uint32
	ValueSize  uint32
	MaxEntries uint32
	Flags      uint32
	PinPath    string
}

// CreateMap creates a BPF map from the given spec.
func CreateMap(spec *MapSpec) (*ebpf.Map, error) {
	mapSpec := &ebpf.MapSpec{
		Name:       spec.Name,
		Type:       spec.Type,
		KeySize:    spec.KeySize,
		ValueSize:  spec.ValueSize,
		MaxEntries: spec.MaxEntries,
		Flags:      spec.Flags,
	}

	if spec.PinPath != "" {
		mapSpec.Pinning = ebpf.PinByName
	}

	m, err := ebpf.NewMap(mapSpec)
	if err != nil {
		return nil, fmt.Errorf("create map %q: %w", spec.Name, err)
	}

	if spec.PinPath != "" {
		if err := PinMap(m, spec.PinPath); err != nil {
			m.Close()
			return nil, err
		}
	}

	return m, nil
}

// PinMap pins a BPF map to the filesystem.
func PinMap(m *ebpf.Map, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %q: %w", dir, err)
	}
	if err := m.Pin(path); err != nil {
		return fmt.Errorf("pin map to %q: %w", path, err)
	}
	return nil
}

// LoadPinnedMap loads a BPF map that was previously pinned.
func LoadPinnedMap(path string) (*ebpf.Map, error) {
	m, err := ebpf.LoadPinnedMap(path, nil)
	if err != nil {
		return nil, fmt.Errorf("load pinned map %q: %w", path, err)
	}
	return m, nil
}

// LoadPinnedProgram loads a BPF program that was previously pinned.
func LoadPinnedProgram(path string) (*ebpf.Program, error) {
	p, err := ebpf.LoadPinnedProgram(path, nil)
	if err != nil {
		return nil, fmt.Errorf("load pinned program %q: %w", path, err)
	}
	return p, nil
}

// PinProgram pins a BPF program to the filesystem.
func PinProgram(p *ebpf.Program, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("mkdir %q: %w", dir, err)
	}
	if err := p.Pin(path); err != nil {
		return fmt.Errorf("pin program to %q: %w", path, err)
	}
	return nil
}

// MapLookup performs a lookup on a BPF map.
func MapLookup(m *ebpf.Map, key, value interface{}) error {
	return m.Lookup(key, value)
}

// MapUpdate inserts or updates an entry in a BPF map.
func MapUpdate(m *ebpf.Map, key, value interface{}, flags ebpf.MapUpdateFlags) error {
	return m.Update(key, value, flags)
}

// MapDelete removes an entry from a BPF map.
func MapDelete(m *ebpf.Map, key interface{}) error {
	return m.Delete(key)
}

// MapIterator returns an iterator over a BPF map's entries.
func MapIterator(m *ebpf.Map) *ebpf.MapIterator {
	return m.Iterate()
}

// CleanupPins removes all straitKubegateway BPF pins from the filesystem.
func CleanupPins() error {
	return os.RemoveAll(PinBasePath)
}

// EnsurePinDirectory creates the pin directory structure.
func EnsurePinDirectory() error {
	dirs := []string{
		PinBasePath,
		filepath.Join(PinBasePath, "maps"),
		filepath.Join(PinBasePath, "progs"),
		filepath.Join(PinBasePath, "links"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("mkdir %q: %w", dir, err)
		}
	}
	return nil
}

// MapPinPath returns the standard pin path for a named BPF map.
func MapPinPath(name string) string {
	return filepath.Join(PinBasePath, "maps", name)
}

// ProgPinPath returns the standard pin path for a named BPF program.
func ProgPinPath(name string) string {
	return filepath.Join(PinBasePath, "progs", name)
}
