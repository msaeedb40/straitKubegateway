// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/straitkubegateway/straitkubegateway/service"
)

// ServiceReconciler reconciles Kubernetes Service and EndpointSlice objects.
//
// Architectural invariant: this controller produces IR only —
// it never touches BPF maps directly.
type ServiceReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Log     *zap.Logger
	Service *service.Manager
}

// Reconcile is called for every Service and EndpointSlice change.
// It is idempotent: reconciling the same object multiple times is safe.
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Debug("reconciling Service", zap.String("service", req.String()))

	// Check whether the service still exists. If it was deleted, remove it
	// from the service manager and let the dataplane compiler clean up.
	var svc corev1.Service
	if err := r.Get(ctx, req.NamespacedName, &svc); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info("Service deleted — removing from dataplane",
				zap.String("namespace", req.Namespace),
				zap.String("name", req.Name),
			)
			r.Service.Delete(req.NamespacedName)
			return ctrl.Result{}, nil
		}
		r.Log.Error("failed to get Service", zap.Error(err))
		return ctrl.Result{}, err
	}

	if _, err := r.Service.Reconcile(ctx, req.NamespacedName); err != nil {
		r.Log.Error("service reconciliation failed", zap.Error(err))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager registers the ServiceReconciler with the controller manager.
// It watches both Service objects and EndpointSlice objects.
// When an EndpointSlice changes, it enqueues the owning Service for reconciliation.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// mapEndpointSliceToService converts an EndpointSlice event into a
	// reconcile.Request for the owning Service.
	mapEndpointSliceToService := func(_ context.Context, obj client.Object) []reconcile.Request {
		labels := obj.GetLabels()
		svcName, ok := labels[discoveryv1.LabelServiceName]
		if !ok || svcName == "" {
			return nil
		}
		return []reconcile.Request{
			{NamespacedName: types.NamespacedName{
				Namespace: obj.GetNamespace(),
				Name:      svcName,
			}},
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Service{}).
		Watches(
			&discoveryv1.EndpointSlice{},
			handler.EnqueueRequestsFromMapFunc(mapEndpointSliceToService),
		).
		Complete(r)
}
