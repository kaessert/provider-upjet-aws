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

// QueuePolicyRAWSpec defines the desired state of QueuePolicyRAW (namespaced scope).
type QueuePolicyRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.policy) || (has(self.initProvider) && has(self.initProvider.policy))",message="spec.forProvider.policy is a required parameter"
	ForProvider clusternative.QueuePolicyRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider clusternative.QueuePolicyRAWInitParameters `json:"initProvider,omitempty"`
}

// QueuePolicyRAWStatus defines the observed state of QueuePolicyRAW.
type QueuePolicyRAWStatus struct {
	xpv1.ConditionedStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.QueuePolicyRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// QueuePolicyRAW is the native (non-Terraform) Schema for AWS SQS Queue Policies API (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type QueuePolicyRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

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
func (q *QueuePolicyRAW) GetForProvider() *clusternative.QueuePolicyRAWParameters {
	return &q.Spec.ForProvider
}

// GetInitProvider returns the InitProvider parameters.
func (q *QueuePolicyRAW) GetInitProvider() *clusternative.QueuePolicyRAWInitParameters {
	return &q.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (q *QueuePolicyRAW) GetAtProvider() clusternative.QueuePolicyRAWObservation {
	return q.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (q *QueuePolicyRAW) SetAtProvider(o clusternative.QueuePolicyRAWObservation) {
	q.Status.AtProvider = o
}
