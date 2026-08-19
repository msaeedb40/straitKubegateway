package bgp_test

import (
	"net/netip"
	"testing"
	"time"

	"github.com/straitKubegateway/straitKubegateway/routing/bgp"
)

func TestBGPPeerStateAndAdvertisement(t *testing.T) {
	speaker := bgp.NewSpeaker(65000, netip.MustParseAddr("10.0.0.1"), false)

	peerCfg := bgp.PeerConfig{
		PeerASN:           65001,
		PeerAddress:       netip.MustParseAddr("10.0.0.2"),
		HoldTime:          90 * time.Second,
		KeepaliveInterval: 30 * time.Second,
	}

	peer := speaker.AddPeer(peerCfg)
	if !peer.IsEstablished() {
		t.Errorf("expected peer state Established, got %s", peer.State)
	}

	// Advertise PodCIDR
	podCIDR := netip.MustParsePrefix("10.244.1.0/24")
	speaker.AdvertisePrefix(podCIDR, netip.MustParseAddr("10.0.0.1"), []uint32{65000})

	if peer.TxCount != 1 {
		t.Errorf("expected 1 advertised route on peer, got %d", peer.TxCount)
	}
}

func TestBGPBestPathSelection(t *testing.T) {
	speaker := bgp.NewSpeaker(65000, netip.MustParseAddr("10.0.0.1"), false)

	peer1 := speaker.AddPeer(bgp.PeerConfig{
		PeerASN:     65001,
		PeerAddress: netip.MustParseAddr("10.0.0.2"),
	})
	peer2 := speaker.AddPeer(bgp.PeerConfig{
		PeerASN:     65002,
		PeerAddress: netip.MustParseAddr("10.0.0.3"),
	})

	dest := netip.MustParsePrefix("198.51.100.0/24")

	// Route 1 from Peer 1: LocalPref 100, AS-Path [65001]
	route1 := bgp.Route{
		Prefix: dest,
		Attributes: bgp.PathAttributes{
			LocalPref: 100,
			ASPath:    []uint32{65001},
			NextHop:   netip.MustParseAddr("10.0.0.2"),
		},
	}
	speaker.IngestRoute(peer1.Config.PeerAddress, route1)

	// Route 2 from Peer 2: LocalPref 200 (Higher), AS-Path [65002, 65003]
	route2 := bgp.Route{
		Prefix: dest,
		Attributes: bgp.PathAttributes{
			LocalPref: 200,
			ASPath:    []uint32{65002, 65003},
			NextHop:   netip.MustParseAddr("10.0.0.3"),
		},
	}
	speaker.IngestRoute(peer2.Config.PeerAddress, route2)

	// Best path should select Route 2 due to higher LocalPref
	best, err := speaker.GetBestRoute(dest)
	if err != nil {
		t.Fatalf("GetBestRoute failed: %v", err)
	}
	if best.Attributes.NextHop.String() != "10.0.0.3" {
		t.Errorf("expected best path NextHop 10.0.0.3 (higher LocalPref), got %s", best.Attributes.NextHop)
	}
}

func TestBGPRouteReflection(t *testing.T) {
	// Enable Route Reflector
	speaker := bgp.NewSpeaker(65000, netip.MustParseAddr("10.0.0.1"), true)

	peer1 := speaker.AddPeer(bgp.PeerConfig{
		PeerASN:     65000,
		PeerAddress: netip.MustParseAddr("10.0.0.10"),
	})
	peer2 := speaker.AddPeer(bgp.PeerConfig{
		PeerASN:     65000,
		PeerAddress: netip.MustParseAddr("10.0.0.20"),
	})

	dest := netip.MustParsePrefix("10.244.2.0/24")
	route := bgp.Route{
		Prefix: dest,
		Attributes: bgp.PathAttributes{
			LocalPref: 100,
			ASPath:    []uint32{65000},
			NextHop:   netip.MustParseAddr("10.0.0.10"),
		},
	}

	speaker.IngestRoute(peer1.Config.PeerAddress, route)

	// Peer 2 should have received the reflected route
	if peer2.TxCount != 1 {
		t.Errorf("expected Peer 2 to receive 1 reflected route, got %d", peer2.TxCount)
	}
}
