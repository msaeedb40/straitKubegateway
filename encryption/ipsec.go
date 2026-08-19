package encryption

import (
	"fmt"
	"net/netip"
	"sync"
)

// IPsecMode represents the IPsec mode (Tunnel or Transport).
type IPsecMode string

const (
	IPsecModeTunnel    IPsecMode = "Tunnel"
	IPsecModeTransport IPsecMode = "Transport"
)

// IPsecPolicy represents an IPsec Security Policy (SP).
type IPsecPolicy struct {
	ID       uint32
	SrcCIDR  netip.Prefix
	DstCIDR  netip.Prefix
	Mode     IPsecMode
	Action   string // Protect, Pass, Discard
	Priority uint32
}

// IPsecSA represents an IPsec Security Association (SA).
type IPsecSA struct {
	SPI        uint32
	LocalIP    netip.Addr
	RemoteIP   netip.Addr
	AuthKey    []byte
	EncryptKey []byte
	Mode       IPsecMode
	AuthAlgo   string // HMAC-SHA256
	CipherAlgo string // AES-GCM-256
}

// IPsecManager manages IPsec policies and security associations.
type IPsecManager struct {
	mu       sync.RWMutex
	policies map[uint32]*IPsecPolicy
	sas      map[uint32]*IPsecSA // SPI -> SA
}

// NewIPsecManager creates a new IPsec manager.
func NewIPsecManager() *IPsecManager {
	return &IPsecManager{
		policies: make(map[uint32]*IPsecPolicy),
		sas:      make(map[uint32]*IPsecSA),
	}
}

// AddPolicy adds an IPsec policy.
func (m *IPsecManager) AddPolicy(policy IPsecPolicy) {
	m.mu.Lock()
	defer m.mu.Unlock()
	pCopy := policy
	m.policies[policy.ID] = &pCopy
}

// AddSA adds an IPsec Security Association.
func (m *IPsecManager) AddSA(sa IPsecSA) {
	m.mu.Lock()
	defer m.mu.Unlock()
	saCopy := sa
	m.sas[sa.SPI] = &saCopy
}

// LookupPolicy finds a matching IPsec policy for source and destination IP.
func (m *IPsecManager) LookupPolicy(srcIP, dstIP netip.Addr) (*IPsecPolicy, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, p := range m.policies {
		if p.SrcCIDR.Contains(srcIP) && p.DstCIDR.Contains(dstIP) {
			return p, nil
		}
	}

	return nil, fmt.Errorf("no matching IPsec policy for %s -> %s", srcIP, dstIP)
}
