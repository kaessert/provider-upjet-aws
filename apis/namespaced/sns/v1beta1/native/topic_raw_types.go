// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native"
)

// TopicRAWParameters defines the namespaced configuration parameters for a native SNS Topic.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type TopicRAWParameters struct {
	// IAM role for failure feedback for application endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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

// TopicRAWInitParameters defines the namespaced init parameters for a native SNS Topic.
type TopicRAWInitParameters struct {
	// IAM role for failure feedback for application endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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

	// Enables higher throughput for FIFO topics.
	// +kubebuilder:validation:Optional
	FifoThroughputScope *string `json:"fifoThroughputScope,omitempty"`

	// Boolean indicating whether or not to create a FIFO topic.
	// +kubebuilder:validation:Optional
	FifoTopic *bool `json:"fifoTopic,omitempty"`

	// IAM role for failure feedback for Firehose endpoints.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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

// TopicRAWSpec defines the desired state of TopicRAW (namespaced scope).
type TopicRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider TopicRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider TopicRAWInitParameters `json:"initProvider,omitempty"`
}

// TopicRAWStatus defines the observed state of TopicRAW (namespaced scope).
type TopicRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.TopicRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// TopicRAW is the native (non-Terraform) Schema for AWS SNS Topics API (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type TopicRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicRAWSpec   `json:"spec"`
	Status TopicRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopicRAWList contains a list of TopicRAW resources (namespaced scope).
type TopicRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TopicRAW `json:"items"`
}

// Repository type metadata for TopicRAW (namespaced).
var (
	TopicRAW_Kind             = "TopicRAW"
	TopicRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: TopicRAW_Kind}.String()
	TopicRAW_KindAPIVersion   = TopicRAW_Kind + "." + CRDGroupVersion.String()
	TopicRAW_GroupVersionKind = CRDGroupVersion.WithKind(TopicRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&TopicRAW{}, &TopicRAWList{})
}

// GetForProvider returns a cluster-scoped TopicRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (t *TopicRAW) GetForProvider() *clusternative.TopicRAWParameters {
	return &clusternative.TopicRAWParameters{
		ApplicationFailureFeedbackRoleArn:    t.Spec.ForProvider.ApplicationFailureFeedbackRoleArn,
		ApplicationSuccessFeedbackRoleArn:    t.Spec.ForProvider.ApplicationSuccessFeedbackRoleArn,
		ApplicationSuccessFeedbackSampleRate: t.Spec.ForProvider.ApplicationSuccessFeedbackSampleRate,
		ArchivePolicy:                        t.Spec.ForProvider.ArchivePolicy,
		ContentBasedDeduplication:            t.Spec.ForProvider.ContentBasedDeduplication,
		DeliveryPolicy:                       t.Spec.ForProvider.DeliveryPolicy,
		DisplayName:                          t.Spec.ForProvider.DisplayName,
		FifoThroughputScope:                  t.Spec.ForProvider.FifoThroughputScope,
		FifoTopic:                            t.Spec.ForProvider.FifoTopic,
		FirehoseFailureFeedbackRoleArn:       t.Spec.ForProvider.FirehoseFailureFeedbackRoleArn,
		FirehoseSuccessFeedbackRoleArn:       t.Spec.ForProvider.FirehoseSuccessFeedbackRoleArn,
		FirehoseSuccessFeedbackSampleRate:    t.Spec.ForProvider.FirehoseSuccessFeedbackSampleRate,
		HTTPFailureFeedbackRoleArn:           t.Spec.ForProvider.HTTPFailureFeedbackRoleArn,
		HTTPSuccessFeedbackRoleArn:           t.Spec.ForProvider.HTTPSuccessFeedbackRoleArn,
		HTTPSuccessFeedbackSampleRate:        t.Spec.ForProvider.HTTPSuccessFeedbackSampleRate,
		KMSMasterKeyID:                       t.Spec.ForProvider.KMSMasterKeyID,
		LambdaFailureFeedbackRoleArn:         t.Spec.ForProvider.LambdaFailureFeedbackRoleArn,
		LambdaSuccessFeedbackRoleArn:         t.Spec.ForProvider.LambdaSuccessFeedbackRoleArn,
		LambdaSuccessFeedbackSampleRate:      t.Spec.ForProvider.LambdaSuccessFeedbackSampleRate,
		Policy:                               t.Spec.ForProvider.Policy,
		Region:                               t.Spec.ForProvider.Region,
		SignatureVersion:                     t.Spec.ForProvider.SignatureVersion,
		SqsFailureFeedbackRoleArn:            t.Spec.ForProvider.SqsFailureFeedbackRoleArn,
		SqsSuccessFeedbackRoleArn:            t.Spec.ForProvider.SqsSuccessFeedbackRoleArn,
		SqsSuccessFeedbackSampleRate:         t.Spec.ForProvider.SqsSuccessFeedbackSampleRate,
		Tags:                                 t.Spec.ForProvider.Tags,
		TracingConfig:                        t.Spec.ForProvider.TracingConfig,
	}
}

