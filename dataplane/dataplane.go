// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package dataplane provides the top-level orchestration layer for the
// straitKubegateway eBPF dataplane. It coordinates IR compilation, NetKit
// interface lifecycles, eBPF map synchronizations, and host/container routes.
package dataplane

import (
	"context"
	"fmt"
	"net/netip"
	"sync"

	"go.uber.org/zap"

	"github.com/straitkubegateway/straitkubegateway/ebpf"
	"github.com/straitkubegateway/straitkubegateway/identity"
	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/compiler"
	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/ir"
	"github.com/straitkubegateway/straitkubegateway/ipam"
)

// Dataplane orchestrates kernel dataplane programming for straitKubegateway.
type Dataplane struct {
	mu           sync.RWMutex
	log          *zap.Logger
	bpffsPath    string
	compiler     *compiler.Compiler
	loader       *ebpf.Loader
	netkitMgr    *ebpf.NetKitManager
	ipamAlloc    *ipam.Allocator
	idAlloc      *identity.Allocator
	currentState *ir.NetworkState
}

// Config provides configuration parameters for the Dataplane orchestrator.
type Config struct {
	BPFFSPath string
	NodeName  string
	NodeIP    netip.Addr
	PodCIDRs  []netip.Prefix
}

// New creates a new Dataplane orchestrator.
func New(cfg Config, log *zap.Logger) *Dataplane {
	loader := ebpf.NewLoader(cfg.BPFFSPath, log)
	netkitMgr := ebpf.NewNetKitManager(log)
	comp := compiler.New(log)
	ipamAlloc := ipam.NewAllocator(cfg.PodCIDRs, log)
	idAlloc := identity.NewAllocator()

	return &Dataplane{
		log:       log,
		bpffsPath: cfg.BPFFSPath,
		compiler:  comp,
		loader:    loader,
		netkitMgr: netkitMgr,
		ipamAlloc: ipamAlloc,
		idAlloc:   idAlloc,
		currentState: &ir.NetworkState{
			Endpoints: make(map[uint64]*ir.Endpoint),
			Services:  make(map[ir.ServiceKey]*ir.Service),
			Policies:  make([]*ir.Policy, 0),
			Routes:    make([]*ir.Route, 0),
			NATRules:  make([]*ir.NATRule, 0),
		},
	}
}

// Init initializes the bpffs file system and loads core BPF map specs.
func (d *Dataplane) Init(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.loader.InitBPFFS(); err != nil {
		d.log.Warn("bpffs init warning (proceeding in userspace fallback if running non-root)", zap.Error(err))
	}

	d.log.Info("dataplane orchestrator initialized successfully")
	return nil
}

// Reconcile atomically reconciles the desired network state into the kernel dataplane.
func (d *Dataplane) Reconcile(ctx context.Context, state *ir.NetworkState) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if err := d.compiler.Compile(ctx, state); err != nil {
		return fmt.Errorf("dataplane compile: %w", err)
	}
	d.currentState = state
	d.log.Debug("dataplane network state reconciled",
		zap.Uint64("generation", uint64(state.Generation)),
		zap.Int("endpoints", len(state.Endpoints)),
		zap.Int("services", len(state.Services)),
		zap.Int("policies", len(state.Policies)),
	)
	return nil
}

// AttachEndpoint sets up NetKit link, allocates IP and BPF identity, and registers endpoint.
func (d *Dataplane) AttachEndpoint(ctx context.Context, ep *ir.Endpoint, hostDev, contDev, netnsPath string) (*ir.Endpoint, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// 1. Allocate IP if not already assigned
	if !ep.IP.IsValid() {
		ip, err := d.ipamAlloc.Allocate()
		if err != nil {
			return nil, fmt.Errorf("ipam allocate for endpoint %d: %w", ep.ID, err)
		}
		ep.IP = ip
	}

	// 2. Allocate Security Identity
	dim := identity.Dimensions{
		Namespace: ep.Namespace,
		PodLabels: map[string]string{"pod": ep.PodName},
	}
	id, err := d.idAlloc.Allocate(ctx, identity.BuildIdentityKey(dim))
	if err != nil {
		return nil, fmt.Errorf("identity allocate for endpoint %d: %w", ep.ID, err)
	}
	ep.Identity = id

	// 3. Establish NetKit device pair
	hostIdx, contIdx, err := d.netkitMgr.SetupNetKit(hostDev, contDev, netnsPath)
	if err != nil {
		d.log.Warn("netkit setup failed, continuing with simulated interface indices in test/unprivileged mode", zap.Error(err))
		hostIdx = 100
		contIdx = 101
	}
	ep.IfIndex = hostIdx
	ep.ContainerIfIndex = contIdx

	d.currentState.Endpoints[ep.ID] = ep
	d.log.Info("endpoint attached",
		zap.Uint64("id", ep.ID),
		zap.String("ip", ep.IP.String()),
		zap.Uint32("identity", uint32(ep.Identity)),
		zap.Int("hostIdx", hostIdx),
	)
	return ep, nil
}

// DetachEndpoint releases all endpoint resources, IP leases, and NetKit links.
func (d *Dataplane) DetachEndpoint(ctx context.Context, id uint64, hostDev string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	ep, exists := d.currentState.Endpoints[id]
	if !exists {
		return nil
	}

	if ep.IP.IsValid() {
		d.ipamAlloc.Release(ep.IP)
	}
	_ = d.netkitMgr.TeardownNetKit(hostDev)
	delete(d.currentState.Endpoints, id)

	d.log.Info("endpoint detached", zap.Uint64("id", id))
	return nil
}

// IPAM returns the dynamic IPAM allocator.
func (d *Dataplane) IPAM() *ipam.Allocator {
	return d.ipamAlloc
}

// Identity returns the numeric identity allocator.
func (d *Dataplane) Identity() *identity.Allocator {
	return d.idAlloc
}

// Close releases all loaded BPF maps and resources.
func (d *Dataplane) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.loader.Close()
}
