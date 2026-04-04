// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native"
)

// SecretPolicyRAWParameters defines the namespaced configuration parameters for a SecretPolicyRAW.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type SecretPolicyRAWParameters struct {

	// Makes an optional API call to Zelkova to validate the Resource Policy to prevent broad access to your secret.
	// +kubebuilder:validation:Optional
	BlockPublicPolicy *bool `json:"blockPublicPolicy,omitempty"`

	// Valid JSON document representing a resource policy.
	// +kubebuilder:validation:Optional
	Policy *string `json:"policy,omitempty"`

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Secret ARN.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/secretsmanager/v1beta1/native.SecretRAW
	// +kubebuilder:validation:Optional
	SecretArn *string `json:"secretArn,omitempty"`

	// Reference to a SecretRAW in secretsmanager to populate secretArn.
	// +kubebuilder:validation:Optional
	SecretArnRef *xpv1.NamespacedReference `json:"secretArnRef,omitempty"`

	// Selector for a SecretRAW in secretsmanager to populate secretArn.
	// +kubebuilder:validation:Optional
	SecretArnSelector *xpv1.NamespacedSelector `json:"secretArnSelector,omitempty"`
}

// SecretPolicyRAWInitParameters defines the init parameters for a namespaced SecretPolicyRAW.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type SecretPolicyRAWInitParameters struct {

	// Makes an optional API call to Zelkova to validate the Resource Policy to prevent broad access to your secret.
	// +kubebuilder:validation:Optional
	BlockPublicPolicy *bool `json:"blockPublicPolicy,omitempty"`

	// Valid JSON document representing a resource policy.
	// +kubebuilder:validation:Optional
	Policy *string `json:"policy,omitempty"`

	// Secret ARN.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/secretsmanager/v1beta1/native.SecretRAW
	// +kubebuilder:validation:Optional
	SecretArn *string `json:"secretArn,omitempty"`

	// Reference to a SecretRAW in secretsmanager to populate secretArn.
	// +kubebuilder:validation:Optional
	SecretArnRef *xpv1.NamespacedReference `json:"secretArnRef,omitempty"`

	// Selector for a SecretRAW in secretsmanager to populate secretArn.
	// +kubebuilder:validation:Optional
	SecretArnSelector *xpv1.NamespacedSelector `json:"secretArnSelector,omitempty"`
}

// SecretPolicyRAWSpec defines the desired state of the namespaced SecretPolicyRAW.
type SecretPolicyRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider SecretPolicyRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider SecretPolicyRAWInitParameters `json:"initProvider,omitempty"`
}

// SecretPolicyRAWStatus defines the observed state of the namespaced SecretPolicyRAW.
type SecretPolicyRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          clusternative.SecretPolicyRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}

// SecretPolicyRAW is the Schema for the native Secrets Manager Secret Policy API (namespaced scope).
type SecretPolicyRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecretPolicyRAWSpec   `json:"spec"`
	Status            SecretPolicyRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretPolicyRAWList contains a list of namespaced SecretPolicyRAW resources.
type SecretPolicyRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretPolicyRAW `json:"items"`
}

// Repository type metadata.
var (
	SecretPolicyRAW_Kind             = "SecretPolicyRAW"
	SecretPolicyRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: SecretPolicyRAW_Kind}.String()
	SecretPolicyRAW_KindAPIVersion   = SecretPolicyRAW_Kind + "." + CRDGroupVersion.String()
	SecretPolicyRAW_GroupVersionKind = CRDGroupVersion.WithKind(SecretPolicyRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&SecretPolicyRAW{}, &SecretPolicyRAWList{})
}

// GetForProvider returns a cluster-scoped SecretPolicyRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (s *SecretPolicyRAW) GetForProvider() *clusternative.SecretPolicyRAWParameters {
	return &clusternative.SecretPolicyRAWParameters{
		BlockPublicPolicy: s.Spec.ForProvider.BlockPublicPolicy,
		Policy:            s.Spec.ForProvider.Policy,
		Region:            s.Spec.ForProvider.Region,
		SecretArn:         s.Spec.ForProvider.SecretArn,
	}
}

// SetForProvider copies cluster-scoped SecretPolicyRAWParameters back to the
// namespaced spec. Required by the SecretPolicyCR interface for late-init write-back.
// Namespaced GetForProvider() returns a freshly-allocated copy, so any
// mutations made by the shared CRUD late-init logic must be written back via
// this method to actually persist in the spec.
func (s *SecretPolicyRAW) SetForProvider(p clusternative.SecretPolicyRAWParameters) {
	s.Spec.ForProvider.BlockPublicPolicy = p.BlockPublicPolicy
	s.Spec.ForProvider.Policy = p.Policy
	s.Spec.ForProvider.Region = p.Region
	s.Spec.ForProvider.SecretArn = p.SecretArn
}

// GetInitProvider returns a cluster-scoped SecretPolicyRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (s *SecretPolicyRAW) GetInitProvider() *clusternative.SecretPolicyRAWInitParameters {
	return &clusternative.SecretPolicyRAWInitParameters{
		BlockPublicPolicy: s.Spec.InitProvider.BlockPublicPolicy,
		Policy:            s.Spec.InitProvider.Policy,
		SecretArn:         s.Spec.InitProvider.SecretArn,
	}
}

// GetAtProvider returns the current observed state.
func (s *SecretPolicyRAW) GetAtProvider() clusternative.SecretPolicyRAWObservation {
	return s.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (s *SecretPolicyRAW) SetAtProvider(o clusternative.SecretPolicyRAWObservation) {
	s.Status.AtProvider = o
}
