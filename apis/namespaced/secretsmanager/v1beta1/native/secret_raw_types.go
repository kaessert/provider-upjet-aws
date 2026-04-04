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

// SecretRAWReplicaInitParameters defines the init parameters for a namespaced secret replica.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type SecretRAWReplicaInitParameters struct {
	// ARN, Key ID, or Alias of the AWS KMS key within the region secret is replicated to.
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`
}

// SecretRAWReplica defines the writable configuration for a namespaced secret replica.
type SecretRAWReplica struct {
	// ARN, Key ID, or Alias of the AWS KMS key within the region secret is replicated to.
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Region for replicating the secret.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`
}

// SecretRAWInitParameters defines the init parameters for a namespaced SecretRAW.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type SecretRAWInitParameters struct {
	// Description of the secret.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// Accepts boolean value to specify whether to overwrite a secret with the same name in the destination Region.
	// +kubebuilder:validation:Optional
	ForceOverwriteReplicaSecret *bool `json:"forceOverwriteReplicaSecret,omitempty"`

	// ARN or Id of the AWS KMS key to be used to encrypt the secret values in the versions stored in this secret.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/kms/v1beta1.Key
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Reference to a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDRef *xpv1.NamespacedReference `json:"kmsKeyIdRef,omitempty"`

	// Selector for a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDSelector *xpv1.NamespacedSelector `json:"kmsKeyIdSelector,omitempty"`

	// Friendly name of the new secret.
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty"`

	// Number of days that AWS Secrets Manager waits before it can delete the secret.
	// +kubebuilder:validation:Optional
	RecoveryWindowInDays *float64 `json:"recoveryWindowInDays,omitempty"`

	// Configuration block to support secret replication.
	// +kubebuilder:validation:Optional
	Replica []SecretRAWReplicaInitParameters `json:"replica,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// SecretRAWParameters defines the namespaced configuration parameters for a SecretRAW.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type SecretRAWParameters struct {
	// Description of the secret.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// Accepts boolean value to specify whether to overwrite a secret with the same name in the destination Region.
	// +kubebuilder:validation:Optional
	ForceOverwriteReplicaSecret *bool `json:"forceOverwriteReplicaSecret,omitempty"`

	// ARN or Id of the AWS KMS key to be used to encrypt the secret values in the versions stored in this secret.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/kms/v1beta1.Key
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Reference to a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDRef *xpv1.NamespacedReference `json:"kmsKeyIdRef,omitempty"`

	// Selector for a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDSelector *xpv1.NamespacedSelector `json:"kmsKeyIdSelector,omitempty"`

	// Friendly name of the new secret.
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty"`

	// Number of days that AWS Secrets Manager waits before it can delete the secret.
	// +kubebuilder:validation:Optional
	RecoveryWindowInDays *float64 `json:"recoveryWindowInDays,omitempty"`

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Configuration block to support secret replication.
	// +kubebuilder:validation:Optional
	Replica []SecretRAWReplica `json:"replica,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// SecretRAWSpec defines the desired state of the namespaced SecretRAW.
type SecretRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider SecretRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider SecretRAWInitParameters `json:"initProvider,omitempty"`
}

// SecretRAWStatus defines the observed state of the namespaced SecretRAW.
type SecretRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          clusternative.SecretRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}

// SecretRAW is the Schema for the native Secrets Manager Secret API (namespaced scope).
type SecretRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecretRAWSpec   `json:"spec"`
	Status            SecretRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretRAWList contains a list of namespaced SecretRAW resources.
type SecretRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretRAW `json:"items"`
}

// Repository type metadata.
var (
	SecretRAW_Kind             = "SecretRAW"
	SecretRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: SecretRAW_Kind}.String()
	SecretRAW_KindAPIVersion   = SecretRAW_Kind + "." + CRDGroupVersion.String()
	SecretRAW_GroupVersionKind = CRDGroupVersion.WithKind(SecretRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&SecretRAW{}, &SecretRAWList{})
}

// GetForProvider returns a cluster-scoped SecretRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (s *SecretRAW) GetForProvider() *clusternative.SecretRAWParameters {
	return &clusternative.SecretRAWParameters{
		Description:                 s.Spec.ForProvider.Description,
		ForceOverwriteReplicaSecret: s.Spec.ForProvider.ForceOverwriteReplicaSecret,
		KMSKeyID:                    s.Spec.ForProvider.KMSKeyID,
		Name:                        s.Spec.ForProvider.Name,
		RecoveryWindowInDays:        s.Spec.ForProvider.RecoveryWindowInDays,
		Region:                      s.Spec.ForProvider.Region,
		Replica:                     namespacedReplicaToCluster(s.Spec.ForProvider.Replica),
		Tags:                        s.Spec.ForProvider.Tags,
	}
}

// SetForProvider copies cluster-scoped SecretRAWParameters back to the
// namespaced spec. Required by the SecretCR interface for late-init write-back.
// Namespaced GetForProvider() returns a freshly-allocated copy, so any
// mutations made by the shared CRUD late-init logic must be written back via
// this method to actually persist in the spec.
func (s *SecretRAW) SetForProvider(p clusternative.SecretRAWParameters) {
	s.Spec.ForProvider.Description = p.Description
	s.Spec.ForProvider.ForceOverwriteReplicaSecret = p.ForceOverwriteReplicaSecret
	s.Spec.ForProvider.KMSKeyID = p.KMSKeyID
	s.Spec.ForProvider.Name = p.Name
	s.Spec.ForProvider.RecoveryWindowInDays = p.RecoveryWindowInDays
	s.Spec.ForProvider.Region = p.Region
	s.Spec.ForProvider.Replica = clusterReplicaToNamespaced(p.Replica)
	s.Spec.ForProvider.Tags = p.Tags
}

// GetInitProvider returns a cluster-scoped SecretRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (s *SecretRAW) GetInitProvider() *clusternative.SecretRAWInitParameters {
	return &clusternative.SecretRAWInitParameters{
		Description:                 s.Spec.InitProvider.Description,
		ForceOverwriteReplicaSecret: s.Spec.InitProvider.ForceOverwriteReplicaSecret,
		KMSKeyID:                    s.Spec.InitProvider.KMSKeyID,
		Name:                        s.Spec.InitProvider.Name,
		RecoveryWindowInDays:        s.Spec.InitProvider.RecoveryWindowInDays,
		Replica:                     namespacedReplicaInitToCluster(s.Spec.InitProvider.Replica),
		Tags:                        s.Spec.InitProvider.Tags,
	}
}

// GetAtProvider returns the current observed state.
func (s *SecretRAW) GetAtProvider() clusternative.SecretRAWObservation {
	return s.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (s *SecretRAW) SetAtProvider(o clusternative.SecretRAWObservation) {
	s.Status.AtProvider = o
}

// namespacedReplicaToCluster converts namespaced replica params to cluster-scoped params.
func namespacedReplicaToCluster(replicas []SecretRAWReplica) []clusternative.SecretRAWReplica {
	if replicas == nil {
		return nil
	}
	out := make([]clusternative.SecretRAWReplica, len(replicas))
	for i, r := range replicas {
		out[i] = clusternative.SecretRAWReplica{
			KMSKeyID: r.KMSKeyID,
			Region:   r.Region,
		}
	}
	return out
}

// clusterReplicaToNamespaced converts cluster-scoped replica params to namespaced params.
func clusterReplicaToNamespaced(replicas []clusternative.SecretRAWReplica) []SecretRAWReplica {
	if replicas == nil {
		return nil
	}
	out := make([]SecretRAWReplica, len(replicas))
	for i, r := range replicas {
		out[i] = SecretRAWReplica{
			KMSKeyID: r.KMSKeyID,
			Region:   r.Region,
		}
	}
	return out
}

// namespacedReplicaInitToCluster converts namespaced replica init params to cluster-scoped.
func namespacedReplicaInitToCluster(replicas []SecretRAWReplicaInitParameters) []clusternative.SecretRAWReplicaInitParameters {
	if replicas == nil {
		return nil
	}
	out := make([]clusternative.SecretRAWReplicaInitParameters, len(replicas))
	for i, r := range replicas {
		out[i] = clusternative.SecretRAWReplicaInitParameters{
			KMSKeyID: r.KMSKeyID,
		}
	}
	return out
}
