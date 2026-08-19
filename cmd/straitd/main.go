// Package main implements the straitd node runtime agent daemon.
package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/straitKubegateway/straitKubegateway/dataplane"
	"github.com/straitKubegateway/straitKubegateway/internal/version"
	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/linux"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
	initutil "github.com/straitKubegateway/straitKubegateway/platform/init"
	"github.com/straitKubegateway/straitKubegateway/platform/kernel"
	"github.com/straitKubegateway/straitKubegateway/platform/process"
	"github.com/straitKubegateway/straitKubegateway/platform/resource"
)

type NodeState struct {
	CNIReady     bool
	ServiceReady bool
	PolicyReady  bool
	GatewayReady bool
}

func main() {
	podCIDR := flag.String("pod-cidr", "10.244.0.0/16", "Node Pod CIDR pool")
	apiServerAddr := flag.String("apiserver", "kubernetes.default.svc:443", "Kubernetes API Server address for bootstrap check")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Get().String())
		os.Exit(0)
	}

	logger := logging.DefaultLogger()
	ctx := &types.Metadata{
		ServiceName:    "straitd",
		ServiceVersion: version.Version,
		Component:      "node-agent",
		NodeID:         os.Getenv("NODE_NAME"),
	}

	logger.Info("starting straitd node agent daemon", ctx)

	metricsMux := http.NewServeMux()
	metricsMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	metricsMux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	metricsServer := &http.Server{
		Addr:    ":9090",
		Handler: metricsMux,
	}
	go func() {
		if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server error", &types.ErrorInfo{Code: "METRICS_ERR", Message: err.Error()}, ctx)
		}
	}()
	defer func() {
		_ = metricsServer.Close()
	}()

	// 1. Detect environment and validate kernel prerequisites
	env, err := kernel.DetectEnvironment()
	if err != nil {
		logger.Error("kernel environment detection failed", &types.ErrorInfo{Code: "KERNEL_ERR", Message: err.Error()}, ctx)
	} else {
		logger.Info(fmt.Sprintf("detected host kernel version: %s (6.7+ compatible: %v, cgroup v2: %v)",
			env.KernelVersion, env.IsKernel67Plus, env.CgroupV2), ctx)
	}

	// 2. Ensure bpffs is mounted
	if err := linux.EnsureBPFFilesystem("/sys/fs/bpf"); err != nil {
		logger.Error("failed to mount bpffs", &types.ErrorInfo{Code: "BPFFS_ERR", Message: err.Error()}, ctx)
	}

	// 3. Initialize cgroup v2 containment
	if resource.IsCgroupV2Available() {
		cgroupMgr := resource.NewManager("")
		_ = cgroupMgr.EnsureCgroup()
		_ = cgroupMgr.AddProcess(os.Getpid())
	}

	// 4. Initialize Dataplane Manager
	dpManager, err := dataplane.NewManager(dataplane.Config{
		PodCIDR: *podCIDR,
	})
	if err != nil {
		logger.Error("failed to initialize dataplane manager", &types.ErrorInfo{Code: "DP_INIT_ERR", Message: err.Error()}, ctx)
		os.Exit(1)
	}

	// 5. Start CNI daemon socket server for persistent IPAM/dataplane state
	cniServer := dataplane.NewServer(dpManager, dataplane.DefaultSocketPath)
	if err := cniServer.Start(); err != nil {
		logger.Error("failed to start CNI daemon socket", &types.ErrorInfo{Code: "CNI_SOCK_ERR", Message: err.Error()}, ctx)
		// Non-fatal: CNI plugin can fall back to local mode
	}
	defer cniServer.Stop()

	nodeState := NodeState{
		CNIReady:     true, // CNI dataplane initialized
		ServiceReady: false,
		PolicyReady:  false,
		GatewayReady: false,
	}

	logger.Info(fmt.Sprintf("straitd CNI initialized. Readiness state: CNI=%v, Service=%v, Policy=%v, Gateway=%v",
		nodeState.CNIReady, nodeState.ServiceReady, nodeState.PolicyReady, nodeState.GatewayReady), ctx)

	// 6. Bootstrap verification in background (non-blocking)
	sup := process.NewSupervisor(5 * time.Second)
	sup.Go(func(gCtx context.Context) {
		verifyBootstrapInvariants(gCtx, *apiServerAddr, logger, ctx)
	})

	sup.Go(func(gCtx context.Context) {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-gCtx.Done():
				return
			case <-ticker.C:
				eps := dpManager.ListEndpoints()
				logger.Debug(fmt.Sprintf("straitd heartbeat: %d active pod endpoints", len(eps)), ctx)
			}
		}
	})

	logger.Info(fmt.Sprintf("straitd running under %s init system", initutil.DetectInitSystem()), ctx)
	sup.WaitForSignals()
	logger.Info("straitd stopped gracefully", ctx)
}

// verifyBootstrapInvariants checks reachability of API server without circular blocking
func verifyBootstrapInvariants(ctx context.Context, apiServerAddr string, logger *logging.Logger, meta *types.Metadata) {
	conn, err := net.DialTimeout("tcp", apiServerAddr, 2*time.Second)
	if err != nil {
		logger.Debug(fmt.Sprintf("bootstrap check: API server %s is not yet directly reachable: %v", apiServerAddr, err), meta)
		return
	}
	conn.Close()
	logger.Info(fmt.Sprintf("bootstrap check passed: API server %s reachable", apiServerAddr), meta)
}
