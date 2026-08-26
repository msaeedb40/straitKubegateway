// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"net/netip"
	"testing"

	"go.uber.org/zap"

	"github.com/straitkubegateway/straitkubegateway/cni"
	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/compiler"
	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/ir"
	"github.com/straitkubegateway/straitkubegateway/pkg/types"
)

func TestE2EDataplaneLifecycle(t *testing.T) {
	log := zap.NewNop()
	comp := compiler.New(log)
	ctx := context.Background()

	// 1. Define desired network state
	state := &ir.NetworkState{
		Generation: 1,
		Endpoints: map[uint64]*ir.Endpoint{
			1: {
				ID:        1,
				Identity:  types.Identity(200),
				IP:        netip.MustParseAddr("10.244.1.5"),
				Namespace: "default",
				PodName:   "nginx-pod",
			},
		},
		Services: map[ir.ServiceKey]*ir.Service{
			{Namespace: "default", Name: "nginx-svc"}: {
				Key:       ir.ServiceKey{Namespace: "default", Name: "nginx-svc"},
				Type:      ir.ServiceTypeClusterIP,
				ClusterIP: netip.MustParseAddr("10.96.0.100"),
				Ports: []ir.ServicePort{
					{Protocol: types.ProtocolTCP, Port: 80, TargetPort: 80},
				},
				Backends: []ir.Backend{
					{ID: 1, IP: netip.MustParseAddr("10.244.1.5"), Port: 80, Healthy: true},
				},
				Algorithm: types.LBAlgorithmMaglev,
			},
		},
		Policies: []*ir.Policy{
			{
				ID:        "default/allow-http",
				Priority:  10,
				Direction: ir.PolicyDirectionIngress,
				Action:    types.PolicyActionAllow,
			},
		},
	}

	// 2. Compile to dataplane
	if err := comp.Compile(ctx, state); err != nil {
		t.Fatalf("failed to compile dataplane state: %v", err)
	}

	// 3. Verify CNI plugin instance
	plugin := cni.New(log)
	if plugin == nil {
		t.Fatalf("expected non-nil CNI plugin")
	}
}
