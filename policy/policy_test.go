package policy_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/straitKubegateway/straitKubegateway/api/v1alpha1"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
	"github.com/straitKubegateway/straitKubegateway/policy"
)

func TestCompilerRuleNoDiscard(t *testing.T) {
	compiler := policy.NewCompiler()

	p := &v1alpha1.StraitNetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{Name: "test-policy", Namespace: "default"},
		Spec: v1alpha1.StraitNetworkPolicySpec{
			Scope:    "Namespace",
			Priority: 10,
			Ingress: []v1alpha1.PolicyRule{
				{RuleNo: 0, Action: "Allow"}, // Discard rule!
				{RuleNo: 1, Action: "Allow"}, // Valid rule
				{RuleNo: 2, Action: "Deny"},  // Valid rule
			},
		},
	}

	rules := compiler.Compile(p, nil, nil)
	if len(rules) != 2 {
		t.Fatalf("expected 2 compiled rules (ruleNo=0 discarded), got %d", len(rules))
	}
	if rules[0].RuleNo != 1 || rules[1].RuleNo != 2 {
		t.Errorf("expected rules 1 and 2, got %d and %d", rules[0].RuleNo, rules[1].RuleNo)
	}
}

func TestScopeHierarchy(t *testing.T) {
	// Cluster policy (Scope=Cluster, Priority=50, RuleNo=1, Action=Deny)
	// Segment policy (Scope=Segment, Priority=10, RuleNo=1, Action=Allow)
	// Namespace policy (Scope=Namespace, Priority=5, RuleNo=1, Action=Allow)
	// Hierarchy: Cluster > Segment > Namespace
	// Even though Namespace has Priority 5, Cluster (ScopeRank 1) takes precedence!
	clusterRule := policy.CompiledRule{
		PolicyName: "cluster-policy",
		Scope:      policy.ScopeCluster,
		Priority:   50,
		RuleNo:     1,
		Action:     policy.ActionDeny,
		Direction:  types.DirectionIngress,
		Protocol:   types.ProtocolTCP,
		DstPort:    80,
	}

	namespaceRule := policy.CompiledRule{
		PolicyName: "ns-policy",
		Scope:      policy.ScopeNamespace,
		Priority:   5,
		RuleNo:     1,
		Action:     policy.ActionAllow,
		Direction:  types.DirectionIngress,
		Protocol:   types.ProtocolTCP,
		DstPort:    80,
	}

	rules := []policy.CompiledRule{namespaceRule, clusterRule}
	evaluator := policy.NewEvaluator(rules)

	pkt := policy.PacketContext{
		Direction: types.DirectionIngress,
		Protocol:  types.ProtocolTCP,
		DstPort:   80,
	}

	decision := evaluator.Evaluate(pkt)
	if decision.Action != policy.ActionDeny {
		t.Errorf("expected Cluster scope rule (Deny) to take precedence over Namespace rule, got %v", decision.Action)
	}
	if decision.MatchingRule.PolicyName != "cluster-policy" {
		t.Errorf("expected matched policy 'cluster-policy', got %s", decision.MatchingRule.PolicyName)
	}
}

func TestDefaultIngressAndEgressActions(t *testing.T) {
	evaluator := policy.NewEvaluator(nil) // No rules configured

	// Ingress -> Default Deny-all
	ingressPkt := policy.PacketContext{
		Direction: types.DirectionIngress,
		Protocol:  types.ProtocolTCP,
		DstPort:   443,
	}
	ingressDecision := evaluator.Evaluate(ingressPkt)
	if ingressDecision.Action != policy.ActionDeny {
		t.Errorf("expected default Ingress action to be Deny, got %v", ingressDecision.Action)
	}

	// Egress -> Default Allow-all
	egressPkt := policy.PacketContext{
		Direction: types.DirectionEgress,
		Protocol:  types.ProtocolTCP,
		DstPort:   443,
	}
	egressDecision := evaluator.Evaluate(egressPkt)
	if egressDecision.Action != policy.ActionAllow {
		t.Errorf("expected default Egress action to be Allow, got %v", egressDecision.Action)
	}
}

func TestPriorityAndRuleNoOrdering(t *testing.T) {
	// Rule 1: Priority 10, RuleNo 1 -> Action Deny
	// Rule 2: Priority 10, RuleNo 2 -> Action Allow
	// At same priority, RuleNo 1 (Deny) evaluated before RuleNo 2
	ruleDeny := policy.CompiledRule{
		Scope:     policy.ScopeNamespace,
		Priority:  10,
		RuleNo:    1,
		Action:    policy.ActionDeny,
		Direction: types.DirectionIngress,
		Protocol:  types.ProtocolTCP,
		DstPort:   8080,
	}
	ruleAllow := policy.CompiledRule{
		Scope:     policy.ScopeNamespace,
		Priority:  10,
		RuleNo:    2,
		Action:    policy.ActionAllow,
		Direction: types.DirectionIngress,
		Protocol:  types.ProtocolTCP,
		DstPort:   8080,
	}

	evaluator := policy.NewEvaluator([]policy.CompiledRule{ruleAllow, ruleDeny})

	pkt := policy.PacketContext{
		Direction: types.DirectionIngress,
		Protocol:  types.ProtocolTCP,
		DstPort:   8080,
	}

	decision := evaluator.Evaluate(pkt)
	if decision.Action != policy.ActionDeny {
		t.Errorf("expected RuleNo 1 (Deny) to be selected over RuleNo 2, got %v", decision.Action)
	}
}
