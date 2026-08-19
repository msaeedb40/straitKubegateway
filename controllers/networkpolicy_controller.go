package controllers

import (
	"context"
	"fmt"
	"sync"

	"github.com/straitKubegateway/straitKubegateway/api/v1alpha1"
	"github.com/straitKubegateway/straitKubegateway/observability/logging"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
	"github.com/straitKubegateway/straitKubegateway/policy"
)

// NetworkPolicyController reconciles StraitNetworkPolicy resources and syncs them to the policy manager.
type NetworkPolicyController struct {
	mu            sync.RWMutex
	policyManager *policy.Manager
	logger        *logging.Logger
}

// NewNetworkPolicyController creates a new NetworkPolicy controller.
func NewNetworkPolicyController(policyMgr *policy.Manager) *NetworkPolicyController {
	return &NetworkPolicyController{
		policyManager: policyMgr,
		logger:        logging.DefaultLogger(),
	}
}

// ReconcilePolicy updates or adds a StraitNetworkPolicy.
func (c *NetworkPolicyController) ReconcilePolicy(ctx context.Context, p *v1alpha1.StraitNetworkPolicy) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.policyManager.UpsertPolicy(p)
	c.logger.Info(fmt.Sprintf("reconciled StraitNetworkPolicy %s/%s (scope=%s, priority=%d, ingress_rules=%d, egress_rules=%d)",
		p.Namespace, p.Name, p.Spec.Scope, p.Spec.Priority, len(p.Spec.Ingress), len(p.Spec.Egress)), &types.Metadata{
		Component: "networkpolicy-controller",
		Namespace: p.Namespace,
		PolicyID:  string(p.UID),
	})

	return nil
}

// DeletePolicy removes a StraitNetworkPolicy.
func (c *NetworkPolicyController) DeletePolicy(ctx context.Context, ns, name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.policyManager.DeletePolicy(ns, name)
	c.logger.Info(fmt.Sprintf("deleted StraitNetworkPolicy %s/%s", ns, name), &types.Metadata{
		Component: "networkpolicy-controller",
		Namespace: ns,
	})
}
