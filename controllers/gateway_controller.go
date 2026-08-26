// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/straitkubegateway/straitkubegateway/gateway"
)

// GatewayReconciler reconciles Gateway API Gateway and HTTPRoute objects.
type GatewayReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Log     *zap.Logger
	Gateway *gateway.Manager
}

func (r *GatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciling Gateway API resource", zap.String("resource", req.String()))

	if err := r.Gateway.ReconcileGateway(ctx, req.NamespacedName); err != nil {
		r.Log.Error("gateway reconciliation failed", zap.Error(err))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *GatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gwv1.Gateway{}).
		Watches(&gwv1.HTTPRoute{}, nil).
		Complete(r)
}
