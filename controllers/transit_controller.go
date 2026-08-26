// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/straitkubegateway/straitkubegateway/api/v1alpha1"
	"github.com/straitkubegateway/straitkubegateway/transit"
)

// TransitGatewayReconciler reconciles TransitGateway and TransitSegmentAttachment objects.
type TransitGatewayReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Log     *zap.Logger
	Transit *transit.Manager
}

func (r *TransitGatewayReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciling TransitGateway", zap.String("name", req.Name))

	if err := r.Transit.ReconcileGateway(ctx, req.NamespacedName); err != nil {
		r.Log.Error("transit gateway reconcile failed", zap.Error(err))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *TransitGatewayReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.TransitGateway{}).
		Complete(r)
}
