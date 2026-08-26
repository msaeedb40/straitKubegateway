// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package bgp implements BGP route advertisement, neighbor peering, and
// BFD (Bidirectional Forwarding Detection) sub-second failure detection.
package bgp

import (
	"net/netip"
	"sync"
	"time"

	"go.uber.org/zap"
)

// BFDState represents the state of a BFD session (RFC 5880).
type BFDState string

const (
	BFDStateAdminDown BFDState = "AdminDown"
	BFDStateDown      BFDState = "Down"
	BFDStateInit      BFDState = "Init"
	BFDStateUp        BFDState = "Up"
)

// BFDSession represents a single BFD tracking session for a BGP peer.
type BFDSession struct {
	mu                   sync.Mutex
	Peer                 netip.Addr
	State                BFDState
	DesiredMinTxInterval time.Duration
	RequiredMinRxInterval time.Duration
	DetectMultiplier     uint8
	missedPackets        uint8
	lastRx               time.Time
}

// BFDManager manages fast failure detection sessions for routing peers.
type BFDManager struct {
	mu       sync.RWMutex
	log      *zap.Logger
	sessions map[netip.Addr]*BFDSession
	onDown   func(peer netip.Addr)
}

// NewBFDManager creates a new BFD manager.
func NewBFDManager(onDown func(peer netip.Addr), log *zap.Logger) *BFDManager {
	return &BFDManager{
		log:      log,
		sessions: make(map[netip.Addr]*BFDSession),
		onDown:   onDown,
	}
}

// AddSession registers a BFD session for a peer.
func (m *BFDManager) AddSession(peer netip.Addr, txInterval, rxInterval time.Duration, multiplier uint8) *BFDSession {
	m.mu.Lock()
	defer m.mu.Unlock()

	s := &BFDSession{
		Peer:                 peer,
		State:                BFDStateInit,
		DesiredMinTxInterval: txInterval,
		RequiredMinRxInterval: rxInterval,
		DetectMultiplier:     multiplier,
		lastRx:               time.Now(),
	}
	m.sessions[peer] = s
	m.log.Info("BFD session added",
		zap.String("peer", peer.String()),
		zap.Duration("minTx", txInterval),
		zap.Duration("minRx", rxInterval),
		zap.Uint8("multiplier", multiplier),
	)
	return s
}

// ReceiveHeartbeat records a received BFD control frame from peer.
func (m *BFDManager) ReceiveHeartbeat(peer netip.Addr) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, ok := m.sessions[peer]
	if !ok {
		return
	}

	s.mu.Lock()
	s.lastRx = time.Now()
	s.missedPackets = 0
	if s.State != BFDStateUp {
		s.State = BFDStateUp
		m.log.Info("BFD session transitioned to UP", zap.String("peer", peer.String()))
	}
	s.mu.Unlock()
}

// Tick checks all sessions for timeouts based on the detect multiplier.
func (m *BFDManager) Tick(now time.Time) []netip.Addr {
	m.mu.Lock()
	defer m.mu.Unlock()

	var downedPeers []netip.Addr
	for peer, s := range m.sessions {
		s.mu.Lock()
		if s.State == BFDStateUp {
			timeout := s.RequiredMinRxInterval * time.Duration(s.DetectMultiplier)
			if now.Sub(s.lastRx) > timeout {
				s.State = BFDStateDown
				downedPeers = append(downedPeers, peer)
				m.log.Warn("BFD session detected peer DOWN (sub-second failover)",
					zap.String("peer", peer.String()),
					zap.Duration("elapsed", now.Sub(s.lastRx)),
					zap.Duration("timeout", timeout),
				)
				if m.onDown != nil {
					m.onDown(peer)
				}
			}
		}
		s.mu.Unlock()
	}
	return downedPeers
}

// GetSession returns session state for a peer.
func (m *BFDManager) GetSession(peer netip.Addr) (*BFDSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[peer]
	return s, ok
}
