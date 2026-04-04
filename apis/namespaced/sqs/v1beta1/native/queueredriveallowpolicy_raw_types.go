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

// QueueRedriveAllowPolicyRAWParameters defines the namespaced configuration
// parameters for a native SQS Queue Redrive Allow Policy.  The QueueURL
// reference annotation points to the namespaced QueueRAW type so that
// angryjet generates a namespaced resolver.
type QueueRedriveAllowPolicyRAWParameters struct {
	// The URL of the SQS Queue to which to attach the redrive allow policy.
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

	// The JSON redrive allow policy for the SQS queue. Learn more in the
	// Amazon SQS dead-letter queues documentation.
	// +kubebuilder:validation:Optional
	RedriveAllowPolicy *string `json:"redriveAllowPolicy,omitempty"`

	// Region where this resource will be managed. Required for credential resolution.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`
}

// QueueRedriveAllowPolicyRAWInitParameters defines the namespaced init
// parameters for a native SQS Queue Redrive Allow Policy.
type QueueRedriveAllowPolicyRAWInitParameters struct {
	// The URL of the SQS Queue to which to attach the redrive allow policy.
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

	// The JSON redrive allow policy for the SQS queue.
	// +kubebuilder:validation:Optional
	RedriveAllowPolicy *string `json:"redriveAllowPolicy,omitempty"`
}

// QueueRedriveAllowPolicyRAWSpec defines the desired state of QueueRedriveAllowPolicyRAW (namespaced scope).
type QueueRedriveAllowPolicyRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider QueueRedriveAllowPolicyRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider QueueRedriveAllowPolicyRAWInitParameters `json:"initProvider,omitempty"`
}

// QueueRedriveAllowPolicyRAWStatus defines the observed state of QueueRedriveAllowPolicyRAW.
// Note: using xpv1.ResourceStatus (not ConditionedStatus) so that angryjet's
// ManagedV2() matcher recognises this as a v2-style managed resource and
// generates zz_generated.resolvers.go for this package.
type QueueRedriveAllowPolicyRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.QueueRedriveAllowPolicyRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// QueueRedriveAllowPolicyRAW is the native (non-Terraform) Schema for AWS SQS Queue Redrive Allow Policies API (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type QueueRedriveAllowPolicyRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.redriveAllowPolicy) || (has(self.initProvider) && has(self.initProvider.redriveAllowPolicy))",message="spec.forProvider.redriveAllowPolicy is a required parameter"
	Spec   QueueRedriveAllowPolicyRAWSpec   `json:"spec"`
	Status QueueRedriveAllowPolicyRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// QueueRedriveAllowPolicyRAWList contains a list of QueueRedriveAllowPolicyRAW resources.
type QueueRedriveAllowPolicyRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QueueRedriveAllowPolicyRAW `json:"items"`
}

// Repository type metadata for QueueRedriveAllowPolicyRAW.
var (
	QueueRedriveAllowPolicyRAW_Kind             = "QueueRedriveAllowPolicyRAW"
	QueueRedriveAllowPolicyRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: QueueRedriveAllowPolicyRAW_Kind}.String()
	QueueRedriveAllowPolicyRAW_KindAPIVersion   = QueueRedriveAllowPolicyRAW_Kind + "." + CRDGroupVersion.String()
	QueueRedriveAllowPolicyRAW_GroupVersionKind = CRDGroupVersion.WithKind(QueueRedriveAllowPolicyRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&QueueRedriveAllowPolicyRAW{}, &QueueRedriveAllowPolicyRAWList{})
}

// GetForProvider converts the namespaced ForProvider params to the cluster
// type required by the shared QueueRedriveAllowPolicyCR interface.  The shared
// CRUD code only reads from the returned value (no late-init writes), so
// returning a freshly allocated cluster struct is safe.
func (q *QueueRedriveAllowPolicyRAW) GetForProvider() *clusternative.QueueRedriveAllowPolicyRAWParameters {
	p := &clusternative.QueueRedriveAllowPolicyRAWParameters{
		QueueURL:           q.Spec.ForProvider.QueueURL,
		RedriveAllowPolicy: q.Spec.ForProvider.RedriveAllowPolicy,
		Region:             q.Spec.ForProvider.Region,
	}
	if q.Spec.ForProvider.QueueURLRef != nil {
		p.QueueURLRef = namespacedRefToRef(q.Spec.ForProvider.QueueURLRef)
	}
	if q.Spec.ForProvider.QueueURLSelector != nil {
		p.QueueURLSelector = namespacedSelectorToSelector(q.Spec.ForProvider.QueueURLSelector)
	}
	return p
}

// GetInitProvider converts the namespaced InitProvider params to the cluster type.
func (q *QueueRedriveAllowPolicyRAW) GetInitProvider() *clusternative.QueueRedriveAllowPolicyRAWInitParameters {
	p := &clusternative.QueueRedriveAllowPolicyRAWInitParameters{
		QueueURL:           q.Spec.InitProvider.QueueURL,
		RedriveAllowPolicy: q.Spec.InitProvider.RedriveAllowPolicy,
	}
	if q.Spec.InitProvider.QueueURLRef != nil {
		p.QueueURLRef = namespacedRefToRef(q.Spec.InitProvider.QueueURLRef)
	}
	if q.Spec.InitProvider.QueueURLSelector != nil {
		p.QueueURLSelector = namespacedSelectorToSelector(q.Spec.InitProvider.QueueURLSelector)
	}
	return p
}

// GetAtProvider returns the current observed state.
func (q *QueueRedriveAllowPolicyRAW) GetAtProvider() clusternative.QueueRedriveAllowPolicyRAWObservation {
	return q.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (q *QueueRedriveAllowPolicyRAW) SetAtProvider(o clusternative.QueueRedriveAllowPolicyRAWObservation) {
	q.Status.AtProvider = o
}

// SetForProvider copies cluster-scoped QueueRedriveAllowPolicyRAWParameters
// back to the namespaced spec. Added for interface consistency.
func (q *QueueRedriveAllowPolicyRAW) SetForProvider(p clusternative.QueueRedriveAllowPolicyRAWParameters) {
	q.Spec.ForProvider.QueueURL = p.QueueURL
	q.Spec.ForProvider.QueueURLRef = refToNamespacedRef(p.QueueURLRef)
	q.Spec.ForProvider.QueueURLSelector = selectorToNamespacedSelector(p.QueueURLSelector)
	q.Spec.ForProvider.RedriveAllowPolicy = p.RedriveAllowPolicy
	q.Spec.ForProvider.Region = p.Region
}
