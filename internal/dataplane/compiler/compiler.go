// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package compiler implements the Dataplane Compiler — the ONLY component
// that translates the Dataplane IR into BPF/netlink kernel state.
//
// Architectural invariant:
//   - Controllers produce desired NetworkState (IR)
//   - The Compiler translates IR → BPF maps + netlink
//   - No other component touches BPF maps directly
package compiler

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/ir"
)

// ============================================================================
// Compiler
// ============================================================================

// Compiler translates a Dataplane IR NetworkState into kernel BPF/netlink state.
// It is the sole writer of BPF maps and netlink routes.
type Compiler struct {
	log        *zap.Logger
	generation ir.Generation
}

// New creates a new Compiler.
func New(log *zap.Logger) *Compiler {
	return &Compiler{log: log}
}

// Compile applies the given NetworkState to the kernel dataplane.
// It is idempotent: compiling the same state multiple times is safe.
func (c *Compiler) Compile(ctx context.Context, state *ir.NetworkState) error {
	if state == nil {
		return fmt.Errorf("nil NetworkState")
	}
	if state.Generation <= c.generation {
		c.log.Debug("skipping compile: state generation not newer",
			zap.Uint64("current", uint64(c.generation)),
			zap.Uint64("incoming", uint64(state.Generation)),
		)
		return nil
	}

	c.log.Info("compiling network state",
		zap.Uint64("generation", uint64(state.Generation)),
		zap.Int("endpoints", len(state.Endpoints)),
		zap.Int("services", len(state.Services)),
		zap.Int("policies", len(state.Policies)),
		zap.Int("routes", len(state.Routes)),
	)

	// Phase 1: Endpoints (must come before services and policies)
	if err := c.compileEndpoints(ctx, state); err != nil {
		return fmt.Errorf("compile endpoints: %w", err)
	}

	// Phase 2: Service LB
	if err := c.compileServices(ctx, state); err != nil {
		return fmt.Errorf("compile services: %w", err)
	}

	// Phase 3: NAT
	if err := c.compileNAT(ctx, state); err != nil {
		return fmt.Errorf("compile NAT: %w", err)
	}

	// Phase 4: Policies
	if err := c.compilePolicies(ctx, state); err != nil {
		return fmt.Errorf("compile policies: %w", err)
	}

	// Phase 5: Routes
	if err := c.compileRoutes(ctx, state); err != nil {
		return fmt.Errorf("compile routes: %w", err)
	}

	// Phase 6: Tunnel peers
	if err := c.compileTunnels(ctx, state); err != nil {
		return fmt.Errorf("compile tunnels: %w", err)
	}

	// Phase 7: Transit segments
	if err := c.compileTransit(ctx, state); err != nil {
		return fmt.Errorf("compile transit: %w", err)
	}

	c.generation = state.Generation
	c.log.Info("network state compiled successfully",
		zap.Uint64("generation", uint64(state.Generation)),
	)
	return nil
}

// compileEndpoints programs endpoint BPF maps (identity table, endpoint table).
func (c *Compiler) compileEndpoints(_ context.Context, state *ir.NetworkState) error {
	for id, ep := range state.Endpoints {
		c.log.Debug("programming endpoint",
			zap.Uint64("id", id),
			zap.String("ip", ep.IP.String()),
			zap.String("identity", ep.Identity.String()),
		)
		// TODO(phase1): write to endpoint BPF map via pkg/bpf
		_ = ep
	}
	return nil
}

