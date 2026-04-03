// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/kinesis/v1beta2/native"
	v1beta1native "github.com/upbound/provider-aws/v2/apis/namespaced/kinesis/v1beta1/native"
)

// StreamModeDetailsRAWInitParameters defines the init parameters for stream mode details (namespaced v1beta2).
type StreamModeDetailsRAWInitParameters struct {
	// Specifies the capacity mode of the stream. Must be either PROVISIONED or ON_DEMAND.
	// +kubebuilder:validation:Optional
	StreamMode *string `json:"streamMode,omitempty"`
}

// StreamModeDetailsRAWParameters defines the parameters for stream mode details (namespaced v1beta2).
type StreamModeDetailsRAWParameters struct {
	// Specifies the capacity mode of the stream. Must be either PROVISIONED or ON_DEMAND.
	// +kubebuilder:validation:Required
	StreamMode *string `json:"streamMode"`
}

// StreamRAWParameters defines the namespaced v1beta2 configuration parameters for a native Kinesis Stream.
// Schema is identical to v1beta1; this version exists so that manifests submitted as
// kinesis.aws.m.upbound.io/v1beta2/StreamRAW are accepted and stored as v1beta1.
type StreamRAWParameters struct {
	// The encryption type to use. The only acceptable values are NONE or KMS.
	// +kubebuilder:validation:Optional
	EncryptionType *string `json:"encryptionType,omitempty"`

	// A boolean that indicates all registered consumers should be deregistered.
	// +kubebuilder:validation:Optional
	EnforceConsumerDeletion *bool `json:"enforceConsumerDeletion,omitempty"`

	// The GUID for the customer-managed KMS key to use for encryption.
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

// StreamRAWInitParameters defines the namespaced v1beta2 init parameters for a native Kinesis Stream.
type StreamRAWInitParameters struct {
	// The encryption type to use.
	// +kubebuilder:validation:Optional
	EncryptionType *string `json:"encryptionType,omitempty"`

	// A boolean that indicates all registered consumers should be deregistered.
	// +kubebuilder:validation:Optional
	EnforceConsumerDeletion *bool `json:"enforceConsumerDeletion,omitempty"`

	// The GUID for the customer-managed KMS key to use for encryption.
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

// StreamRAWSpec defines the desired state of StreamRAW (namespaced scope, v1beta2).
type StreamRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider StreamRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider StreamRAWInitParameters `json:"initProvider,omitempty"`
}

// StreamRAWStatus defines the observed state of StreamRAW (namespaced scope, v1beta2).
type StreamRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.StreamRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// StreamRAW is the native (non-Terraform) Schema for AWS Kinesis Streams API
// (namespaced scope, v1beta2). This version is served but not stored; CRs
// submitted as v1beta2 are stored as v1beta1 by the API server.
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

// ConvertTo converts StreamRAW v1beta2 to the hub version (v1beta1).
// Since v1beta1 and v1beta2 have identical JSON schemas, the conversion is
// done by marshalling to JSON and unmarshalling into the hub type.
func (src *StreamRAW) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*v1beta1native.StreamRAW)
	if !ok {
		return fmt.Errorf("expected *v1beta1native.StreamRAW, got %T", dstRaw)
	}
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// ConvertFrom converts from the hub version (v1beta1) to StreamRAW v1beta2.
func (dst *StreamRAW) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1beta1native.StreamRAW)
	if !ok {
		return fmt.Errorf("expected *v1beta1native.StreamRAW, got %T", srcRaw)
	}
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}
