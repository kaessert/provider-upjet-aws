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

// QueueRAWSpec defines the desired state of QueueRAW (namespaced scope).
// It reuses the cluster-native Parameters and InitParameters types so that both
// cluster and namespaced QueueRAW implement the shared QueueCR interface.
type QueueRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider clusternative.QueueRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider clusternative.QueueRAWInitParameters `json:"initProvider,omitempty"`
}

// QueueRAWStatus defines the observed state of QueueRAW.
type QueueRAWStatus struct {
	xpv1.ConditionedStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.QueueRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// QueueRAW is the native (non-Terraform) Schema for AWS SQS Queues API (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type QueueRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   QueueRAWSpec   `json:"spec"`
	Status QueueRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// QueueRAWList contains a list of QueueRAW resources.
type QueueRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []QueueRAW `json:"items"`
}

// Repository type metadata for QueueRAW.
var (
	QueueRAW_Kind             = "QueueRAW"
	QueueRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: QueueRAW_Kind}.String()
	QueueRAW_KindAPIVersion   = QueueRAW_Kind + "." + CRDGroupVersion.String()
	QueueRAW_GroupVersionKind = CRDGroupVersion.WithKind(QueueRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&QueueRAW{}, &QueueRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (q *QueueRAW) GetForProvider() *clusternative.QueueRAWParameters { return &q.Spec.ForProvider }

// GetInitProvider returns the InitProvider parameters.
func (q *QueueRAW) GetInitProvider() *clusternative.QueueRAWInitParameters {
	return &q.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (q *QueueRAW) GetAtProvider() clusternative.QueueRAWObservation { return q.Status.AtProvider }

// SetAtProvider sets the observed state.
func (q *QueueRAW) SetAtProvider(o clusternative.QueueRAWObservation) { q.Status.AtProvider = o }
