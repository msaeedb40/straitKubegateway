// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/straitkubegateway/straitkubegateway/api/v1alpha1"
)

// StraitNodeReconciler reconciles StraitNode and corev1.Node objects.
type StraitNodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    *zap.Logger
}

func (r *StraitNodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var snode apiv1.StraitNode
	if err := r.Get(ctx, req.NamespacedName, &snode); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.Log.Debug("reconciling StraitNode",
		zap.String("node", snode.Spec.NodeName),
		zap.String("phase", snode.Status.Phase),
		zap.Bool("cniReady", snode.Status.CNIReady),
		zap.Bool("serviceReady", snode.Status.ServiceReady),
	)

	return ctrl.Result{}, nil
}

func (r *StraitNodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.StraitNode{}).
		Watches(&corev1.Node{}, nil).
		Complete(r)
}
