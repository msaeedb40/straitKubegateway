package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// BGPPeer represents a BGP peering configuration.
type BGPPeer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BGPPeerSpec   `json:"spec,omitempty"`
	Status BGPPeerStatus `json:"status,omitempty"`
}

// BGPPeerSpec defines the desired BGP peer configuration.
type BGPPeerSpec struct {
	// PeerASN is the remote peer's autonomous system number.
	PeerASN uint32 `json:"peerAsn"`

	// LocalASN is the local autonomous system number.
	LocalASN uint32 `json:"localAsn"`

	// PeerAddress is the remote peer's IP address.
	PeerAddress string `json:"peerAddress"`

	// LocalAddress is the local IP address to peer from.
	LocalAddress string `json:"localAddress,omitempty"`

	// HoldTime is the BGP hold time in seconds.
	// +kubebuilder:default=90
	HoldTime int32 `json:"holdTime,omitempty"`

	// KeepaliveInterval is the keepalive interval in seconds.
	// +kubebuilder:default=30
	KeepaliveInterval int32 `json:"keepaliveInterval,omitempty"`

	// BFDEnabled enables Bidirectional Forwarding Detection.
	BFDEnabled bool `json:"bfdEnabled,omitempty"`

	// AdvertisedPrefixes is the list of prefixes to advertise.
	AdvertisedPrefixes []string `json:"advertisedPrefixes,omitempty"`

	// RouteFilters are filters applied to received routes.
	RouteFilters []BGPRouteFilter `json:"routeFilters,omitempty"`

	// NodeSelector selects which nodes establish this peering.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
}

// BGPRouteFilter defines a filter for BGP route advertisement or learning.
type BGPRouteFilter struct {
	// Prefix is the CIDR prefix to match.
	Prefix string `json:"prefix"`

	// Action is the filter action (accept, reject).
	// +kubebuilder:validation:Enum=accept;reject
	Action string `json:"action"`

	// MatchType is how the prefix is matched (exact, prefix, longer).
	// +kubebuilder:validation:Enum=exact;prefix;longer
	// +kubebuilder:default=exact
	MatchType string `json:"matchType,omitempty"`
}

// BGPPeerStatus defines the observed state of a BGP peer.
type BGPPeerStatus struct {
	// SessionState is the BGP session state.
	// +kubebuilder:validation:Enum=Idle;Connect;Active;OpenSent;OpenConfirm;Established
	SessionState string `json:"sessionState,omitempty"`

	// ReceivedPrefixes is the number of received prefixes.
	ReceivedPrefixes int32 `json:"receivedPrefixes,omitempty"`

	// AdvertisedPrefixes is the number of advertised prefixes.
	AdvertisedPrefixes int32 `json:"advertisedPrefixes,omitempty"`

	// UptimeSeconds is how long the session has been established.
	UptimeSeconds int64 `json:"uptimeSeconds,omitempty"`

	// Conditions represent the peer's current conditions.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the most recently observed generation.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true

// BGPPeerList contains a list of BGPPeers.
type BGPPeerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BGPPeer `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// IPAMPool represents an IP address pool for pod CIDR allocation.
type IPAMPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IPAMPoolSpec   `json:"spec,omitempty"`
	Status IPAMPoolStatus `json:"status,omitempty"`
}

// IPAMPoolSpec defines the desired IPAM pool configuration.
type IPAMPoolSpec struct {
	// CIDRs is the list of CIDRs available for allocation.
	CIDRs []string `json:"cidrs"`

	// PerNodeMaskSize is the prefix length allocated to each node.
	// +kubebuilder:validation:Minimum=8
	// +kubebuilder:validation:Maximum=30
	PerNodeMaskSize int32 `json:"perNodeMaskSize"`

	// AddressFamily is the address family (IPv4, IPv6).
	// +kubebuilder:validation:Enum=IPv4;IPv6
	AddressFamily string `json:"addressFamily"`
}

// IPAMPoolStatus defines the observed state.
type IPAMPoolStatus struct {
	// AllocatedCIDRs tracks which CIDRs are allocated to which nodes.
	AllocatedCIDRs map[string]string `json:"allocatedCidrs,omitempty"`

	// AvailableIPs is the number of remaining allocatable IPs.
	AvailableIPs int64 `json:"availableIps,omitempty"`

	// Conditions represent the pool's current conditions.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ObservedGeneration is the most recently observed generation.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true

// IPAMPoolList contains a list of IPAMPools.
type IPAMPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IPAMPool `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// ClusterLink represents a federated cluster link for multi-cluster networking.
type ClusterLink struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterLinkSpec   `json:"spec,omitempty"`
	Status ClusterLinkStatus `json:"status,omitempty"`
}

// ClusterLinkSpec defines the desired cluster link configuration.
type ClusterLinkSpec struct {
	// ClusterID is the unique identifier for the remote cluster.
	ClusterID string `json:"clusterId"`

	// APIEndpoint is the Kubernetes API endpoint of the remote cluster.
	APIEndpoint string `json:"apiEndpoint"`

	// PodCIDRs are the pod CIDRs of the remote cluster.
	PodCIDRs []string `json:"podCidrs,omitempty"`

	// ServiceCIDRs are the service CIDRs of the remote cluster.
	ServiceCIDRs []string `json:"serviceCidrs,omitempty"`

	// SecretRef references the secret containing kubeconfig for the remote cluster.
	SecretRef string `json:"secretRef,omitempty"`
}

// ClusterLinkStatus defines the observed state.
type ClusterLinkStatus struct {
	Phase              string             `json:"phase,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	Connected          bool               `json:"connected,omitempty"`
	LastHeartbeat      *metav1.Time       `json:"lastHeartbeat,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterLinkList contains a list of ClusterLinks.
type ClusterLinkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterLink `json:"items"`
}
