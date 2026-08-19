package controllers

import (
	"context"
	"fmt"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
	"github.com/straitKubegateway/straitKubegateway/service"
)

// KubeProxyReconciler reconciles all Kubernetes Services for complete kube-proxy replacement.
type KubeProxyReconciler struct {
	mu       sync.RWMutex
	proxyMgr *service.KubeProxyManager
	logger   *logging.Logger
}

// NewKubeProxyReconciler creates a new kube-proxy reconciler.
func NewKubeProxyReconciler(proxyMgr *service.KubeProxyManager) *KubeProxyReconciler {
	return &KubeProxyReconciler{
		proxyMgr: proxyMgr,
		logger:   logging.DefaultLogger(),
	}
}

// ReconcileKubeService updates the dataplane for a Kubernetes Service.
func (r *KubeProxyReconciler) ReconcileKubeService(ctx context.Context, svc service.KubeProxyService, endpoints []service.BackendEndpoint) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.proxyMgr.UpsertKubeService(svc, endpoints)
	r.logger.Info(fmt.Sprintf("reconciled kube-proxy service %s/%s (Type=%s, ClusterIP=%s, NodePorts=%v, LBIPs=%v, Backends=%d)",
		svc.Namespace, svc.Name, svc.Type, svc.ClusterIP, svc.NodePorts, svc.LoadBalancerIPs, len(endpoints)), &types.Metadata{
		Component:   "kubeproxy-controller",
		Namespace:   svc.Namespace,
		ServiceName: svc.Name,
	})

	return nil
}

// DeleteKubeService removes a Service from the kube-proxy dataplane.
func (r *KubeProxyReconciler) DeleteKubeService(ctx context.Context, ns, name string, port uint16, proto types.Protocol) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.proxyMgr.DeleteKubeService(ns, name, port, proto)
	r.logger.Info(fmt.Sprintf("deleted kube-proxy service %s/%s (%d/%s)", ns, name, port, proto), &types.Metadata{
		Component:   "kubeproxy-controller",
		Namespace:   ns,
		ServiceName: name,
	})
}
