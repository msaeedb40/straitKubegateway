// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package encryption implements WireGuard and IPsec encryption for
// pod-to-pod and node-to-node traffic encryption in straitKubegateway.
package encryption

import (
	"crypto/ecdh"
	"crypto/rand"
	"fmt"
	"net/netip"
	"sync"

	"go.uber.org/zap"
	"golang.org/x/sys/unix"
)

// ============================================================================
// WireGuard
// ============================================================================

// WireGuardKeySize is the size of a WireGuard key (Curve25519).
const WireGuardKeySize = 32

// WireGuardKey is a WireGuard public or private key.
type WireGuardKey [WireGuardKeySize]byte

// WireGuardManager manages WireGuard tunnel state for pod-to-pod encryption.
type WireGuardManager struct {
	mu         sync.RWMutex
	log        *zap.Logger
	privateKey WireGuardKey
	publicKey  WireGuardKey
	peers      map[netip.Addr]*WireGuardPeer
	ifName     string
	listenPort uint16
}

// WireGuardPeer represents a WireGuard remote peer.
type WireGuardPeer struct {
	// PublicKey is the peer's WireGuard public key.
	PublicKey WireGuardKey
	// Endpoint is the peer's WireGuard UDP endpoint.
	Endpoint netip.AddrPort
	// AllowedIPs are the CIDRs routed through this peer.
	AllowedIPs []netip.Prefix
	// NodeIP is the Kubernetes node IP of this peer.
	NodeIP netip.Addr
}

// NewWireGuardManager creates a new WireGuard manager.
func NewWireGuardManager(ifName string, listenPort uint16, log *zap.Logger) (*WireGuardManager, error) {
	priv, pub, err := generateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("generate WireGuard key pair: %w", err)
	}
	return &WireGuardManager{
		log:        log,
		privateKey: priv,
		publicKey:  pub,
		peers:      make(map[netip.Addr]*WireGuardPeer),
		ifName:     ifName,
		listenPort: listenPort,
	}, nil
}

// PublicKey returns the node's WireGuard public key.
func (m *WireGuardManager) PublicKey() WireGuardKey { return m.publicKey }

// UpsertPeer adds or updates a WireGuard peer.
func (m *WireGuardManager) UpsertPeer(peer *WireGuardPeer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.peers[peer.NodeIP] = peer
	m.log.Info("WireGuard peer upserted",
		zap.String("nodeIP", peer.NodeIP.String()),
		zap.String("endpoint", peer.Endpoint.String()),
		zap.Int("allowedIPs", len(peer.AllowedIPs)),
	)
	// TODO: apply via WireGuard netlink API (golang.zx2c4.com/wireguard/wgctrl)
	return nil
}

// DeletePeer removes a WireGuard peer.
func (m *WireGuardManager) DeletePeer(nodeIP netip.Addr) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.peers, nodeIP)
	m.log.Info("WireGuard peer deleted", zap.String("nodeIP", nodeIP.String()))
}

// generateKeyPair generates a WireGuard Curve25519 key pair.
func generateKeyPair() (priv, pub WireGuardKey, err error) {
	if _, err = rand.Read(priv[:]); err != nil {
		return priv, pub, fmt.Errorf("rand read: %w", err)
	}
	// Clamp the private key (RFC 7748)
	priv[0] &= 248
	priv[31] &= 127
	priv[31] |= 64

	k, err := ecdh.X25519().NewPrivateKey(priv[:])
	if err != nil {
		return priv, pub, fmt.Errorf("x25519 key derivation: %w", err)
	}
	copy(pub[:], k.PublicKey().Bytes())
	return priv, pub, nil
}

// ============================================================================
// IPsec
// ============================================================================

// IPsecSA is an IPsec Security Association.
type IPsecSA struct {
	// SPI is the Security Parameter Index.
	SPI uint32
	// LocalIP is the local tunnel endpoint.
	LocalIP netip.Addr
	// RemoteIP is the remote tunnel endpoint.
	RemoteIP netip.Addr
	// AuthKey is the authentication key (HMAC-SHA256).
	AuthKey []byte
	// EncKey is the encryption key (AES-128-GCM).
	EncKey []byte
	// Proto is the IPsec protocol: ESP or AH.
	Proto string
	// Mode is the IPsec mode: tunnel or transport.
	Mode string
}

// IPsecManager manages IPsec Security Associations.
type IPsecManager struct {
	mu  sync.RWMutex
	log *zap.Logger
	sas map[uint32]*IPsecSA
}

// NewIPsecManager creates a new IPsec manager.
func NewIPsecManager(log *zap.Logger) *IPsecManager {
	return &IPsecManager{
		log: log,
		sas: make(map[uint32]*IPsecSA),
	}
}

// AddSA programs an IPsec Security Association via the kernel XFRM interface.
func (m *IPsecManager) AddSA(sa *IPsecSA) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sas[sa.SPI] = sa
	m.log.Info("IPsec SA added",
		zap.Uint32("spi", sa.SPI),
		zap.String("local", sa.LocalIP.String()),
		zap.String("remote", sa.RemoteIP.String()),
		zap.String("proto", sa.Proto),
		zap.String("mode", sa.Mode),
	)
	// TODO: use XFRM netlink to program the SA into the kernel
	// netlink.XfrmStateAdd(...)
	_ = unix.AF_INET // ensure unix import is used
	return nil
}

// DeleteSA removes an IPsec Security Association.
func (m *IPsecManager) DeleteSA(spi uint32) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if sa, ok := m.sas[spi]; ok {
		m.log.Info("IPsec SA deleted",
			zap.Uint32("spi", sa.SPI),
			zap.String("remote", sa.RemoteIP.String()),
		)
		delete(m.sas, spi)
	}
	// TODO: XFRM netlink delete
	return nil
}

// GetSA returns an SA by SPI.
func (m *IPsecManager) GetSA(spi uint32) (*IPsecSA, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sa, ok := m.sas[spi]
	return sa, ok
}

// ============================================================================
// mTLS — Kubernetes CA-based mutual TLS
// ============================================================================

// KubernetesCAPath is the default Kubernetes service account CA path.
const KubernetesCAPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

// KubernetesTokenPath is the default service account token path.
const KubernetesTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
