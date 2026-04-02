// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
)

// QueueRedrivePolicyRAWParameters defines the namespaced configuration
// parameters for a native SQS Queue Redrive Policy.  The QueueURL reference
// annotation points to the namespaced QueueRAW type so that angryjet generates
// a namespaced resolver.
type QueueRedrivePolicyRAWParameters struct {
	// The URL of the SQS Queue to which to attach the redrive policy.
	//
	// +optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/sqs/v1beta1/native.QueueRAW
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

// QueueRedrivePolicyRAWInitParameters defines the namespaced init parameters
// for a native SQS Queue Redrive Policy.
type QueueRedrivePolicyRAWInitParameters struct {
	// The URL of the SQS Queue to which to attach the redrive policy.
	// +optional
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/sqs/v1beta1/native.QueueRAW
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

// QueueRedrivePolicyRAWSpec defines the desired state of QueueRedrivePolicyRAW (namespaced scope).
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
// Note: using xpv1.ResourceStatus (not ConditionedStatus) so that angryjet's
// ManagedV2() matcher recognises this as a v2-style managed resource and
// generates zz_generated.resolvers.go for this package.
type QueueRedrivePolicyRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.QueueRedrivePolicyRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// QueueRedrivePolicyRAW is the native (non-Terraform) Schema for AWS SQS Queue Redrive Policies API (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
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

// GetForProvider converts the namespaced ForProvider params to the cluster
// type required by the shared QueueRedrivePolicyCR interface.  The shared CRUD
// code only reads from the returned value (no late-init writes), so returning a
// freshly allocated cluster struct is safe.
func (q *QueueRedrivePolicyRAW) GetForProvider() *clusternative.QueueRedrivePolicyRAWParameters {
	return &clusternative.QueueRedrivePolicyRAWParameters{
		QueueURL:         q.Spec.ForProvider.QueueURL,
		QueueURLRef:      q.Spec.ForProvider.QueueURLRef,
		QueueURLSelector: q.Spec.ForProvider.QueueURLSelector,
		RedrivePolicy:    q.Spec.ForProvider.RedrivePolicy,
		Region:           q.Spec.ForProvider.Region,
	}
}

// GetInitProvider converts the namespaced InitProvider params to the cluster type.
func (q *QueueRedrivePolicyRAW) GetInitProvider() *clusternative.QueueRedrivePolicyRAWInitParameters {
	return &clusternative.QueueRedrivePolicyRAWInitParameters{
		QueueURL:         q.Spec.InitProvider.QueueURL,
		QueueURLRef:      q.Spec.InitProvider.QueueURLRef,
		QueueURLSelector: q.Spec.InitProvider.QueueURLSelector,
		RedrivePolicy:    q.Spec.InitProvider.RedrivePolicy,
	}
}

// GetAtProvider returns the current observed state.
func (q *QueueRedrivePolicyRAW) GetAtProvider() clusternative.QueueRedrivePolicyRAWObservation {
	return q.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (q *QueueRedrivePolicyRAW) SetAtProvider(o clusternative.QueueRedrivePolicyRAWObservation) {
	q.Status.AtProvider = o
}
