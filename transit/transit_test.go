package transit_test

import (
	"net/netip"
	"testing"

	"github.com/straitKubegateway/straitKubegateway/encryption"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
	"github.com/straitKubegateway/straitKubegateway/transit"
)

func TestTransitGatewaySegmentIsolationAndRouting(t *testing.T) {
	tg := transit.NewTransitGatewayEngine("tgw-01", "main-transit", 64512)

	// Attachment 1: Segment 100 (Prod)
	attProd := transit.Attachment{
		ID:        "att-prod",
		Name:      "prod-vpc",
		Type:      transit.AttachmentTypeVPC,
		SegmentID: 100,
		Subnets: []netip.Prefix{
			netip.MustParsePrefix("10.100.0.0/16"),
		},
	}
	tg.Attach(attProd)

	// Attachment 2: Segment 200 (Dev)
	attDev := transit.Attachment{
		ID:        "att-dev",
		Name:      "dev-vpc",
		Type:      transit.AttachmentTypeVPC,
		SegmentID: 200,
		Subnets: []netip.Prefix{
			netip.MustParsePrefix("10.200.0.0/16"),
		},
	}
	tg.Attach(attDev)

	// 1. Prod lookup Prod destination -> Match!
	routeProd, err := tg.LookupRoute(100, netip.MustParseAddr("10.100.1.5"))
	if err != nil {
		t.Fatalf("Prod lookup failed: %v", err)
	}
	if routeProd.AttachmentID != "att-prod" {
		t.Errorf("expected attachment att-prod, got %s", routeProd.AttachmentID)
	}

	// 2. Prod lookup Dev destination -> Isolated, Should fail!
	_, errIsolated := tg.LookupRoute(100, netip.MustParseAddr("10.200.1.5"))
	if errIsolated == nil {
		t.Errorf("expected segment isolation error when accessing Dev from Prod")
	}

	// 3. Explicit inter-segment route added
	tg.AddRoute(transit.TransitRoute{
		Destination:  netip.MustParsePrefix("10.200.0.0/16"),
		AttachmentID: "att-dev",
		SegmentID:    100, // Explicitly route into Prod segment
		Metric:       10,
	})

	routeCross, err := tg.LookupRoute(100, netip.MustParseAddr("10.200.1.5"))
	if err != nil {
		t.Fatalf("cross-segment route lookup failed: %v", err)
	}
	if routeCross.AttachmentID != "att-dev" {
		t.Errorf("expected attachment att-dev, got %s", routeCross.AttachmentID)
	}
}

func TestWireGuardConfigAndTunnelManager(t *testing.T) {
	tm, err := transit.NewTunnelManager("wg-test0", 51820)
	if err != nil {
		t.Fatalf("failed to create TunnelManager: %v", err)
	}

	privKey, pubKey, err := encryption.GenerateKeyPair()
	if err != nil || privKey == "" || pubKey == "" {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}

	link := transit.PeeringLink{
		ID:             "peer-cluster-b",
		LocalCluster:   "cluster-a",
		RemoteCluster:  "cluster-b",
		RemoteEndpoint: netip.MustParseAddrPort("203.0.113.10:51820"),
		Encryption:     "WireGuard",
		Segments:       []types.SegmentID{100, 200},
		Healthy:        true,
	}

	tm.AddWireGuardPeer(link, pubKey, []netip.Prefix{
		netip.MustParsePrefix("10.244.0.0/16"),
	})

	foundLink, err := tm.LookupTunnel("cluster-b")
	if err != nil || foundLink.ID != "peer-cluster-b" {
		t.Errorf("LookupTunnel failed to find remote cluster peer")
	}

	peers := tm.WireGuardConfig().GetPeers()
	if len(peers) != 1 {
		t.Fatalf("expected 1 WireGuard peer, got %d", len(peers))
	}
}

func TestGeneveHeaderEncapsulation(t *testing.T) {
	hdr := transit.EncapsulateGeneveHeader(100)
	if hdr[2] != 0x65 || hdr[3] != 0x58 {
		t.Errorf("invalid Geneve protocol type")
	}
	// VNI for segment 100 in bytes 4..6
	vni := (uint32(hdr[4]) << 16) | (uint32(hdr[5]) << 8) | uint32(hdr[6])
	if vni != 100 {
		t.Errorf("expected VNI 100, got %d", vni)
	}
}
