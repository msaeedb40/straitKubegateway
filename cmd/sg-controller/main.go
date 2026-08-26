// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// sg-controller is the centralized control plane controller manager for straitKubegateway.
// It manages reconcilers for Gateway API, Services, Endpoints, NetworkPolicies,
// Multi-Cluster Transit, IPAM, and Cluster Federation.
package main

import (
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"

	apiv1 "github.com/straitkubegateway/straitkubegateway/api/v1alpha1"
	"github.com/straitkubegateway/straitkubegateway/controllers"
	"github.com/straitkubegateway/straitkubegateway/gateway"
	"github.com/straitkubegateway/straitkubegateway/internal/version"
	"github.com/straitkubegateway/straitkubegateway/ipam"
	"github.com/straitkubegateway/straitkubegateway/observability"
	"github.com/straitkubegateway/straitkubegateway/pkg/identity"
	"github.com/straitkubegateway/straitkubegateway/policy"
	"github.com/straitkubegateway/straitkubegateway/service"
	"github.com/straitkubegateway/straitkubegateway/transit"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(apiv1.AddToScheme(scheme))
	utilruntime.Must(gwv1.AddToScheme(scheme))
	utilruntime.Must(discoveryv1.AddToScheme(scheme))
	utilruntime.Must(corev1.AddToScheme(scheme))
}

func main() {
	var metricsAddr string
	var probeAddr string
	var enableLeaderElection bool
	var debug bool

	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false, "Enable leader election for controller manager.")
	flag.BoolVar(&debug, "debug", false, "Enable verbose debug logging.")
	flag.Parse()

	// Logger setup
	zapLogger, err := observability.NewLogger(debug)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = zapLogger.Sync() }()

	ctrl.SetLogger(ctrlzap.New(ctrlzap.UseDevMode(debug)))

	vInfo := version.Get()
	zapLogger.Info("starting sg-controller manager",
		zap.String("version", vInfo.Version),
		zap.String("commit", vInfo.GitCommit),
		zap.String("built", vInfo.BuildDate),
	)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress: metricsAddr,
		},
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "straitkubegateway-controller-leader",
	})
	if err != nil {
		zapLogger.Fatal("unable to start manager", zap.Error(err))
	}

	// Subsystem managers
	identityAlloc := identity.NewAllocator()
	discoverer := ipam.NewDiscoverer(mgr.GetClient(), zapLogger)
	serviceMgr := service.NewManager(mgr.GetClient(), zapLogger)
	policyEngine := policy.NewEngine(mgr.GetClient(), identityAlloc, zapLogger)
	transitMgr := transit.NewManager(mgr.GetClient(), zapLogger)
	gatewayMgr := gateway.NewManager(mgr.GetClient(), zapLogger)

	// Set up Reconcilers
	if err := (&controllers.ServiceReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Log:     zapLogger,
		Service: serviceMgr,
	}).SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("unable to create Service controller", zap.Error(err))
	}

	if err := (&controllers.TransitGatewayReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Log:     zapLogger,
		Transit: transitMgr,
	}).SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("unable to create TransitGateway controller", zap.Error(err))
	}

	if err := (&controllers.GatewayReconciler{
		Client:  mgr.GetClient(),
		Scheme:  mgr.GetScheme(),
		Log:     zapLogger,
		Gateway: gatewayMgr,
	}).SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("unable to create Gateway API controller", zap.Error(err))
	}

	if err := (&controllers.StraitPolicyReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    zapLogger,
		Engine: policyEngine,
	}).SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("unable to create NetworkPolicy controller", zap.Error(err))
	}

	if err := (&controllers.StraitNodeReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
		Log:    zapLogger,
	}).SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("unable to create Node controller", zap.Error(err))
	}

	if err := (&controllers.StraitNetworkReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		Log:        zapLogger,
		Discoverer: discoverer,
	}).SetupWithManager(mgr); err != nil {
		zapLogger.Fatal("unable to create Network controller", zap.Error(err))
	}

	// Health and readiness checks
	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		zapLogger.Fatal("unable to set up health check", zap.Error(err))
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		zapLogger.Fatal("unable to set up ready check", zap.Error(err))
	}

	// Perform dynamic CIDR discovery on startup in background
	ctx := ctrl.SetupSignalHandler()
	go func() {
		disc := ipam.NewDiscoverer(mgr.GetClient(), zapLogger)
		cidrs, err := disc.Discover(ctx)
		if err != nil {
			zapLogger.Warn("dynamic CIDR discovery encountered issues", zap.Error(err))
		} else {
			zapLogger.Info("initial dynamic CIDR discovery complete",
				zap.Int("podCIDRs", len(cidrs.PodCIDRs)),
				zap.Int("serviceCIDRs", len(cidrs.ServiceCIDRs)),
			)
		}
	}()

	zapLogger.Info("starting controller manager")
	if err := mgr.Start(ctx); err != nil {
		zapLogger.Fatal("problem running manager", zap.Error(err))
	}
}
