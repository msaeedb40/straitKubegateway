// Package main implements the sg-controller Kubernetes control plane daemon.
package main

import (
	"context"
	"flag"
	"fmt"
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
	metricsAddr := flag.String("metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
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

	logger.Info("starting sg-controller control plane daemon", ctx)
	logger.Info(fmt.Sprintf("metrics server listening on %s", *metricsAddr), ctx)

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
