// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package cni

import (
	"context"
	"net/netip"
	"strings"
	"testing"

	"go.uber.org/zap"

	"github.com/straitkubegateway/straitkubegateway/identity"
	"github.com/straitkubegateway/straitkubegateway/ipam"
)

func TestParseConfig(t *testing.T) {
	raw := []byte(`{
		"cniVersion": "1.1.0",
		"name": "strait-net",
		"type": "straitkubegateway",
		"mtu": 1500,
		"bpffsPath": "/sys/fs/bpf"
	}`)

	cfg, err := ParseConfig(raw)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if cfg.Name != "strait-net" {
		t.Errorf("got name %q, want strait-net", cfg.Name)
	}
	if cfg.CNIVersion != "1.1.0" {
		t.Errorf("got cniVersion %q, want 1.1.0", cfg.CNIVersion)
	}
	if cfg.MTU != 1500 {
		t.Errorf("got MTU %d, want 1500", cfg.MTU)
	}
}

func TestPluginVersion(t *testing.T) {
	plugin := New(zap.NewNop())
	versions := plugin.VERSION()
	if len(versions) == 0 {
		t.Fatal("expected non-empty CNI version list")
	}

	found110 := false
	for _, v := range versions {
		if v == "1.1.0" {
			found110 = true
			break
		}
	}
	if !found110 {
		t.Errorf("expected 1.1.0 in SupportedVersions: %v", versions)
	}
}

func TestPluginLifecycle(t *testing.T) {
	log := zap.NewNop()
	cidrs := []netip.Prefix{netip.MustParsePrefix("10.244.1.0/24")}
	ipamAlloc := ipam.NewAllocator(cidrs, log)
	identAlloc := identity.NewAllocator()

	plugin := New(log).WithIPAM(ipamAlloc).WithIdentity(identAlloc)

	cfg := &NetworkConfig{
		CNIVersion: "1.1.0",
		Name:       "strait-net",
		Type:       "straitkubegateway",
	}

	containerID := "container-test-01"
	netns := "/proc/1/ns/net"
	ifName := "eth0"

	// 1. ADD
	res, err := plugin.ADD(cfg, netns, containerID, ifName)
	if err != nil {
		// NetKit/veth creation requires CAP_NET_ADMIN — skip in unprivileged environments.
		if strings.Contains(err.Error(), "operation not permitted") ||
			strings.Contains(err.Error(), "not permitted") {
			t.Skipf("skipping: veth creation requires CAP_NET_ADMIN (root): %v", err)
		}
		t.Fatalf("CNI ADD failed: %v", err)
	}

	if !res.IP.IsValid() {
		t.Errorf("invalid allocated IP")
	}
	if !cidrs[0].Contains(res.IP) {
		t.Errorf("allocated IP %s not in pod CIDR %s", res.IP, cidrs[0])
	}
	if res.Gateway.String() != "10.244.1.1" {
		t.Errorf("got gateway %s, want 10.244.1.1", res.Gateway)
	}

	// 2. CHECK
	if err := plugin.CHECK(cfg, netns, containerID, ifName); err != nil {
		t.Errorf("CNI CHECK failed: %v", err)
	}

	// 3. GC with active container
	ctx := context.Background()
	if err := plugin.GC(ctx, map[string]bool{containerID: true}); err != nil {
		t.Errorf("CNI GC failed: %v", err)
	}

	// 4. DEL
	if err := plugin.DEL(cfg, netns, containerID, ifName); err != nil {
		t.Errorf("CNI DEL failed: %v", err)
	}
}
