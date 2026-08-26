// Copyright 2026 straitKubegateway authors.
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"
	"fmt"
	"testing"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	apiv1 "github.com/straitkubegateway/straitkubegateway/api/v1alpha1"
	"github.com/straitkubegateway/straitkubegateway/internal/dataplane/ir"
	identitypkg "github.com/straitkubegateway/straitkubegateway/pkg/identity"
	sgtypes "github.com/straitkubegateway/straitkubegateway/pkg/types"
)

// ============================================================================
// Unit Tests
// ============================================================================

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

	// Explicit egress rule
	egressRules := []apiv1.EgressRule{{Action: apiv1.PolicyActionDeny}}
	actExplicitEgress := defaultAction(ir.PolicyDirectionEgress, nil, egressRules)
	if actExplicitEgress != sgtypes.PolicyActionDeny {
		t.Errorf("got %s, want Deny for explicit egress", actExplicitEgress)
	}
}

func TestPolicySorting(t *testing.T) {
	e := &Engine{
		policies: []*ir.Policy{
			{ID: "low-priority", Priority: 200, Action: sgtypes.PolicyActionAllow},
			{ID: "high-priority", Priority: 50, Action: sgtypes.PolicyActionAllow},
			{ID: "equal-allow", Priority: 100, Action: sgtypes.PolicyActionAllow},
			{ID: "equal-deny", Priority: 100, Action: sgtypes.PolicyActionDeny},
			{ID: "equal-reject", Priority: 100, Action: sgtypes.PolicyActionReject},
		},
	}

	e.sortPolicies()

	// Priority order: 50 -> 100 (Deny/Reject before Allow) -> 200
	if e.policies[0].ID != "high-priority" {
		t.Errorf("expected high-priority first, got %s", e.policies[0].ID)
	}
	if !isDeny(e.policies[1].Action) || !isDeny(e.policies[2].Action) {
		t.Errorf("expected deny/reject before allow at priority 100")
	}
	if e.policies[3].ID != "equal-allow" {
		t.Errorf("expected equal-allow fourth, got %s", e.policies[3].ID)
	}
	if e.policies[4].ID != "low-priority" {
		t.Errorf("expected low-priority last, got %s", e.policies[4].ID)
	}
}

func TestIsDeny(t *testing.T) {
	if !isDeny(sgtypes.PolicyActionDeny) {
		t.Error("expected PolicyActionDeny to be deny")
	}
	if !isDeny(sgtypes.PolicyActionReject) {
		t.Error("expected PolicyActionReject to be deny")
	}
	if isDeny(sgtypes.PolicyActionAllow) {
		t.Error("expected PolicyActionAllow to NOT be deny")
	}
}

func TestPolicyID(t *testing.T) {
	if got := policyID("default", "web"); got != "default/web" {
		t.Errorf("got %s, want default/web", got)
	}
	if got := policyID("", "cluster-policy"); got != "cluster-policy" {
		t.Errorf("got %s, want cluster-policy", got)
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

func TestResolveIdentitiesAllSelectors(t *testing.T) {
	alloc := identitypkg.NewAllocator()
	engine := &Engine{identity: alloc}
	ctx := context.Background()

	sel := apiv1.EndpointSelector{
		PodSelector:       &metav1.LabelSelector{MatchLabels: map[string]string{"app": "api"}},
		NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"tier": "backend"}},
		ClusterSelector:   &metav1.LabelSelector{MatchLabels: map[string]string{"region": "us-east"}},
		SegmentSelector:   &apiv1.SegmentSelector{SegmentID: 100},
		GatewaySelector:   &metav1.LabelSelector{MatchLabels: map[string]string{"gw": "edge"}},
		HTTPRouteSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"http": "v2"}},
		TCPRouteSelector:  &metav1.LabelSelector{MatchLabels: map[string]string{"tcp": "db"}},
		UDPRouteSelector:  &metav1.LabelSelector{MatchLabels: map[string]string{"udp": "dns"}},
		GRPCRouteSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"grpc": "auth"}},
		TLSRouteSelector:  &metav1.LabelSelector{MatchLabels: map[string]string{"tls": "ingress"}},
	}

	ids, err := engine.resolveIdentities(ctx, "prod", sel)
	if err != nil {
		t.Fatalf("resolveIdentities failed: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("expected 1 identity, got %d", len(ids))
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

func TestCompileDefaultDenyIngress(t *testing.T) {
	alloc := identitypkg.NewAllocator()
	engine := &Engine{identity: alloc}
	ctx := context.Background()

	snp := &apiv1.StraitNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "deny-all-ingress",
		},
		Spec: apiv1.StraitNetworkPolicySpec{
			PolicyTypes: []apiv1.PolicyType{apiv1.PolicyTypeIngress},
			EndpointSelector: apiv1.EndpointSelector{
				PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "isolated"}},
			},
		},
	}

	compiled, err := engine.compile(ctx, snp)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if len(compiled) != 1 {
		t.Fatalf("expected 1 default-deny rule, got %d", len(compiled))
	}
	if compiled[0].Action != sgtypes.PolicyActionDeny {
		t.Errorf("expected Deny action, got %s", compiled[0].Action)
	}
	if compiled[0].Direction != ir.PolicyDirectionIngress {
		t.Errorf("expected Ingress direction, got %s", compiled[0].Direction)
	}
}

