package types

import "fmt"

// SegmentID represents a network segment identifier.
// Segments provide network isolation within and across clusters.
// All segment IDs are 32-bit unsigned integers (range 0–4294967295).
type SegmentID uint32

const (
	// SegmentBackbone is the default backbone segment.
	// All segments are connected through the backbone segment.
	SegmentBackbone SegmentID = 0

	// SegmentMax is the maximum valid segment ID.
	SegmentMax SegmentID = 4294967295
)

// IsBackbone reports whether this is the backbone segment.
func (s SegmentID) IsBackbone() bool {
	return s == SegmentBackbone
}

// Uint32 returns the segment ID as a uint32 for BPF map operations.
func (s SegmentID) Uint32() uint32 {
	return uint32(s)
}

// String returns a human-readable representation of the segment ID.
func (s SegmentID) String() string {
	if s.IsBackbone() {
		return "backbone(0)"
	}
	return fmt.Sprintf("segment(%d)", s)
}

// SegmentPolicy defines the isolation and connectivity rules for a segment.
type SegmentPolicy struct {
	// ID is the unique segment identifier.
	ID SegmentID

	// Isolated indicates whether the segment is isolated by default.
	// All segments are isolated by default.
	Isolated bool

	// BackboneConnected indicates whether the segment is connected to the
	// backbone segment (segment 0). All segments are connected by default.
	BackboneConnected bool
}

// DefaultSegmentPolicy returns the default policy for a new segment.
// Segments are isolated by default and connected to the backbone.
func DefaultSegmentPolicy(id SegmentID) SegmentPolicy {
	return SegmentPolicy{
		ID:                id,
		Isolated:          true,
		BackboneConnected: true,
	}
}
