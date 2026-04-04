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

// StreamConsumerRAWParameters defines the configuration parameters for a native Kinesis Stream Consumer.
type StreamConsumerRAWParameters struct {
	// Name of the stream consumer.
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty"`

	// Region where this resource will be managed. Required for credential resolution.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Amazon Resource Name (ARN) of the data stream the consumer is registered with.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kinesis/v1beta2/native.StreamRAW
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractResourceID()
	// +kubebuilder:validation:Optional
	StreamArn *string `json:"streamArn,omitempty"`

	// Reference to a StreamRAW in kinesis to populate streamArn.
	// +kubebuilder:validation:Optional
	StreamArnRef *xpv1.NamespacedReference `json:"streamArnRef,omitempty"`

	// Selector for a StreamRAW in kinesis to populate streamArn.
	// +kubebuilder:validation:Optional
	StreamArnSelector *xpv1.NamespacedSelector `json:"streamArnSelector,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// StreamConsumerRAWInitParameters defines the init parameters for a native Kinesis Stream Consumer.
type StreamConsumerRAWInitParameters struct {
	// Name of the stream consumer.
	// +kubebuilder:validation:Optional
	Name *string `json:"name,omitempty"`

	// Amazon Resource Name (ARN) of the data stream the consumer is registered with.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kinesis/v1beta2/native.StreamRAW
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractResourceID()
	// +kubebuilder:validation:Optional
	StreamArn *string `json:"streamArn,omitempty"`

	// Reference to a StreamRAW in kinesis to populate streamArn.
	// +kubebuilder:validation:Optional
	StreamArnRef *xpv1.NamespacedReference `json:"streamArnRef,omitempty"`

	// Selector for a StreamRAW in kinesis to populate streamArn.
	// +kubebuilder:validation:Optional
	StreamArnSelector *xpv1.NamespacedSelector `json:"streamArnSelector,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`
}

// StreamConsumerRAWObservation defines the observed state of a native Kinesis Stream Consumer.
type StreamConsumerRAWObservation struct {
	// Amazon Resource Name (ARN) of the stream consumer.
	Arn *string `json:"arn,omitempty"`

	// Approximate timestamp in RFC3339 format of when the stream consumer was created.
	CreationTimestamp *string `json:"creationTimestamp,omitempty"`

	// The unique Stream Consumer id (same as ARN).
	ID *string `json:"id,omitempty"`

	// A map of tags assigned to the resource, including those inherited from the provider.
	// +mapType=granular
	TagsAll map[string]*string `json:"tagsAll,omitempty"`
}

// StreamConsumerRAWSpec defines the desired state of StreamConsumerRAW.
type StreamConsumerRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider StreamConsumerRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider StreamConsumerRAWInitParameters `json:"initProvider,omitempty"`
}

// StreamConsumerRAWStatus defines the observed state of StreamConsumerRAW.
type StreamConsumerRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider StreamConsumerRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// StreamConsumerRAW is the native (non-Terraform) Schema for AWS Kinesis Stream Consumers API.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type StreamConsumerRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.name) || (has(self.initProvider) && has(self.initProvider.name))",message="spec.forProvider.name is a required parameter"
	Spec   StreamConsumerRAWSpec   `json:"spec"`
	Status StreamConsumerRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StreamConsumerRAWList contains a list of StreamConsumerRAW resources.
type StreamConsumerRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StreamConsumerRAW `json:"items"`
}

// Repository type metadata for StreamConsumerRAW.
var (
	StreamConsumerRAW_Kind             = "StreamConsumerRAW"
	StreamConsumerRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: StreamConsumerRAW_Kind}.String()
	StreamConsumerRAW_KindAPIVersion   = StreamConsumerRAW_Kind + "." + CRDGroupVersion.String()
	StreamConsumerRAW_GroupVersionKind = CRDGroupVersion.WithKind(StreamConsumerRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&StreamConsumerRAW{}, &StreamConsumerRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (s *StreamConsumerRAW) GetForProvider() *StreamConsumerRAWParameters {
	return &s.Spec.ForProvider
}

// SetForProvider sets spec.forProvider to the given parameters.
// Required by the StreamConsumerCR interface for consistency with other native CR types.
func (s *StreamConsumerRAW) SetForProvider(p StreamConsumerRAWParameters) {
	s.Spec.ForProvider = p
}

// GetInitProvider returns the InitProvider parameters.
func (s *StreamConsumerRAW) GetInitProvider() *StreamConsumerRAWInitParameters {
	return &s.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (s *StreamConsumerRAW) GetAtProvider() StreamConsumerRAWObservation {
	return s.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (s *StreamConsumerRAW) SetAtProvider(o StreamConsumerRAWObservation) {
	s.Status.AtProvider = o
}
