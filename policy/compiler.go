// Package policy provides the NetworkPolicy compiler, evaluator, and BPF map
// generator for straitKubegateway.
//
// Policy Semantics:
//   - Priority: uint32 (lower value = higher priority)
//   - RuleNo: uint32 (1-based index; RuleNo=0 indicates discarded rule)
//   - Deny overrides allow based on Priority + RuleNo positioning
//   - Scope hierarchy: Cluster (1) > Segment (2) > Namespace (3)
//   - Default Ingress: Deny-all
//   - Default Egress: Allow-all
package policy

import (
	"sort"
	"strings"

	"github.com/straitKubegateway/straitKubegateway/api/v1alpha1"
	"github.com/straitKubegateway/straitKubegateway/pkg/types"
)

// Action represents a policy enforcement action.
type Action uint32

const (
	ActionUnspecified Action = 0
	ActionAllow       Action = 1
	ActionDeny        Action = 2
	ActionReject      Action = 3
)

// String returns the action name.
func (a Action) String() string {
	switch a {
	case ActionAllow:
		return "Allow"
	case ActionDeny:
		return "Deny"
	case ActionReject:
		return "Reject"
	default:
		return "Unknown"
	}
}

// ParseAction converts a string to an Action.
func ParseAction(s string) Action {
	switch strings.ToLower(s) {
	case "allow":
		return ActionAllow
	case "deny":
		return ActionDeny
	case "reject":
		return ActionReject
	default:
		return ActionDeny
	}
}

// ScopeRank represents the hierarchical authority of a policy.
type ScopeRank uint8

const (
	ScopeCluster   ScopeRank = 1 // Highest authority
	ScopeSegment   ScopeRank = 2 // Intermediate authority
	ScopeNamespace ScopeRank = 3 // Standard namespace authority
)

// ParseScopeRank returns the ScopeRank for a scope string.
func ParseScopeRank(s string) ScopeRank {
	switch strings.ToLower(s) {
	case "cluster":
		return ScopeCluster
	case "segment":
		return ScopeSegment
	default:
		return ScopeNamespace
	}
}

// CompiledRule represents a single compiled BPF-ready policy rule.
type CompiledRule struct {
	PolicyID    string
	PolicyName  string
	Namespace   string
	Scope       ScopeRank
	Priority    uint32
	RuleNo      uint32
	Action      Action
	SrcIdentity types.Identity
	DstIdentity types.Identity
	DstPort     uint16
	Protocol    types.Protocol
	Direction   types.Direction
	Log         bool
}

// Compiler compiles StraitNetworkPolicy CRDs into sorted, BPF-ready rule sets.
type Compiler struct{}

// NewCompiler creates a new policy compiler.
func NewCompiler() *Compiler {
	return &Compiler{}
}

// Compile translates a StraitNetworkPolicy into a slice of CompiledRules.
// Rules with RuleNo=0 are discarded.
func (c *Compiler) Compile(p *v1alpha1.StraitNetworkPolicy, srcIDMap, dstIDMap map[string]types.Identity) []CompiledRule {
	scope := ParseScopeRank(p.Spec.Scope)
	priority := p.Spec.Priority

	var rules []CompiledRule

	// 1. Compile Ingress Rules
	for _, rule := range p.Spec.Ingress {
		if rule.RuleNo == 0 {
			continue // Discard rule as specified: RuleNo 0 is discarded
		}

		action := ParseAction(rule.Action)
		ports := extractPorts(rule.Ports)

		for _, portProto := range ports {
			r := CompiledRule{
				PolicyID:    string(p.UID),
				PolicyName:  p.Name,
				Namespace:   p.Namespace,
				Scope:       scope,
				Priority:    priority,
				RuleNo:      rule.RuleNo,
				Action:      action,
				SrcIdentity: types.IdentityWorld, // Default or resolved from selectors
				DstIdentity: types.IdentityUnknown,
				DstPort:     portProto.Port,
				Protocol:    portProto.Protocol,
				Direction:   types.DirectionIngress,
				Log:         rule.Log,
			}
			rules = append(rules, r)
		}
	}

	// 2. Compile Egress Rules
	for _, rule := range p.Spec.Egress {
		if rule.RuleNo == 0 {
			continue
		}

		action := ParseAction(rule.Action)
		ports := extractPorts(rule.Ports)

		for _, portProto := range ports {
			r := CompiledRule{
				PolicyID:    string(p.UID),
				PolicyName:  p.Name,
				Namespace:   p.Namespace,
				Scope:       scope,
				Priority:    priority,
				RuleNo:      rule.RuleNo,
				Action:      action,
				SrcIdentity: types.IdentityUnknown,
				DstIdentity: types.IdentityWorld,
				DstPort:     portProto.Port,
				Protocol:    portProto.Protocol,
				Direction:   types.DirectionEgress,
				Log:         rule.Log,
			}
			rules = append(rules, r)
		}
	}

	// Sort rules according to hierarchy (Scope -> Priority -> RuleNo)
	SortRules(rules)
	return rules
}

// SortRules orders rules by:
// 1. Scope (Cluster < Segment < Namespace in rank, where 1=Cluster has highest priority)
// 2. Priority (lower value = higher priority)
// 3. RuleNo (1-based index)
func SortRules(rules []CompiledRule) {
	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].Scope != rules[j].Scope {
			return rules[i].Scope < rules[j].Scope // ScopeCluster (1) < ScopeSegment (2) < ScopeNamespace (3)
		}
		if rules[i].Priority != rules[j].Priority {
			return rules[i].Priority < rules[j].Priority
		}
		return rules[i].RuleNo < rules[j].RuleNo
	})
}

type portProtocol struct {
	Port     uint16
	Protocol types.Protocol
}

func extractPorts(ports []v1alpha1.PolicyPort) []portProtocol {
	if len(ports) == 0 {
		return []portProtocol{{Port: 0, Protocol: types.ProtocolTCP}}
	}

	var res []portProtocol
	for _, p := range ports {
		var proto types.Protocol
		switch strings.ToUpper(p.Protocol) {
		case "UDP":
			proto = types.ProtocolUDP
		case "ICMP":
			proto = types.ProtocolICMP
		default:
			proto = types.ProtocolTCP
		}

		var port uint16
		if p.Port != nil {
			port = uint16(*p.Port)
		}

		res = append(res, portProtocol{
			Port:     port,
			Protocol: proto,
		})
	}
	return res
}
