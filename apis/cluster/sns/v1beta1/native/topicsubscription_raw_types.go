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

// TopicSubscriptionRAWParameters defines the configuration parameters for a native SNS TopicSubscription.
type TopicSubscriptionRAWParameters struct {
	// Integer indicating number of minutes to wait in retrying mode for fetching subscription arn.
	// Only applicable for http and https protocols.
	// +kubebuilder:validation:Optional
	ConfirmationTimeoutInMinutes *float64 `json:"confirmationTimeoutInMinutes,omitempty"`

	// JSON String with the delivery policy that will be used in the subscription.
	// +kubebuilder:validation:Optional
	DeliveryPolicy *string `json:"deliveryPolicy,omitempty"`

	// Endpoint to send data to. The contents vary with the protocol.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1.Queue
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	Endpoint *string `json:"endpoint,omitempty"`

	// Reference to a Queue in sqs to populate endpoint.
	// +kubebuilder:validation:Optional
	EndpointRef *xpv1.NamespacedReference `json:"endpointRef,omitempty"`

	// Selector for a Queue in sqs to populate endpoint.
	// +kubebuilder:validation:Optional
	EndpointSelector *xpv1.NamespacedSelector `json:"endpointSelector,omitempty"`

	// Whether the endpoint is capable of auto confirming subscription.
	// +kubebuilder:validation:Optional
	EndpointAutoConfirms *bool `json:"endpointAutoConfirms,omitempty"`

	// JSON String with the filter policy that will be used in the subscription to filter messages.
	// +kubebuilder:validation:Optional
	FilterPolicy *string `json:"filterPolicy,omitempty"`

	// Whether the filter_policy applies to MessageAttributes (default) or MessageBody.
	// +kubebuilder:validation:Optional
	FilterPolicyScope *string `json:"filterPolicyScope,omitempty"`

	// Protocol to use. Valid values are: sqs, sms, lambda, firehose, and application.
	// +kubebuilder:validation:Optional
	Protocol *string `json:"protocol,omitempty"`

	// Whether to enable raw message delivery.
	// +kubebuilder:validation:Optional
	RawMessageDelivery *bool `json:"rawMessageDelivery,omitempty"`

	// JSON String with the redrive policy that will be used in the subscription.
	// +kubebuilder:validation:Optional
	RedrivePolicy *string `json:"redrivePolicy,omitempty"`

	// Region where this resource will be managed. Required for credential resolution.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// JSON String with the archived message replay policy that will be used in the subscription.
	// +kubebuilder:validation:Optional
	ReplayPolicy *string `json:"replayPolicy,omitempty"`

	// ARN of the IAM role to publish to Kinesis Data Firehose delivery stream.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	SubscriptionRoleArn *string `json:"subscriptionRoleArn,omitempty"`

	// Reference to a Role in iam to populate subscriptionRoleArn.
	// +kubebuilder:validation:Optional
	SubscriptionRoleArnRef *xpv1.NamespacedReference `json:"subscriptionRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate subscriptionRoleArn.
	// +kubebuilder:validation:Optional
	SubscriptionRoleArnSelector *xpv1.NamespacedSelector `json:"subscriptionRoleArnSelector,omitempty"`

	// ARN of the SNS topic to subscribe to.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native.TopicRAW
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	TopicArn *string `json:"topicArn,omitempty"`

	// Reference to a TopicRAW in sns to populate topicArn.
	// +kubebuilder:validation:Optional
	TopicArnRef *xpv1.NamespacedReference `json:"topicArnRef,omitempty"`

	// Selector for a TopicRAW in sns to populate topicArn.
	// +kubebuilder:validation:Optional
	TopicArnSelector *xpv1.NamespacedSelector `json:"topicArnSelector,omitempty"`
}

// TopicSubscriptionRAWInitParameters defines the init parameters for a native SNS TopicSubscription.
type TopicSubscriptionRAWInitParameters struct {
	// Integer indicating number of minutes to wait in retrying mode for fetching subscription arn.
	// +kubebuilder:validation:Optional
	ConfirmationTimeoutInMinutes *float64 `json:"confirmationTimeoutInMinutes,omitempty"`

	// JSON String with the delivery policy.
	// +kubebuilder:validation:Optional
	DeliveryPolicy *string `json:"deliveryPolicy,omitempty"`

	// Endpoint to send data to.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1.Queue
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	Endpoint *string `json:"endpoint,omitempty"`

	// Reference to a Queue in sqs to populate endpoint.
	// +kubebuilder:validation:Optional
	EndpointRef *xpv1.NamespacedReference `json:"endpointRef,omitempty"`

	// Selector for a Queue in sqs to populate endpoint.
	// +kubebuilder:validation:Optional
	EndpointSelector *xpv1.NamespacedSelector `json:"endpointSelector,omitempty"`

	// Whether the endpoint is capable of auto confirming subscription.
	// +kubebuilder:validation:Optional
	EndpointAutoConfirms *bool `json:"endpointAutoConfirms,omitempty"`

	// JSON String with the filter policy.
	// +kubebuilder:validation:Optional
	FilterPolicy *string `json:"filterPolicy,omitempty"`

	// Whether the filter_policy applies to MessageAttributes or MessageBody.
	// +kubebuilder:validation:Optional
	FilterPolicyScope *string `json:"filterPolicyScope,omitempty"`

	// Protocol to use.
	// +kubebuilder:validation:Optional
	Protocol *string `json:"protocol,omitempty"`

	// Whether to enable raw message delivery.
	// +kubebuilder:validation:Optional
	RawMessageDelivery *bool `json:"rawMessageDelivery,omitempty"`

	// JSON String with the redrive policy.
	// +kubebuilder:validation:Optional
	RedrivePolicy *string `json:"redrivePolicy,omitempty"`

	// JSON String with the archived message replay policy.
	// +kubebuilder:validation:Optional
	ReplayPolicy *string `json:"replayPolicy,omitempty"`

	// ARN of the IAM role to publish to Kinesis Data Firehose delivery stream.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	SubscriptionRoleArn *string `json:"subscriptionRoleArn,omitempty"`

	// Reference to a Role in iam to populate subscriptionRoleArn.
	// +kubebuilder:validation:Optional
	SubscriptionRoleArnRef *xpv1.NamespacedReference `json:"subscriptionRoleArnRef,omitempty"`

	// Selector for a Role in iam to populate subscriptionRoleArn.
	// +kubebuilder:validation:Optional
	SubscriptionRoleArnSelector *xpv1.NamespacedSelector `json:"subscriptionRoleArnSelector,omitempty"`

	// ARN of the SNS topic to subscribe to.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native.TopicRAW
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	TopicArn *string `json:"topicArn,omitempty"`

	// Reference to a TopicRAW in sns to populate topicArn.
	// +kubebuilder:validation:Optional
	TopicArnRef *xpv1.NamespacedReference `json:"topicArnRef,omitempty"`

	// Selector for a TopicRAW in sns to populate topicArn.
	// +kubebuilder:validation:Optional
	TopicArnSelector *xpv1.NamespacedSelector `json:"topicArnSelector,omitempty"`
}

// TopicSubscriptionRAWObservation defines the observed state of a native SNS TopicSubscription.
type TopicSubscriptionRAWObservation struct {
	// ARN of the subscription.
	Arn *string `json:"arn,omitempty"`

	// Integer indicating number of minutes to wait in retrying mode.
	ConfirmationTimeoutInMinutes *float64 `json:"confirmationTimeoutInMinutes,omitempty"`

	// Whether the subscription confirmation request was authenticated.
	ConfirmationWasAuthenticated *bool `json:"confirmationWasAuthenticated,omitempty"`

	// JSON String with the delivery policy.
	DeliveryPolicy *string `json:"deliveryPolicy,omitempty"`

	// Endpoint to send data to.
	Endpoint *string `json:"endpoint,omitempty"`

	// Whether the endpoint is capable of auto confirming subscription.
	EndpointAutoConfirms *bool `json:"endpointAutoConfirms,omitempty"`

	// JSON String with the filter policy.
	FilterPolicy *string `json:"filterPolicy,omitempty"`

	// Whether the filter_policy applies to MessageAttributes or MessageBody.
	FilterPolicyScope *string `json:"filterPolicyScope,omitempty"`

	// ARN of the subscription (same as Arn).
	ID *string `json:"id,omitempty"`

	// AWS account ID of the subscription's owner.
	OwnerID *string `json:"ownerId,omitempty"`

	// Whether the subscription has not been confirmed.
	PendingConfirmation *bool `json:"pendingConfirmation,omitempty"`

	// Protocol to use.
	Protocol *string `json:"protocol,omitempty"`

	// Whether raw message delivery is enabled.
	RawMessageDelivery *bool `json:"rawMessageDelivery,omitempty"`

	// JSON String with the redrive policy.
	RedrivePolicy *string `json:"redrivePolicy,omitempty"`

	// Region where this resource is managed.
	Region *string `json:"region,omitempty"`

	// JSON String with the archived message replay policy.
	ReplayPolicy *string `json:"replayPolicy,omitempty"`

	// ARN of the IAM role to publish to Kinesis Data Firehose delivery stream.
	SubscriptionRoleArn *string `json:"subscriptionRoleArn,omitempty"`

	// ARN of the SNS topic to subscribe to.
	TopicArn *string `json:"topicArn,omitempty"`
}

// TopicSubscriptionRAWSpec defines the desired state of TopicSubscriptionRAW.
type TopicSubscriptionRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider TopicSubscriptionRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider TopicSubscriptionRAWInitParameters `json:"initProvider,omitempty"`
}

// TopicSubscriptionRAWStatus defines the observed state of TopicSubscriptionRAW.
type TopicSubscriptionRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider TopicSubscriptionRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// TopicSubscriptionRAW is the native (non-Terraform) Schema for AWS SNS TopicSubscriptions API.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type TopicSubscriptionRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicSubscriptionRAWSpec   `json:"spec"`
	Status TopicSubscriptionRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopicSubscriptionRAWList contains a list of TopicSubscriptionRAW resources.
type TopicSubscriptionRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TopicSubscriptionRAW `json:"items"`
}

// Repository type metadata for TopicSubscriptionRAW.
var (
	TopicSubscriptionRAW_Kind             = "TopicSubscriptionRAW"
	TopicSubscriptionRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: TopicSubscriptionRAW_Kind}.String()
	TopicSubscriptionRAW_KindAPIVersion   = TopicSubscriptionRAW_Kind + "." + CRDGroupVersion.String()
	TopicSubscriptionRAW_GroupVersionKind = CRDGroupVersion.WithKind(TopicSubscriptionRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&TopicSubscriptionRAW{}, &TopicSubscriptionRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (t *TopicSubscriptionRAW) GetForProvider() *TopicSubscriptionRAWParameters {
	return &t.Spec.ForProvider
}

// SetForProvider sets spec.forProvider to the given parameters.
// Required by the TopicSubscriptionCR interface for late-initialization write-back.
func (t *TopicSubscriptionRAW) SetForProvider(p TopicSubscriptionRAWParameters) {
	t.Spec.ForProvider = p
}

// GetInitProvider returns the InitProvider parameters.
func (t *TopicSubscriptionRAW) GetInitProvider() *TopicSubscriptionRAWInitParameters {
	return &t.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (t *TopicSubscriptionRAW) GetAtProvider() TopicSubscriptionRAWObservation {
	return t.Status.AtProvider
}

// SetForProviderFilterPolicyScope sets FilterPolicyScope for late-initialization.
func (t *TopicSubscriptionRAW) SetForProviderFilterPolicyScope(v *string) {
	t.Spec.ForProvider.FilterPolicyScope = v
}

// SetAtProvider sets the observed state.
func (t *TopicSubscriptionRAW) SetAtProvider(o TopicSubscriptionRAWObservation) {
	t.Status.AtProvider = o
}
