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
	"github.com/straitkubegateway/straitkubegateway/ipam"
)

// StraitNetworkReconciler reconciles a StraitNetwork object.
type StraitNetworkReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	Log        *zap.Logger
	Discoverer *ipam.Discoverer
}

// Reconcile reads the state of StraitNetwork and ensures cluster CIDR configuration is synced.
func (r *StraitNetworkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var snet apiv1.StraitNetwork
	if err := r.Get(ctx, req.NamespacedName, &snet); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	r.Log.Info("reconciling StraitNetwork",
		zap.String("name", req.Name),
		zap.String("tunnelMode", snet.Spec.TunnelMode),
	)

	// Discover CIDRs dynamically if not provided
	if snet.Spec.PodCIDR == "" || snet.Spec.ServiceCIDR == "" {
		cidrs, err := r.Discoverer.Discover(ctx)
		if err == nil && len(cidrs.PodCIDRs) > 0 {
			snet.Status.DiscoveredPodCIDR = cidrs.PodCIDRs[0].String()
			if len(cidrs.ServiceCIDRs) > 0 {
				snet.Status.DiscoveredServiceCIDR = cidrs.ServiceCIDRs[0].String()
			}
			snet.Status.Phase = "Ready"
			_ = r.Status().Update(ctx, &snet)
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StraitNetworkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.StraitNetwork{}).
		Complete(r)
}
