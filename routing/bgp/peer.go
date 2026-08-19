package bgp

import (
	"sync"
	"time"
)

// Peer represents an individual BGP peering session.
type Peer struct {
	mu             sync.RWMutex
	Config         PeerConfig
	State          SessionState
	Uptime         time.Time
	LastSeen       time.Time
	AdjRIBIn       []Route
	AdjRIBOut      []Route
	RxCount        uint64
	TxCount        uint64
	keepaliveTimer *time.Timer
	holdTimer      *time.Timer
}

// NewPeer creates a new BGP peer.
func NewPeer(cfg PeerConfig) *Peer {
	if cfg.HoldTime <= 0 {
		cfg.HoldTime = 90 * time.Second
	}
	if cfg.KeepaliveInterval <= 0 {
		cfg.KeepaliveInterval = 30 * time.Second
	}

	return &Peer{
		Config:    cfg,
		State:     StateIdle,
		AdjRIBIn:  make([]Route, 0),
		AdjRIBOut: make([]Route, 0),
	}
}

// Start begins the peering session transition to Established.
func (p *Peer) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.State = StateEstablished
	p.Uptime = time.Now()
	p.LastSeen = time.Now()
}

// Stop terminates the peering session.
func (p *Peer) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.State = StateIdle
	p.AdjRIBIn = nil
	p.AdjRIBOut = nil
}

// ReceiveRoute adds a route to the peer's Adj-RIB-In.
func (p *Peer) ReceiveRoute(route Route) {
	p.mu.Lock()
	defer p.mu.Unlock()

	route.ReceivedAt = time.Now()
	route.PeerAddr = p.Config.PeerAddress
	p.AdjRIBIn = append(p.AdjRIBIn, route)
	p.RxCount++
	p.LastSeen = time.Now()
}

// GetAdjRIBIn returns the inbound routes received from this peer.
func (p *Peer) GetAdjRIBIn() []Route {
	p.mu.RLock()
	defer p.mu.RUnlock()

	res := make([]Route, len(p.AdjRIBIn))
	copy(res, p.AdjRIBIn)
	return res
}

// AdvertiseRoute adds a route to the peer's Adj-RIB-Out.
func (p *Peer) AdvertiseRoute(route Route) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.AdjRIBOut = append(p.AdjRIBOut, route)
	p.TxCount++
}

// IsEstablished returns true if the BGP session is currently Established.
func (p *Peer) IsEstablished() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.State == StateEstablished
}
