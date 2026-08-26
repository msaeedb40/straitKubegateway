// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/netip"
	"testing"
)

func TestMaglevLookup(t *testing.T) {
	table := make([]uint32, MaglevTableSize)
	for i := range table {
		table[i] = uint32(i % 3) // 3 backends
	}

	srcIP := netip.MustParseAddr("10.244.1.2")
	dstIP := netip.MustParseAddr("10.96.0.10")

	// Consistent hashing: same flow 5-tuple must always pick the same backend
	backend1 := MaglevLookup(table, srcIP, dstIP, 40000, 80, 6)
	backend2 := MaglevLookup(table, srcIP, dstIP, 40000, 80, 6)

	if backend1 != backend2 {
		t.Errorf("maglev hashing not consistent: got %d and %d", backend1, backend2)
	}

	if backend1 < 0 || backend1 >= 3 {
		t.Errorf("invalid backend index %d", backend1)
	}
}
