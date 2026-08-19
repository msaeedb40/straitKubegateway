// Package v1alpha1 contains the API types for straitKubegateway CRDs.
// +kubebuilder:object:generate=true
// +groupName=straitkubegateway.io
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	// GroupName is the API group name.
	GroupName = "straitkubegateway.io"

	// Version is the API version.
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is the group version used to register these objects.
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionResource scheme.
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

// Resource takes an unqualified resource and returns a Group qualified GroupResource.
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Gateway{},
		&GatewayList{},
		&StraitNetworkPolicy{},
		&StraitNetworkPolicyList{},
		&TransitGateway{},
		&TransitGatewayList{},
		&TransitAttachment{},
		&TransitAttachmentList{},
		&Segment{},
		&SegmentList{},
		&BGPPeer{},
		&BGPPeerList{},
		&IPAMPool{},
		&IPAMPoolList{},
		&ClusterLink{},
		&ClusterLinkList{},
		&TransitSegmentAttachment{},
		&TransitSegmentAttachmentList{},
		&TransitSegmentRoute{},
		&TransitSegmentRouteList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}
