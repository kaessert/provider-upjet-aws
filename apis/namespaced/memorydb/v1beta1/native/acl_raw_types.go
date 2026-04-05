// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/memorydb/v1beta1/native"
)

// ACLRAWParameters defines the namespaced configuration parameters for a native MemoryDB ACL.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type ACLRAWParameters struct {
	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// Set of MemoryDB user names to be included in this ACL.
	// +kubebuilder:validation:Optional
	// +listType=set
	UserNames []*string `json:"userNames,omitempty"`
}

// ACLRAWInitParameters defines the namespaced init parameters for ACLRAW.
type ACLRAWInitParameters struct {
	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// Set of MemoryDB user names to be included in this ACL.
	// +kubebuilder:validation:Optional
	// +listType=set
	UserNames []*string `json:"userNames,omitempty"`
}

// ACLRAWSpec defines the desired state of namespaced ACLRAW.
type ACLRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider ACLRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider ACLRAWInitParameters `json:"initProvider,omitempty"`
}

// ACLRAWStatus defines the observed state of namespaced ACLRAW.
type ACLRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.ACLRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// ACLRAW is the native (non-Terraform) Schema for AWS MemoryDB ACL (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type ACLRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ACLRAWSpec   `json:"spec"`
	Status ACLRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ACLRAWList contains a list of ACLRAW resources (namespaced scope).
type ACLRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ACLRAW `json:"items"`
}

// Repository type metadata for namespaced ACLRAW.
var (
	ACLRAW_Kind             = "ACLRAW"
	ACLRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ACLRAW_Kind}.String()
	ACLRAW_KindAPIVersion   = ACLRAW_Kind + "." + CRDGroupVersion.String()
	ACLRAW_GroupVersionKind = CRDGroupVersion.WithKind(ACLRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&ACLRAW{}, &ACLRAWList{})
}

// GetForProvider returns a cluster-scoped ACLRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (a *ACLRAW) GetForProvider() *clusternative.ACLRAWParameters {
	return &clusternative.ACLRAWParameters{
		Region:    a.Spec.ForProvider.Region,
		Tags:      a.Spec.ForProvider.Tags,
		UserNames: a.Spec.ForProvider.UserNames,
	}
}

// SetForProvider writes back a cluster-scoped ACLRAWParameters to this
// namespaced resource's ForProvider fields. This is the symmetric inverse of
// GetForProvider, required so that late-initialized fields are persisted.
func (a *ACLRAW) SetForProvider(p clusternative.ACLRAWParameters) {
	a.Spec.ForProvider.Region = p.Region
	a.Spec.ForProvider.Tags = p.Tags
	a.Spec.ForProvider.UserNames = p.UserNames
}

// GetInitProvider returns a cluster-scoped ACLRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (a *ACLRAW) GetInitProvider() *clusternative.ACLRAWInitParameters {
	return &clusternative.ACLRAWInitParameters{
		Tags:      a.Spec.InitProvider.Tags,
		UserNames: a.Spec.InitProvider.UserNames,
	}
}

// GetAtProvider returns the current observed state.
func (a *ACLRAW) GetAtProvider() clusternative.ACLRAWObservation {
	return a.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (a *ACLRAW) SetAtProvider(o clusternative.ACLRAWObservation) {
	a.Status.AtProvider = o
}
