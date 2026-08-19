// Package types provides common type definitions shared across the
// straitKubegateway codebase.
package types

// Identity represents a unique BPF-level identity for a network endpoint.
// It is a 32-bit unsigned integer allocated by the identity controller and
// compiled into BPF maps for fast-path policy and routing decisions.
type Identity uint32

const (
	// IdentityUnknown is the zero value indicating no identity has been assigned.
	IdentityUnknown Identity = 0

	// IdentityHost identifies traffic originating from or destined to the host network namespace.
	IdentityHost Identity = 1

	// IdentityWorld identifies traffic from external (non-cluster) sources.
	IdentityWorld Identity = 2

	// IdentityHealth identifies health-check probes.
	IdentityHealth Identity = 3

	// IdentityKubeAPIServer identifies traffic from/to the Kubernetes API server.
	IdentityKubeAPIServer Identity = 4

	// IdentityCoreDNS identifies traffic from/to CoreDNS.
	IdentityCoreDNS Identity = 5

	// IdentityReservedMax is the upper bound of reserved identities.
	// User-allocated identities start above this value.
	IdentityReservedMax Identity = 255
)

// IsReserved reports whether the identity is in the reserved range.
func (id Identity) IsReserved() bool {
	return id <= IdentityReservedMax
}

// IsValid reports whether the identity is a non-zero, allocated value.
func (id Identity) IsValid() bool {
	return id != IdentityUnknown
}

// Uint32 returns the identity as a uint32 for BPF map operations.
func (id Identity) Uint32() uint32 {
	return uint32(id)
}
