package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Gateway represents a straitKubegateway gateway instance.
type Gateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GatewaySpec   `json:"spec,omitempty"`
	Status GatewayStatus `json:"status,omitempty"`
}

// GatewaySpec defines the desired state of a Gateway.
type GatewaySpec struct {
	// Mode is the gateway operating mode (standalone, hub, spoke, mesh).
	// +kubebuilder:validation:Enum=standalone;hub;spoke;mesh
	Mode string `json:"mode"`

	// SegmentID is the network segment this gateway belongs to.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	SegmentID uint32 `json:"segmentId,omitempty"`

	// Listeners defines the network listeners for this gateway.
	Listeners []GatewayListener `json:"listeners,omitempty"`

	// NodeSelector selects which nodes run this gateway.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`

	// Encryption configures transit encryption.
	Encryption *EncryptionConfig `json:"encryption,omitempty"`
}

// GatewayListener defines a network listener on the gateway.
type GatewayListener struct {
	// Name is a unique name for this listener.
	Name string `json:"name"`

	// Protocol is the protocol to listen on (TCP, UDP, HTTP, HTTPS, TLS, gRPC).
	// +kubebuilder:validation:Enum=TCP;UDP;HTTP;HTTPS;TLS;gRPC
	Protocol string `json:"protocol"`

	// Port is the port number to listen on.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
}

// EncryptionConfig specifies encryption settings for transit.
type EncryptionConfig struct {
	// Type is the encryption type (wireguard, ipsec, none).
	// +kubebuilder:validation:Enum=wireguard;ipsec;none
	Type string `json:"type"`

	// KeyRotationInterval is the key rotation interval in seconds.
	KeyRotationInterval int32 `json:"keyRotationInterval,omitempty"`
}

// GatewayStatus defines the observed state of a Gateway.
type GatewayStatus struct {
	// Phase is the current lifecycle phase.
	// +kubebuilder:validation:Enum=Pending;Running;Degraded;Failed;Terminating
	Phase string `json:"phase,omitempty"`

	// Conditions represent the gateway's current conditions.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// ActiveListeners is the number of active listeners.
	ActiveListeners int32 `json:"activeListeners,omitempty"`

	// ConnectedPeers is the number of connected peer gateways.
	ConnectedPeers int32 `json:"connectedPeers,omitempty"`

	// ObservedGeneration is the most recently observed generation.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true

// GatewayList contains a list of Gateways.
type GatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gateway `json:"items"`
}
