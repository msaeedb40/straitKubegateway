package controllers

import (
	"context"
	"fmt"
	"net/netip"
	"sync"
	"time"

	"github.com/straitKubegateway/straitKubegateway/api/v1alpha1"
	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
	"github.com/straitKubegateway/straitKubegateway/routing/bgp"
)

// BGPController reconciles BGPPeer CRDs and configures the BGP Speaker.
type BGPController struct {
	mu      sync.RWMutex
	speaker *bgp.Speaker
	logger  *logging.Logger
}

// NewBGPController creates a new BGP controller.
func NewBGPController(speaker *bgp.Speaker) *BGPController {
	return &BGPController{
		speaker: speaker,
		logger:  logging.DefaultLogger(),
	}
}

// ReconcileBGPPeer reconciles a BGPPeer CRD.
func (c *BGPController) ReconcileBGPPeer(ctx context.Context, peer *v1alpha1.BGPPeer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	peerIP, err := netip.ParseAddr(peer.Spec.PeerAddress)
	if err != nil {
		return fmt.Errorf("invalid BGP peer address %q: %w", peer.Spec.PeerAddress, err)
	}

	var localIP netip.Addr
	if peer.Spec.LocalAddress != "" {
		localIP, _ = netip.ParseAddr(peer.Spec.LocalAddress)
	}

	var prefixes []netip.Prefix
	for _, pStr := range peer.Spec.AdvertisedPrefixes {
		p, err := netip.ParsePrefix(pStr)
		if err == nil {
			prefixes = append(prefixes, p)
		}
	}

	cfg := bgp.PeerConfig{
		PeerASN:            peer.Spec.PeerASN,
		LocalASN:           peer.Spec.LocalASN,
		PeerAddress:        peerIP,
		LocalAddress:       localIP,
		HoldTime:           time.Duration(peer.Spec.HoldTime) * time.Second,
		KeepaliveInterval:  time.Duration(peer.Spec.KeepaliveInterval) * time.Second,
		BFDEnabled:         peer.Spec.BFDEnabled,
		AdvertisedPrefixes: prefixes,
	}

	c.speaker.AddPeer(cfg)
	c.logger.Info(fmt.Sprintf("reconciled BGPPeer %s (PeerASN=%d, PeerAddress=%s, BFD=%v)",
		peer.Name, peer.Spec.PeerASN, peer.Spec.PeerAddress, peer.Spec.BFDEnabled), &types.Metadata{
		Component: "bgp-controller",
	})

	return nil
}

// DeleteBGPPeer removes a BGP peer from the speaker.
func (c *BGPController) DeleteBGPPeer(ctx context.Context, peerAddress string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ip, err := netip.ParseAddr(peerAddress); err == nil {
		c.speaker.RemovePeer(ip)
		c.logger.Info(fmt.Sprintf("deleted BGPPeer %s", peerAddress), &types.Metadata{
			Component: "bgp-controller",
		})
	}
}
