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

// QueueRedrivePolicyRAWParameters defines the configuration parameters for a native SQS Queue Redrive Policy.
type QueueRedrivePolicyRAWParameters struct {
	// The URL of the SQS Queue to which to attach the redrive policy.
	//
	// +optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native.QueueRAW
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractResourceID()
	QueueURL *string `json:"queueUrl,omitempty"`

	// Reference to a QueueRAW to populate queueUrl.
	// +kubebuilder:validation:Optional
	QueueURLRef *xpv1.NamespacedReference `json:"queueUrlRef,omitempty"`

	// Selector for a QueueRAW to populate queueUrl.
	// +kubebuilder:validation:Optional
	QueueURLSelector *xpv1.NamespacedSelector `json:"queueUrlSelector,omitempty"`

	// The JSON redrive policy for the SQS queue. Accepts two key/val pairs:
	// deadLetterTargetArn and maxReceiveCount.
	// +kubebuilder:validation:Optional
	RedrivePolicy *string `json:"redrivePolicy,omitempty"`

	// Region where this resource will be managed. Required for credential resolution.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`
}

// QueueRedrivePolicyRAWInitParameters defines the init parameters for a native SQS Queue Redrive Policy.
type QueueRedrivePolicyRAWInitParameters struct {
	// The URL of the SQS Queue to which to attach the redrive policy.
	// +optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native.QueueRAW
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractResourceID()
	QueueURL *string `json:"queueUrl,omitempty"`

	// Reference to a QueueRAW to populate queueUrl.
	// +kubebuilder:validation:Optional
	QueueURLRef *xpv1.NamespacedReference `json:"queueUrlRef,omitempty"`

	// Selector for a QueueRAW to populate queueUrl.
	// +kubebuilder:validation:Optional
	QueueURLSelector *xpv1.NamespacedSelector `json:"queueUrlSelector,omitempty"`

	// The JSON redrive policy for the SQS queue.
	// +kubebuilder:validation:Optional
	RedrivePolicy *string `json:"redrivePolicy,omitempty"`
}

// QueueRedrivePolicyRAWObservation defines the observed state of a native SQS Queue Redrive Policy.
type QueueRedrivePolicyRAWObservation struct {
	// ID is the queue URL (external name).
	ID *string `json:"id,omitempty"`

	// URL of the SQS Queue to which the redrive policy is attached.
	QueueURL *string `json:"queueUrl,omitempty"`

	// The JSON redrive policy for the SQS queue.
	RedrivePolicy *string `json:"redrivePolicy,omitempty"`
}

// QueueRedrivePolicyRAWSpec defines the desired state of QueueRedrivePolicyRAW.
type QueueRedrivePolicyRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider QueueRedrivePolicyRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider QueueRedrivePolicyRAWInitParameters `json:"initProvider,omitempty"`
}

// QueueRedrivePolicyRAWStatus defines the observed state of QueueRedrivePolicyRAW.
type QueueRedrivePolicyRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider QueueRedrivePolicyRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// QueueRedrivePolicyRAW is the native (non-Terraform) Schema for AWS SQS Queue Redrive Policies API.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type QueueRedrivePolicyRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.redrivePolicy) || (has(self.initProvider) && has(self.initProvider.redrivePolicy))",message="spec.forProvider.redrivePolicy is a required parameter"
	Spec   QueueRedrivePolicyRAWSpec   `json:"spec"`
	Status QueueRedrivePolicyRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// QueueRedrivePolicyRAWList contains a list of QueueRedrivePolicyRAW resources.
type QueueRedrivePolicyRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QueueRedrivePolicyRAW `json:"items"`
}

// Repository type metadata for QueueRedrivePolicyRAW.
var (
	QueueRedrivePolicyRAW_Kind             = "QueueRedrivePolicyRAW"
	QueueRedrivePolicyRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: QueueRedrivePolicyRAW_Kind}.String()
	QueueRedrivePolicyRAW_KindAPIVersion   = QueueRedrivePolicyRAW_Kind + "." + CRDGroupVersion.String()
	QueueRedrivePolicyRAW_GroupVersionKind = CRDGroupVersion.WithKind(QueueRedrivePolicyRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&QueueRedrivePolicyRAW{}, &QueueRedrivePolicyRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (q *QueueRedrivePolicyRAW) GetForProvider() *QueueRedrivePolicyRAWParameters {
	return &q.Spec.ForProvider
}

// GetInitProvider returns the InitProvider parameters.
func (q *QueueRedrivePolicyRAW) GetInitProvider() *QueueRedrivePolicyRAWInitParameters {
	return &q.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (q *QueueRedrivePolicyRAW) GetAtProvider() QueueRedrivePolicyRAWObservation {
	return q.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (q *QueueRedrivePolicyRAW) SetAtProvider(o QueueRedrivePolicyRAWObservation) {
	q.Status.AtProvider = o
}
