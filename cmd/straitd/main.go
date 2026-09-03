// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// straitd is the per-node agent daemon for straitKubegateway.
// It orchestrates CNI, eBPF loaders, NetKit, TCX, XDP, Service LB,
// NAT, Routing, WireGuard, Policy, cgroup v2, systemd, and Metrics.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"github.com/straitkubegateway/straitkubegateway/cni"
	"github.com/straitkubegateway/straitkubegateway/encryption"
	"github.com/straitkubegateway/straitkubegateway/identity"
	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/compiler"
	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/ir"
	"github.com/straitkubegateway/straitkubegateway/internal/version"
	"github.com/straitkubegateway/straitkubegateway/nat"
	"github.com/straitkubegateway/straitkubegateway/observability"
	"github.com/straitkubegateway/straitkubegateway/pkg/linux"
	"github.com/straitkubegateway/straitkubegateway/platform"
	"github.com/straitkubegateway/straitkubegateway/routing"
)

type config struct {
	nodeName      string
	nodeIP        string
	podCIDR       string
	serviceCIDR   string
	apiServerAddr string
	apiServerPort int
	metricsAddr   string
	debug         bool
	bpffsPath     string
	tunnelMode    string
	enableIPv6    bool
	wireguardPort int
	kubeProxyMode string
	cniConfDir    string
	cniBinDir     string
}

func parseFlags() *config {
	cfg := &config{}
	flag.StringVar(&cfg.nodeName, "node-name", os.Getenv("NODE_NAME"), "Kubernetes node name")
	flag.StringVar(&cfg.nodeIP, "node-ip", os.Getenv("NODE_IP"), "Node IP address")
	flag.StringVar(&cfg.podCIDR, "pod-cidr", "", "Pod CIDR allocated to this node (auto-discovered if empty)")
	flag.StringVar(&cfg.serviceCIDR, "service-cidr", "", "Cluster service CIDR (auto-discovered if empty)")
	flag.StringVar(&cfg.apiServerAddr, "api-server-addr", os.Getenv("KUBERNETES_SERVICE_HOST"), "Kubernetes API server IP")
	flag.IntVar(&cfg.apiServerPort, "api-server-port", 6443, "Kubernetes API server port")
	flag.StringVar(&cfg.metricsAddr, "metrics-addr", ":9090", "Prometheus metrics server address")
	flag.BoolVar(&cfg.debug, "debug", false, "Enable debug logging")
	flag.StringVar(&cfg.bpffsPath, "bpffs-path", linux.BPFFSDefaultPath, "Mount path for bpffs")
	flag.StringVar(&cfg.tunnelMode, "tunnel-mode", "vxlan", "Overlay tunnel mode (vxlan, geneve, gre, disabled)")
	flag.BoolVar(&cfg.enableIPv6, "enable-ipv6", false, "Enable IPv6 dual-stack")
	flag.IntVar(&cfg.wireguardPort, "wireguard-port", 51820, "WireGuard listen UDP port")
	flag.StringVar(&cfg.kubeProxyMode, "kube-proxy-mode", "none", "Kube-proxy replacement mode (none, partial)")
	flag.StringVar(&cfg.cniConfDir, "cni-conf-dir", cni.DefaultCNIConfDir, "CNI configuration directory")
	flag.StringVar(&cfg.cniBinDir, "cni-bin-dir", cni.DefaultCNIBinDir, "CNI plugin binary directory")
	flag.Parse()
	return cfg
}

