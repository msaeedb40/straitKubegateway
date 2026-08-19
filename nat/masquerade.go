package nat

import (
	"net/netip"
	"sync"
)

// MasqueradeRule specifies whether outbound traffic matching non-cluster destinations should be masqueraded.
type MasqueradeRule struct {
	PodCIDR       netip.Prefix
	ExcludedCIDRs []netip.Prefix // e.g. internal cluster subnets (ClusterIPs, other PodCIDRs)
	EgressNodeIP  netip.Addr
}

// MasqueradeEngine determines whether a packet should undergo SNAT masquerading to the host IP.
type MasqueradeEngine struct {
	mu    sync.RWMutex
	rules []MasqueradeRule
}

// NewMasqueradeEngine creates a masquerade rule engine.
func NewMasqueradeEngine() *MasqueradeEngine {
	return &MasqueradeEngine{
		rules: make([]MasqueradeRule, 0),
	}
}

// AddRule registers a masquerade rule for a pod subnet.
func (m *MasqueradeEngine) AddRule(rule MasqueradeRule) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rules = append(m.rules, rule)
}

// ShouldMasquerade evaluates whether traffic from srcIP to dstIP should be masqueraded.
// Returns true and the egress node IP if masquerade applies.
func (m *MasqueradeEngine) ShouldMasquerade(srcIP, dstIP netip.Addr) (bool, netip.Addr) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, r := range m.rules {
		if !r.PodCIDR.Contains(srcIP) {
			continue
		}

		// If destination is in excluded list (e.g. cluster pod or service network), don't masquerade
		isExcluded := false
		for _, exc := range r.ExcludedCIDRs {
			if exc.Contains(dstIP) {
				isExcluded = true
				break
			}
		}

		if !isExcluded {
			return true, r.EgressNodeIP
		}
	}

	return false, netip.Addr{}
}
