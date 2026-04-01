// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sfn/v1beta2/native"
)

// StateMachineRAWSpec defines the desired state of StateMachineRAW (namespaced scope).
// It reuses the cluster-native Parameters and InitParameters types so that both
// cluster and namespaced StateMachineRAW implement the shared StateMachineCR
// interface declared in the shared CRUD package.
type StateMachineRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider clusternative.StateMachineRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider clusternative.StateMachineRAWInitParameters `json:"initProvider,omitempty"`
}

// StateMachineRAWStatus defines the observed state of StateMachineRAW.
type StateMachineRAWStatus struct {
	xpv1.ConditionedStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.StateMachineRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// StateMachineRAW is the native (non-Terraform) Schema for AWS Step Functions
// State Machines API (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type StateMachineRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.definition) || (has(self.initProvider) && has(self.initProvider.definition))",message="spec.forProvider.definition is a required parameter"
	Spec   StateMachineRAWSpec   `json:"spec"`
	Status StateMachineRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StateMachineRAWList contains a list of StateMachineRAW resources.
type StateMachineRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StateMachineRAW `json:"items"`
}

// Repository type metadata for StateMachineRAW.
var (
	StateMachineRAW_Kind             = "StateMachineRAW"
	StateMachineRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: StateMachineRAW_Kind}.String()
	StateMachineRAW_KindAPIVersion   = StateMachineRAW_Kind + "." + CRDGroupVersion.String()
	StateMachineRAW_GroupVersionKind = CRDGroupVersion.WithKind(StateMachineRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&StateMachineRAW{}, &StateMachineRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (sm *StateMachineRAW) GetForProvider() *clusternative.StateMachineRAWParameters {
	return &sm.Spec.ForProvider
}

// GetInitProvider returns the InitProvider parameters.
func (sm *StateMachineRAW) GetInitProvider() *clusternative.StateMachineRAWInitParameters {
	return &sm.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (sm *StateMachineRAW) GetAtProvider() clusternative.StateMachineRAWObservation {
	return sm.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (sm *StateMachineRAW) SetAtProvider(o clusternative.StateMachineRAWObservation) {
	sm.Status.AtProvider = o
}
