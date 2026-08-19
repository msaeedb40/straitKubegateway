package transit

import (
	"fmt"
	"net/netip"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/encryption"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// TunnelManager coordinates multi-cluster encrypted tunnels.
type TunnelManager struct {
	mu        sync.RWMutex
	wgConfig  *encryption.WireGuardConfig
	ipsecMgr  *encryption.IPsecManager
	peerLinks map[string]*PeeringLink
}

// NewTunnelManager creates a new tunnel manager.
func NewTunnelManager(wgDevice string, wgPort uint16) (*TunnelManager, error) {
	wg, err := encryption.NewWireGuardConfig(wgDevice, wgPort)
	if err != nil {
		return nil, err
	}

	return &TunnelManager{
		wgConfig:  wg,
		ipsecMgr:  encryption.NewIPsecManager(),
		peerLinks: make(map[string]*PeeringLink),
	}, nil
}

// AddWireGuardPeer connects a remote cluster over WireGuard.
func (tm *TunnelManager) AddWireGuardPeer(link PeeringLink, pubKey string, allowedCIDRs []netip.Prefix) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.peerLinks[link.ID] = &link

	peer := encryption.WireGuardPeer{
		PublicKey:           pubKey,
		Endpoint:            link.RemoteEndpoint,
		AllowedIPs:          allowedCIDRs,
		PersistentKeepalive: 25,
	}
	tm.wgConfig.AddPeer(peer)
}

// WireGuardConfig returns the active WireGuard configuration.
func (tm *TunnelManager) WireGuardConfig() *encryption.WireGuardConfig {
	return tm.wgConfig
}

// LookupTunnel returns the peering link for a remote cluster.
func (tm *TunnelManager) LookupTunnel(remoteCluster string) (*PeeringLink, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	for _, l := range tm.peerLinks {
		if l.RemoteCluster == remoteCluster {
			return l, nil
		}
	}
	return nil, fmt.Errorf("no peering tunnel found for remote cluster %s", remoteCluster)
}

// EncapsulateGeneveHeader creates a simulated Geneve overlay header with SegmentID in VNI.
func EncapsulateGeneveHeader(segmentID types.SegmentID) [8]byte {
	var hdr [8]byte
	hdr[0] = 0    // Version = 0
	hdr[1] = 0    // Options len = 0
	hdr[2] = 0x65 // Protocol = 0x6558 (Ethernet)
	hdr[3] = 0x58

	// VNI is 24-bit in bytes 4..6
	vni := uint32(segmentID) & 0x00FFFFFF
	hdr[4] = byte(vni >> 16)
	hdr[5] = byte(vni >> 8)
	hdr[6] = byte(vni)
	hdr[7] = 0 // Reserved

	return hdr
}
