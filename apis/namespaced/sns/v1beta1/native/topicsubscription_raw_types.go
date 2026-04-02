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

// TopicSubscriptionRAWParameters defines the namespaced configuration parameters for a native SNS TopicSubscription.
type TopicSubscriptionRAWParameters struct {
	// Integer indicating number of minutes to wait in retrying mode for fetching subscription arn.
	// +kubebuilder:validation:Optional
	ConfirmationTimeoutInMinutes *float64 `json:"confirmationTimeoutInMinutes,omitempty"`

	// JSON String with the delivery policy.
	// +kubebuilder:validation:Optional
	DeliveryPolicy *string `json:"deliveryPolicy,omitempty"`

	// Endpoint to send data to.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/sqs/v1beta1.Queue
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/namespaced/common.ARNExtractor()
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

	// Region where this resource will be managed. Required for credential resolution.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// JSON String with the archived message replay policy.
	// +kubebuilder:validation:Optional
	ReplayPolicy *string `json:"replayPolicy,omitempty"`

	// ARN of the IAM role to publish to Kinesis Data Firehose delivery stream.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/sns/v1beta1/native.TopicRAW
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/namespaced/common.ARNExtractor()
	TopicArn *string `json:"topicArn,omitempty"`

	// Reference to a TopicRAW in sns to populate topicArn.
	// +kubebuilder:validation:Optional
	TopicArnRef *xpv1.NamespacedReference `json:"topicArnRef,omitempty"`

	// Selector for a TopicRAW in sns to populate topicArn.
	// +kubebuilder:validation:Optional
	TopicArnSelector *xpv1.NamespacedSelector `json:"topicArnSelector,omitempty"`
}

// TopicSubscriptionRAWInitParameters defines the namespaced init parameters for a native SNS TopicSubscription.
type TopicSubscriptionRAWInitParameters struct {
	// Integer indicating number of minutes to wait in retrying mode.
	// +kubebuilder:validation:Optional
	ConfirmationTimeoutInMinutes *float64 `json:"confirmationTimeoutInMinutes,omitempty"`

	// JSON String with the delivery policy.
	// +kubebuilder:validation:Optional
	DeliveryPolicy *string `json:"deliveryPolicy,omitempty"`

	// Endpoint to send data to.
	// +kubebuilder:validation:Optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/sqs/v1beta1.Queue
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/namespaced/common.ARNExtractor()
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/sns/v1beta1/native.TopicRAW
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/namespaced/common.ARNExtractor()
	TopicArn *string `json:"topicArn,omitempty"`

	// Reference to a TopicRAW in sns to populate topicArn.
	// +kubebuilder:validation:Optional
	TopicArnRef *xpv1.NamespacedReference `json:"topicArnRef,omitempty"`

	// Selector for a TopicRAW in sns to populate topicArn.
	// +kubebuilder:validation:Optional
	TopicArnSelector *xpv1.NamespacedSelector `json:"topicArnSelector,omitempty"`
}

// TopicSubscriptionRAWSpec defines the desired state of TopicSubscriptionRAW (namespaced scope).
type TopicSubscriptionRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider TopicSubscriptionRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider TopicSubscriptionRAWInitParameters `json:"initProvider,omitempty"`
}

// TopicSubscriptionRAWStatus defines the observed state of TopicSubscriptionRAW (namespaced scope).
type TopicSubscriptionRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.TopicSubscriptionRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// TopicSubscriptionRAW is the native (non-Terraform) Schema for AWS SNS TopicSubscriptions API (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type TopicSubscriptionRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicSubscriptionRAWSpec   `json:"spec"`
	Status TopicSubscriptionRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopicSubscriptionRAWList contains a list of TopicSubscriptionRAW resources (namespaced scope).
type TopicSubscriptionRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TopicSubscriptionRAW `json:"items"`
}

