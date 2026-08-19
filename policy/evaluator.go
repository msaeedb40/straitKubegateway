package policy

import (
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// DecisionResult represents the outcome of evaluating a packet against active policies.
type DecisionResult struct {
	Action       Action
	MatchingRule *CompiledRule
	Reason       string
}

// PacketContext represents the packet attributes evaluated against policies.
type PacketContext struct {
	SrcIdentity types.Identity
	DstIdentity types.Identity
	DstPort     uint16
	Protocol    types.Protocol
	Direction   types.Direction
	Namespace   string
	SegmentID   types.SegmentID
}

// Evaluator simulates and evaluates policy rules against packet contexts in memory.
type Evaluator struct {
	rules []CompiledRule
}

// NewEvaluator creates an evaluator with a given set of compiled rules.
func NewEvaluator(rules []CompiledRule) *Evaluator {
	sorted := make([]CompiledRule, len(rules))
	copy(sorted, rules)
	SortRules(sorted)
	return &Evaluator{rules: sorted}
}

// Evaluate determines the policy decision for a packet.
// Evaluates rules in sorted order (Cluster > Segment > Namespace, Priority, RuleNo).
// Defaults: Ingress -> Deny (Deny-all), Egress -> Allow (Allow-all).
func (e *Evaluator) Evaluate(pkt PacketContext) DecisionResult {
	for i := range e.rules {
		r := &e.rules[i]

		// Check traffic direction
		if r.Direction != pkt.Direction {
			continue
		}

		// Check protocol (if specified)
		if r.Protocol != 0 && r.Protocol != pkt.Protocol {
			continue
		}

		// Check port (if specified)
		if r.DstPort != 0 && r.DstPort != pkt.DstPort {
			continue
		}

		// Rule matched!
		return DecisionResult{
			Action:       r.Action,
			MatchingRule: r,
			Reason:       "matched rule",
		}
	}

	// Default fallback:
	// Ingress: Deny-all
	// Egress: Allow-all
	if pkt.Direction == types.DirectionIngress {
		return DecisionResult{
			Action: ActionDeny,
			Reason: "default ingress deny-all",
		}
	}

	return DecisionResult{
		Action: ActionAllow,
		Reason: "default egress allow-all",
	}
}
