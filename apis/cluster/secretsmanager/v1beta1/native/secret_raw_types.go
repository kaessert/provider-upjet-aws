// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// SecretRAWReplicaInitParameters defines the init parameters for a secret replica.
type SecretRAWReplicaInitParameters struct {
	// ARN, Key ID, or Alias of the AWS KMS key within the region secret is replicated to.
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`
}

// SecretRAWReplica defines the writable configuration for a secret replica.
type SecretRAWReplica struct {
	// ARN, Key ID, or Alias of the AWS KMS key within the region secret is replicated to.
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Region for replicating the secret.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`
}

// SecretRAWReplicaObservation holds the observed state of a secret replica.
type SecretRAWReplicaObservation struct {
	// ARN, Key ID, or Alias of the AWS KMS key within the region secret is replicated to.
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Date that you last accessed the secret in the Region.
	LastAccessedDate *string `json:"lastAccessedDate,omitempty"`

	// Region for replicating the secret.
	Region *string `json:"region,omitempty"`

	// Status can be InProgress, Failed, or InSync.
	Status *string `json:"status,omitempty"`

	// Message such as Replication succeeded or Secret with this name already exists in this region.
	StatusMessage *string `json:"statusMessage,omitempty"`
}

// SecretRAWRotationRules holds the observed rotation rules for a secret.
type SecretRAWRotationRules struct {
	// Number of days between automatic scheduled rotations of the secret.
	AutomaticallyAfterDays *float64 `json:"automaticallyAfterDays,omitempty"`

	// The length of the rotation window in hours, e.g. 3h.
	Duration *string `json:"duration,omitempty"`

	// A cron() or rate() expression that defines the schedule for rotating your secret.
	ScheduleExpression *string `json:"scheduleExpression,omitempty"`
}

// SecretRAWInitParameters defines the init parameters for a SecretRAW.
type SecretRAWInitParameters struct {
	// Description of the secret.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// Accepts boolean value to specify whether to overwrite a secret with the same name in the destination Region.
	// +kubebuilder:validation:Optional
	ForceOverwriteReplicaSecret *bool `json:"forceOverwriteReplicaSecret,omitempty"`

	// ARN or Id of the AWS KMS key to be used to encrypt the secret values in the versions stored in this secret.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kms/v1beta1.Key
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Reference to a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDRef *xpv1.Reference `json:"kmsKeyIdRef,omitempty"`

	// Selector for a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDSelector *xpv1.Selector `json:"kmsKeyIdSelector,omitempty"`

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

// SecretRAWParameters defines the configuration parameters for a SecretRAW.
type SecretRAWParameters struct {
	// Description of the secret.
	// +kubebuilder:validation:Optional
	Description *string `json:"description,omitempty"`

	// Accepts boolean value to specify whether to overwrite a secret with the same name in the destination Region.
	// +kubebuilder:validation:Optional
	ForceOverwriteReplicaSecret *bool `json:"forceOverwriteReplicaSecret,omitempty"`

	// ARN or Id of the AWS KMS key to be used to encrypt the secret values in the versions stored in this secret.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kms/v1beta1.Key
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Reference to a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDRef *xpv1.Reference `json:"kmsKeyIdRef,omitempty"`

	// Selector for a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDSelector *xpv1.Selector `json:"kmsKeyIdSelector,omitempty"`

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

// SecretRAWObservation holds the observed state of a SecretRAW.
type SecretRAWObservation struct {
	// ARN of the secret.
	ARN *string `json:"arn,omitempty"`

	// ARN of the secret (same as ARN; set as the external-name).
	ID *string `json:"id,omitempty"`

	// Valid JSON document representing a resource policy.
	Policy *string `json:"policy,omitempty"`

	// Configuration block showing replica status.
	Replica []SecretRAWReplicaObservation `json:"replica,omitempty"`

	// ARN of the Lambda function that can rotate the secret.
	RotationLambdaARN *string `json:"rotationLambdaArn,omitempty"`

	// A structure that defines the rotation configuration for this secret.
	RotationRules []SecretRAWRotationRules `json:"rotationRules,omitempty"`

	// Map of tags assigned to the resource, including those inherited from the provider default_tags configuration block.
	// +mapType=granular
	TagsAll map[string]*string `json:"tagsAll,omitempty"`
}

// SecretRAWSpec defines the desired state of SecretRAW.
type SecretRAWSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SecretRAWParameters `json:"forProvider"`
	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider SecretRAWInitParameters `json:"initProvider,omitempty"`
}

// SecretRAWStatus defines the observed state of SecretRAW.
type SecretRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SecretRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}

// SecretRAW is the Schema for the native Secrets Manager Secret API.
type SecretRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SecretRAWSpec   `json:"spec"`
	Status            SecretRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretRAWList contains a list of SecretRAW resources.
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

// GetForProvider returns a pointer to the ForProvider parameters.
func (s *SecretRAW) GetForProvider() *SecretRAWParameters { return &s.Spec.ForProvider }

// SetForProvider sets spec.forProvider to the given parameters.
// Required by the SecretCR interface for late-initialization write-back.
func (s *SecretRAW) SetForProvider(p SecretRAWParameters) { s.Spec.ForProvider = p }

// GetInitProvider returns the InitProvider parameters.
func (s *SecretRAW) GetInitProvider() *SecretRAWInitParameters { return &s.Spec.InitProvider }

// GetAtProvider returns the current observed state.
func (s *SecretRAW) GetAtProvider() SecretRAWObservation { return s.Status.AtProvider }

// SetAtProvider sets the observed state.
func (s *SecretRAW) SetAtProvider(o SecretRAWObservation) { s.Status.AtProvider = o }