func TestCompileDefaultAllowEgress(t *testing.T) {
	alloc := identitypkg.NewAllocator()
	engine := &Engine{identity: alloc}
	ctx := context.Background()

	snp := &apiv1.StraitNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "allow-all-egress",
		},
		Spec: apiv1.StraitNetworkPolicySpec{
			PolicyTypes: []apiv1.PolicyType{apiv1.PolicyTypeEgress},
			EndpointSelector: apiv1.EndpointSelector{
				PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "open"}},
			},
		},
	}

	compiled, err := engine.compile(ctx, snp)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if len(compiled) != 1 {
		t.Fatalf("expected 1 default-allow rule, got %d", len(compiled))
	}
	if compiled[0].Action != sgtypes.PolicyActionAllow {
		t.Errorf("expected Allow action, got %s", compiled[0].Action)
	}
	if compiled[0].Direction != ir.PolicyDirectionEgress {
		t.Errorf("expected Egress direction, got %s", compiled[0].Direction)
	}
}

func TestCompilePolicyTypesInference(t *testing.T) {
	alloc := identitypkg.NewAllocator()
	engine := &Engine{identity: alloc}
	ctx := context.Background()

	// Ingress inferred from Ingress rules
	snpIngress := &apiv1.StraitNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "infer-ingress"},
		Spec: apiv1.StraitNetworkPolicySpec{
			EndpointSelector: apiv1.EndpointSelector{
				PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			},
			Ingress: []apiv1.IngressRule{{Action: apiv1.PolicyActionAllow}},
		},
	}
	compiled, err := engine.compile(ctx, snpIngress)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if len(compiled) != 1 || compiled[0].Direction != ir.PolicyDirectionIngress {
		t.Errorf("expected 1 ingress rule, got %v", compiled)
	}

	// Egress inferred from Egress rules
	snpEgress := &apiv1.StraitNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "infer-egress"},
		Spec: apiv1.StraitNetworkPolicySpec{
			EndpointSelector: apiv1.EndpointSelector{
				PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			},
			Egress: []apiv1.EgressRule{{Action: apiv1.PolicyActionAllow}},
		},
	}
	compiledEgress, err := engine.compile(ctx, snpEgress)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}
	if len(compiledEgress) != 1 || compiledEgress[0].Direction != ir.PolicyDirectionEgress {
		t.Errorf("expected 1 egress rule, got %v", compiledEgress)
	}
}

func TestReconcileAndGetAll(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = apiv1.AddToScheme(scheme)

	snp := &apiv1.StraitNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "policy-a",
		},
		Spec: apiv1.StraitNetworkPolicySpec{
			Priority: 10,
			PolicyTypes: []apiv1.PolicyType{apiv1.PolicyTypeIngress},
			EndpointSelector: apiv1.EndpointSelector{
				PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "a"}},
			},
			Ingress: []apiv1.IngressRule{
				{Action: apiv1.PolicyActionAllow},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(snp).Build()
	alloc := identitypkg.NewAllocator()
	engine := NewEngine(client, alloc, zap.NewNop())
	ctx := context.Background()

	key := types.NamespacedName{Namespace: "default", Name: "policy-a"}
	if err := engine.Reconcile(ctx, key); err != nil {
		t.Fatalf("Reconcile failed: %v", err)
	}

	policies := engine.GetAll()
	if len(policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(policies))
	}
	if policies[0].Priority != 10 {
		t.Errorf("expected priority 10, got %d", policies[0].Priority)
	}
}

