// Package main implements the sg-controller Kubernetes control plane daemon.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/straitKubegateway/straitKubegateway/controllers"
	"github.com/straitKubegateway/straitKubegateway/internal/version"
	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
	"github.com/straitKubegateway/straitKubegateway/platform/process"
)

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	metricsAddr := flag.String("metrics-bind-address", ":9090", "The address the metric endpoint binds to.")
	metricsPort := flag.Int("metrics-port", 9090, "The metrics port")
	_ = flag.Bool("leader-elect", true, "Enable leader election")
	_ = flag.Int("grpc-port", 50051, "gRPC port")
	_ = flag.String("config", "/etc/strait/config.json", "Config file path")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Get().String())
		os.Exit(0)
	}

	logger := logging.DefaultLogger()
	ctx := &types.Metadata{
		ServiceName:    "sg-controller",
		ServiceVersion: version.Version,
		Component:      "control-plane",
	}

	addr := *metricsAddr
	if *metricsPort != 9090 && addr == ":9090" {
		addr = fmt.Sprintf(":%d", *metricsPort)
	}

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
		Addr:    addr,
		Handler: metricsMux,
	}
	go func() {
		_ = metricsServer.ListenAndServe()
	}()
	defer func() {
		_ = metricsServer.Close()
	}()

	logger.Info("starting sg-controller control plane daemon", ctx)
	logger.Info(fmt.Sprintf("metrics server listening on %s", addr), ctx)

	_ = controllers.NewIPAMReconciler()
	_ = controllers.NewEndpointReconciler()

	sup := process.NewSupervisor(5 * time.Second)
	sup.Go(func(gCtx context.Context) {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-gCtx.Done():
				return
			case <-ticker.C:
				logger.Debug("sg-controller leader heartbeat active", ctx)
			}
		}
	})

	sup.WaitForSignals()
	logger.Info("sg-controller stopped gracefully", ctx)
}
