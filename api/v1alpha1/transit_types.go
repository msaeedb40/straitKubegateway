package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// TransitGateway represents a multi-cluster transit gateway.
type TransitGateway struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitGatewaySpec   `json:"spec,omitempty"`
	Status TransitGatewayStatus `json:"status,omitempty"`
}

// TransitGatewaySpec defines the desired state of a TransitGateway.
type TransitGatewaySpec struct {
	// Topology is the transit gateway topology (hub-spoke, mesh, peer-to-peer).
	// +kubebuilder:validation:Enum=hub-spoke;mesh;peer-to-peer
	Topology string `json:"topology"`

	// SegmentID is the backbone segment for this transit gateway.
	// Default is 0 (backbone).
	// +kubebuilder:default=0
	SegmentID uint32 `json:"segmentId,omitempty"`

	// Encryption configures transit encryption.
	Encryption *EncryptionConfig `json:"encryption,omitempty"`

	// TunnelType specifies the tunnel encapsulation type.
	// +kubebuilder:validation:Enum=vxlan;geneve;gre;wireguard
	TunnelType string `json:"tunnelType,omitempty"`
}

// TransitGatewayStatus defines the observed state.
type TransitGatewayStatus struct {
	Phase              string             `json:"phase,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	AttachedClusters   int32              `json:"attachedClusters,omitempty"`
	ActiveTunnels      int32              `json:"activeTunnels,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true

// TransitGatewayList contains a list of TransitGateways.
type TransitGatewayList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitGateway `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TransitAttachment represents a cluster's attachment to a transit gateway.
type TransitAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitAttachmentSpec   `json:"spec,omitempty"`
	Status TransitAttachmentStatus `json:"status,omitempty"`
}

// TransitAttachmentSpec defines the desired attachment configuration.
type TransitAttachmentSpec struct {
	// TransitGatewayRef references the parent TransitGateway.
	TransitGatewayRef string `json:"transitGatewayRef"`

	// ClusterID is the unique ID of the attaching cluster.
	ClusterID string `json:"clusterId"`

	// SegmentID is the segment this attachment belongs to.
	SegmentID uint32 `json:"segmentId"`

	// PodCIDRs are the pod CIDRs advertised by this cluster.
	PodCIDRs []string `json:"podCidrs,omitempty"`

	// ServiceCIDRs are the service CIDRs advertised by this cluster.
	ServiceCIDRs []string `json:"serviceCidrs,omitempty"`

	// Routes are the static routes advertised through this attachment.
	Routes []TransitRoute `json:"routes,omitempty"`
}

// TransitRoute defines a route advertised through a transit attachment.
type TransitRoute struct {
	// CIDR is the destination CIDR (e.g., "10.0.0.0/8" or "0.0.0.0/0").
	CIDR string `json:"cidr"`

	// NextHop is the next hop for this route (transit attachment name).
	NextHop string `json:"nextHop,omitempty"`
}

// TransitAttachmentStatus defines the observed state.
type TransitAttachmentStatus struct {
	Phase              string             `json:"phase,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	TunnelEndpoint     string             `json:"tunnelEndpoint,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true

// TransitAttachmentList contains a list of TransitAttachments.
type TransitAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitAttachment `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Segment represents a network segment for isolation.
type Segment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SegmentSpec   `json:"spec,omitempty"`
	Status SegmentStatus `json:"status,omitempty"`
}

// SegmentSpec defines the desired state of a Segment.
type SegmentSpec struct {
	// ID is the 32-bit segment identifier (0–4294967295).
	// Segment 0 is the backbone segment.
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=4294967295
	ID uint32 `json:"id"`

	// Isolated indicates whether the segment is isolated by default.
	// +kubebuilder:default=true
	Isolated bool `json:"isolated,omitempty"`

	// BackboneConnected indicates connectivity to segment 0.
	// +kubebuilder:default=true
	BackboneConnected bool `json:"backboneConnected,omitempty"`
}

// SegmentStatus defines the observed state.
type SegmentStatus struct {
	Phase              string             `json:"phase,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	EndpointCount      int32              `json:"endpointCount,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
}

// +kubebuilder:object:root=true

