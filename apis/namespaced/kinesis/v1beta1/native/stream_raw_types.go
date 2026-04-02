// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/kinesis/v1beta2/native"
)

// StreamModeDetailsRAWInitParameters defines the init parameters for stream mode details (namespaced).
// Field layout matches cluster type exactly for CRUD interface compatibility.
type StreamModeDetailsRAWInitParameters struct {
	// Specifies the capacity mode of the stream. Must be either PROVISIONED or ON_DEMAND.
	// +kubebuilder:validation:Optional
	StreamMode *string `json:"streamMode,omitempty"`
}

// StreamModeDetailsRAWParameters defines the parameters for stream mode details (namespaced).
// Field layout matches cluster type exactly for CRUD interface compatibility.
type StreamModeDetailsRAWParameters struct {
	// Specifies the capacity mode of the stream. Must be either PROVISIONED or ON_DEMAND.
	// +kubebuilder:validation:Required
	StreamMode *string `json:"streamMode"`
}

// StreamRAWParameters defines the namespaced configuration parameters for a native Kinesis Stream.
// Fields are identical to the cluster-scoped type; defined locally for angryjet compatibility.
type StreamRAWParameters struct {
	// The encryption type to use. The only acceptable values are NONE or KMS.
	// +kubebuilder:validation:Optional
	EncryptionType *string `json:"encryptionType,omitempty"`

	// A boolean that indicates all registered consumers should be deregistered.
	// Note: this field is ignored for late-initialization purposes.
	// +kubebuilder:validation:Optional
	EnforceConsumerDeletion *bool `json:"enforceConsumerDeletion,omitempty"`

	// The GUID for the customer-managed KMS key to use for encryption.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/kms/v1beta1.Key
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Reference to a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDRef *xpv1.NamespacedReference `json:"kmsKeyIdRef,omitempty"`

	// Selector for a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDSelector *xpv1.NamespacedSelector `json:"kmsKeyIdSelector,omitempty"`

	// The maximum size for a single data record in KiB.
	// +kubebuilder:validation:Optional
	MaxRecordSizeInKib *float64 `json:"maxRecordSizeInKib,omitempty"`

	// Region where this resource will be managed. Required for credential resolution.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Length of time data records are accessible after they are added to the stream.
	// +kubebuilder:validation:Optional
	RetentionPeriod *float64 `json:"retentionPeriod,omitempty"`

	// The number of shards that the stream will use.
	// +kubebuilder:validation:Optional
	ShardCount *float64 `json:"shardCount,omitempty"`

	// A list of shard-level CloudWatch metrics which can be enabled for the stream.
	// +kubebuilder:validation:Optional
	// +listType=set
	ShardLevelMetrics []*string `json:"shardLevelMetrics,omitempty"`

	// Indicates the capacity mode of the data stream.
	// +kubebuilder:validation:Optional
	StreamModeDetails *StreamModeDetailsRAWParameters `json:"streamModeDetails,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// StreamRAWInitParameters defines the namespaced init parameters for a native Kinesis Stream.
type StreamRAWInitParameters struct {
	// The encryption type to use.
	// +kubebuilder:validation:Optional
	EncryptionType *string `json:"encryptionType,omitempty"`

	// A boolean that indicates all registered consumers should be deregistered.
	// +kubebuilder:validation:Optional
	EnforceConsumerDeletion *bool `json:"enforceConsumerDeletion,omitempty"`

	// The GUID for the customer-managed KMS key to use for encryption.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/kms/v1beta1.Key
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Reference to a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDRef *xpv1.NamespacedReference `json:"kmsKeyIdRef,omitempty"`

	// Selector for a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDSelector *xpv1.NamespacedSelector `json:"kmsKeyIdSelector,omitempty"`

	// The maximum size for a single data record in KiB.
	// +kubebuilder:validation:Optional
	MaxRecordSizeInKib *float64 `json:"maxRecordSizeInKib,omitempty"`

	// Length of time data records are accessible after they are added to the stream.
	// +kubebuilder:validation:Optional
	RetentionPeriod *float64 `json:"retentionPeriod,omitempty"`

	// The number of shards that the stream will use.
	// +kubebuilder:validation:Optional
	ShardCount *float64 `json:"shardCount,omitempty"`

	// A list of shard-level CloudWatch metrics which can be enabled for the stream.
	// +kubebuilder:validation:Optional
	// +listType=set
	ShardLevelMetrics []*string `json:"shardLevelMetrics,omitempty"`

	// Indicates the capacity mode of the data stream.
	// +kubebuilder:validation:Optional
	StreamModeDetails *StreamModeDetailsRAWInitParameters `json:"streamModeDetails,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// StreamRAWSpec defines the desired state of StreamRAW (namespaced scope).
type StreamRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider StreamRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider StreamRAWInitParameters `json:"initProvider,omitempty"`
}

// StreamRAWStatus defines the observed state of StreamRAW (namespaced scope).
// Note: using xpv1.ResourceStatus (not ConditionedStatus) so that angryjet's
// ManagedV2() matcher recognises this as a v2-style managed resource.
type StreamRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.StreamRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// StreamRAW is the native (non-Terraform) Schema for AWS Kinesis Streams API (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type StreamRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StreamRAWSpec   `json:"spec"`
	Status StreamRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StreamRAWList contains a list of StreamRAW resources.
type StreamRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StreamRAW `json:"items"`
}

// Repository type metadata for StreamRAW.
var (
	StreamRAW_Kind             = "StreamRAW"
	StreamRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: StreamRAW_Kind}.String()
	StreamRAW_KindAPIVersion   = StreamRAW_Kind + "." + CRDGroupVersion.String()
	StreamRAW_GroupVersionKind = CRDGroupVersion.WithKind(StreamRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&StreamRAW{}, &StreamRAWList{})
}

