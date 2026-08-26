package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ============================================================================
// StraitNetwork — cluster-wide network configuration
// ============================================================================

// StraitNetworkSpec defines the desired network configuration for the cluster.
type StraitNetworkSpec struct {
	// PodCIDR is the cluster pod network CIDR. Discovered dynamically if empty.
	// +optional
	PodCIDR string `json:"podCIDR,omitempty"`

	// ServiceCIDR is the cluster service network CIDR. Discovered dynamically if empty.
	// +optional
	ServiceCIDR string `json:"serviceCIDR,omitempty"`

	// IPv6PodCIDR is the IPv6 pod network CIDR for dual-stack. Discovered dynamically if empty.
	// +optional
	IPv6PodCIDR string `json:"ipv6PodCIDR,omitempty"`

	// IPv6ServiceCIDR is the IPv6 service network CIDR for dual-stack. Discovered dynamically if empty.
	// +optional
	IPv6ServiceCIDR string `json:"ipv6ServiceCIDR,omitempty"`

	// TunnelMode specifies the overlay tunnel type: vxlan, geneve, gre, or disabled.
	// +kubebuilder:validation:Enum=vxlan;geneve;gre;disabled
	// +kubebuilder:default=vxlan
	TunnelMode string `json:"tunnelMode,omitempty"`

	// EnableIPv6 enables IPv6 dual-stack networking.
	// +kubebuilder:default=false
	EnableIPv6 bool `json:"enableIPv6,omitempty"`

	// KubeProxyReplacement enables complete kube-proxy replacement.
	// +kubebuilder:default=true
	KubeProxyReplacement bool `json:"kubeProxyReplacement,omitempty"`

	// MTU is the network MTU. Auto-discovered if 0.
	// +optional
	MTU int `json:"mtu,omitempty"`

	// Encryption configures pod-to-pod encryption.
	// +optional
	Encryption *EncryptionSpec `json:"encryption,omitempty"`
}

// EncryptionSpec defines encryption settings.
type EncryptionSpec struct {
	// Type specifies the encryption type: wireguard or ipsec.
	// +kubebuilder:validation:Enum=wireguard;ipsec;disabled
	// +kubebuilder:default=disabled
	Type string `json:"type,omitempty"`
}

// StraitNetworkStatus defines the observed state of StraitNetwork.
type StraitNetworkStatus struct {
	// Phase is the current lifecycle phase: Pending, Ready, Error.
	Phase string `json:"phase,omitempty"`

	// ObservedGeneration is the last generation observed by the controller.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent the latest available observations.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// DiscoveredPodCIDR is the dynamically discovered pod CIDR.
	DiscoveredPodCIDR string `json:"discoveredPodCIDR,omitempty"`

	// DiscoveredServiceCIDR is the dynamically discovered service CIDR.
	DiscoveredServiceCIDR string `json:"discoveredServiceCIDR,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// StraitNetwork is the Schema for the cluster-wide network configuration.
type StraitNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StraitNetworkSpec   `json:"spec,omitempty"`
	Status StraitNetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StraitNetworkList contains a list of StraitNetwork.
type StraitNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StraitNetwork `json:"items"`
}

// ============================================================================
// StraitNode — per-node agent state
// ============================================================================

// StraitNodeSpec defines the desired state for a node's networking.
type StraitNodeSpec struct {
	// NodeName is the Kubernetes node name.
	NodeName string `json:"nodeName"`

	// PodCIDR is the pod CIDR allocated to this node.
	PodCIDR string `json:"podCIDR,omitempty"`

	// IPv6PodCIDR is the IPv6 pod CIDR allocated to this node.
	// +optional
	IPv6PodCIDR string `json:"ipv6PodCIDR,omitempty"`

	// InternalIP is the node's internal IP address.
	InternalIP string `json:"internalIP,omitempty"`
}

// StraitNodeStatus defines the observed state of a node agent.
type StraitNodeStatus struct {
	// Phase is the current phase: Initializing, Ready, Error.
	Phase string `json:"phase,omitempty"`

	// ObservedGeneration is the last generation observed.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent independent readiness conditions.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// CNIReady indicates the CNI is ready to handle ADD/DEL.
	CNIReady bool `json:"cniReady,omitempty"`

	// ServiceReady indicates the service dataplane is ready.
	ServiceReady bool `json:"serviceReady,omitempty"`

	// PolicyReady indicates the policy engine is ready.
	PolicyReady bool `json:"policyReady,omitempty"`

	// GatewayReady indicates the gateway dataplane is ready.
	GatewayReady bool `json:"gatewayReady,omitempty"`

	// TransitReady indicates the transit gateway is ready.
	TransitReady bool `json:"transitReady,omitempty"`

	// BGPReady indicates the BGP subsystem is ready.
	BGPReady bool `json:"bgpReady,omitempty"`

	// AllocatedIdentities is the count of BPF identities allocated on this node.
	AllocatedIdentities int `json:"allocatedIdentities,omitempty"`

	// KernelVersion is the running kernel version.
	KernelVersion string `json:"kernelVersion,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Node",type=string,JSONPath=`.spec.nodeName`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="CNI",type=boolean,JSONPath=`.status.cniReady`
// +kubebuilder:printcolumn:name="Service",type=boolean,JSONPath=`.status.serviceReady`
// +kubebuilder:printcolumn:name="Policy",type=boolean,JSONPath=`.status.policyReady`
// +kubebuilder:printcolumn:name="Gateway",type=boolean,JSONPath=`.status.gatewayReady`

// StraitNode is the Schema for per-node agent state.
type StraitNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StraitNodeSpec   `json:"spec,omitempty"`
	Status StraitNodeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StraitNodeList contains a list of StraitNode.
type StraitNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StraitNode `json:"items"`
}

