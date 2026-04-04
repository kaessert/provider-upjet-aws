// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// SecretVersionRAWInitParameters defines the init parameters for a SecretVersionRAW.
type SecretVersionRAWInitParameters struct {

	// Specifies binary data that you want to encrypt and store in this version of the secret.
	// This is required if secret_string or secret_string_wo is not set.
	// Needs to be encoded to base64.
	// +kubebuilder:validation:Optional
	SecretBinarySecretRef *xpv1.SecretKeySelector `json:"secretBinarySecretRef,omitempty"`

	// Specifies the secret to which you want to add a new version. You can specify either
	// the Amazon Resource Name (ARN) or the friendly name of the secret.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native.SecretRAW
	// +kubebuilder:validation:Optional
	SecretID *string `json:"secretId,omitempty"`

	// Reference to a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDRef *xpv1.Reference `json:"secretIdRef,omitempty"`

	// Selector for a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDSelector *xpv1.Selector `json:"secretIdSelector,omitempty"`

	// Specifies text data that you want to encrypt and store in this version of the secret.
	// This is required if secret_binary or secret_string_wo is not set.
	// +kubebuilder:validation:Optional
	SecretStringSecretRef *xpv1.SecretKeySelector `json:"secretStringSecretRef,omitempty"`

	// Specifies text data that you want to encrypt and store in this version of the secret.
	// This is required if secret_binary or secret_string is not set.
	// +kubebuilder:validation:Optional
	SecretStringWoSecretRef *xpv1.SecretKeySelector `json:"secretStringWoSecretRef,omitempty"`

	// Used together with secret_string_wo to trigger an update. Increment this value when
	// an update to secret_string_wo is required.
	// +kubebuilder:validation:Optional
	SecretStringWoVersion *float64 `json:"secretStringWoVersion,omitempty"`

	// Specifies a list of staging labels that are attached to this version of the secret.
	// +kubebuilder:validation:Optional
	// +listType=set
	VersionStages []*string `json:"versionStages,omitempty"`
}

// SecretVersionRAWParameters defines the configuration parameters for a SecretVersionRAW.
type SecretVersionRAWParameters struct {

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Specifies binary data that you want to encrypt and store in this version of the secret.
	// This is required if secret_string or secret_string_wo is not set.
	// Needs to be encoded to base64.
	// +kubebuilder:validation:Optional
	SecretBinarySecretRef *xpv1.SecretKeySelector `json:"secretBinarySecretRef,omitempty"`

	// Specifies the secret to which you want to add a new version. You can specify either
	// the Amazon Resource Name (ARN) or the friendly name of the secret.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native.SecretRAW
	// +kubebuilder:validation:Optional
	SecretID *string `json:"secretId,omitempty"`

	// Reference to a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDRef *xpv1.Reference `json:"secretIdRef,omitempty"`

	// Selector for a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDSelector *xpv1.Selector `json:"secretIdSelector,omitempty"`

	// Specifies text data that you want to encrypt and store in this version of the secret.
	// This is required if secret_binary or secret_string_wo is not set.
	// +kubebuilder:validation:Optional
	SecretStringSecretRef *xpv1.SecretKeySelector `json:"secretStringSecretRef,omitempty"`

	// Specifies text data that you want to encrypt and store in this version of the secret.
	// This is required if secret_binary or secret_string is not set.
	// +kubebuilder:validation:Optional
	SecretStringWoSecretRef *xpv1.SecretKeySelector `json:"secretStringWoSecretRef,omitempty"`

	// Used together with secret_string_wo to trigger an update. Increment this value when
	// an update to secret_string_wo is required.
	// +kubebuilder:validation:Optional
	SecretStringWoVersion *float64 `json:"secretStringWoVersion,omitempty"`

	// Specifies a list of staging labels that are attached to this version of the secret.
	// +kubebuilder:validation:Optional
	// +listType=set
	VersionStages []*string `json:"versionStages,omitempty"`
}

// SecretVersionRAWObservation holds the observed state of a SecretVersionRAW.
type SecretVersionRAWObservation struct {

	// The ARN of the secret.
	Arn *string `json:"arn,omitempty"`

	// Specifies text data that you want to encrypt and store in this version of the secret.
	HasSecretStringWo *bool `json:"hasSecretStringWo,omitempty"`

	// A pipe delimited combination of secret ID and version ID.
	ID *string `json:"id,omitempty"`

	// Region where this resource will be managed.
	Region *string `json:"region,omitempty"`

	// Specifies the secret to which you want to add a new version.
	SecretID *string `json:"secretId,omitempty"`

	// Used together with secret_string_wo to trigger an update.
	SecretStringWoVersion *float64 `json:"secretStringWoVersion,omitempty"`

	// The unique identifier of the version of the secret.
	VersionID *string `json:"versionId,omitempty"`

	// Specifies a list of staging labels that are attached to this version of the secret.
	// +listType=set
	VersionStages []*string `json:"versionStages,omitempty"`
}

// SecretVersionRAWSpec defines the desired state of SecretVersionRAW.
type SecretVersionRAWSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SecretVersionRAWParameters `json:"forProvider"`
	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider SecretVersionRAWInitParameters `json:"initProvider,omitempty"`
}

// SecretVersionRAWStatus defines the observed state of SecretVersionRAW.
type SecretVersionRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SecretVersionRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}

// SecretVersionRAW is the Schema for the native Secrets Manager Secret Version API.
type SecretVersionRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecretVersionRAWSpec   `json:"spec"`
	Status            SecretVersionRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretVersionRAWList contains a list of SecretVersionRAW resources.
type SecretVersionRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretVersionRAW `json:"items"`
}

// Repository type metadata.
var (
	SecretVersionRAW_Kind             = "SecretVersionRAW"
	SecretVersionRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: SecretVersionRAW_Kind}.String()
	SecretVersionRAW_KindAPIVersion   = SecretVersionRAW_Kind + "." + CRDGroupVersion.String()
	SecretVersionRAW_GroupVersionKind = CRDGroupVersion.WithKind(SecretVersionRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&SecretVersionRAW{}, &SecretVersionRAWList{})
}

// GetForProvider returns a pointer to the ForProvider parameters.
func (s *SecretVersionRAW) GetForProvider() *SecretVersionRAWParameters { return &s.Spec.ForProvider }

// SetForProvider sets spec.forProvider to the given parameters.
// Required by the SecretVersionCR interface for late-initialization write-back.
func (s *SecretVersionRAW) SetForProvider(p SecretVersionRAWParameters) { s.Spec.ForProvider = p }

// GetInitProvider returns the InitProvider parameters.
func (s *SecretVersionRAW) GetInitProvider() *SecretVersionRAWInitParameters {
	return &s.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (s *SecretVersionRAW) GetAtProvider() SecretVersionRAWObservation { return s.Status.AtProvider }

// SetAtProvider sets the observed state.
func (s *SecretVersionRAW) SetAtProvider(o SecretVersionRAWObservation) { s.Status.AtProvider = o }
