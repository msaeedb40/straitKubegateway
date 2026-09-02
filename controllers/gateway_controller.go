// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	// Map HTTPRoute events to parent Gateway reconcile requests
	mapRouteToGateways := func(_ context.Context, obj client.Object) []reconcile.Request {
		route, ok := obj.(*gwv1.HTTPRoute)
		if !ok {
			return nil
		}
		var reqs []reconcile.Request
		for _, parent := range route.Spec.ParentRefs {
			ns := route.Namespace
			if parent.Namespace != nil {
				ns = string(*parent.Namespace)
			}
			reqs = append(reqs, reconcile.Request{
				NamespacedName: types.NamespacedName{
					Namespace: ns,
					Name:      string(parent.Name),
				},
			})
		}
		return reqs
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&gwv1.Gateway{}).
		Watches(&gwv1.HTTPRoute{}, handler.EnqueueRequestsFromMapFunc(mapRouteToGateways)).
		Complete(r)
}