// GetForProvider returns a cluster-scoped StreamRAWParameters populated from
// this namespaced resource's ForProvider fields.
// A field-by-field copy is necessary because the namespaced package defines its
// own StreamRAWParameters struct for angryjet compatibility while the shared
// CRUD interface expects the cluster-scoped parameter type.
func (s *StreamRAW) GetForProvider() *clusternative.StreamRAWParameters {
	return &clusternative.StreamRAWParameters{
		EncryptionType:          s.Spec.ForProvider.EncryptionType,
		EnforceConsumerDeletion: s.Spec.ForProvider.EnforceConsumerDeletion,
		KMSKeyID:                s.Spec.ForProvider.KMSKeyID,
		KMSKeyIDRef:             s.Spec.ForProvider.KMSKeyIDRef,
		KMSKeyIDSelector:        s.Spec.ForProvider.KMSKeyIDSelector,
		MaxRecordSizeInKib:      s.Spec.ForProvider.MaxRecordSizeInKib,
		Region:                  s.Spec.ForProvider.Region,
		RetentionPeriod:         s.Spec.ForProvider.RetentionPeriod,
		ShardCount:              s.Spec.ForProvider.ShardCount,
		ShardLevelMetrics:       s.Spec.ForProvider.ShardLevelMetrics,
		StreamModeDetails:       streamModeDetailsToCluster(s.Spec.ForProvider.StreamModeDetails),
		Tags:                    s.Spec.ForProvider.Tags,
	}
}

// GetInitProvider returns a cluster-scoped StreamRAWInitParameters populated
// from this namespaced resource's InitProvider fields.
func (s *StreamRAW) GetInitProvider() *clusternative.StreamRAWInitParameters {
	return &clusternative.StreamRAWInitParameters{
		EncryptionType:          s.Spec.InitProvider.EncryptionType,
		EnforceConsumerDeletion: s.Spec.InitProvider.EnforceConsumerDeletion,
		KMSKeyID:                s.Spec.InitProvider.KMSKeyID,
		KMSKeyIDRef:             s.Spec.InitProvider.KMSKeyIDRef,
		KMSKeyIDSelector:        s.Spec.InitProvider.KMSKeyIDSelector,
		MaxRecordSizeInKib:      s.Spec.InitProvider.MaxRecordSizeInKib,
		RetentionPeriod:         s.Spec.InitProvider.RetentionPeriod,
		ShardCount:              s.Spec.InitProvider.ShardCount,
		ShardLevelMetrics:       s.Spec.InitProvider.ShardLevelMetrics,
		StreamModeDetails:       streamModeDetailsInitToCluster(s.Spec.InitProvider.StreamModeDetails),
		Tags:                    s.Spec.InitProvider.Tags,
	}
}

// GetAtProvider returns the current observed state.
func (s *StreamRAW) GetAtProvider() clusternative.StreamRAWObservation { return s.Status.AtProvider }

// SetAtProvider sets the observed state.
func (s *StreamRAW) SetAtProvider(o clusternative.StreamRAWObservation) { s.Status.AtProvider = o }

// SetForProviderEncryptionType sets spec.forProvider.encryptionType.
// Used by late-initialization so that mutations propagate back to the
// namespaced spec (GetForProvider() returns a field-copied struct).
func (s *StreamRAW) SetForProviderEncryptionType(v *string) { s.Spec.ForProvider.EncryptionType = v }

// SetForProviderRetentionPeriod sets spec.forProvider.retentionPeriod.
func (s *StreamRAW) SetForProviderRetentionPeriod(v *float64) {
	s.Spec.ForProvider.RetentionPeriod = v
}

// SetForProviderShardCount sets spec.forProvider.shardCount.
func (s *StreamRAW) SetForProviderShardCount(v *float64) { s.Spec.ForProvider.ShardCount = v }

// SetForProviderMaxRecordSizeInKib sets spec.forProvider.maxRecordSizeInKib.
func (s *StreamRAW) SetForProviderMaxRecordSizeInKib(v *float64) {
	s.Spec.ForProvider.MaxRecordSizeInKib = v
}

// SetForProviderStreamModeDetails sets spec.forProvider.streamModeDetails.
// Converts from the cluster-scoped type to the namespaced local type.
func (s *StreamRAW) SetForProviderStreamModeDetails(v *clusternative.StreamModeDetailsRAWParameters) {
	if v == nil {
		s.Spec.ForProvider.StreamModeDetails = nil
		return
	}
	s.Spec.ForProvider.StreamModeDetails = &StreamModeDetailsRAWParameters{StreamMode: v.StreamMode}
}

// streamModeDetailsToCluster converts a namespaced StreamModeDetailsRAWParameters
// to the cluster-scoped type.
func streamModeDetailsToCluster(p *StreamModeDetailsRAWParameters) *clusternative.StreamModeDetailsRAWParameters {
	if p == nil {
		return nil
	}
	return &clusternative.StreamModeDetailsRAWParameters{
		StreamMode: p.StreamMode,
	}
}

// streamModeDetailsInitToCluster converts a namespaced StreamModeDetailsRAWInitParameters
// to the cluster-scoped type.
func streamModeDetailsInitToCluster(p *StreamModeDetailsRAWInitParameters) *clusternative.StreamModeDetailsRAWInitParameters {
	if p == nil {
		return nil
	}
	return &clusternative.StreamModeDetailsRAWInitParameters{
		StreamMode: p.StreamMode,
	}
}