func TestReconcileNotFound(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = apiv1.AddToScheme(scheme)

	client := fake.NewClientBuilder().WithScheme(scheme).Build()
	alloc := identitypkg.NewAllocator()
	engine := NewEngine(client, alloc, zap.NewNop())
	ctx := context.Background()

	key := types.NamespacedName{Namespace: "default", Name: "nonexistent"}
	if err := engine.Reconcile(ctx, key); err != nil {
		t.Fatalf("expected nil error for not found, got: %v", err)
	}
}

func TestDeletePolicy(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = apiv1.AddToScheme(scheme)

	snp := &apiv1.StraitNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "to-delete",
		},
		Spec: apiv1.StraitNetworkPolicySpec{
			PolicyTypes: []apiv1.PolicyType{apiv1.PolicyTypeIngress},
			EndpointSelector: apiv1.EndpointSelector{
				PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "del"}},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(snp).Build()
	alloc := identitypkg.NewAllocator()
	engine := NewEngine(client, alloc, zap.NewNop())
	ctx := context.Background()

	key := types.NamespacedName{Namespace: "default", Name: "to-delete"}
	_ = engine.Reconcile(ctx, key)
	if len(engine.GetAll()) == 0 {
		t.Fatal("expected policy before delete")
	}

	engine.Delete(key)
	if len(engine.GetAll()) != 0 {
		t.Errorf("expected 0 policies after delete, got %d", len(engine.GetAll()))
	}
}

func TestUpsertPolicy(t *testing.T) {
	engine := &Engine{}

	p1 := &ir.Policy{ID: "pol-1", Priority: 100, Action: sgtypes.PolicyActionAllow}
	engine.upsertPolicy(p1)
	if len(engine.policies) != 1 {
		t.Fatalf("expected 1 policy, got %d", len(engine.policies))
	}

	// Update existing policy
	p1Updated := &ir.Policy{ID: "pol-1", Priority: 50, Action: sgtypes.PolicyActionDeny}
	engine.upsertPolicy(p1Updated)
	if len(engine.policies) != 1 {
		t.Fatalf("expected 1 policy after update, got %d", len(engine.policies))
	}
	if engine.policies[0].Priority != 50 {
		t.Errorf("expected priority 50, got %d", engine.policies[0].Priority)
	}
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkCompilePolicy(b *testing.B) {
	for _, ruleCount := range []int{1, 10, 100} {
		b.Run(fmt.Sprintf("rules=%d", ruleCount), func(b *testing.B) {
			alloc := identitypkg.NewAllocator()
			engine := &Engine{identity: alloc}
			ctx := context.Background()

			ingressRules := make([]apiv1.IngressRule, ruleCount)
			for i := range ingressRules {
				ingressRules[i] = apiv1.IngressRule{
					Action: apiv1.PolicyActionAllow,
					From: []apiv1.EndpointSelector{
						{PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": fmt.Sprintf("src-%d", i)}}},
					},
				}
			}

			snp := &apiv1.StraitNetworkPolicy{
				ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "bench-policy"},
				Spec: apiv1.StraitNetworkPolicySpec{
					Priority:    100,
					PolicyTypes: []apiv1.PolicyType{apiv1.PolicyTypeIngress},
					EndpointSelector: apiv1.EndpointSelector{
						PodSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "bench"}},
					},
					Ingress: ingressRules,
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = engine.compile(ctx, snp)
			}
		})
	}
}

func BenchmarkSortPolicies(b *testing.B) {
	for _, count := range []int{100, 1000, 10000} {
		b.Run(fmt.Sprintf("count=%d", count), func(b *testing.B) {
			policies := make([]*ir.Policy, count)
			for i := range policies {
				action := sgtypes.PolicyActionAllow
				if i%2 == 0 {
					action = sgtypes.PolicyActionDeny
				}
				policies[i] = &ir.Policy{
					ID:       fmt.Sprintf("pol-%d", i),
					Priority: uint8(i % 256),
					Action:   action,
				}
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				e := &Engine{policies: make([]*ir.Policy, count)}
				copy(e.policies, policies)
				e.sortPolicies()
			}
		})
	}
}

func BenchmarkResolveIdentities(b *testing.B) {
	alloc := identitypkg.NewAllocator()
	engine := &Engine{identity: alloc}
	ctx := context.Background()

	sel := apiv1.EndpointSelector{
		PodSelector:       &metav1.LabelSelector{MatchLabels: map[string]string{"app": "bench", "tier": "web"}},
		NamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"env": "prod"}},
		SegmentSelector:   &apiv1.SegmentSelector{SegmentID: 100},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.resolveIdentities(ctx, "prod", sel)
	}
}
