// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native"
)

// ParameterGroupRAWParameters defines the namespaced configuration parameters for a native ElastiCache Parameter Group.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type ParameterGroupRAWParameters struct {
	// The description of the ElastiCache parameter group.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// The family of the ElastiCache parameter group.
	// +kubebuilder:validation:Optional
	Family *string `json:"family,omitempty"`

	// The name of the ElastiCache parameter group.
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty"`

	// A list of ElastiCache parameters to apply.
	// +kubebuilder:validation:Optional
	Parameter []clusternative.ParameterRAWParameters `json:"parameter,omitempty"`

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// ParameterGroupRAWInitParameters defines the namespaced init parameters for ParameterGroupRAW.
type ParameterGroupRAWInitParameters struct {
	// The description of the ElastiCache parameter group.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// The family of the ElastiCache parameter group.
	// +kubebuilder:validation:Optional
	Family *string `json:"family,omitempty"`

	// The name of the ElastiCache parameter group.
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty"`

	// A list of ElastiCache parameters to apply.
	// +kubebuilder:validation:Optional
	Parameter []clusternative.ParameterRAWInitParameters `json:"parameter,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// ParameterGroupRAWSpec defines the desired state of namespaced ParameterGroupRAW.
type ParameterGroupRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider ParameterGroupRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider ParameterGroupRAWInitParameters `json:"initProvider,omitempty"`
}

// ParameterGroupRAWStatus defines the observed state of namespaced ParameterGroupRAW.
type ParameterGroupRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.ParameterGroupRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// ParameterGroupRAW is the native (non-Terraform) Schema for AWS ElastiCache Parameter Group (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type ParameterGroupRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ParameterGroupRAWSpec   `json:"spec"`
	Status ParameterGroupRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ParameterGroupRAWList contains a list of ParameterGroupRAW resources (namespaced scope).
type ParameterGroupRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ParameterGroupRAW `json:"items"`
}

// Repository type metadata for namespaced ParameterGroupRAW.
var (
	ParameterGroupRAW_Kind             = "ParameterGroupRAW"
	ParameterGroupRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ParameterGroupRAW_Kind}.String()
	ParameterGroupRAW_KindAPIVersion   = ParameterGroupRAW_Kind + "." + CRDGroupVersion.String()
	ParameterGroupRAW_GroupVersionKind = CRDGroupVersion.WithKind(ParameterGroupRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&ParameterGroupRAW{}, &ParameterGroupRAWList{})
}

// GetForProvider returns a cluster-scoped ParameterGroupRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (p *ParameterGroupRAW) GetForProvider() *clusternative.ParameterGroupRAWParameters {
	return &clusternative.ParameterGroupRAWParameters{
		Description: p.Spec.ForProvider.Description,
		Family:      p.Spec.ForProvider.Family,
		Name:        p.Spec.ForProvider.Name,
		Parameter:   p.Spec.ForProvider.Parameter,
		Region:      p.Spec.ForProvider.Region,
		Tags:        p.Spec.ForProvider.Tags,
	}
}

// SetForProvider writes back a cluster-scoped ParameterGroupRAWParameters to this
// namespaced resource's ForProvider fields. Symmetric inverse of GetForProvider.
func (p *ParameterGroupRAW) SetForProvider(pp clusternative.ParameterGroupRAWParameters) {
	p.Spec.ForProvider.Description = pp.Description
	p.Spec.ForProvider.Family = pp.Family
	p.Spec.ForProvider.Name = pp.Name
	p.Spec.ForProvider.Parameter = pp.Parameter
	p.Spec.ForProvider.Region = pp.Region
	p.Spec.ForProvider.Tags = pp.Tags
}

// GetInitProvider returns a cluster-scoped ParameterGroupRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (p *ParameterGroupRAW) GetInitProvider() *clusternative.ParameterGroupRAWInitParameters {
	return &clusternative.ParameterGroupRAWInitParameters{
		Description: p.Spec.InitProvider.Description,
		Family:      p.Spec.InitProvider.Family,
		Name:        p.Spec.InitProvider.Name,
		Parameter:   p.Spec.InitProvider.Parameter,
		Tags:        p.Spec.InitProvider.Tags,
	}
}

// GetAtProvider returns the current observed state.
func (p *ParameterGroupRAW) GetAtProvider() clusternative.ParameterGroupRAWObservation {
	return p.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (p *ParameterGroupRAW) SetAtProvider(o clusternative.ParameterGroupRAWObservation) {
	p.Status.AtProvider = o
}
