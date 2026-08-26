// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

// Package policy implements the straitKubegateway network policy engine.
//
// Policy evaluation semantics (deterministic compiler rules):
//   - Rules are evaluated in priority order: lower value = higher priority (0-255)
//   - Default ingress action: Deny (deny-by-default)
//   - Default egress action: Allow (allow-by-default)
//   - At equal priority: Deny overrides Allow
//   - Namespace policy cannot silently override cluster/segment policy
package policy

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/straitkubegateway/straitkubegateway/api/v1alpha1"
	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/ir"
	identitypkg "github.com/straitkubegateway/straitkubegateway/pkg/identity"
	sgtypes "github.com/straitkubegateway/straitkubegateway/pkg/types"
)

// ============================================================================
// Policy Engine
// ============================================================================

// Engine compiles StraitNetworkPolicy CRs into the IR policy list.
// Thread-safe.
type Engine struct {
	mu        sync.RWMutex
	client    client.Client
	identity  *identitypkg.Allocator
	log       *zap.Logger
	policies  []*ir.Policy
}

// NewEngine creates a new policy engine.
func NewEngine(c client.Client, identity *identitypkg.Allocator, log *zap.Logger) *Engine {
	return &Engine{
		client:   c,
		identity: identity,
		log:      log,
	}
}

// ============================================================================
// Reconcile
// ============================================================================

// Reconcile fetches the StraitNetworkPolicy and compiles it into IR policies.
func (e *Engine) Reconcile(ctx context.Context, key types.NamespacedName) error {
	var snp apiv1.StraitNetworkPolicy
	if err := e.client.Get(ctx, key, &snp); err != nil {
		return client.IgnoreNotFound(err)
	}

	compiledList, err := e.compile(ctx, &snp)
	if err != nil {
		return fmt.Errorf("compile policy %s: %w", key, err)
	}

	e.mu.Lock()
	prefix := policyID(key.Namespace, key.Name)
	remaining := e.policies[:0]
	for _, p := range e.policies {
		if p.ID != prefix && !strings.HasPrefix(p.ID, prefix+"/") {
			remaining = append(remaining, p)
		}
	}
	e.policies = append(remaining, compiledList...)
	e.sortPolicies()
	e.mu.Unlock()

	e.log.Info("policy reconciled",
		zap.String("policy", key.String()),
		zap.Int("compiledRules", len(compiledList)),
	)
	return nil
}

// Delete removes a compiled policy.
func (e *Engine) Delete(key types.NamespacedName) {
	prefix := policyID(key.Namespace, key.Name)
	e.mu.Lock()
	defer e.mu.Unlock()
	remaining := e.policies[:0]
	for _, p := range e.policies {
		if p.ID != prefix && !strings.HasPrefix(p.ID, prefix+"/") {
			remaining = append(remaining, p)
		}
	}
	e.policies = remaining
}

// GetAll returns the current compiled policy list, sorted by priority.
func (e *Engine) GetAll() []*ir.Policy {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]*ir.Policy, len(e.policies))
	copy(out, e.policies)
	return out
}

// ============================================================================
// Compiler
// ============================================================================

func (e *Engine) compile(ctx context.Context, snp *apiv1.StraitNetworkPolicy) ([]*ir.Policy, error) {
	baseID := policyID(snp.Namespace, snp.Name)
	priority := uint8(snp.Spec.Priority)
	if snp.Spec.Priority == 0 && priority == 0 {
		priority = 100
	}

	// Resolve subject identities for the policy endpoint selector
	subjectIDs, err := e.resolveIdentities(ctx, snp.Namespace, snp.Spec.EndpointSelector)
	if err != nil {
		return nil, fmt.Errorf("resolve endpoint selector: %w", err)
	}

	policyTypes := snp.Spec.PolicyTypes
	if len(policyTypes) == 0 {
		if len(snp.Spec.Ingress) > 0 {
			policyTypes = append(policyTypes, apiv1.PolicyTypeIngress)
		}
		if len(snp.Spec.Egress) > 0 {
			policyTypes = append(policyTypes, apiv1.PolicyTypeEgress)
		}
		if len(policyTypes) == 0 {
			policyTypes = []apiv1.PolicyType{apiv1.PolicyTypeIngress}
		}
	}

	var results []*ir.Policy

	for _, pt := range policyTypes {
		switch pt {
		case apiv1.PolicyTypeIngress:
			if len(snp.Spec.Ingress) == 0 {
				results = append(results, &ir.Policy{
					ID:             fmt.Sprintf("%s/ingress/default-deny", baseID),
					Priority:       priority,
					Direction:      ir.PolicyDirectionIngress,
					DestIdentities: subjectIDs,
					Action:         sgtypes.PolicyActionDeny,
				})
			} else {
				for i, rule := range snp.Spec.Ingress {
					action := sgtypes.PolicyAction(rule.Action)
					if action == "" {
						action = sgtypes.PolicyActionAllow
					}
					var srcIDs []sgtypes.Identity
					for _, fromSel := range rule.From {
						ids, err := e.resolveIdentities(ctx, snp.Namespace, fromSel)
						if err == nil {
							srcIDs = append(srcIDs, ids...)
						}
					}
					results = append(results, &ir.Policy{
						ID:               fmt.Sprintf("%s/ingress/%d", baseID, i),
						Priority:         priority,
						Direction:        ir.PolicyDirectionIngress,
						SourceIdentities: srcIDs,
						DestIdentities:   subjectIDs,
						Action:           action,
					})
				}
			}

		case apiv1.PolicyTypeEgress:
			if len(snp.Spec.Egress) == 0 {
				results = append(results, &ir.Policy{
					ID:               fmt.Sprintf("%s/egress/default-allow", baseID),
					Priority:         priority,
					Direction:        ir.PolicyDirectionEgress,
					SourceIdentities: subjectIDs,
					Action:           sgtypes.PolicyActionAllow,
				})
			} else {
				for i, rule := range snp.Spec.Egress {
					action := sgtypes.PolicyAction(rule.Action)
					if action == "" {
						action = sgtypes.PolicyActionAllow
					}
					var dstIDs []sgtypes.Identity
					for _, toSel := range rule.To {
						ids, err := e.resolveIdentities(ctx, snp.Namespace, toSel)
						if err == nil {
							dstIDs = append(dstIDs, ids...)
						}
					}
					results = append(results, &ir.Policy{
						ID:               fmt.Sprintf("%s/egress/%d", baseID, i),
						Priority:         priority,
						Direction:        ir.PolicyDirectionEgress,
						SourceIdentities: subjectIDs,
						DestIdentities:   dstIDs,
						Action:           action,
					})
				}
			}
		}
	}

	return results, nil
}