// ============================================================================
// StraitNetworkPolicy — enhanced Kubernetes NetworkPolicy
// ============================================================================

// StraitNetworkPolicySpec defines a network policy with extended selectors.
type StraitNetworkPolicySpec struct {
	// PolicyType specifies whether this policy applies to Ingress, Egress, or both.
	// +kubebuilder:validation:MinItems=1
	PolicyTypes []PolicyType `json:"policyTypes"`

	// Priority determines rule evaluation order. Lower value = higher priority.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=255
	// +kubebuilder:default=100
	Priority int `json:"priority,omitempty"`

	// EndpointSelector selects the pods to which this policy applies.
	EndpointSelector EndpointSelector `json:"endpointSelector"`

	// Ingress defines the ingress rules.
	// +optional
	Ingress []IngressRule `json:"ingress,omitempty"`

	// Egress defines the egress rules.
	// +optional
	Egress []EgressRule `json:"egress,omitempty"`
}

// PolicyType represents the direction of traffic.
// +kubebuilder:validation:Enum=Ingress;Egress
type PolicyType string

const (
	PolicyTypeIngress PolicyType = "Ingress"
	PolicyTypeEgress  PolicyType = "Egress"
)

// PolicyAction defines what to do when a rule matches.
// +kubebuilder:validation:Enum=Allow;Deny;Reject
type PolicyAction string

const (
	PolicyActionAllow  PolicyAction = "Allow"
	PolicyActionDeny   PolicyAction = "Deny"
	PolicyActionReject PolicyAction = "Reject"
)

// EndpointSelector selects endpoints using extended selectors.
type EndpointSelector struct {
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty"`
	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty"`
	// +optional
	ClusterSelector *metav1.LabelSelector `json:"clusterSelector,omitempty"`
	// +optional
	SegmentSelector *SegmentSelector `json:"segmentSelector,omitempty"`
	// +optional
	GatewaySelector *metav1.LabelSelector `json:"gatewaySelector,omitempty"`
	// +optional
	HTTPRouteSelector *metav1.LabelSelector `json:"httprouteSelector,omitempty"`
	// +optional
	TCPRouteSelector *metav1.LabelSelector `json:"tcprouteSelector,omitempty"`
	// +optional
	UDPRouteSelector *metav1.LabelSelector `json:"udprouteSelector,omitempty"`
	// +optional
	GRPCRouteSelector *metav1.LabelSelector `json:"grpcrouteSelector,omitempty"`
	// +optional
	TLSRouteSelector *metav1.LabelSelector `json:"tlsrouteSelector,omitempty"`
}

// SegmentSelector selects by transit segment ID.
type SegmentSelector struct {
	// SegmentID is the 32-bit transit segment identifier.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	SegmentID uint32 `json:"segmentID"`
}

// IngressRule defines an ingress policy rule.
type IngressRule struct {
	// Action is the action to take when this rule matches.
	// +kubebuilder:default=Allow
	Action PolicyAction `json:"action,omitempty"`

	// From defines the sources allowed.
	// +optional
	From []EndpointSelector `json:"from,omitempty"`

	// Ports defines the ports/protocols allowed.
	// +optional
	Ports []PolicyPort `json:"ports,omitempty"`
}

// EgressRule defines an egress policy rule.
type EgressRule struct {
	// Action is the action to take when this rule matches.
	// +kubebuilder:default=Allow
	Action PolicyAction `json:"action,omitempty"`

	// To defines the destinations allowed.
	// +optional
	To []EndpointSelector `json:"to,omitempty"`

	// Ports defines the ports/protocols allowed.
	// +optional
	Ports []PolicyPort `json:"ports,omitempty"`
}

// PolicyPort defines a protocol/port combination.
type PolicyPort struct {
	// Protocol is the network protocol: TCP, UDP, or ICMP.
	// +kubebuilder:validation:Enum=TCP;UDP;ICMP
	// +kubebuilder:default=TCP
	Protocol string `json:"protocol,omitempty"`

	// Port is the port number.
	// +optional
	Port *int32 `json:"port,omitempty"`

	// EndPort is the end of a port range (inclusive).
	// +optional
	EndPort *int32 `json:"endPort,omitempty"`
}