// GetInitProvider returns a cluster-scoped TopicRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (t *TopicRAW) GetInitProvider() *clusternative.TopicRAWInitParameters {
	return &clusternative.TopicRAWInitParameters{
		ApplicationFailureFeedbackRoleArn:    t.Spec.InitProvider.ApplicationFailureFeedbackRoleArn,
		ApplicationSuccessFeedbackRoleArn:    t.Spec.InitProvider.ApplicationSuccessFeedbackRoleArn,
		ApplicationSuccessFeedbackSampleRate: t.Spec.InitProvider.ApplicationSuccessFeedbackSampleRate,
		ArchivePolicy:                        t.Spec.InitProvider.ArchivePolicy,
		ContentBasedDeduplication:            t.Spec.InitProvider.ContentBasedDeduplication,
		DeliveryPolicy:                       t.Spec.InitProvider.DeliveryPolicy,
		DisplayName:                          t.Spec.InitProvider.DisplayName,
		FifoThroughputScope:                  t.Spec.InitProvider.FifoThroughputScope,
		FifoTopic:                            t.Spec.InitProvider.FifoTopic,
		FirehoseFailureFeedbackRoleArn:       t.Spec.InitProvider.FirehoseFailureFeedbackRoleArn,
		FirehoseSuccessFeedbackRoleArn:       t.Spec.InitProvider.FirehoseSuccessFeedbackRoleArn,
		FirehoseSuccessFeedbackSampleRate:    t.Spec.InitProvider.FirehoseSuccessFeedbackSampleRate,
		HTTPFailureFeedbackRoleArn:           t.Spec.InitProvider.HTTPFailureFeedbackRoleArn,
		HTTPSuccessFeedbackRoleArn:           t.Spec.InitProvider.HTTPSuccessFeedbackRoleArn,
		HTTPSuccessFeedbackSampleRate:        t.Spec.InitProvider.HTTPSuccessFeedbackSampleRate,
		KMSMasterKeyID:                       t.Spec.InitProvider.KMSMasterKeyID,
		LambdaFailureFeedbackRoleArn:         t.Spec.InitProvider.LambdaFailureFeedbackRoleArn,
		LambdaSuccessFeedbackRoleArn:         t.Spec.InitProvider.LambdaSuccessFeedbackRoleArn,
		LambdaSuccessFeedbackSampleRate:      t.Spec.InitProvider.LambdaSuccessFeedbackSampleRate,
		Policy:                               t.Spec.InitProvider.Policy,
		SignatureVersion:                     t.Spec.InitProvider.SignatureVersion,
		SqsFailureFeedbackRoleArn:            t.Spec.InitProvider.SqsFailureFeedbackRoleArn,
		SqsSuccessFeedbackRoleArn:            t.Spec.InitProvider.SqsSuccessFeedbackRoleArn,
		SqsSuccessFeedbackSampleRate:         t.Spec.InitProvider.SqsSuccessFeedbackSampleRate,
		Tags:                                 t.Spec.InitProvider.Tags,
		TracingConfig:                        t.Spec.InitProvider.TracingConfig,
	}
}

// GetAtProvider returns the current observed state.
func (t *TopicRAW) GetAtProvider() clusternative.TopicRAWObservation {
	return t.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (t *TopicRAW) SetAtProvider(o clusternative.TopicRAWObservation) {
	t.Status.AtProvider = o
}