// resolveIdentities resolves an EndpointSelector to numeric identities.
func (e *Engine) resolveIdentities(ctx context.Context, namespace string, sel apiv1.EndpointSelector) ([]sgtypes.Identity, error) {
	labels := identitypkg.Labels{}

	// Add namespace as a label dimension
	if namespace != "" {
		labels["k8s:io.kubernetes.pod.namespace"] = namespace
	}

	if sel.PodSelector != nil {
		for k, v := range sel.PodSelector.MatchLabels {
			labels[k] = v
		}
	}
	if sel.NamespaceSelector != nil {
		for k, v := range sel.NamespaceSelector.MatchLabels {
			labels["ns:"+k] = v
		}
	}
	if sel.ClusterSelector != nil {
		for k, v := range sel.ClusterSelector.MatchLabels {
			labels["cluster:"+k] = v
		}
	}
	if sel.SegmentSelector != nil {
		labels["segment"] = fmt.Sprintf("%d", sel.SegmentSelector.SegmentID)
	}
	if sel.GatewaySelector != nil {
		for k, v := range sel.GatewaySelector.MatchLabels {
			labels["gw:"+k] = v
		}
	}
	if sel.HTTPRouteSelector != nil {
		for k, v := range sel.HTTPRouteSelector.MatchLabels {
			labels["httproute:"+k] = v
		}
	}
	if sel.TCPRouteSelector != nil {
		for k, v := range sel.TCPRouteSelector.MatchLabels {
			labels["tcproute:"+k] = v
		}
	}
	if sel.UDPRouteSelector != nil {
		for k, v := range sel.UDPRouteSelector.MatchLabels {
			labels["udproute:"+k] = v
		}
	}
	if sel.GRPCRouteSelector != nil {
		for k, v := range sel.GRPCRouteSelector.MatchLabels {
			labels["grpcroute:"+k] = v
		}
	}
	if sel.TLSRouteSelector != nil {
		for k, v := range sel.TLSRouteSelector.MatchLabels {
			labels["tlsroute:"+k] = v
		}
	}

	id, err := e.identity.Allocate(ctx, labels.Key())
	if err != nil {
		return nil, err
	}
	return []sgtypes.Identity{id}, nil
}

// defaultAction returns the policy action following deterministic semantics.
// Default ingress action: Deny. Default egress action: Allow.
func defaultAction(
	direction ir.PolicyDirection,
	ingress []apiv1.IngressRule,
	egress []apiv1.EgressRule,
) sgtypes.PolicyAction {
	if direction == ir.PolicyDirectionIngress {
		if len(ingress) == 0 {
			return sgtypes.PolicyActionDeny
		}
		return sgtypes.PolicyAction(ingress[0].Action)
	}
	// Egress
	if len(egress) == 0 {
		return sgtypes.PolicyActionAllow
	}
	return sgtypes.PolicyAction(egress[0].Action)
}

// upsertPolicy inserts or replaces a compiled policy by ID, then re-sorts.
func (e *Engine) upsertPolicy(p *ir.Policy) {
	for i, existing := range e.policies {
		if existing.ID == p.ID {
			e.policies[i] = p
			e.sortPolicies()
			return
		}
	}
	e.policies = append(e.policies, p)
	e.sortPolicies()
}

// sortPolicies sorts by priority ascending (lower value = evaluated first).
// At equal priority, Deny overrides Allow.
func (e *Engine) sortPolicies() {
	sort.SliceStable(e.policies, func(i, j int) bool {
		pi, pj := e.policies[i], e.policies[j]
		if pi.Priority != pj.Priority {
			return pi.Priority < pj.Priority
		}
		// Equal priority: Deny before Allow
		return isDeny(pi.Action) && !isDeny(pj.Action)
	})
}

func isDeny(a sgtypes.PolicyAction) bool {
	return a == sgtypes.PolicyActionDeny || a == sgtypes.PolicyActionReject
}

func policyID(namespace, name string) string {
	if namespace == "" {
		return name
	}
	return namespace + "/" + name
}