// StraitNetworkPolicyStatus defines the observed state.
type StraitNetworkPolicyStatus struct {
	// Phase is the policy lifecycle phase.
	Phase string `json:"phase,omitempty"`

	// ObservedGeneration is the last generation observed.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent the latest available observations.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// EnforcedNodes is the count of nodes enforcing this policy.
	EnforcedNodes int `json:"enforcedNodes,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Priority",type=integer,JSONPath=`.spec.priority`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// StraitNetworkPolicy is the Schema for enhanced network policies.
type StraitNetworkPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StraitNetworkPolicySpec   `json:"spec,omitempty"`
	Status StraitNetworkPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StraitNetworkPolicyList contains a list of StraitNetworkPolicy.
type StraitNetworkPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StraitNetworkPolicy `json:"items"`
}

// ============================================================================
// TransitGateway — multi-cluster transit gateway
// ============================================================================

// TransitGatewaySpec defines the desired state of a transit gateway.
type TransitGatewaySpec struct {
	// SegmentID is the transit segment this gateway belongs to.
	// Segment 0 is the backbone segment.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	// +kubebuilder:default=0
	SegmentID uint32 `json:"segmentID,omitempty"`

	// Topology is the gateway topology: hub-spoke, mesh, or peer-to-peer.
	// +kubebuilder:validation:Enum=hub-spoke;mesh;peer-to-peer
	// +kubebuilder:default=hub-spoke
	Topology string `json:"topology,omitempty"`

	// Peers defines the cluster peers for this gateway.
	// +optional
	Peers []TransitPeer `json:"peers,omitempty"`

	// Encryption configures tunnel encryption.
	// +optional
	Encryption *EncryptionSpec `json:"encryption,omitempty"`
}

// TransitPeer defines a remote cluster peer.
type TransitPeer struct {
	// ClusterName is the name of the remote cluster.
	ClusterName string `json:"clusterName"`

	// Endpoint is the remote cluster's gateway endpoint address.
	Endpoint string `json:"endpoint"`

	// Port is the remote cluster's gateway port.
	// +kubebuilder:default=4789
	Port int32 `json:"port,omitempty"`

	// PodCIDRs are the remote cluster's pod CIDRs.
	PodCIDRs []string `json:"podCIDRs,omitempty"`

	// ServiceCIDRs are the remote cluster's service CIDRs.
	ServiceCIDRs []string `json:"serviceCIDRs,omitempty"`
}

// TransitGatewayStatus defines the observed state of a transit gateway.
type TransitGatewayStatus struct {
	// Phase is the current lifecycle phase.
	Phase string `json:"phase,omitempty"`

	// ObservedGeneration is the last generation observed.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions represent the latest available observations.
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ConnectedPeers is the count of connected peers.
	ConnectedPeers int `json:"connectedPeers,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="Segment",type=integer,JSONPath=`.spec.segmentID`
// +kubebuilder:printcolumn:name="Topology",type=string,JSONPath=`.spec.topology`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`

// TransitGateway is the Schema for multi-cluster transit gateways.
type TransitGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitGatewaySpec   `json:"spec,omitempty"`
	Status TransitGatewayStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TransitGatewayList contains a list of TransitGateway.
type TransitGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitGateway `json:"items"`
}

// ============================================================================
// TransitSegmentAttachment — inter-segment communication
// ============================================================================

// TransitSegmentAttachmentSpec defines an attachment between two segments.
type TransitSegmentAttachmentSpec struct {
	// SegmentID is the segment this attachment belongs to.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	SegmentID uint32 `json:"segmentID"`

	// PeerSegmentID is the segment to peer with.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	PeerSegmentID uint32 `json:"peerSegmentID"`

	// Routes defines the routing table entries for this attachment.
	// +optional
	Routes []TransitSegmentRoute `json:"routes,omitempty"`
}

// TransitSegmentRoute defines a route for inter-segment communication.
type TransitSegmentRoute struct {
	// CIDR is the destination CIDR for this route.
	CIDR string `json:"cidr"`

	// NextHop is the next-hop address or attachment name.
	NextHop string `json:"nextHop"`
}

// TransitSegmentAttachmentStatus defines the observed state.
type TransitSegmentAttachmentStatus struct {
	Phase              string             `json:"phase,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// TransitSegmentAttachment is the Schema for inter-segment attachments.
type TransitSegmentAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitSegmentAttachmentSpec   `json:"spec,omitempty"`
	Status TransitSegmentAttachmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TransitSegmentAttachmentList contains a list of TransitSegmentAttachment.
type TransitSegmentAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitSegmentAttachment `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&StraitNetwork{}, &StraitNetworkList{},
		&StraitNode{}, &StraitNodeList{},
		&StraitNetworkPolicy{}, &StraitNetworkPolicyList{},
		&TransitGateway{}, &TransitGatewayList{},
		&TransitSegmentAttachment{}, &TransitSegmentAttachmentList{},
	)
}
