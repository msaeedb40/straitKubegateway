package bgp

import (
	"fmt"
	"net/netip"
	"sort"
	"sync"
)

// Speaker represents a BGP Speaker daemon managing the Routing Information Base (RIB).
type Speaker struct {
	mu          sync.RWMutex
	localASN    uint32
	routerID    netip.Addr
	peers       map[string]*Peer   // "peerAddr" -> Peer
	locRIB      map[string][]Route // "prefix" -> candidates
	bestRIB     map[string]*Route  // "prefix" -> Best Route
	isReflector bool
}

// NewSpeaker creates a new BGP Speaker.
func NewSpeaker(localASN uint32, routerID netip.Addr, isReflector bool) *Speaker {
	return &Speaker{
		localASN:    localASN,
		routerID:    routerID,
		peers:       make(map[string]*Peer),
		locRIB:      make(map[string][]Route),
		bestRIB:     make(map[string]*Route),
		isReflector: isReflector,
	}
}

// AddPeer registers a BGP peer and starts session negotiation.
func (s *Speaker) AddPeer(cfg PeerConfig) *Peer {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg.LocalASN = s.localASN
	peer := NewPeer(cfg)
	peer.Start()
	s.peers[cfg.PeerAddress.String()] = peer
	return peer
}

// RemovePeer removes a BGP peer and clears its learned routes.
func (s *Speaker) RemovePeer(peerAddr netip.Addr) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if peer, exists := s.peers[peerAddr.String()]; exists {
		peer.Stop()
		delete(s.peers, peerAddr.String())
		s.recomputeLocRIBLocked()
	}
}

// IngestRoute processes a newly received route from a peer.
func (s *Speaker) IngestRoute(peerAddr netip.Addr, route Route) {
	s.mu.Lock()
	defer s.mu.Unlock()

	peer, exists := s.peers[peerAddr.String()]
	if !exists || !peer.IsEstablished() {
		return
	}

	peer.ReceiveRoute(route)
	prefixKey := route.Prefix.String()
	s.locRIB[prefixKey] = append(s.locRIB[prefixKey], route)

	// Recompute best path for this prefix
	s.recomputePrefixLocked(prefixKey)

	// If Route Reflector, reflect to other peers
	if s.isReflector {
		s.reflectRouteLocked(peerAddr, route)
	}
}

// AdvertisePrefix advertises a local prefix (PodCIDR, VIP) to all established peers.
func (s *Speaker) AdvertisePrefix(prefix netip.Prefix, nextHop netip.Addr, communities []uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()

	route := Route{
		Prefix: prefix,
		Attributes: PathAttributes{
			Origin:      0, // IGP
			ASPath:      []uint32{s.localASN},
			NextHop:     nextHop,
			LocalPref:   100,
			Communities: communities,
		},
		PeerAddr: s.routerID,
		IsBest:   true,
	}

	for _, peer := range s.peers {
		if peer.IsEstablished() {
			peer.AdvertiseRoute(route)
		}
	}
}

// GetBestRoute returns the best BGP route for a prefix.
func (s *Speaker) GetBestRoute(prefix netip.Prefix) (*Route, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	route, exists := s.bestRIB[prefix.String()]
	if !exists {
		return nil, fmt.Errorf("no best route found for %s", prefix)
	}
	return route, nil
}

// GetAllBestRoutes returns all active best paths in the Loc-RIB.
func (s *Speaker) GetAllBestRoutes() []*Route {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*Route, 0, len(s.bestRIB))
	for _, r := range s.bestRIB {
		res = append(res, r)
	}
	return res
}

func (s *Speaker) recomputePrefixLocked(prefixKey string) {
	candidates := s.locRIB[prefixKey]
	if len(candidates) == 0 {
		delete(s.bestRIB, prefixKey)
		return
	}

	// Best-path selection:
	// 1. Highest LocalPref
	// 2. Shortest AS_PATH
	// 3. Lowest MED
	sort.Slice(candidates, func(i, j int) bool {
		a := candidates[i].Attributes
		b := candidates[j].Attributes

		if a.LocalPref != b.LocalPref {
			return a.LocalPref > b.LocalPref // Higher is better
		}
		if len(a.ASPath) != len(b.ASPath) {
			return len(a.ASPath) < len(b.ASPath) // Shorter is better
		}
		return a.MultiExitDisc < b.MultiExitDisc // Lower is better
	})

	best := candidates[0]
	best.IsBest = true
	s.bestRIB[prefixKey] = &best
}

func (s *Speaker) recomputeLocRIBLocked() {
	s.locRIB = make(map[string][]Route)
	s.bestRIB = make(map[string]*Route)

	for _, peer := range s.peers {
		for _, r := range peer.GetAdjRIBIn() {
			k := r.Prefix.String()
			s.locRIB[k] = append(s.locRIB[k], r)
		}
	}

	for k := range s.locRIB {
		s.recomputePrefixLocked(k)
	}
}

func (s *Speaker) reflectRouteLocked(sender netip.Addr, route Route) {
	for _, peer := range s.peers {
		// Do not reflect back to sender
		if peer.Config.PeerAddress == sender {
			continue
		}
		if peer.IsEstablished() {
			peer.AdvertiseRoute(route)
		}
	}
}
