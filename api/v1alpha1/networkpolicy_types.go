package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="PolicyType",type=string,JSONPath=`.spec.policyType`
// +kubebuilder:printcolumn:name="DefaultAction",type=string,JSONPath=`.spec.defaultAction`
// +kubebuilder:printcolumn:name="Priority",type=integer,JSONPath=`.spec.priority`

// StraitNetworkPolicy is the enhanced network policy CRD for straitKubegateway.
// It extends Kubernetes NetworkPolicy with clusterSelector, segmentSelector,
// and Gateway API route selectors.
//
// Policy evaluation semantics:
//   - Priority: uint32 (lower value = higher priority)
//   - Rule ordering within policy: RuleNo (1-based uint32; RuleNo=0 indicates discarded rule)
//   - Deny overrides allow based on Priority + RuleNo positioning
//   - Default ingress action: Deny-all
//   - Default egress action: Allow-all
//   - Policy hierarchy: Cluster > Segment > Namespace
//   - Cluster policies cannot be overridden by namespace policies
//   - Segment policies can override namespace policies
type StraitNetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StraitNetworkPolicySpec   `json:"spec,omitempty"`
	Status StraitNetworkPolicyStatus `json:"status,omitempty"`
}

// StraitNetworkPolicySpec defines the desired network policy.
type StraitNetworkPolicySpec struct {
	// PolicyType is the type of traffic this policy applies to.
	// +kubebuilder:validation:Enum=Ingress;Egress;Both
	PolicyType string `json:"policyType"`

	// Priority determines the evaluation order. Lower value = higher priority.
	// +kubebuilder:validation:Minimum=0
	Priority uint32 `json:"priority"`

	// DefaultAction is the action when no rule matches.
	// Defaults to Deny for Ingress, Allow for Egress.
	// +kubebuilder:validation:Enum=Allow;Deny;Reject
	DefaultAction string `json:"defaultAction,omitempty"`

	// Scope defines whether this policy is cluster-scoped, segment-scoped,
	// or namespace-scoped. Hierarchy: Cluster > Segment > Namespace.
	// +kubebuilder:validation:Enum=Cluster;Segment;Namespace
	// +kubebuilder:default=Namespace
	Scope string `json:"scope,omitempty"`

	// Ingress specifies ingress rules evaluated in order of RuleNo.
	Ingress []PolicyRule `json:"ingress,omitempty"`

	// Egress specifies egress rules evaluated in order of RuleNo.
	Egress []PolicyRule `json:"egress,omitempty"`
}

// PolicyRule defines a single policy rule for ingress or egress traffic.
type PolicyRule struct {
	// RuleNo is the 1-based incrementing rule index (RuleNo 0 = discard rule).
	// +kubebuilder:validation:Minimum=0
	RuleNo uint32 `json:"ruleNo"`

	// Action is the action to take when this rule matches (Allow, Deny, Reject).
	// +kubebuilder:validation:Enum=Allow;Deny;Reject
	Action string `json:"action"`

	// From specifies the source selectors for this rule (ingress).
	From []PolicyPeer `json:"from,omitempty"`

	// To specifies the destination selectors for this rule (egress).
	To []PolicyPeer `json:"to,omitempty"`

	// Ports specifies the ports and protocols this rule applies to.
	Ports []PolicyPort `json:"ports,omitempty"`

	// Log specifies whether traffic matching this rule should be logged.
	Log bool `json:"log,omitempty"`
}

// PolicyPeer defines a peer selector for network policy rules.
// It extends standard Kubernetes selectors with cluster, segment,
// and Gateway API route selectors.
type PolicyPeer struct {
	// PodSelector selects pods by labels.
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty"`

	// NamespaceSelector selects namespaces by labels.
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`

	// ClusterSelector selects clusters by labels (multi-cluster).
	ClusterSelector *metav1.LabelSelector `json:"clusterSelector,omitempty"`

	// SegmentSelector selects network segments by labels.
	SegmentSelector *metav1.LabelSelector `json:"segmentSelector,omitempty"`

	// GatewaySelector selects gateways by labels.
	GatewaySelector *metav1.LabelSelector `json:"gatewaySelector,omitempty"`

	// HTTPRouteSelector selects HTTPRoute resources by labels.
	HTTPRouteSelector *metav1.LabelSelector `json:"httprouteSelector,omitempty"`

	// TCPRouteSelector selects TCPRoute resources by labels.
	TCPRouteSelector *metav1.LabelSelector `json:"tcprouteSelector,omitempty"`

	// UDPRouteSelector selects UDPRoute resources by labels.
	UDPRouteSelector *metav1.LabelSelector `json:"udprouteSelector,omitempty"`

	// GRPCRouteSelector selects GRPCRoute resources by labels.
	GRPCRouteSelector *metav1.LabelSelector `json:"grpcrouteSelector,omitempty"`

	// TLSRouteSelector selects TLSRoute resources by labels.
	TLSRouteSelector *metav1.LabelSelector `json:"tlsrouteSelector,omitempty"`
}

// PolicyPort defines a port and protocol for a policy rule.
type PolicyPort struct {
	// Protocol is the network protocol.
	// +kubebuilder:validation:Enum=TCP;UDP;ICMP
	Protocol string `json:"protocol"`

	// Port is the port number.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port *int32 `json:"port,omitempty"`

	// EndPort is the end of a port range (inclusive).
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	EndPort *int32 `json:"endPort,omitempty"`
}

// StraitNetworkPolicyStatus defines the observed state.
type StraitNetworkPolicyStatus struct {
	// Phase is the current policy state.
	Phase string `json:"phase,omitempty"`

	// Conditions represent the policy's current conditions.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// EnforcedEndpoints is the number of endpoints enforcing this policy.
	EnforcedEndpoints int32 `json:"enforcedEndpoints,omitempty"`

	// ActiveRulesCount is the count of active (non-discarded) compiled rules.
	ActiveRulesCount int32 `json:"activeRulesCount,omitempty"`

	// ObservedGeneration is the most recently observed generation.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true

// StraitNetworkPolicyList contains a list of StraitNetworkPolicies.
type StraitNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StraitNetworkPolicy `json:"items"`
}
