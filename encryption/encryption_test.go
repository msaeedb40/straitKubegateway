// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package encryption

import (
	"bytes"
	"crypto/ecdh"
	"testing"

	"go.uber.org/zap"
)

func TestWireGuardKeyPairGeneration(t *testing.T) {
	mgr, err := NewWireGuardManager("sg-wg0", 51820, zap.NewNop())
	if err != nil {
		t.Fatalf("failed to create WireGuard manager: %v", err)
	}

	pub := mgr.PublicKey()
	priv := mgr.privateKey

	// Assert public key is not equal to private key
	if bytes.Equal(pub[:], priv[:]) {
		t.Fatalf("critical error: WireGuard public key equals private key")
	}

	// Verify public key can be parsed as valid X25519 public key
	_, err = ecdh.X25519().NewPublicKey(pub[:])
	if err != nil {
		t.Fatalf("generated public key is invalid X25519 key: %v", err)
	}
}
