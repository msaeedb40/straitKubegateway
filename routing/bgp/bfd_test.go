// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package bgp

import (
	"net/netip"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestBFDStateTransitionAndTimeout(t *testing.T) {
	log := zap.NewNop()
	peer := netip.MustParseAddr("192.168.1.1")

	downed := false
	bfd := NewBFDManager(func(p netip.Addr) {
		if p == peer {
			downed = true
		}
	}, log)

	session := bfd.AddSession(peer, 50*time.Millisecond, 50*time.Millisecond, 3)
	if session.State != BFDStateInit {
		t.Fatalf("expected state Init, got %s", session.State)
	}

	// Receive heartbeat -> UP
	bfd.ReceiveHeartbeat(peer)
	if session.State != BFDStateUp {
		t.Fatalf("expected state Up after heartbeat, got %s", session.State)
	}

	// Fast forward time past 3 * 50ms = 150ms timeout
	future := time.Now().Add(200 * time.Millisecond)
	downedPeers := bfd.Tick(future)

	if len(downedPeers) != 1 || downedPeers[0] != peer {
		t.Fatalf("expected peer %s in downed peers, got %v", peer, downedPeers)
	}
	if session.State != BFDStateDown {
		t.Fatalf("expected state Down after timeout, got %s", session.State)
	}
	if !downed {
		t.Fatalf("expected onDown callback to be invoked")
	}
}
