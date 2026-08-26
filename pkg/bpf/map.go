// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package bpf provides typed Go wrappers for eBPF maps and helpers
// used by the straitKubegateway dataplane.
package bpf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"golang.org/x/sys/unix"
)

// ============================================================================
// Map type constants (mirrors kernel bpf_map_type)
// ============================================================================

// MapType identifies the kernel eBPF map type.
type MapType uint32

const (
	MapTypeHash           MapType = 1
	MapTypeArray          MapType = 2
	MapTypeProgramArray   MapType = 3
	MapTypePerfEventArray MapType = 4
	MapTypePerCPUHash     MapType = 5
	MapTypePerCPUArray    MapType = 6
	MapTypeLRUHash        MapType = 9
	MapTypeLPMTrie        MapType = 11
	MapTypeRingBuf        MapType = 27
)

// ============================================================================
// Map version contract
// ============================================================================

// MapVersion is a monotonically increasing version for BPF map layouts.
// Invariant: BPF map layouts are versioned contracts — changing a map's
// key/value layout requires incrementing the version.
type MapVersion uint32

const (
	// MapVersionCurrent is the current straitKubegateway BPF map schema version.
	MapVersionCurrent MapVersion = 1
)

// ============================================================================
// Map descriptor
// ============================================================================

// MapSpec describes an eBPF map before it is created.
type MapSpec struct {
	// Name is the BPF map name (max 15 chars for kernel).
	Name string
	// Type is the kernel BPF map type.
	Type MapType
	// KeySize is the key size in bytes.
	KeySize uint32
	// ValueSize is the value size in bytes.
	ValueSize uint32
	// MaxEntries is the maximum number of entries.
	MaxEntries uint32
	// Flags are the BPF map creation flags.
	Flags uint32
	// Version is the map schema version (straitKubegateway extension).
	Version MapVersion
	// PinPath is where this map is pinned in bpffs (empty = not pinned).
	PinPath string
}

// ============================================================================
// Map — a handle to a live eBPF map
// ============================================================================

// Map represents a live eBPF map file descriptor.
type Map struct {
	fd   int
	spec MapSpec
}

// FD returns the file descriptor of the map.
func (m *Map) FD() int { return m.fd }

// Spec returns the specification used to create the map.
func (m *Map) Spec() MapSpec { return m.spec }

// Close closes the map file descriptor.
func (m *Map) Close() error {
	if m.fd < 0 {
		return nil
	}
	err := unix.Close(m.fd)
	m.fd = -1
	return err
}

// CreateMap creates a new eBPF map according to spec.
func CreateMap(spec MapSpec) (*Map, error) {
	attr := bpfMapCreateAttr{
		mapType:    uint32(spec.Type),
		keySize:    spec.KeySize,
		valueSize:  spec.ValueSize,
		maxEntries: spec.MaxEntries,
		mapFlags:   spec.Flags,
	}
	// Copy name (up to 15 chars + NUL)
	name := spec.Name
	if len(name) > 15 {
		name = name[:15]
	}
	copy(attr.mapName[:], name)

	fd, _, errno := unix.Syscall(unix.SYS_BPF,
		uintptr(bpfCmdMapCreate),
		uintptr(unsafe.Pointer(&attr)),
		unsafe.Sizeof(attr),
	)
	if errno != 0 {
		return nil, fmt.Errorf("BPF_MAP_CREATE %q: %w", spec.Name, errno)
	}

	m := &Map{fd: int(fd), spec: spec}

	if spec.PinPath != "" {
		if err := m.Pin(spec.PinPath); err != nil {
			_ = m.Close()
			return nil, fmt.Errorf("pin map %q at %q: %w", spec.Name, spec.PinPath, err)
		}
	}
	return m, nil
}

// OpenPinnedMap opens an already-pinned eBPF map from bpffs.
func OpenPinnedMap(path string) (*Map, error) {
	fd, err := bpfObjGet(path)
	if err != nil {
		return nil, fmt.Errorf("open pinned map %q: %w", path, err)
	}
	return &Map{fd: fd, spec: MapSpec{PinPath: path}}, nil
}

// Pin pins the map to a bpffs path.
func (m *Map) Pin(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return bpfObjPin(path, m.fd)
}

// Unpin removes the map's bpffs pin.
func (m *Map) Unpin() error {
	if m.spec.PinPath == "" {
		return nil
	}
	return os.Remove(m.spec.PinPath)
}

// Lookup retrieves the value for key. key and value must be pointers to
// correctly-sized types matching the map's KeySize and ValueSize.
func (m *Map) Lookup(key, value unsafe.Pointer) error {
	attr := bpfMapOpAttr{
		mapFD: uint32(m.fd),
		key:   uint64(uintptr(key)),
		value: uint64(uintptr(value)),
	}
	_, _, errno := unix.Syscall(unix.SYS_BPF,
		uintptr(bpfCmdMapLookupElem),
		uintptr(unsafe.Pointer(&attr)),
		unsafe.Sizeof(attr),
	)
	if errno != 0 {
		if errors.Is(errno, unix.ENOENT) {
			return ErrKeyNotFound
		}
		return fmt.Errorf("map lookup: %w", errno)
	}
	return nil
}