func main() {
	cfg := parseFlags()

	// 1. Structured Logging
	logger, err := observability.NewLogger(cfg.debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = logger.Sync() }()

	vInfo := version.Get()
	logger.Info("starting straitd node agent",
		zap.String("version", vInfo.Version),
		zap.String("commit", vInfo.GitCommit),
		zap.String("node", cfg.nodeName),
	)

	// 2. Kernel version verification (>= 6.7 minimum, 6.12 LTS baseline)
	if err := linux.CheckKernelVersion(); err != nil {
		logger.Warn("kernel version check warning", zap.Error(err))
	} else {
		kv, _ := linux.GetKernelVersion()
		logger.Info("running on supported Linux kernel", zap.String("kernel", kv.String()))
	}

	// 3. Platform & System Setup (sysctls, cgroup v2, bpffs)
	if err := platform.NetworkSysctls(); err != nil {
		logger.Warn("could not set all network sysctls", zap.Error(err))
	}

	if err := linux.MountBPFFS(cfg.bpffsPath); err != nil {
		logger.Warn("bpffs mount check", zap.String("path", cfg.bpffsPath), zap.Error(err))
	}

	cgroupMgr := platform.NewCgroupManager("straitd", logger)
	if platform.IsCgroupV2() {
		if err := cgroupMgr.Create(); err == nil {
			_ = cgroupMgr.AddCurrentPID()
			_ = cgroupMgr.SetCPUWeight(1000)
			logger.Info("attached straitd to cgroup v2", zap.String("cgroup", cgroupMgr.Path()))
		}
	}

	// 4. Observability & Metrics Server
	reg := prometheus.NewRegistry()
	metrics := observability.NewMetrics(reg)
	metricsSrv := observability.NewServer(cfg.metricsAddr, reg, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := metricsSrv.Serve(ctx); err != nil {
			logger.Error("metrics server error", zap.Error(err))
		}
	}()

	// 5. Initialize Core Subsystems & Install CNI Configuration
	if err := cni.InstallConfig(cfg.cniConfDir, cfg.podCIDR, logger); err != nil {
		logger.Warn("could not install CNI configuration", zap.Error(err))
	}
	identityAlloc := identity.NewAllocator()
	dataplaneCompiler := compiler.New(logger)
	routingMgr := routing.NewManager(logger, 0)
	natMgr := nat.NewManager(logger)
	cniPlugin := cni.New(logger)
	_ = cniPlugin

	// 6. Bootstrap API Path Check
	// Invariant: Bootstrap API path (straitd -> API:6443) must not depend on Service LB
	if cfg.apiServerAddr != "" {
		if apiAddr, err := netip.ParseAddr(cfg.apiServerAddr); err == nil {
			logger.Info("verifying bootstrap API route isolation",
				zap.String("apiServer", apiAddr.String()),
				zap.Int("port", cfg.apiServerPort),
			)
			_ = routingMgr.EnsureAPIServerRoute(apiAddr, "eth0")
		}
	}

	// 7. Initialize WireGuard if configured
	if cfg.wireguardPort > 0 {
		wgMgr, err := encryption.NewWireGuardManager("sg-wg0", uint16(cfg.wireguardPort), logger)
		if err != nil {
			logger.Warn("wireguard initialization deferred", zap.Error(err))
		} else {
			logger.Info("wireguard manager initialized", zap.String("pubkey", fmt.Sprintf("%x", wgMgr.PublicKey())))
		}
	}

	// 8. Systemd Notification (READY=1)
	systemdNotifier := platform.NewSystemdNotifier(logger)
	_ = systemdNotifier.Ready()
	_ = systemdNotifier.Status("straitd node agent running")

	// 9. Periodic Dataplane Reconciliation Loop
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		var gen uint64 = 1
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = systemdNotifier.Watchdog()
				gen++
				state := &ir.NetworkState{
					Generation: ir.Generation(gen),
					Endpoints:  make(map[uint64]*ir.Endpoint),
					Services:   make(map[ir.ServiceKey]*ir.Service),
					Policies:   make([]*ir.Policy, 0),
					Routes:     make([]*ir.Route, 0),
					NATRules:   natMgr.GetRules(),
				}
				if err := dataplaneCompiler.Compile(ctx, state); err != nil {
					logger.Error("periodic dataplane compile error", zap.Error(err))
				}
				metrics.IdentitiesTotal.Set(float64(identityAlloc.Count()))
			}
		}
	}()

	logger.Info("straitd node agent initialization complete, listening for events")

	// 10. Handle OS Signals for Graceful Shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	logger.Info("shutting down straitd", zap.String("signal", sig.String()))
	_ = systemdNotifier.Stopping()
	cancel()

	// Allow goroutines graceful cleanup time
	time.Sleep(500 * time.Millisecond)
	logger.Info("straitd shutdown complete")
}
