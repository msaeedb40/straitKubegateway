package nat

import (
	"net/netip"
	"sync"
	"time"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// Manager coordinates stateful connection tracking, SNAT, DNAT, Masquerade, and NAT64.
type Manager struct {
	mu         sync.RWMutex
	conntrack  *ConntrackTable
	snat       *SNATEngine
	dnat       *DNATEngine
	masquerade *MasqueradeEngine
	nat64      *NAT64Engine
}

// Config holds configuration parameters for the NAT Manager.
type Config struct {
	TCPTimeout    time.Duration
	UDPTimeout    time.Duration
	PortMin       uint16
	PortMax       uint16
	NAT64Prefix   string
	NAT64IPv4Pool string
}

// NewManager creates a fully initialized NAT manager.
func NewManager(cfg Config) (*Manager, error) {
	nat64Eng, err := NewNAT64Engine(cfg.NAT64Prefix, cfg.NAT64IPv4Pool)
	if err != nil {
		return nil, err
	}

	return &Manager{
		conntrack:  NewConntrackTable(cfg.TCPTimeout, cfg.UDPTimeout),
		snat:       NewSNATEngine(cfg.PortMin, cfg.PortMax),
		dnat:       NewDNATEngine(),
		masquerade: NewMasqueradeEngine(),
		nat64:      nat64Eng,
	}, nil
}

// Conntrack returns the connection tracking table.
func (m *Manager) Conntrack() *ConntrackTable {
	return m.conntrack
}

// SNAT returns the SNAT engine.
func (m *Manager) SNAT() *SNATEngine {
	return m.snat
}

// DNAT returns the DNAT engine.
func (m *Manager) DNAT() *DNATEngine {
	return m.dnat
}

// Masquerade returns the Masquerade engine.
func (m *Manager) Masquerade() *MasqueradeEngine {
	return m.masquerade
}

// NAT64 returns the NAT64 engine.
func (m *Manager) NAT64() *NAT64Engine {
	return m.nat64
}

// ProcessEgressPacket applies NAT rules (SNAT or Masquerade) and registers conntrack state.
func (m *Manager) ProcessEgressPacket(srcIP, dstIP netip.Addr, srcPort, dstPort uint16, proto types.Protocol) (netip.Addr, uint16, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	fwdTuple := Tuple{
		SrcIP:    srcIP,
		DstIP:    dstIP,
		SrcPort:  srcPort,
		DstPort:  dstPort,
		Protocol: proto,
	}

	// 1. Check existing connection in Conntrack
	if entry, exists := m.conntrack.LookupForward(fwdTuple); exists {
		return entry.ReverseTuple.DstIP, entry.ReverseTuple.DstPort, true, nil
	}

	// 2. Check Masquerade
	if shouldMasq, nodeIP := m.masquerade.ShouldMasquerade(srcIP, dstIP); shouldMasq {
		// Use SNAT engine on nodeIP
		m.snat.AddRule(SNATRule{
			SourceCIDR: netip.PrefixFrom(srcIP, 32),
			EgressIP:   nodeIP,
		})

		natIP, natPort, err := m.snat.Translate(srcIP, srcPort, proto)
		if err != nil {
			return srcIP, srcPort, false, err
		}

		revTuple := Tuple{
			SrcIP:    dstIP,
			DstIP:    natIP,
			SrcPort:  dstPort,
			DstPort:  natPort,
			Protocol: proto,
		}

		m.conntrack.Track(fwdTuple, revTuple, StateEstablished, 1)
		return natIP, natPort, true, nil
	}

	return srcIP, srcPort, false, nil
}
