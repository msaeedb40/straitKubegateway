// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package dataplane

import (
	"context"
	"net/netip"
	"testing"

	"go.uber.org/zap"

	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/ir"
	"github.com/straitkubegateway/straitkubegateway/pkg/types"
)

func TestDataplaneOrchestrator(t *testing.T) {
	ctx := context.Background()
	log := zap.NewNop()

	cfg := Config{
		BPFFSPath: "/sys/fs/bpf",
		NodeName:  "worker-01",
		NodeIP:    netip.MustParseAddr("192.168.1.10"),
		PodCIDRs:  []netip.Prefix{netip.MustParsePrefix("10.244.1.0/24")},
	}

	dp := New(cfg, log)
	if err := dp.Init(ctx); err != nil {
		t.Fatalf("dp init failed: %v", err)
	}

	// Attach endpoint
	ep := &ir.Endpoint{
		ID:        1001,
		Namespace: "default",
		PodName:   "nginx-demo",
	}

	attached, err := dp.AttachEndpoint(ctx, ep, "nk-h-1001", "nk-c-1001", "")
	if err != nil {
		t.Fatalf("attach endpoint failed: %v", err)
	}

	if !attached.IP.IsValid() {
		t.Fatalf("expected valid allocated IP")
	}
	if attached.Identity < types.IdentityMin {
		t.Fatalf("expected valid allocated identity >= %d, got %d", types.IdentityMin, attached.Identity)
	}

	// Reconcile state with service
	state := &ir.NetworkState{
		Generation: 1,
		Endpoints: map[uint64]*ir.Endpoint{
			attached.ID: attached,
		},
		Services: map[ir.ServiceKey]*ir.Service{
			{Namespace: "default", Name: "nginx-svc"}: {
				Key:       ir.ServiceKey{Namespace: "default", Name: "nginx-svc"},
				Type:      ir.ServiceTypeClusterIP,
				ClusterIP: netip.MustParseAddr("10.96.0.1"),
				Ports: []ir.ServicePort{
					{Protocol: types.ProtocolTCP, Port: 80, TargetPort: 80},
				},
				Backends: []ir.Backend{
					{ID: 1, IP: attached.IP, Port: 80, Healthy: true},
				},
			},
		},
	}

	if err := dp.Reconcile(ctx, state); err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	// Detach endpoint
	if err := dp.DetachEndpoint(ctx, 1001, "nk-h-1001"); err != nil {
		t.Fatalf("detach failed: %v", err)
	}

	_ = dp.Close()
}
