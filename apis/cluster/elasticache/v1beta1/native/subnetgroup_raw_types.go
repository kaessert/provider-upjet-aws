// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"
)

// SubnetGroupRAWParameters defines the configuration parameters for a native ElastiCache Subnet Group.
type SubnetGroupRAWParameters struct {
	// Description for the cache subnet group.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// References to Subnet in ec2 to populate subnetIds.
	// +kubebuilder:validation:Optional
	SubnetIDRefs []xpv1.NamespacedReference `json:"subnetIdRefs,omitempty"`

	// Selector for a list of Subnet in ec2 to populate subnetIds.
	// +kubebuilder:validation:Optional
	SubnetIDSelector *xpv1.NamespacedSelector `json:"subnetIdSelector,omitempty"`

	// List of VPC Subnet IDs for the cache subnet group.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetIDRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	SubnetIds []*string `json:"subnetIds,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// SubnetGroupRAWInitParameters defines the init parameters for SubnetGroupRAW.
type SubnetGroupRAWInitParameters struct {
	// Description for the cache subnet group.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// References to Subnet in ec2 to populate subnetIds.
	// +kubebuilder:validation:Optional
	SubnetIDRefs []xpv1.NamespacedReference `json:"subnetIdRefs,omitempty"`

	// Selector for a list of Subnet in ec2 to populate subnetIds.
	// +kubebuilder:validation:Optional
	SubnetIDSelector *xpv1.NamespacedSelector `json:"subnetIdSelector,omitempty"`

	// List of VPC Subnet IDs for the cache subnet group.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/ec2/v1beta1.Subnet
	// +crossplane:generate:reference:refFieldName=SubnetIDRefs
	// +crossplane:generate:reference:selectorFieldName=SubnetIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	SubnetIds []*string `json:"subnetIds,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// SubnetGroupRAWObservation defines the observed state of SubnetGroupRAW.
type SubnetGroupRAWObservation struct {
	// The ARN of the cache subnet group.
	Arn *string `json:"arn,omitempty"`

	// Description for the cache subnet group.
	Description *string `json:"description,omitempty"`

	// The ID (name) of the subnet group.
	ID *string `json:"id,omitempty"`

	// List of VPC Subnet IDs for the cache subnet group.
	// +listType=set
	SubnetIds []*string `json:"subnetIds,omitempty"`

	// Tags assigned to the resource.
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// The Amazon Virtual Private Cloud identifier (VPC ID) of the cache subnet group.
	VPCID *string `json:"vpcId,omitempty"`
}

// SubnetGroupRAWSpec defines the desired state of SubnetGroupRAW.
type SubnetGroupRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider SubnetGroupRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider SubnetGroupRAWInitParameters `json:"initProvider,omitempty"`
}

// SubnetGroupRAWStatus defines the observed state of SubnetGroupRAW.
type SubnetGroupRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider SubnetGroupRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// SubnetGroupRAW is the native (non-Terraform) Schema for AWS ElastiCache Subnet Group.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type SubnetGroupRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubnetGroupRAWSpec   `json:"spec"`
	Status SubnetGroupRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SubnetGroupRAWList contains a list of SubnetGroupRAW resources.
type SubnetGroupRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubnetGroupRAW `json:"items"`
}

// Repository type metadata for SubnetGroupRAW.
var (
	SubnetGroupRAW_Kind             = "SubnetGroupRAW"
	SubnetGroupRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: SubnetGroupRAW_Kind}.String()
	SubnetGroupRAW_KindAPIVersion   = SubnetGroupRAW_Kind + "." + CRDGroupVersion.String()
	SubnetGroupRAW_GroupVersionKind = CRDGroupVersion.WithKind(SubnetGroupRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&SubnetGroupRAW{}, &SubnetGroupRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (s *SubnetGroupRAW) GetForProvider() *SubnetGroupRAWParameters { return &s.Spec.ForProvider }

// SetForProvider writes back the ForProvider parameters.
func (s *SubnetGroupRAW) SetForProvider(p SubnetGroupRAWParameters) { s.Spec.ForProvider = p }

// GetInitProvider returns the InitProvider parameters.
func (s *SubnetGroupRAW) GetInitProvider() *SubnetGroupRAWInitParameters {
	return &s.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (s *SubnetGroupRAW) GetAtProvider() SubnetGroupRAWObservation { return s.Status.AtProvider }

// SetAtProvider sets the observed state.
func (s *SubnetGroupRAW) SetAtProvider(o SubnetGroupRAWObservation) { s.Status.AtProvider = o }
