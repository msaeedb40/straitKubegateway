// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/straitkubegateway/straitkubegateway/service"
)

// ServiceReconciler reconciles Kubernetes Service and EndpointSlice objects.
type ServiceReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Log     *zap.Logger
	Service *service.Manager
}

func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Debug("reconciling Service", zap.String("service", req.String()))

	if _, err := r.Service.Reconcile(ctx, req.NamespacedName); err != nil {
		r.Log.Error("service reconciliation failed", zap.Error(err))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Watches(&discoveryv1.EndpointSlice{}, nil).
		Complete(r)
}
