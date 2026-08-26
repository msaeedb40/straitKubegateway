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
	"github.com/straitkubegateway/straitkubegateway/policy"
)

// StraitPolicyReconciler reconciles StraitNetworkPolicy objects.
type StraitPolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    *zap.Logger
	Engine *policy.Engine
}

func (r *StraitPolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Info("reconciling StraitNetworkPolicy", zap.String("policy", req.String()))

	if err := r.Engine.Reconcile(ctx, req.NamespacedName); err != nil {
		r.Log.Error("policy engine reconcile failed", zap.Error(err))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *StraitPolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.StraitNetworkPolicy{}).
		Complete(r)
}