// SegmentList contains a list of Segments.
type SegmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Segment `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TransitSegmentAttachment connects multiple segments together for inter-segment communication.
type TransitSegmentAttachment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitSegmentAttachmentSpec   `json:"spec,omitempty"`
	Status TransitSegmentAttachmentStatus `json:"status,omitempty"`
}

// TransitSegmentAttachmentSpec defines the segments participating in this attachment.
type TransitSegmentAttachmentSpec struct {
	// SourceSegmentID is the source segment ID.
	SourceSegmentID uint32 `json:"sourceSegmentId"`

	// TargetSegmentID is the target segment ID.
	TargetSegmentID uint32 `json:"targetSegmentId"`

	// TransitGatewayRef references the parent TransitGateway.
	TransitGatewayRef string `json:"transitGatewayRef"`
}

// TransitSegmentAttachmentStatus defines observed status.
type TransitSegmentAttachmentStatus struct {
	Phase      string             `json:"phase,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// TransitSegmentAttachmentList contains a list of TransitSegmentAttachments.
type TransitSegmentAttachmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitSegmentAttachment `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TransitSegmentRoute defines an inter-segment routing rule.
type TransitSegmentRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TransitSegmentRouteSpec   `json:"spec,omitempty"`
	Status TransitSegmentRouteStatus `json:"status,omitempty"`
}

// TransitSegmentRouteSpec defines destination CIDR and next-hop attachment.
type TransitSegmentRouteSpec struct {
	// CIDR is the destination CIDR prefix (e.g. 0.0.0.0/0).
	CIDR string `json:"cidr"`

	// NextHopAttachment references the target TransitAttachment or TransitSegmentAttachment.
	NextHopAttachment string `json:"nextHopAttachment"`
}

// TransitSegmentRouteStatus defines observed status.
type TransitSegmentRouteStatus struct {
	Phase      string             `json:"phase,omitempty"`
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// TransitSegmentRouteList contains a list of TransitSegmentRoutes.
type TransitSegmentRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TransitSegmentRoute `json:"items"`
}

// DeepCopyInto is an autogenerated deepcopy function.
func (in *TransitSegmentAttachment) DeepCopyInto(out *TransitSegmentAttachment) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy creates a new TransitSegmentAttachment.
func (in *TransitSegmentAttachment) DeepCopy() *TransitSegmentAttachment {
	if in == nil {
		return nil
	}
	out := new(TransitSegmentAttachment)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject creates a new runtime.Object.
func (in *TransitSegmentAttachment) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function.
func (in *TransitSegmentAttachmentStatus) DeepCopyInto(out *TransitSegmentAttachmentStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopyInto is an autogenerated deepcopy function.
func (in *TransitSegmentAttachmentList) DeepCopyInto(out *TransitSegmentAttachmentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TransitSegmentAttachment, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy creates a new TransitSegmentAttachmentList.
func (in *TransitSegmentAttachmentList) DeepCopy() *TransitSegmentAttachmentList {
	if in == nil {
		return nil
	}
	out := new(TransitSegmentAttachmentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject creates a new runtime.Object.
func (in *TransitSegmentAttachmentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function.
func (in *TransitSegmentRoute) DeepCopyInto(out *TransitSegmentRoute) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy creates a new TransitSegmentRoute.
func (in *TransitSegmentRoute) DeepCopy() *TransitSegmentRoute {
	if in == nil {
		return nil
	}
	out := new(TransitSegmentRoute)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject creates a new runtime.Object.
func (in *TransitSegmentRoute) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function.
func (in *TransitSegmentRouteStatus) DeepCopyInto(out *TransitSegmentRouteStatus) {
	*out = *in
	if in.Conditions != nil {
		in, out := &in.Conditions, &out.Conditions
		*out = make([]metav1.Condition, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopyInto is an autogenerated deepcopy function.
func (in *TransitSegmentRouteList) DeepCopyInto(out *TransitSegmentRouteList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TransitSegmentRoute, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy creates a new TransitSegmentRouteList.
func (in *TransitSegmentRouteList) DeepCopy() *TransitSegmentRouteList {
	if in == nil {
		return nil
	}
	out := new(TransitSegmentRouteList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject creates a new runtime.Object.
func (in *TransitSegmentRouteList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
