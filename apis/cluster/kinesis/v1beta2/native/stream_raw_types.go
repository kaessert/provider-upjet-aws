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

// StreamModeDetailsRAWInitParameters defines the init parameters for stream mode details.
type StreamModeDetailsRAWInitParameters struct {
	// Specifies the capacity mode of the stream. Must be either PROVISIONED or ON_DEMAND.
	// +kubebuilder:validation:Optional
	StreamMode *string `json:"streamMode,omitempty"`
}

// StreamModeDetailsRAWParameters defines the parameters for stream mode details.
type StreamModeDetailsRAWParameters struct {
	// Specifies the capacity mode of the stream. Must be either PROVISIONED or ON_DEMAND.
	// +kubebuilder:validation:Required
	StreamMode *string `json:"streamMode"`
}

// StreamModeDetailsRAWObservation defines the observed stream mode details.
type StreamModeDetailsRAWObservation struct {
	// Specifies the capacity mode of the stream.
	StreamMode *string `json:"streamMode,omitempty"`
}

// StreamRAWParameters defines the configuration parameters for a native Kinesis Stream.
type StreamRAWParameters struct {
	// The encryption type to use. The only acceptable values are NONE or KMS.
	// +kubebuilder:validation:Optional
	EncryptionType *string `json:"encryptionType,omitempty"`

	// A boolean that indicates all registered consumers should be deregistered
	// from the stream so that the stream can be destroyed without error.
	// Note: this field is ignored for late-initialization purposes.
	// +kubebuilder:validation:Optional
	EnforceConsumerDeletion *bool `json:"enforceConsumerDeletion,omitempty"`

	// The GUID for the customer-managed KMS key to use for encryption.
	// You can also use a Kinesis-owned master key by specifying the alias alias/aws/kinesis.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kms/v1beta1.Key
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

// StreamRAWInitParameters defines the init parameters for a native Kinesis Stream.
type StreamRAWInitParameters struct {
	// The encryption type to use. The only acceptable values are NONE or KMS.
	// +kubebuilder:validation:Optional
	EncryptionType *string `json:"encryptionType,omitempty"`

	// A boolean that indicates all registered consumers should be deregistered.
	// +kubebuilder:validation:Optional
	EnforceConsumerDeletion *bool `json:"enforceConsumerDeletion,omitempty"`

	// The GUID for the customer-managed KMS key to use for encryption.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kms/v1beta1.Key
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

// StreamRAWObservation defines the observed state of a native Kinesis Stream.
type StreamRAWObservation struct {
	// The Amazon Resource Name (ARN) specifying the Stream (moved from spec.forProvider).
	Arn *string `json:"arn,omitempty"`

	// The unique Stream id.
	ID *string `json:"id,omitempty"`

	// A map of tags assigned to the resource, including those inherited from the provider.
	// +mapType=granular
	TagsAll map[string]*string `json:"tagsAll,omitempty"`
}

// StreamRAWSpec defines the desired state of StreamRAW.
type StreamRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider StreamRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider StreamRAWInitParameters `json:"initProvider,omitempty"`
}

// StreamRAWStatus defines the observed state of StreamRAW.
type StreamRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider StreamRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// StreamRAW is the native (non-Terraform) Schema for AWS Kinesis Streams API.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
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

// GetForProvider returns the ForProvider parameters.
func (s *StreamRAW) GetForProvider() *StreamRAWParameters { return &s.Spec.ForProvider }

// SetForProvider sets spec.forProvider to the given parameters.
// Required by the StreamCR interface for late-initialization write-back.
func (s *StreamRAW) SetForProvider(p StreamRAWParameters) { s.Spec.ForProvider = p }

// GetInitProvider returns the InitProvider parameters.
func (s *StreamRAW) GetInitProvider() *StreamRAWInitParameters { return &s.Spec.InitProvider }

// GetAtProvider returns the current observed state.
func (s *StreamRAW) GetAtProvider() StreamRAWObservation { return s.Status.AtProvider }

// SetAtProvider sets the observed state.
func (s *StreamRAW) SetAtProvider(o StreamRAWObservation) { s.Status.AtProvider = o }

// SetForProviderEncryptionType sets spec.forProvider.encryptionType.
// Used by late-initialization so that cluster and namespaced scope types both
// update their persisted spec correctly.
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
func (s *StreamRAW) SetForProviderStreamModeDetails(v *StreamModeDetailsRAWParameters) {
	s.Spec.ForProvider.StreamModeDetails = v
}
