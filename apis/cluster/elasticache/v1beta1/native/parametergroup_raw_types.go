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

// ParameterRAWInitParameters defines the init parameters for a single cache parameter.
type ParameterRAWInitParameters struct {
	// The name of the parameter.
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty"`

	// The value of the parameter.
	// +kubebuilder:validation:Optional
	Value *string `json:"value,omitempty"`
}

// ParameterRAWParameters defines the configuration for a single cache parameter.
type ParameterRAWParameters struct {
	// The name of the parameter.
	// +kubebuilder:validation:Optional
	Name *string `json:"name"`

	// The value of the parameter.
	// +kubebuilder:validation:Optional
	Value *string `json:"value"`
}

// ParameterRAWObservation defines the observed state of a single cache parameter.
type ParameterRAWObservation struct {
	// The name of the parameter.
	Name *string `json:"name,omitempty"`

	// The value of the parameter.
	Value *string `json:"value,omitempty"`
}

// ParameterGroupRAWParameters defines the configuration parameters for a native ElastiCache Parameter Group.
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
	Parameter []ParameterRAWParameters `json:"parameter,omitempty"`

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// ParameterGroupRAWInitParameters defines the init parameters for ParameterGroupRAW.
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
	Parameter []ParameterRAWInitParameters `json:"parameter,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// ParameterGroupRAWObservation defines the observed state of ParameterGroupRAW.
type ParameterGroupRAWObservation struct {
	// The AWS ARN associated with the parameter group.
	Arn *string `json:"arn,omitempty"`

	// The description of the ElastiCache parameter group.
	Description *string `json:"description,omitempty"`

	// The family of the ElastiCache parameter group.
	Family *string `json:"family,omitempty"`

	// The ElastiCache parameter group name.
	ID *string `json:"id,omitempty"`

	// The name of the ElastiCache parameter group.
	Name *string `json:"name,omitempty"`

	// A list of ElastiCache parameters.
	Parameter []ParameterRAWObservation `json:"parameter,omitempty"`

	// Tags assigned to the resource.
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// ParameterGroupRAWSpec defines the desired state of ParameterGroupRAW.
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

// ParameterGroupRAWStatus defines the observed state of ParameterGroupRAW.
type ParameterGroupRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider ParameterGroupRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// ParameterGroupRAW is the native (non-Terraform) Schema for AWS ElastiCache Parameter Group.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type ParameterGroupRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ParameterGroupRAWSpec   `json:"spec"`
	Status ParameterGroupRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ParameterGroupRAWList contains a list of ParameterGroupRAW resources.
type ParameterGroupRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ParameterGroupRAW `json:"items"`
}

// Repository type metadata for ParameterGroupRAW.
var (
	ParameterGroupRAW_Kind             = "ParameterGroupRAW"
	ParameterGroupRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ParameterGroupRAW_Kind}.String()
	ParameterGroupRAW_KindAPIVersion   = ParameterGroupRAW_Kind + "." + CRDGroupVersion.String()
	ParameterGroupRAW_GroupVersionKind = CRDGroupVersion.WithKind(ParameterGroupRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&ParameterGroupRAW{}, &ParameterGroupRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (p *ParameterGroupRAW) GetForProvider() *ParameterGroupRAWParameters { return &p.Spec.ForProvider }

// GetInitProvider returns the InitProvider parameters.
func (p *ParameterGroupRAW) GetInitProvider() *ParameterGroupRAWInitParameters {
	return &p.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (p *ParameterGroupRAW) GetAtProvider() ParameterGroupRAWObservation { return p.Status.AtProvider }

// SetAtProvider sets the observed state.
func (p *ParameterGroupRAW) SetAtProvider(o ParameterGroupRAWObservation) { p.Status.AtProvider = o }