// Update inserts or updates a key→value pair.
// flags: 0=any, BPF_NOEXIST=insert-only, BPF_EXIST=update-only.
func (m *Map) Update(key, value unsafe.Pointer, flags uint64) error {
	attr := bpfMapOpAttr{
		mapFD: uint32(m.fd),
		key:   uint64(uintptr(key)),
		value: uint64(uintptr(value)),
		flags: flags,
	}
	_, _, errno := unix.Syscall(unix.SYS_BPF,
		uintptr(bpfCmdMapUpdateElem),
		uintptr(unsafe.Pointer(&attr)),
		unsafe.Sizeof(attr),
	)
	if errno != 0 {
		return fmt.Errorf("map update: %w", errno)
	}
	return nil
}

// Delete removes a key from the map.
func (m *Map) Delete(key unsafe.Pointer) error {
	attr := bpfMapOpAttr{
		mapFD: uint32(m.fd),
		key:   uint64(uintptr(key)),
	}
	_, _, errno := unix.Syscall(unix.SYS_BPF,
		uintptr(bpfCmdMapDeleteElem),
		uintptr(unsafe.Pointer(&attr)),
		unsafe.Sizeof(attr),
	)
	if errno != 0 {
		if errors.Is(errno, unix.ENOENT) {
			return ErrKeyNotFound
		}
		return fmt.Errorf("map delete: %w", errno)
	}
	return nil
}

// ErrKeyNotFound is returned when a lookup or delete finds no matching key.
var ErrKeyNotFound = errors.New("key not found")

// ============================================================================
// BPF syscall structures and helpers
// ============================================================================

const (
	bpfCmdMapCreate     = 0
	bpfCmdMapLookupElem = 1
	bpfCmdMapUpdateElem = 2
	bpfCmdMapDeleteElem = 3
	bpfCmdObjPin        = 8
	bpfCmdObjGet        = 7
)

//nolint:structcheck,unused
type bpfMapCreateAttr struct {
	mapType    uint32
	keySize    uint32
	valueSize  uint32
	maxEntries uint32
	mapFlags   uint32
	innerMapFD uint32
	numaNode   uint32
	mapName    [16]byte
	mapIfindex uint32
	btfFD      uint32
	btfKeyID   uint32
	btfValueID uint32
}

//nolint:structcheck,unused
type bpfMapOpAttr struct {
	mapFD uint32
	_     uint32
	key   uint64
	value uint64
	flags uint64
}

type bpfObjAttr struct {
	pathname uint64
	bpfFD    uint32
	fileFlags uint32
}

func bpfObjPin(path string, fd int) error {
	pathBytes, err := unix.BytePtrFromString(path)
	if err != nil {
		return err
	}
	attr := bpfObjAttr{
		pathname: uint64(uintptr(unsafe.Pointer(pathBytes))),
		bpfFD:   uint32(fd),
	}
	_, _, errno := unix.Syscall(unix.SYS_BPF,
		uintptr(bpfCmdObjPin),
		uintptr(unsafe.Pointer(&attr)),
		unsafe.Sizeof(attr),
	)
	if errno != 0 {
		return fmt.Errorf("bpf obj pin %q: %w", path, errno)
	}
	return nil
}

func bpfObjGet(path string) (int, error) {
	pathBytes, err := unix.BytePtrFromString(path)
	if err != nil {
		return -1, err
	}
	attr := bpfObjAttr{
		pathname: uint64(uintptr(unsafe.Pointer(pathBytes))),
	}
	fd, _, errno := unix.Syscall(unix.SYS_BPF,
		uintptr(bpfCmdObjGet),
		uintptr(unsafe.Pointer(&attr)),
		unsafe.Sizeof(attr),
	)
	if errno != 0 {
		return -1, fmt.Errorf("bpf obj get %q: %w", path, errno)
	}
	return int(fd), nil
}

// ============================================================================
// Byte-order helpers (eBPF maps store network byte order)
// ============================================================================

var isBigEndian = binary.NativeEndian.Uint16([]byte{1, 0}) == 0x0100

// Htons converts a uint16 from host to network byte order.
func Htons(v uint16) uint16 {
	if isBigEndian {
		return v
	}
	return (v << 8) | (v >> 8)
}

// Htonl converts a uint32 from host to network byte order.
func Htonl(v uint32) uint32 {
	if isBigEndian {
		return v
	}
	return (v << 24) | ((v & 0x0000ff00) << 8) | ((v & 0x00ff0000) >> 8) | (v >> 24)
}

// Ntohs converts a uint16 from network to host byte order.
func Ntohs(v uint16) uint16 { return Htons(v) }

// Ntohl converts a uint32 from network to host byte order.
func Ntohl(v uint32) uint32 { return Htonl(v) }
