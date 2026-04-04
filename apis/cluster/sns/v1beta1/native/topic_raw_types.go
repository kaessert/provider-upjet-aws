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

// TopicRAWParameters defines the configuration parameters for a native SNS Topic.
type TopicRAWParameters struct {
	// IAM role for failure feedback for application endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	ApplicationFailureFeedbackRoleArn *string `json:"applicationFailureFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate applicationFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	ApplicationFailureFeedbackRoleArnRef *xpv1.NamespacedReference `json:"applicationFailureFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate applicationFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	ApplicationFailureFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"applicationFailureFeedbackRoleArnSelector,omitempty"`

	// IAM role permitted to receive success feedback for application endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	ApplicationSuccessFeedbackRoleArn *string `json:"applicationSuccessFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate applicationSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	ApplicationSuccessFeedbackRoleArnRef *xpv1.NamespacedReference `json:"applicationSuccessFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate applicationSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	ApplicationSuccessFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"applicationSuccessFeedbackRoleArnSelector,omitempty"`

	// Percentage of success to sample for application endpoints.
	// +kubebuilder:validation:Optional
	ApplicationSuccessFeedbackSampleRate *float64 `json:"applicationSuccessFeedbackSampleRate,omitempty"`

	// Message archive policy for FIFO topics.
	// +kubebuilder:validation:Optional
	ArchivePolicy *string `json:"archivePolicy,omitempty"`

	// Enables content-based deduplication for FIFO topics.
	// +kubebuilder:validation:Optional
	ContentBasedDeduplication *bool `json:"contentBasedDeduplication,omitempty"`

	// SNS delivery policy.
	// +kubebuilder:validation:Optional
	DeliveryPolicy *string `json:"deliveryPolicy,omitempty"`

	// Display name for the topic.
	// +kubebuilder:validation:Optional
	DisplayName *string `json:"displayName,omitempty"`

	// Enables higher throughput for FIFO topics by adjusting the scope of deduplication.
	// +kubebuilder:validation:Optional
	FifoThroughputScope *string `json:"fifoThroughputScope,omitempty"`

	// Boolean indicating whether or not to create a FIFO (first-in-first-out) topic.
	// +kubebuilder:validation:Optional
	FifoTopic *bool `json:"fifoTopic,omitempty"`

	// IAM role for failure feedback for Firehose endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	FirehoseFailureFeedbackRoleArn *string `json:"firehoseFailureFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate firehoseFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	FirehoseFailureFeedbackRoleArnRef *xpv1.NamespacedReference `json:"firehoseFailureFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate firehoseFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	FirehoseFailureFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"firehoseFailureFeedbackRoleArnSelector,omitempty"`

	// IAM role permitted to receive success feedback for Firehose endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	FirehoseSuccessFeedbackRoleArn *string `json:"firehoseSuccessFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate firehoseSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	FirehoseSuccessFeedbackRoleArnRef *xpv1.NamespacedReference `json:"firehoseSuccessFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate firehoseSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	FirehoseSuccessFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"firehoseSuccessFeedbackRoleArnSelector,omitempty"`

	// Percentage of success to sample for Firehose endpoints.
	// +kubebuilder:validation:Optional
	FirehoseSuccessFeedbackSampleRate *float64 `json:"firehoseSuccessFeedbackSampleRate,omitempty"`

	// IAM role for failure feedback for HTTP endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	HTTPFailureFeedbackRoleArn *string `json:"httpFailureFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate httpFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	HTTPFailureFeedbackRoleArnRef *xpv1.NamespacedReference `json:"httpFailureFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate httpFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	HTTPFailureFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"httpFailureFeedbackRoleArnSelector,omitempty"`

	// IAM role permitted to receive success feedback for HTTP endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	HTTPSuccessFeedbackRoleArn *string `json:"httpSuccessFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate httpSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	HTTPSuccessFeedbackRoleArnRef *xpv1.NamespacedReference `json:"httpSuccessFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate httpSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	HTTPSuccessFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"httpSuccessFeedbackRoleArnSelector,omitempty"`

	// Percentage of success to sample for HTTP endpoints.
	// +kubebuilder:validation:Optional
	HTTPSuccessFeedbackSampleRate *float64 `json:"httpSuccessFeedbackSampleRate,omitempty"`

	// ID of an AWS-managed customer master key (CMK) for Amazon SNS or a custom CMK.
	// +kubebuilder:validation:Optional
	KMSMasterKeyID *string `json:"kmsMasterKeyId,omitempty"`

	// IAM role for failure feedback for Lambda endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	LambdaFailureFeedbackRoleArn *string `json:"lambdaFailureFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate lambdaFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	LambdaFailureFeedbackRoleArnRef *xpv1.NamespacedReference `json:"lambdaFailureFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate lambdaFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	LambdaFailureFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"lambdaFailureFeedbackRoleArnSelector,omitempty"`

	// IAM role permitted to receive success feedback for Lambda endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	LambdaSuccessFeedbackRoleArn *string `json:"lambdaSuccessFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate lambdaSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	LambdaSuccessFeedbackRoleArnRef *xpv1.NamespacedReference `json:"lambdaSuccessFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate lambdaSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	LambdaSuccessFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"lambdaSuccessFeedbackRoleArnSelector,omitempty"`

	// Percentage of success to sample for Lambda endpoints.
	// +kubebuilder:validation:Optional
	LambdaSuccessFeedbackSampleRate *float64 `json:"lambdaSuccessFeedbackSampleRate,omitempty"`

	// The fully-formed AWS policy as JSON.
	// +kubebuilder:validation:Optional
	Policy *string `json:"policy,omitempty"`

	// Region where this resource will be managed. Required for credential resolution.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// If SignatureVersion should be 1 (SHA1) or 2 (SHA256).
	// +kubebuilder:validation:Optional
	SignatureVersion *float64 `json:"signatureVersion,omitempty"`

	// IAM role for failure feedback for SQS endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	SqsFailureFeedbackRoleArn *string `json:"sqsFailureFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate sqsFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	SqsFailureFeedbackRoleArnRef *xpv1.NamespacedReference `json:"sqsFailureFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate sqsFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	SqsFailureFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"sqsFailureFeedbackRoleArnSelector,omitempty"`

	// IAM role permitted to receive success feedback for SQS endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	SqsSuccessFeedbackRoleArn *string `json:"sqsSuccessFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate sqsSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	SqsSuccessFeedbackRoleArnRef *xpv1.NamespacedReference `json:"sqsSuccessFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate sqsSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	SqsSuccessFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"sqsSuccessFeedbackRoleArnSelector,omitempty"`

	// Percentage of success to sample for SQS endpoints.
	// +kubebuilder:validation:Optional
	SqsSuccessFeedbackSampleRate *float64 `json:"sqsSuccessFeedbackSampleRate,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// Tracing mode of an Amazon SNS topic. Valid values: "PassThrough", "Active".
	// +kubebuilder:validation:Optional
	TracingConfig *string `json:"tracingConfig,omitempty"`
}

// TopicRAWInitParameters defines the init parameters for a native SNS Topic.
// These fields are merged into ForProvider when the resource is created.
type TopicRAWInitParameters struct {
	// IAM role for failure feedback for application endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	ApplicationFailureFeedbackRoleArn *string `json:"applicationFailureFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate applicationFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	ApplicationFailureFeedbackRoleArnRef *xpv1.NamespacedReference `json:"applicationFailureFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate applicationFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	ApplicationFailureFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"applicationFailureFeedbackRoleArnSelector,omitempty"`

	// IAM role permitted to receive success feedback for application endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	ApplicationSuccessFeedbackRoleArn *string `json:"applicationSuccessFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate applicationSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	ApplicationSuccessFeedbackRoleArnRef *xpv1.NamespacedReference `json:"applicationSuccessFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate applicationSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	ApplicationSuccessFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"applicationSuccessFeedbackRoleArnSelector,omitempty"`

	// Percentage of success to sample for application endpoints.
	// +kubebuilder:validation:Optional
	ApplicationSuccessFeedbackSampleRate *float64 `json:"applicationSuccessFeedbackSampleRate,omitempty"`

	// Message archive policy for FIFO topics.
	// +kubebuilder:validation:Optional
	ArchivePolicy *string `json:"archivePolicy,omitempty"`

	// Enables content-based deduplication for FIFO topics.
	// +kubebuilder:validation:Optional
	ContentBasedDeduplication *bool `json:"contentBasedDeduplication,omitempty"`

	// SNS delivery policy.
	// +kubebuilder:validation:Optional
	DeliveryPolicy *string `json:"deliveryPolicy,omitempty"`

	// Display name for the topic.
	// +kubebuilder:validation:Optional
	DisplayName *string `json:"displayName,omitempty"`

	// Enables higher throughput for FIFO topics by adjusting the scope of deduplication.
	// +kubebuilder:validation:Optional
	FifoThroughputScope *string `json:"fifoThroughputScope,omitempty"`

	// Boolean indicating whether or not to create a FIFO topic.
	// +kubebuilder:validation:Optional
	FifoTopic *bool `json:"fifoTopic,omitempty"`

	// IAM role for failure feedback for Firehose endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	FirehoseFailureFeedbackRoleArn *string `json:"firehoseFailureFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate firehoseFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	FirehoseFailureFeedbackRoleArnRef *xpv1.NamespacedReference `json:"firehoseFailureFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate firehoseFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	FirehoseFailureFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"firehoseFailureFeedbackRoleArnSelector,omitempty"`

	// IAM role permitted to receive success feedback for Firehose endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	FirehoseSuccessFeedbackRoleArn *string `json:"firehoseSuccessFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate firehoseSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	FirehoseSuccessFeedbackRoleArnRef *xpv1.NamespacedReference `json:"firehoseSuccessFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate firehoseSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	FirehoseSuccessFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"firehoseSuccessFeedbackRoleArnSelector,omitempty"`

	// Percentage of success to sample for Firehose endpoints.
	// +kubebuilder:validation:Optional
	FirehoseSuccessFeedbackSampleRate *float64 `json:"firehoseSuccessFeedbackSampleRate,omitempty"`

	// IAM role for failure feedback for HTTP endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	HTTPFailureFeedbackRoleArn *string `json:"httpFailureFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate httpFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	HTTPFailureFeedbackRoleArnRef *xpv1.NamespacedReference `json:"httpFailureFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate httpFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	HTTPFailureFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"httpFailureFeedbackRoleArnSelector,omitempty"`

	// IAM role permitted to receive success feedback for HTTP endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	HTTPSuccessFeedbackRoleArn *string `json:"httpSuccessFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate httpSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	HTTPSuccessFeedbackRoleArnRef *xpv1.NamespacedReference `json:"httpSuccessFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate httpSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	HTTPSuccessFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"httpSuccessFeedbackRoleArnSelector,omitempty"`

	// Percentage of success to sample for HTTP endpoints.
	// +kubebuilder:validation:Optional
	HTTPSuccessFeedbackSampleRate *float64 `json:"httpSuccessFeedbackSampleRate,omitempty"`

	// ID of an AWS-managed customer master key (CMK) for Amazon SNS or a custom CMK.
	// +kubebuilder:validation:Optional
	KMSMasterKeyID *string `json:"kmsMasterKeyId,omitempty"`

	// IAM role for failure feedback for Lambda endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	LambdaFailureFeedbackRoleArn *string `json:"lambdaFailureFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate lambdaFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	LambdaFailureFeedbackRoleArnRef *xpv1.NamespacedReference `json:"lambdaFailureFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate lambdaFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	LambdaFailureFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"lambdaFailureFeedbackRoleArnSelector,omitempty"`

	// IAM role permitted to receive success feedback for Lambda endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	LambdaSuccessFeedbackRoleArn *string `json:"lambdaSuccessFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate lambdaSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	LambdaSuccessFeedbackRoleArnRef *xpv1.NamespacedReference `json:"lambdaSuccessFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate lambdaSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	LambdaSuccessFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"lambdaSuccessFeedbackRoleArnSelector,omitempty"`

	// Percentage of success to sample for Lambda endpoints.
	// +kubebuilder:validation:Optional
	LambdaSuccessFeedbackSampleRate *float64 `json:"lambdaSuccessFeedbackSampleRate,omitempty"`

	// The fully-formed AWS policy as JSON.
	// +kubebuilder:validation:Optional
	Policy *string `json:"policy,omitempty"`

	// If SignatureVersion should be 1 (SHA1) or 2 (SHA256).
	// +kubebuilder:validation:Optional
	SignatureVersion *float64 `json:"signatureVersion,omitempty"`

	// IAM role for failure feedback for SQS endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	SqsFailureFeedbackRoleArn *string `json:"sqsFailureFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate sqsFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	SqsFailureFeedbackRoleArnRef *xpv1.NamespacedReference `json:"sqsFailureFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate sqsFailureFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	SqsFailureFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"sqsFailureFeedbackRoleArnSelector,omitempty"`

	// IAM role permitted to receive success feedback for SQS endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	SqsSuccessFeedbackRoleArn *string `json:"sqsSuccessFeedbackRoleArn,omitempty"`

	// Reference to a Role in iam to populate sqsSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	SqsSuccessFeedbackRoleArnRef *xpv1.NamespacedReference `json:"sqsSuccessFeedbackRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate sqsSuccessFeedbackRoleArn.
	// +kubebuilder:validation:Optional
	SqsSuccessFeedbackRoleArnSelector *xpv1.NamespacedSelector `json:"sqsSuccessFeedbackRoleArnSelector,omitempty"`

	// Percentage of success to sample for SQS endpoints.
	// +kubebuilder:validation:Optional
	SqsSuccessFeedbackSampleRate *float64 `json:"sqsSuccessFeedbackSampleRate,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// Tracing mode of an Amazon SNS topic. Valid values: "PassThrough", "Active".
	// +kubebuilder:validation:Optional
	TracingConfig *string `json:"tracingConfig,omitempty"`
}

// TopicRAWObservation defines the observed state of a native SNS Topic.
type TopicRAWObservation struct {
	// IAM role for failure feedback for application endpoints.
	ApplicationFailureFeedbackRoleArn *string `json:"applicationFailureFeedbackRoleArn,omitempty"`

	// IAM role permitted to receive success feedback for application endpoints.
	ApplicationSuccessFeedbackRoleArn *string `json:"applicationSuccessFeedbackRoleArn,omitempty"`

	// Percentage of success to sample for application endpoints.
	ApplicationSuccessFeedbackSampleRate *float64 `json:"applicationSuccessFeedbackSampleRate,omitempty"`

	// Message archive policy for FIFO topics.
	ArchivePolicy *string `json:"archivePolicy,omitempty"`

	// ARN of the SNS topic.
	Arn *string `json:"arn,omitempty"`

	// The oldest timestamp at which a FIFO topic subscriber can start a replay.
	BeginningArchiveTime *string `json:"beginningArchiveTime,omitempty"`

	// Enables content-based deduplication for FIFO topics.
	ContentBasedDeduplication *bool `json:"contentBasedDeduplication,omitempty"`

	// SNS delivery policy.
	DeliveryPolicy *string `json:"deliveryPolicy,omitempty"`

	// Display name for the topic.
	DisplayName *string `json:"displayName,omitempty"`

	// Enables higher throughput for FIFO topics.
	FifoThroughputScope *string `json:"fifoThroughputScope,omitempty"`

	// Boolean indicating whether or not to create a FIFO topic.
	FifoTopic *bool `json:"fifoTopic,omitempty"`

	// IAM role for failure feedback for Firehose endpoints.
	FirehoseFailureFeedbackRoleArn *string `json:"firehoseFailureFeedbackRoleArn,omitempty"`

	// IAM role permitted to receive success feedback for Firehose endpoints.
	FirehoseSuccessFeedbackRoleArn *string `json:"firehoseSuccessFeedbackRoleArn,omitempty"`

	// Percentage of success to sample for Firehose endpoints.
	FirehoseSuccessFeedbackSampleRate *float64 `json:"firehoseSuccessFeedbackSampleRate,omitempty"`

	// IAM role for failure feedback for HTTP endpoints.
	HTTPFailureFeedbackRoleArn *string `json:"httpFailureFeedbackRoleArn,omitempty"`

	// IAM role permitted to receive success feedback for HTTP endpoints.
	HTTPSuccessFeedbackRoleArn *string `json:"httpSuccessFeedbackRoleArn,omitempty"`

	// Percentage of success to sample for HTTP endpoints.
	HTTPSuccessFeedbackSampleRate *float64 `json:"httpSuccessFeedbackSampleRate,omitempty"`

	// ARN of the SNS topic (same as Arn).
	ID *string `json:"id,omitempty"`

	// ID of an AWS-managed customer master key (CMK) for Amazon SNS or a custom CMK.
	KMSMasterKeyID *string `json:"kmsMasterKeyId,omitempty"`

	// IAM role for failure feedback for Lambda endpoints.
	LambdaFailureFeedbackRoleArn *string `json:"lambdaFailureFeedbackRoleArn,omitempty"`

	// IAM role permitted to receive success feedback for Lambda endpoints.
	LambdaSuccessFeedbackRoleArn *string `json:"lambdaSuccessFeedbackRoleArn,omitempty"`

	// Percentage of success to sample for Lambda endpoints.
	LambdaSuccessFeedbackSampleRate *float64 `json:"lambdaSuccessFeedbackSampleRate,omitempty"`

	// AWS Account ID of the SNS topic owner.
	Owner *string `json:"owner,omitempty"`

	// The fully-formed AWS policy as JSON.
	Policy *string `json:"policy,omitempty"`

	// Region where this resource is managed.
	Region *string `json:"region,omitempty"`

	// SignatureVersion of the topic.
	SignatureVersion *float64 `json:"signatureVersion,omitempty"`

	// IAM role for failure feedback for SQS endpoints.
	SqsFailureFeedbackRoleArn *string `json:"sqsFailureFeedbackRoleArn,omitempty"`

	// IAM role permitted to receive success feedback for SQS endpoints.
	SqsSuccessFeedbackRoleArn *string `json:"sqsSuccessFeedbackRoleArn,omitempty"`

	// Percentage of success to sample for SQS endpoints.
	SqsSuccessFeedbackSampleRate *float64 `json:"sqsSuccessFeedbackSampleRate,omitempty"`

	// Key-value map of resource tags.
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// Map of tags assigned to the resource, including those inherited from the provider.
	// +mapType=granular
	TagsAll map[string]*string `json:"tagsAll,omitempty"`

	// Tracing mode of an Amazon SNS topic.
	TracingConfig *string `json:"tracingConfig,omitempty"`
}

// TopicRAWSpec defines the desired state of TopicRAW.
type TopicRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider TopicRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider TopicRAWInitParameters `json:"initProvider,omitempty"`
}

// TopicRAWStatus defines the observed state of TopicRAW.
type TopicRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider TopicRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// TopicRAW is the native (non-Terraform) Schema for AWS SNS Topics API.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type TopicRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicRAWSpec   `json:"spec"`
	Status TopicRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopicRAWList contains a list of TopicRAW resources.
type TopicRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TopicRAW `json:"items"`
}

// Repository type metadata for TopicRAW.
var (
	TopicRAW_Kind             = "TopicRAW"
	TopicRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: TopicRAW_Kind}.String()
	TopicRAW_KindAPIVersion   = TopicRAW_Kind + "." + CRDGroupVersion.String()
	TopicRAW_GroupVersionKind = CRDGroupVersion.WithKind(TopicRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&TopicRAW{}, &TopicRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (t *TopicRAW) GetForProvider() *TopicRAWParameters { return &t.Spec.ForProvider }

// SetForProvider sets spec.forProvider to the given parameters.
// Required by the TopicCR interface for late-initialization write-back.
func (t *TopicRAW) SetForProvider(p TopicRAWParameters) { t.Spec.ForProvider = p }

// GetInitProvider returns the InitProvider parameters.
func (t *TopicRAW) GetInitProvider() *TopicRAWInitParameters { return &t.Spec.InitProvider }

// GetAtProvider returns the current observed state.
func (t *TopicRAW) GetAtProvider() TopicRAWObservation { return t.Status.AtProvider }

// SetAtProvider sets the observed state.
func (t *TopicRAW) SetAtProvider(o TopicRAWObservation) { t.Status.AtProvider = o }

// SetForProviderFifoThroughputScope sets spec.forProvider.fifoThroughputScope.
// Used by late-initialization so the persisted spec is updated correctly.
func (t *TopicRAW) SetForProviderFifoThroughputScope(v *string) {
	t.Spec.ForProvider.FifoThroughputScope = v
}

// SetForProviderSignatureVersion sets spec.forProvider.signatureVersion.
// Used by late-initialization so the persisted spec is updated correctly.
func (t *TopicRAW) SetForProviderSignatureVersion(v *float64) {
	t.Spec.ForProvider.SignatureVersion = v
}

// SetForProviderTracingConfig sets spec.forProvider.tracingConfig.
// Used by late-initialization so the persisted spec is updated correctly.
func (t *TopicRAW) SetForProviderTracingConfig(v *string) {
	t.Spec.ForProvider.TracingConfig = v
}
