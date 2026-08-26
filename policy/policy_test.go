// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apiv1 "github.com/straitkubegateway/straitkubegateway/api/v1alpha1"
	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/ir"
	identitypkg "github.com/straitkubegateway/straitkubegateway/pkg/identity"
	sgtypes "github.com/straitkubegateway/straitkubegateway/pkg/types"
)

func TestDefaultAction(t *testing.T) {
	// Ingress with no rules defaults to Deny
	actIngress := defaultAction(ir.PolicyDirectionIngress, nil, nil)
	if actIngress != sgtypes.PolicyActionDeny {
		t.Errorf("got %s, want Deny for default ingress", actIngress)
	}

	// Egress with no rules defaults to Allow
	actEgress := defaultAction(ir.PolicyDirectionEgress, nil, nil)
	if actEgress != sgtypes.PolicyActionAllow {
		t.Errorf("got %s, want Allow for default egress", actEgress)
	}

	// Explicit ingress rule
	ingressRules := []apiv1.IngressRule{{Action: apiv1.PolicyActionAllow}}
	actExplicit := defaultAction(ir.PolicyDirectionIngress, ingressRules, nil)
	if actExplicit != sgtypes.PolicyActionAllow {
		t.Errorf("got %s, want Allow", actExplicit)
	}
}

func TestPolicySorting(t *testing.T) {
	e := &Engine{
		policies: []*ir.Policy{
			{ID: "low-priority", Priority: 200, Action: sgtypes.PolicyActionAllow},
			{ID: "high-priority", Priority: 50, Action: sgtypes.PolicyActionAllow},
			{ID: "equal-allow", Priority: 100, Action: sgtypes.PolicyActionAllow},
			{ID: "equal-deny", Priority: 100, Action: sgtypes.PolicyActionDeny},
		},
	}

	e.sortPolicies()

	// Priority order: 50 -> 100 (Deny before Allow) -> 200
	if e.policies[0].ID != "high-priority" {
		t.Errorf("expected high-priority first, got %s", e.policies[0].ID)
	}
	if e.policies[1].ID != "equal-deny" {
		t.Errorf("expected equal-deny second (deny overrides allow), got %s", e.policies[1].ID)
	}
	if e.policies[2].ID != "equal-allow" {
		t.Errorf("expected equal-allow third, got %s", e.policies[2].ID)
	}
	if e.policies[3].ID != "low-priority" {
		t.Errorf("expected low-priority last, got %s", e.policies[3].ID)
	}
}

func TestResolveIdentitiesWithExtendedSelectors(t *testing.T) {
	alloc := identitypkg.NewAllocator()
	engine := &Engine{identity: alloc}
	ctx := context.Background()

	sel := apiv1.EndpointSelector{
		PodSelector:       &metav1.LabelSelector{MatchLabels: map[string]string{"app": "frontend"}},
		NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"env": "prod"}},
		SegmentSelector:   &apiv1.SegmentSelector{SegmentID: 1},
		GatewaySelector:   &metav1.LabelSelector{MatchLabels: map[string]string{"gateway": "strait-gw"}},
		HTTPRouteSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"route": "v1-api"}},
	}

	ids, err := engine.resolveIdentities(ctx, "prod", sel)
	if err != nil {
		t.Fatalf("resolveIdentities failed: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 identity, got %d", len(ids))
	}
	if ids[0] < sgtypes.IdentityMin {
		t.Errorf("expected identity >= IdentityMin (%d), got %d", sgtypes.IdentityMin, ids[0])
	}
}

func TestCompileMultiRulePolicy(t *testing.T) {
	alloc := identitypkg.NewAllocator()
	engine := &Engine{identity: alloc}
	ctx := context.Background()

	snp := &apiv1.StraitNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "web-policy",
		},
		Spec: apiv1.StraitNetworkPolicySpec{
			Priority: 50,
			PolicyTypes: []apiv1.PolicyType{
				apiv1.PolicyTypeIngress,
				apiv1.PolicyTypeEgress,
			},
			EndpointSelector: apiv1.EndpointSelector{
				PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "web"}},
			},
			Ingress: []apiv1.IngressRule{
				{
					Action: apiv1.PolicyActionAllow,
					From: []apiv1.EndpointSelector{
						{PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "frontend"}}},
					},
				},
				{
					Action: apiv1.PolicyActionDeny,
					From: []apiv1.EndpointSelector{
						{PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "untrusted"}}},
					},
				},
			},
			Egress: []apiv1.EgressRule{
				{
					Action: apiv1.PolicyActionAllow,
					To: []apiv1.EndpointSelector{
						{PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "db"}}},
					},
				},
			},
		},
	}

	compiled, err := engine.compile(ctx, snp)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	// Should compile 2 ingress rules + 1 egress rule = 3 IR policies
	if len(compiled) != 3 {
		t.Fatalf("expected 3 compiled IR policies, got %d", len(compiled))
	}

	if compiled[0].Direction != ir.PolicyDirectionIngress || compiled[0].Action != sgtypes.PolicyActionAllow {
		t.Errorf("rule 0 mismatch: %v", compiled[0])
	}
	if compiled[1].Direction != ir.PolicyDirectionIngress || compiled[1].Action != sgtypes.PolicyActionDeny {
		t.Errorf("rule 1 mismatch: %v", compiled[1])
	}
	if compiled[2].Direction != ir.PolicyDirectionEgress || compiled[2].Action != sgtypes.PolicyActionAllow {
		t.Errorf("rule 2 mismatch: %v", compiled[2])
	}
}
