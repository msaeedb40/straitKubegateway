package nat

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// SNATRule defines a Source NAT translation mapping.
type SNATRule struct {
	SourceCIDR netip.Prefix
	EgressIP   netip.Addr
	PortMin    uint16
	PortMax    uint16
}

// SNATEngine handles outbound IP and ephemeral port translation.
type SNATEngine struct {
	mu        sync.Mutex
	rules     []SNATRule
	allocated map[string]uint16 // "proto:egressIP:port" -> allocated
	nextPort  uint16
	portMin   uint16
	portMax   uint16
}

// NewSNATEngine creates a Source NAT engine with ephemeral port ranges.
func NewSNATEngine(portMin, portMax uint16) *SNATEngine {
	if portMin == 0 {
		portMin = 32768
	}
	if portMax == 0 {
		portMax = 65535
	}

	return &SNATEngine{
		rules:     make([]SNATRule, 0),
		allocated: make(map[string]uint16),
		nextPort:  portMin,
		portMin:   portMin,
		portMax:   portMax,
	}
}

// AddRule adds an SNAT rule.
func (e *SNATEngine) AddRule(rule SNATRule) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if rule.PortMin == 0 {
		rule.PortMin = e.portMin
	}
	if rule.PortMax == 0 {
		rule.PortMax = e.portMax
	}
	e.rules = append(e.rules, rule)
}

// Translate allocates a translated egress IP and ephemeral port for a source packet.
func (e *SNATEngine) Translate(srcIP netip.Addr, srcPort uint16, proto types.Protocol) (netip.Addr, uint16, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	var matchingRule *SNATRule
	for _, r := range e.rules {
		if r.SourceCIDR.Contains(srcIP) {
			ruleCopy := r
			matchingRule = &ruleCopy
			break
		}
	}

	if matchingRule == nil {
		return netip.Addr{}, 0, fmt.Errorf("no matching SNAT rule for source IP %s", srcIP)
	}

	// Allocate next available ephemeral port
	portRange := int(matchingRule.PortMax - matchingRule.PortMin + 1)
	for i := 0; i < portRange; i++ {
		port := e.nextPort
		e.nextPort++
		if e.nextPort > matchingRule.PortMax {
			e.nextPort = matchingRule.PortMin
		}

		key := fmt.Sprintf("%s:%s:%d", proto, matchingRule.EgressIP, port)
		if _, inUse := e.allocated[key]; !inUse {
			e.allocated[key] = port
			return matchingRule.EgressIP, port, nil
		}
	}

	return netip.Addr{}, 0, fmt.Errorf("SNAT port pool exhausted on %s", matchingRule.EgressIP)
}

// ReleasePort frees an allocated ephemeral port.
func (e *SNATEngine) ReleasePort(egressIP netip.Addr, port uint16, proto types.Protocol) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := fmt.Sprintf("%s:%s:%d", proto, egressIP, port)
	delete(e.allocated, key)
}