// Repository type metadata for TopicSubscriptionRAW (namespaced).
var (
	TopicSubscriptionRAW_Kind             = "TopicSubscriptionRAW"
	TopicSubscriptionRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: TopicSubscriptionRAW_Kind}.String()
	TopicSubscriptionRAW_KindAPIVersion   = TopicSubscriptionRAW_Kind + "." + CRDGroupVersion.String()
	TopicSubscriptionRAW_GroupVersionKind = CRDGroupVersion.WithKind(TopicSubscriptionRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&TopicSubscriptionRAW{}, &TopicSubscriptionRAWList{})
}

// GetForProvider returns a cluster-scoped TopicSubscriptionRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (t *TopicSubscriptionRAW) GetForProvider() *clusternative.TopicSubscriptionRAWParameters {
	return &clusternative.TopicSubscriptionRAWParameters{
		ConfirmationTimeoutInMinutes: t.Spec.ForProvider.ConfirmationTimeoutInMinutes,
		DeliveryPolicy:               t.Spec.ForProvider.DeliveryPolicy,
		Endpoint:                     t.Spec.ForProvider.Endpoint,
		EndpointAutoConfirms:         t.Spec.ForProvider.EndpointAutoConfirms,
		FilterPolicy:                 t.Spec.ForProvider.FilterPolicy,
		FilterPolicyScope:            t.Spec.ForProvider.FilterPolicyScope,
		Protocol:                     t.Spec.ForProvider.Protocol,
		RawMessageDelivery:           t.Spec.ForProvider.RawMessageDelivery,
		RedrivePolicy:                t.Spec.ForProvider.RedrivePolicy,
		Region:                       t.Spec.ForProvider.Region,
		ReplayPolicy:                 t.Spec.ForProvider.ReplayPolicy,
		SubscriptionRoleArn:          t.Spec.ForProvider.SubscriptionRoleArn,
		TopicArn:                     t.Spec.ForProvider.TopicArn,
	}
}

// GetInitProvider returns a cluster-scoped TopicSubscriptionRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (t *TopicSubscriptionRAW) GetInitProvider() *clusternative.TopicSubscriptionRAWInitParameters {
	return &clusternative.TopicSubscriptionRAWInitParameters{
		ConfirmationTimeoutInMinutes: t.Spec.InitProvider.ConfirmationTimeoutInMinutes,
		DeliveryPolicy:               t.Spec.InitProvider.DeliveryPolicy,
		Endpoint:                     t.Spec.InitProvider.Endpoint,
		EndpointAutoConfirms:         t.Spec.InitProvider.EndpointAutoConfirms,
		FilterPolicy:                 t.Spec.InitProvider.FilterPolicy,
		FilterPolicyScope:            t.Spec.InitProvider.FilterPolicyScope,
		Protocol:                     t.Spec.InitProvider.Protocol,
		RawMessageDelivery:           t.Spec.InitProvider.RawMessageDelivery,
		RedrivePolicy:                t.Spec.InitProvider.RedrivePolicy,
		ReplayPolicy:                 t.Spec.InitProvider.ReplayPolicy,
		SubscriptionRoleArn:          t.Spec.InitProvider.SubscriptionRoleArn,
		TopicArn:                     t.Spec.InitProvider.TopicArn,
	}
}

// GetAtProvider returns the current observed state.
func (t *TopicSubscriptionRAW) GetAtProvider() clusternative.TopicSubscriptionRAWObservation {
	return t.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (t *TopicSubscriptionRAW) SetAtProvider(o clusternative.TopicSubscriptionRAWObservation) {
	t.Status.AtProvider = o
}

// SetForProviderFilterPolicyScope sets FilterPolicyScope for late-initialization.
// This setter is needed because GetForProvider() returns a field-copied struct,
// so mutations through the returned pointer do not propagate back to the spec.
func (t *TopicSubscriptionRAW) SetForProviderFilterPolicyScope(v *string) {
	t.Spec.ForProvider.FilterPolicyScope = v
}