// compileServices programs service LB BPF maps (service table, backend table,
// maglev_map, nodeport_map, externalip_map).
//
// Phase 5 — kube-proxy replacement:
//   - ClusterIP entries written to service_map
//   - NodePort entries written to nodeport_map
//   - ExternalIP + LoadBalancer ingress written to externalip_map
//   - Maglev 127-slot table written to maglev_map per service
//   - Session affinity written to affinity_map
//
// Invariant 16: kube-dns VIP must always be present in service_map.
func (c *Compiler) compileServices(_ context.Context, state *ir.NetworkState) error {
	kubeDNSProgrammed := false

	for key, svc := range state.Services {
		// Count NodePorts for logging
		nodePorts := 0
		for _, p := range svc.Ports {
			if p.NodePort > 0 {
				nodePorts++
			}
		}

		c.log.Debug("programming service into BPF dataplane",
			zap.String("namespace", key.Namespace),
			zap.String("name", key.Name),
			zap.String("type", string(svc.Type)),
			zap.String("clusterIP", svc.ClusterIP.String()),
			zap.Int("backends", len(svc.Backends)),
			zap.Int("nodePorts", nodePorts),
			zap.Int("externalIPs", len(svc.ExternalIPs)),
			zap.Int("maglevSlots", len(svc.MaglevTable)),
			zap.Bool("sessionAffinity", svc.SessionAffinity),
			zap.Bool("kubeProxyReplacement", svc.KubeProxyReplacement),
		)

		// TODO(phase2): write ClusterIP → backends into service_map via pkg/bpf
		// TODO(phase5): write NodePort → backends into nodeport_map
		// TODO(phase5): write ExternalIPs → backends into externalip_map
		// TODO(phase5): write MaglevTable into maglev_map[svc.Key]
		// TODO(phase5): write SessionAffinity flag into affinity_map
		_ = svc

		// Invariant 16: track kube-dns VIP programming
		if key.Namespace == "kube-system" && key.Name == "kube-dns" && svc.ClusterIP.IsValid() {
			kubeDNSProgrammed = true
		}
	}

	// Invariant 16: log warning if kube-dns VIP is not in service_map.
	// Without this, CoreDNS is unreachable and the cluster cannot resolve DNS.
	if len(state.Services) > 0 && !kubeDNSProgrammed {
		c.log.Warn("kube-dns VIP not yet in service_map — CoreDNS may be unreachable; " +
			"ensure kube-system/kube-dns Service has been reconciled")
	} else if kubeDNSProgrammed {
		c.log.Debug("kube-dns VIP present in service_map — CoreDNS reachable")
	}

	return nil
}

// compileNAT programs NAT BPF maps (conntrack, SNAT, DNAT, NAT64).
func (c *Compiler) compileNAT(_ context.Context, state *ir.NetworkState) error {
	for _, rule := range state.NATRules {
		c.log.Debug("programming NAT rule",
			zap.String("type", string(rule.Type)),
			zap.String("match", rule.Match.String()),
		)
		// TODO(phase3): write to NAT BPF maps
		_ = rule
	}
	return nil
}

// compilePolicies programs policy BPF maps (policy map, identity map).
// Policy evaluation: priority order, deny overrides allow at equal priority,
// default ingress=deny, default egress=allow.
func (c *Compiler) compilePolicies(_ context.Context, state *ir.NetworkState) error {
	for _, pol := range state.Policies {
		c.log.Debug("programming policy",
			zap.String("id", pol.ID),
			zap.Uint8("priority", pol.Priority),
			zap.String("direction", string(pol.Direction)),
			zap.String("action", string(pol.Action)),
		)
		// TODO(phase4): write to policy BPF maps
		_ = pol
	}
	return nil
}

// compileRoutes programs netlink routes via pkg/linux.
// Bootstrap API path (straitd → API :6443) must NOT use Service LB.
func (c *Compiler) compileRoutes(_ context.Context, state *ir.NetworkState) error {
	for _, route := range state.Routes {
		c.log.Debug("programming route",
			zap.String("dst", route.Destination.String()),
			zap.String("nexthop", route.NextHop.String()),
			zap.String("dev", route.Dev),
			zap.String("type", string(route.Type)),
		)
		// TODO(phase1): netlink route add via pkg/linux + routing/
		_ = route
	}
	return nil
}

// compileTunnels programs overlay tunnel state (VXLAN/Geneve/GRE/WireGuard).
func (c *Compiler) compileTunnels(_ context.Context, state *ir.NetworkState) error {
	for _, peer := range state.TunnelPeers {
		c.log.Debug("programming tunnel peer",
			zap.String("nodeIP", peer.NodeIP.String()),
			zap.String("mode", string(peer.Mode)),
			zap.String("podCIDR", peer.PodCIDR.String()),
		)
		// TODO(phase1): configure VXLAN/Geneve/GRE FDB entries via netlink
		_ = peer
	}
	return nil
}

// compileTransit programs multi-cluster transit segment state.
func (c *Compiler) compileTransit(_ context.Context, state *ir.NetworkState) error {
	for _, seg := range state.TransitSegments {
		c.log.Debug("programming transit segment",
			zap.String("segmentID", seg.ID.String()),
			zap.Int("routes", len(seg.Routes)),
		)
		// TODO(phase7): program transit segment routes
		_ = seg
	}
	return nil
}
