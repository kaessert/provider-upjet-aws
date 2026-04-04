// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// QueuePolicyRAWParameters defines the configuration parameters for a native SQS Queue Policy.
type QueuePolicyRAWParameters struct {
	// JSON policy for the SQS queue. Ensure that Version = "2012-10-17" is set
	// in the policy or AWS may hang in creating the queue.
	// +kubebuilder:validation:Optional
	Policy *string `json:"policy,omitempty"`

	// URL of the SQS Queue to which to attach the policy.
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

	// Region where this resource will be managed. Required for credential resolution.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`
}

// QueuePolicyRAWInitParameters defines the init parameters for a native SQS Queue Policy.
type QueuePolicyRAWInitParameters struct {
	// JSON policy for the SQS queue.
	// +kubebuilder:validation:Optional
	Policy *string `json:"policy,omitempty"`

	// URL of the SQS Queue to which to attach the policy.
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
}

// QueuePolicyRAWObservation defines the observed state of a native SQS Queue Policy.
type QueuePolicyRAWObservation struct {
	// ID is the queue URL (external name).
	ID *string `json:"id,omitempty"`

	// URL of the SQS Queue to which the policy is attached.
	QueueURL *string `json:"queueUrl,omitempty"`
}

// QueuePolicyRAWSpec defines the desired state of QueuePolicyRAW.
//
// NOTE: cluster-scoped QueuePolicyRAW implements LegacyManaged (not ModernManaged)
// because the TF QueuePolicy CRD uses SecretReference (with namespace) and
// Reference (without kind) for providerConfigRef, matching the TF counterpart.
type QueuePolicyRAWSpec struct {
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
	ForProvider QueuePolicyRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider QueuePolicyRAWInitParameters `json:"initProvider,omitempty"`
}

// QueuePolicyRAWStatus defines the observed state of QueuePolicyRAW.
type QueuePolicyRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider QueuePolicyRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// QueuePolicyRAW is the native (non-Terraform) Schema for AWS SQS Queue Policies API.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type QueuePolicyRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.policy) || (has(self.initProvider) && has(self.initProvider.policy))",message="spec.forProvider.policy is a required parameter"
	Spec   QueuePolicyRAWSpec   `json:"spec"`
	Status QueuePolicyRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// QueuePolicyRAWList contains a list of QueuePolicyRAW resources.
type QueuePolicyRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QueuePolicyRAW `json:"items"`
}

// Repository type metadata for QueuePolicyRAW.
var (
	QueuePolicyRAW_Kind             = "QueuePolicyRAW"
	QueuePolicyRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: QueuePolicyRAW_Kind}.String()
	QueuePolicyRAW_KindAPIVersion   = QueuePolicyRAW_Kind + "." + CRDGroupVersion.String()
	QueuePolicyRAW_GroupVersionKind = CRDGroupVersion.WithKind(QueuePolicyRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&QueuePolicyRAW{}, &QueuePolicyRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (q *QueuePolicyRAW) GetForProvider() *QueuePolicyRAWParameters { return &q.Spec.ForProvider }

// SetForProvider sets spec.forProvider to the given parameters.
func (q *QueuePolicyRAW) SetForProvider(p QueuePolicyRAWParameters) { q.Spec.ForProvider = p }

// GetInitProvider returns the InitProvider parameters.
func (q *QueuePolicyRAW) GetInitProvider() *QueuePolicyRAWInitParameters {
	return &q.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (q *QueuePolicyRAW) GetAtProvider() QueuePolicyRAWObservation { return q.Status.AtProvider }

// SetAtProvider sets the observed state.
func (q *QueuePolicyRAW) SetAtProvider(o QueuePolicyRAWObservation) { q.Status.AtProvider = o }
