package nat

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// DNATRule maps an external VIP and port to an internal destination IP and port.
type DNATRule struct {
	VIP          netip.Addr
	Port         uint16
	Protocol     types.Protocol
	TargetIP     netip.Addr
	TargetPort   uint16
	TargetPrefix *netip.Prefix
}

// DNATEngine handles destination address and port translation.
type DNATEngine struct {
	mu    sync.RWMutex
	rules map[string]DNATRule // "vip:port/proto" -> DNATRule
}

// NewDNATEngine creates a new DNAT engine.
func NewDNATEngine() *DNATEngine {
	return &DNATEngine{
		rules: make(map[string]DNATRule),
	}
}

// AddRule registers a DNAT mapping rule.
func (e *DNATEngine) AddRule(rule DNATRule) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := fmt.Sprintf("%s:%d/%s", rule.VIP, rule.Port, rule.Protocol)
	e.rules[key] = rule
}

// DeleteRule removes a DNAT mapping rule.
func (e *DNATEngine) DeleteRule(vip netip.Addr, port uint16, proto types.Protocol) {
	e.mu.Lock()
	defer e.mu.Unlock()

	key := fmt.Sprintf("%s:%d/%s", vip, port, proto)
	delete(e.rules, key)
}

// Translate checks for a DNAT rule matching the destination and returns the target IP:Port.
func (e *DNATEngine) Translate(dstIP netip.Addr, dstPort uint16, proto types.Protocol) (netip.Addr, uint16, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	key := fmt.Sprintf("%s:%d/%s", dstIP, dstPort, proto)
	rule, exists := e.rules[key]
	if !exists {
		return netip.Addr{}, 0, fmt.Errorf("no DNAT rule matching %s:%d/%s", dstIP, dstPort, proto)
	}

	targetPort := rule.TargetPort
	if targetPort == 0 {
		targetPort = dstPort
	}

	return rule.TargetIP, targetPort, nil
}
