// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
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
	QueueURLRef *xpv1.Reference `json:"queueUrlRef,omitempty"`

	// Selector for a QueueRAW to populate queueUrl.
	// +kubebuilder:validation:Optional
	QueueURLSelector *xpv1.Selector `json:"queueUrlSelector,omitempty"`

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
	QueueURLRef *xpv1.Reference `json:"queueUrlRef,omitempty"`

	// Selector for a QueueRAW to populate queueUrl.
	// +kubebuilder:validation:Optional
	QueueURLSelector *xpv1.Selector `json:"queueUrlSelector,omitempty"`

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
//
// NOTE: cluster-scoped QueueRedrivePolicyRAW implements LegacyManaged (not ModernManaged)
// because the TF QueueRedrivePolicy CRD uses SecretReference (with namespace) and
// Reference (without kind) for providerConfigRef, matching the TF counterpart.
type QueueRedrivePolicyRAWSpec struct {
	// DeletionPolicy specifies what will happen to the underlying external
	// when this managed resource is deleted - either "Delete" or "Orphan" the
	// external resource. This field is planned to be deprecated in favour of
	// the ManagementPolicies field in a future release. Currently, both could be
	// set independently and non-default values would be honored if the feature
	// flag is enabled. See the design doc for more information:
	// https://github.com/crossplane/crossplane/blob/499895a25d1a1a0ba1604944ef98ac7a1a71f197/design/design-doc-observe-only-resources.md?plain=1#L223
	// +optional
	// +kubebuilder:default=Delete
	DeletionPolicy *xpv1.DeletionPolicy `json:"deletionPolicy,omitempty"`

	// WriteConnectionSecretToReference specifies the namespace and name of a
	// Secret to which any connection details for this managed resource should
	// be written. Connection details frequently include the endpoint, username,
	// and password required to connect to the managed resource.
	// +optional
	WriteConnectionSecretToReference *xpv1.SecretReference `json:"writeConnectionSecretToRef,omitempty"`

	// ProviderConfigReference specifies how the provider that will be used to
	// create, observe, update, and delete this managed resource should be
	// configured.
	// +kubebuilder:default={"name": "default"}
	ProviderConfigReference *xpv1.Reference `json:"providerConfigRef,omitempty"`

	// ManagementPolicies specify the array of actions Crossplane is allowed to
	// take on the managed and external resources.
	// +optional
	// +kubebuilder:default={"*"}
	ManagementPolicies xpv1.ManagementPolicies `json:"managementPolicies,omitempty"`

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

// SetForProvider sets spec.forProvider to the given parameters.
func (q *QueueRedrivePolicyRAW) SetForProvider(p QueueRedrivePolicyRAWParameters) {
	q.Spec.ForProvider = p
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
