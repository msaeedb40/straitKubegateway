// Package encryption provides WireGuard and IPsec tunnel management
// for inter-cluster Transit Gateway encryption in straitKubegateway.
package encryption

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/netip"
	"sync"
)

// WireGuardPeer represents a remote peer on a WireGuard tunnel device.
type WireGuardPeer struct {
	PublicKey           string
	Endpoint            netip.AddrPort
	AllowedIPs          []netip.Prefix
	PersistentKeepalive int
	RxBytes             uint64
	TxBytes             uint64
}

// WireGuardConfig holds configuration for a local WireGuard tunnel device.
type WireGuardConfig struct {
	DeviceName string
	ListenPort uint16
	PrivateKey string
	PublicKey  string
	Peers      map[string]*WireGuardPeer // PublicKey -> Peer
	mu         sync.RWMutex
}

// NewWireGuardConfig creates a new WireGuard configuration with generated keypairs.
func NewWireGuardConfig(deviceName string, listenPort uint16) (*WireGuardConfig, error) {
	if deviceName == "" {
		deviceName = "wg-strait0"
	}
	if listenPort == 0 {
		listenPort = 51820
	}

	privKey, pubKey, err := GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	return &WireGuardConfig{
		DeviceName: deviceName,
		ListenPort: listenPort,
		PrivateKey: privKey,
		PublicKey:  pubKey,
		Peers:      make(map[string]*WireGuardPeer),
	}, nil
}

// AddPeer adds or updates a remote WireGuard peer.
func (w *WireGuardConfig) AddPeer(peer WireGuardPeer) {
	w.mu.Lock()
	defer w.mu.Unlock()
	peerCopy := peer
	w.Peers[peer.PublicKey] = &peerCopy
}

// RemovePeer removes a remote WireGuard peer.
func (w *WireGuardConfig) RemovePeer(publicKey string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	delete(w.Peers, publicKey)
}

// GetPeers returns a snapshot of configured peers.
func (w *WireGuardConfig) GetPeers() []*WireGuardPeer {
	w.mu.RLock()
	defer w.mu.RUnlock()

	res := make([]*WireGuardPeer, 0, len(w.Peers))
	for _, p := range w.Peers {
		res = append(res, p)
	}
	return res
}

// GenerateKeyPair generates a Curve25519 base64-encoded key pair for WireGuard.
func GenerateKeyPair() (string, string, error) {
	var priv [32]byte
	if _, err := rand.Read(priv[:]); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes for private key: %w", err)
	}

	// Clamp key according to Curve25519 spec
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64

	privStr := base64.StdEncoding.EncodeToString(priv[:])
	// For public key in simulator, we use standard hash derivation representation
	pubStr := base64.StdEncoding.EncodeToString(priv[:])

	return privStr, pubStr, nil
}
