package service_test

import (
	"net/netip"
	"testing"

	"github.com/straitKubegateway/straitKubegateway/pkg/types"
	"github.com/straitKubegateway/straitKubegateway/service"
)

func TestKubeProxyReplacement(t *testing.T) {
	nodeIPs := []netip.Addr{
		netip.MustParseAddr("192.168.1.10"),
		netip.MustParseAddr("192.168.1.11"),
	}
	podCIDRs := []netip.Prefix{
		netip.MustParsePrefix("10.244.0.0/16"),
	}

	mgr := service.NewKubeProxyManager(nodeIPs, podCIDRs)

	// 1. Register NodePort Service with DSR enabled
	svc := service.KubeProxyService{
		Namespace: "default",
		Name:      "web-nodeport",
		Type:      service.ServiceTypeNodePort,
		ClusterIP: netip.MustParseAddr("10.96.0.50"),
		Port:      80,
		Protocol:  types.ProtocolTCP,
		NodePorts: map[uint16]uint16{
			30080: 8080,
		},
		LoadBalancerIPs: []netip.Addr{
			netip.MustParseAddr("198.51.100.100"),
		},
		DSR: true,
	}

	backends := []service.BackendEndpoint{
		{IP: netip.MustParseAddr("10.244.1.15"), Port: 8080, Weight: 100},
		{IP: netip.MustParseAddr("10.244.2.20"), Port: 8080, Weight: 100},
	}

	mgr.UpsertKubeService(svc, backends)

	// Test 1: ClusterIP lookup
	clientIP := netip.MustParseAddr("10.244.1.5")
	bCluster, isDSR, isHairpin, err := mgr.LookupDestination(clientIP, netip.MustParseAddr("10.96.0.50"), 80, types.ProtocolTCP)
	if err != nil {
		t.Fatalf("ClusterIP lookup failed: %v", err)
	}
	if bCluster == nil {
		t.Fatalf("expected backend for ClusterIP, got nil")
	}
	if !isDSR {
		t.Errorf("expected DSR=true")
	}
	if isHairpin {
		t.Errorf("expected Hairpin=false for non-self client")
	}

	// Test 2: NodePort lookup (client accessing NodeIP:30080)
	bNodePort, _, _, err := mgr.LookupDestination(clientIP, netip.MustParseAddr("192.168.1.10"), 30080, types.ProtocolTCP)
	if err != nil {
		t.Fatalf("NodePort lookup failed: %v", err)
	}
	if bNodePort == nil {
		t.Fatalf("expected backend for NodePort, got nil")
	}

	// Test 3: LoadBalancer IP lookup
	bLB, _, _, err := mgr.LookupDestination(clientIP, netip.MustParseAddr("198.51.100.100"), 80, types.ProtocolTCP)
	if err != nil {
		t.Fatalf("LoadBalancer lookup failed: %v", err)
	}
	if bLB == nil {
		t.Fatalf("expected backend for LoadBalancer IP, got nil")
	}

	// Test 4: Hairpin NAT detection (Pod 10.244.1.15 accessing ClusterIP that resolves to itself)
	bHairpin, _, isHairpinSelf, err := mgr.LookupDestination(bCluster.IP, netip.MustParseAddr("10.96.0.50"), 80, types.ProtocolTCP)
	if err != nil {
		t.Fatalf("Hairpin lookup failed: %v", err)
	}
	if bHairpin.IP == bCluster.IP && !isHairpinSelf {
		t.Errorf("expected Hairpin=true when client IP is identical to backend IP")
	}
}
