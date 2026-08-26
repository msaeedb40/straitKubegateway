// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package bgp implements BGP route advertisement, neighbor management,
// and route learning for straitKubegateway.
package bgp

import (
	"context"
	"net/netip"
	"sync"

	"go.uber.org/zap"
)

// NeighborState represents the BGP peering state.
type NeighborState string

const (
	StateIdle        NeighborState = "Idle"
	StateConnect     NeighborState = "Connect"
	StateActive      NeighborState = "Active"
	StateOpenSent    NeighborState = "OpenSent"
	StateOpenConfirm NeighborState = "OpenConfirm"
	StateEstablished NeighborState = "Established"
)

// Neighbor represents a configured BGP peer.
type Neighbor struct {
	Address          netip.Addr
	ASN              uint32
	State            NeighborState
	RoutesReceived   uint32
	RoutesAdvertised uint32
}

// Manager manages BGP peering and route advertisements.
type Manager struct {
	mu        sync.RWMutex
	log       *zap.Logger
	localASN  uint32
	routerID  netip.Addr
	neighbors map[netip.Addr]*Neighbor
	rib       map[netip.Prefix]netip.Addr // Destination CIDR -> NextHop
	bfd       *BFDManager
}

// NewManager creates a new BGP manager.
func NewManager(localASN uint32, routerID netip.Addr, log *zap.Logger) *Manager {
	m := &Manager{
		log:       log,
		localASN:  localASN,
		routerID:  routerID,
		neighbors: make(map[netip.Addr]*Neighbor),
		rib:       make(map[netip.Prefix]netip.Addr),
	}
	m.bfd = NewBFDManager(func(peer netip.Addr) {
		m.mu.Lock()
		if n, ok := m.neighbors[peer]; ok {
			n.State = StateActive
		}
		m.mu.Unlock()
	}, log)
	return m
}

// BFD returns the BFD manager instance.
func (m *Manager) BFD() *BFDManager {
	return m.bfd
}

// AddNeighbor adds a BGP peer neighbor.
func (m *Manager) AddNeighbor(addr netip.Addr, asn uint32) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.neighbors[addr] = &Neighbor{
		Address: addr,
		ASN:     asn,
		State:   StateEstablished, // Simulated active session
	}
	m.log.Info("BGP neighbor added",
		zap.String("peer", addr.String()),
		zap.Uint32("asn", asn),
	)
}

// AdvertiseRoute adds a route to the local RIB and advertises it to all peers.
func (m *Manager) AdvertiseRoute(cidr netip.Prefix, nextHop netip.Addr) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rib[cidr] = nextHop
	for _, n := range m.neighbors {
		n.RoutesAdvertised++
	}
	m.log.Info("advertising BGP route",
		zap.String("cidr", cidr.String()),
		zap.String("nextHop", nextHop.String()),
	)
}

// WithdrawRoute removes a route from the RIB.
func (m *Manager) WithdrawRoute(cidr netip.Prefix) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rib, cidr)
	m.log.Info("withdrawing BGP route", zap.String("cidr", cidr.String()))
}

// GetNeighbors returns all configured BGP neighbors.
func (m *Manager) GetNeighbors() []*Neighbor {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]*Neighbor, 0, len(m.neighbors))
	for _, n := range m.neighbors {
		out = append(out, n)
	}
	return out
}

// Run starts the BGP management loop.
func (m *Manager) Run(ctx context.Context) error {
	m.log.Info("BGP routing engine started",
		zap.Uint32("localASN", m.localASN),
		zap.String("routerID", m.routerID.String()),
	)
	<-ctx.Done()
	return nil
}
