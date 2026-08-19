// Package bgp provides the BGP-4 and MP-BGP routing control plane,
// peer state machine, best-path selection, route reflection, and prefix advertisement for straitKubegateway.
package bgp

import (
	"fmt"
	"net/netip"
	"time"
)

// SessionState represents BGP finite state machine states (RFC 4271).
type SessionState string

const (
	StateIdle        SessionState = "Idle"
	StateConnect     SessionState = "Connect"
	StateActive      SessionState = "Active"
	StateOpenSent    SessionState = "OpenSent"
	StateOpenConfirm SessionState = "OpenConfirm"
	StateEstablished SessionState = "Established"
)

// PathAttributes represents BGP path attributes.
type PathAttributes struct {
	Origin              uint8 // 0=IGP, 1=EGP, 2=INCOMPLETE
	ASPath              []uint32
	NextHop             netip.Addr
	MultiExitDisc       uint32 // MED
	LocalPref           uint32
	AtomicAggregate     bool
	Communities         []uint32
	ExtendedCommunities []string
}

// Route represents a BGP route entry in the RIB.
type Route struct {
	Prefix     netip.Prefix
	Attributes PathAttributes
	ReceivedAt time.Time
	PeerAddr   netip.Addr
	IsBest     bool
}

// String returns a human-readable representation of a BGP route.
func (r Route) String() string {
	return fmt.Sprintf("%s via %s (ASPath=%v, LocalPref=%d, MED=%d)",
		r.Prefix, r.Attributes.NextHop, r.Attributes.ASPath, r.Attributes.LocalPref, r.Attributes.MultiExitDisc)
}

// PeerConfig holds configuration for a BGP peering session.
type PeerConfig struct {
	PeerASN              uint32
	LocalASN             uint32
	PeerAddress          netip.Addr
	LocalAddress         netip.Addr
	HoldTime             time.Duration
	KeepaliveInterval    time.Duration
	BFDEnabled           bool
	RouteReflectorClient bool
	AdvertisedPrefixes   []netip.Prefix
}
