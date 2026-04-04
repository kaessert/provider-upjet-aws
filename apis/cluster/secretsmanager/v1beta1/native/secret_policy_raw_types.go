// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// SecretPolicyRAWParameters defines the configuration parameters for a SecretPolicyRAW.
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native.SecretRAW
	// +kubebuilder:validation:Optional
	SecretArn *string `json:"secretArn,omitempty"`

	// Reference to a SecretRAW in secretsmanager to populate secretArn.
	// +kubebuilder:validation:Optional
	SecretArnRef *xpv1.Reference `json:"secretArnRef,omitempty"`

	// Selector for a SecretRAW in secretsmanager to populate secretArn.
	// +kubebuilder:validation:Optional
	SecretArnSelector *xpv1.Selector `json:"secretArnSelector,omitempty"`
}

// SecretPolicyRAWObservation holds the observed state of a SecretPolicyRAW.
type SecretPolicyRAWObservation struct {

	// Makes an optional API call to Zelkova to validate the Resource Policy to prevent broad access to your secret.
	BlockPublicPolicy *bool `json:"blockPublicPolicy,omitempty"`

	// Amazon Resource Name (ARN) of the secret.
	ID *string `json:"id,omitempty"`

	// Secret ARN.
	SecretArn *string `json:"secretArn,omitempty"`
}

// SecretPolicyRAWInitParameters defines the init parameters for a SecretPolicyRAW.
type SecretPolicyRAWInitParameters struct {

	// Makes an optional API call to Zelkova to validate the Resource Policy to prevent broad access to your secret.
	// +kubebuilder:validation:Optional
	BlockPublicPolicy *bool `json:"blockPublicPolicy,omitempty"`

	// Valid JSON document representing a resource policy.
	// +kubebuilder:validation:Optional
	Policy *string `json:"policy,omitempty"`

	// Secret ARN.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native.SecretRAW
	// +kubebuilder:validation:Optional
	SecretArn *string `json:"secretArn,omitempty"`

	// Reference to a SecretRAW in secretsmanager to populate secretArn.
	// +kubebuilder:validation:Optional
	SecretArnRef *xpv1.Reference `json:"secretArnRef,omitempty"`

	// Selector for a SecretRAW in secretsmanager to populate secretArn.
	// +kubebuilder:validation:Optional
	SecretArnSelector *xpv1.Selector `json:"secretArnSelector,omitempty"`
}

// SecretPolicyRAWSpec defines the desired state of SecretPolicyRAW.
type SecretPolicyRAWSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SecretPolicyRAWParameters `json:"forProvider"`
	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider SecretPolicyRAWInitParameters `json:"initProvider,omitempty"`
}

// SecretPolicyRAWStatus defines the observed state of SecretPolicyRAW.
type SecretPolicyRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SecretPolicyRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}

// SecretPolicyRAW is the Schema for the native Secrets Manager Secret Policy API.
type SecretPolicyRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecretPolicyRAWSpec   `json:"spec"`
	Status            SecretPolicyRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretPolicyRAWList contains a list of SecretPolicyRAW resources.
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

// GetForProvider returns a pointer to the ForProvider parameters.
func (s *SecretPolicyRAW) GetForProvider() *SecretPolicyRAWParameters { return &s.Spec.ForProvider }

// SetForProvider sets spec.forProvider to the given parameters.
// Required by the SecretPolicyCR interface for late-initialization write-back.
func (s *SecretPolicyRAW) SetForProvider(p SecretPolicyRAWParameters) { s.Spec.ForProvider = p }

// GetInitProvider returns the InitProvider parameters.
func (s *SecretPolicyRAW) GetInitProvider() *SecretPolicyRAWInitParameters {
	return &s.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (s *SecretPolicyRAW) GetAtProvider() SecretPolicyRAWObservation { return s.Status.AtProvider }

// SetAtProvider sets the observed state.
func (s *SecretPolicyRAW) SetAtProvider(o SecretPolicyRAWObservation) { s.Status.AtProvider = o }
