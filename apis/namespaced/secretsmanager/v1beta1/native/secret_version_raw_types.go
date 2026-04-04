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

// SecretVersionRAWInitParameters defines the init parameters for a namespaced SecretVersionRAW.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type SecretVersionRAWInitParameters struct {

	// Specifies binary data that you want to encrypt and store in this version of the secret.
	// +kubebuilder:validation:Optional
	SecretBinarySecretRef *xpv1.SecretKeySelector `json:"secretBinarySecretRef,omitempty"`

	// Specifies the secret to which you want to add a new version.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/secretsmanager/v1beta1/native.SecretRAW
	// +kubebuilder:validation:Optional
	SecretID *string `json:"secretId,omitempty"`

	// Reference to a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDRef *xpv1.NamespacedReference `json:"secretIdRef,omitempty"`

	// Selector for a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDSelector *xpv1.NamespacedSelector `json:"secretIdSelector,omitempty"`

	// Specifies text data that you want to encrypt and store in this version of the secret.
	// +kubebuilder:validation:Optional
	SecretStringSecretRef *xpv1.SecretKeySelector `json:"secretStringSecretRef,omitempty"`

	// Specifies text data that you want to encrypt and store in this version of the secret.
	// +kubebuilder:validation:Optional
	SecretStringWoSecretRef *xpv1.SecretKeySelector `json:"secretStringWoSecretRef,omitempty"`

	// Used together with secret_string_wo to trigger an update.
	// +kubebuilder:validation:Optional
	SecretStringWoVersion *float64 `json:"secretStringWoVersion,omitempty"`

	// Specifies a list of staging labels that are attached to this version of the secret.
	// +kubebuilder:validation:Optional
	// +listType=set
	VersionStages []*string `json:"versionStages,omitempty"`
}

// SecretVersionRAWParameters defines the namespaced configuration parameters for a SecretVersionRAW.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type SecretVersionRAWParameters struct {

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Specifies binary data that you want to encrypt and store in this version of the secret.
	// +kubebuilder:validation:Optional
	SecretBinarySecretRef *xpv1.SecretKeySelector `json:"secretBinarySecretRef,omitempty"`

	// Specifies the secret to which you want to add a new version.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/secretsmanager/v1beta1/native.SecretRAW
	// +kubebuilder:validation:Optional
	SecretID *string `json:"secretId,omitempty"`

	// Reference to a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDRef *xpv1.NamespacedReference `json:"secretIdRef,omitempty"`

	// Selector for a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDSelector *xpv1.NamespacedSelector `json:"secretIdSelector,omitempty"`

	// Specifies text data that you want to encrypt and store in this version of the secret.
	// +kubebuilder:validation:Optional
	SecretStringSecretRef *xpv1.SecretKeySelector `json:"secretStringSecretRef,omitempty"`

	// Specifies text data that you want to encrypt and store in this version of the secret.
	// +kubebuilder:validation:Optional
	SecretStringWoSecretRef *xpv1.SecretKeySelector `json:"secretStringWoSecretRef,omitempty"`

	// Used together with secret_string_wo to trigger an update.
	// +kubebuilder:validation:Optional
	SecretStringWoVersion *float64 `json:"secretStringWoVersion,omitempty"`

	// Specifies a list of staging labels that are attached to this version of the secret.
	// +kubebuilder:validation:Optional
	// +listType=set
	VersionStages []*string `json:"versionStages,omitempty"`
}

// SecretVersionRAWSpec defines the desired state of the namespaced SecretVersionRAW.
type SecretVersionRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider SecretVersionRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider SecretVersionRAWInitParameters `json:"initProvider,omitempty"`
}

// SecretVersionRAWStatus defines the observed state of the namespaced SecretVersionRAW.
type SecretVersionRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          clusternative.SecretVersionRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}

// SecretVersionRAW is the Schema for the native Secrets Manager Secret Version API (namespaced scope).
type SecretVersionRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecretVersionRAWSpec   `json:"spec"`
	Status            SecretVersionRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretVersionRAWList contains a list of namespaced SecretVersionRAW resources.
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

// GetForProvider returns a cluster-scoped SecretVersionRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (s *SecretVersionRAW) GetForProvider() *clusternative.SecretVersionRAWParameters {
	return &clusternative.SecretVersionRAWParameters{
		Region:                  s.Spec.ForProvider.Region,
		SecretBinarySecretRef:   s.Spec.ForProvider.SecretBinarySecretRef,
		SecretID:                s.Spec.ForProvider.SecretID,
		SecretStringSecretRef:   s.Spec.ForProvider.SecretStringSecretRef,
		SecretStringWoSecretRef: s.Spec.ForProvider.SecretStringWoSecretRef,
		SecretStringWoVersion:   s.Spec.ForProvider.SecretStringWoVersion,
		VersionStages:           s.Spec.ForProvider.VersionStages,
	}
}

// SetForProvider copies cluster-scoped SecretVersionRAWParameters back to the
// namespaced spec. Required by the SecretVersionCR interface for late-init write-back.
// Namespaced GetForProvider() returns a freshly-allocated copy, so any
// mutations made by the shared CRUD late-init logic must be written back via
// this method to actually persist in the spec.
func (s *SecretVersionRAW) SetForProvider(p clusternative.SecretVersionRAWParameters) {
	s.Spec.ForProvider.Region = p.Region
	s.Spec.ForProvider.SecretBinarySecretRef = p.SecretBinarySecretRef
	s.Spec.ForProvider.SecretID = p.SecretID
	s.Spec.ForProvider.SecretStringSecretRef = p.SecretStringSecretRef
	s.Spec.ForProvider.SecretStringWoSecretRef = p.SecretStringWoSecretRef
	s.Spec.ForProvider.SecretStringWoVersion = p.SecretStringWoVersion
	s.Spec.ForProvider.VersionStages = p.VersionStages
}

// GetInitProvider returns a cluster-scoped SecretVersionRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (s *SecretVersionRAW) GetInitProvider() *clusternative.SecretVersionRAWInitParameters {
	return &clusternative.SecretVersionRAWInitParameters{
		SecretBinarySecretRef:   s.Spec.InitProvider.SecretBinarySecretRef,
		SecretID:                s.Spec.InitProvider.SecretID,
		SecretStringSecretRef:   s.Spec.InitProvider.SecretStringSecretRef,
		SecretStringWoSecretRef: s.Spec.InitProvider.SecretStringWoSecretRef,
		SecretStringWoVersion:   s.Spec.InitProvider.SecretStringWoVersion,
		VersionStages:           s.Spec.InitProvider.VersionStages,
	}
}

// GetAtProvider returns the current observed state.
func (s *SecretVersionRAW) GetAtProvider() clusternative.SecretVersionRAWObservation {
	return s.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (s *SecretVersionRAW) SetAtProvider(o clusternative.SecretVersionRAWObservation) {
	s.Status.AtProvider = o
}
