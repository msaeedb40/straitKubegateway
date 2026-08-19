// Package nat provides connection tracking, SNAT, DNAT, Masquerading,
// and NAT64 translation for straitKubegateway.
package nat

import (
	"fmt"
	"net/netip"
	"sync"
	"time"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// State represents the connection lifecycle state.
type State uint8

const (
	StateNew         State = 1
	StateEstablished State = 2
	StateReply       State = 3
	StateClosing     State = 4
	StateClosed      State = 5
)

// Tuple represents a 5-tuple identifying a unidirectional flow.
type Tuple struct {
	SrcIP    netip.Addr
	DstIP    netip.Addr
	SrcPort  uint16
	DstPort  uint16
	Protocol types.Protocol
}

// String returns a human-readable 5-tuple string.
func (t Tuple) String() string {
	return fmt.Sprintf("%s:%d -> %s:%d [%s]", t.SrcIP, t.SrcPort, t.DstIP, t.DstPort, t.Protocol)
}

// Entry represents a stateful connection tracking record.
type Entry struct {
	ForwardTuple Tuple
	ReverseTuple Tuple
	State        State
	Flags        uint16
	Created      time.Time
	LastSeen     time.Time
	Packets      uint64
	Bytes        uint64
}

// ConntrackTable manages stateful bidirectional connection entries.
type ConntrackTable struct {
	mu      sync.RWMutex
	forward map[string]*Entry
	reverse map[string]*Entry
	tcpTTL  time.Duration
	udpTTL  time.Duration
}

// NewConntrackTable creates a new connection tracking table.
func NewConntrackTable(tcpTTL, udpTTL time.Duration) *ConntrackTable {
	if tcpTTL <= 0 {
		tcpTTL = 30 * time.Minute
	}
	if udpTTL <= 0 {
		udpTTL = 1 * time.Minute
	}

	return &ConntrackTable{
		forward: make(map[string]*Entry),
		reverse: make(map[string]*Entry),
		tcpTTL:  tcpTTL,
		udpTTL:  udpTTL,
	}
}

// Track registers or updates a connection in the conntrack table.
func (ct *ConntrackTable) Track(fwd, rev Tuple, state State, flags uint16) *Entry {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	now := time.Now()
	fwdKey := fwd.String()
	revKey := rev.String()

	entry, exists := ct.forward[fwdKey]
	if exists {
		entry.State = state
		entry.LastSeen = now
		entry.Packets++
		return entry
	}

	entry = &Entry{
		ForwardTuple: fwd,
		ReverseTuple: rev,
		State:        state,
		Flags:        flags,
		Created:      now,
		LastSeen:     now,
		Packets:      1,
	}

	ct.forward[fwdKey] = entry
	ct.reverse[revKey] = entry
	return entry
}

// LookupForward checks for an existing forward connection.
func (ct *ConntrackTable) LookupForward(t Tuple) (*Entry, bool) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	entry, exists := ct.forward[t.String()]
	return entry, exists
}

// LookupReverse checks for an existing reverse connection.
func (ct *ConntrackTable) LookupReverse(t Tuple) (*Entry, bool) {
	ct.mu.RLock()
	defer ct.mu.RUnlock()

	entry, exists := ct.reverse[t.String()]
	return entry, exists
}

// EvictExpired removes connections exceeding their protocol TTL.
func (ct *ConntrackTable) EvictExpired() int {
	ct.mu.Lock()
	defer ct.mu.Unlock()

	now := time.Now()
	evicted := 0

	for k, entry := range ct.forward {
		ttl := ct.tcpTTL
		if entry.ForwardTuple.Protocol == types.ProtocolUDP {
			ttl = ct.udpTTL
		}

		if now.Sub(entry.LastSeen) > ttl {
			delete(ct.forward, k)
			delete(ct.reverse, entry.ReverseTuple.String())
			evicted++
		}
	}

	return evicted
}

// ActiveConnections returns the total number of tracked connections.
func (ct *ConntrackTable) ActiveConnections() int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return len(ct.forward)
}
