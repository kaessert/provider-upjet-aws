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

// QueueRedriveAllowPolicyRAWSpec defines the desired state of QueueRedriveAllowPolicyRAW (namespaced scope).
type QueueRedriveAllowPolicyRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider clusternative.QueueRedriveAllowPolicyRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider clusternative.QueueRedriveAllowPolicyRAWInitParameters `json:"initProvider,omitempty"`
}

// QueueRedriveAllowPolicyRAWStatus defines the observed state of QueueRedriveAllowPolicyRAW.
type QueueRedriveAllowPolicyRAWStatus struct {
	xpv1.ConditionedStatus `json:",inline"`

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

// GetForProvider returns the ForProvider parameters.
func (q *QueueRedriveAllowPolicyRAW) GetForProvider() *clusternative.QueueRedriveAllowPolicyRAWParameters {
	return &q.Spec.ForProvider
}

// GetInitProvider returns the InitProvider parameters.
func (q *QueueRedriveAllowPolicyRAW) GetInitProvider() *clusternative.QueueRedriveAllowPolicyRAWInitParameters {
	return &q.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (q *QueueRedriveAllowPolicyRAW) GetAtProvider() clusternative.QueueRedriveAllowPolicyRAWObservation {
	return q.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (q *QueueRedriveAllowPolicyRAW) SetAtProvider(o clusternative.QueueRedriveAllowPolicyRAWObservation) {
	q.Status.AtProvider = o
}
