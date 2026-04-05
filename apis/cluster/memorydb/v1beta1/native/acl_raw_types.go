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

// ACLRAWParameters defines the configuration parameters for a native MemoryDB ACL.
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

// ACLRAWInitParameters defines the init parameters for ACLRAW.
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

// ACLRAWObservation defines the observed state of ACLRAW.
type ACLRAWObservation struct {
	// The ARN of the ACL.
	Arn *string `json:"arn,omitempty"`

	// Same as name.
	ID *string `json:"id,omitempty"`

	// The minimum engine version supported by the ACL.
	MinimumEngineVersion *string `json:"minimumEngineVersion,omitempty"`

	// A map of tags assigned to the resource, including those inherited from the
	// provider default_tags configuration block.
	// +mapType=granular
	TagsAll map[string]*string `json:"tagsAll,omitempty"`
}

// ACLRAWSpec defines the desired state of ACLRAW.
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

// ACLRAWStatus defines the observed state of ACLRAW.
type ACLRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider ACLRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// ACLRAW is the native (non-Terraform) Schema for AWS MemoryDB ACL.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type ACLRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ACLRAWSpec   `json:"spec"`
	Status ACLRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ACLRAWList contains a list of ACLRAW resources.
type ACLRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ACLRAW `json:"items"`
}

// Repository type metadata for ACLRAW.
var (
	ACLRAW_Kind             = "ACLRAW"
	ACLRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: ACLRAW_Kind}.String()
	ACLRAW_KindAPIVersion   = ACLRAW_Kind + "." + CRDGroupVersion.String()
	ACLRAW_GroupVersionKind = CRDGroupVersion.WithKind(ACLRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&ACLRAW{}, &ACLRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (a *ACLRAW) GetForProvider() *ACLRAWParameters { return &a.Spec.ForProvider }

// SetForProvider writes back the ForProvider parameters.
func (a *ACLRAW) SetForProvider(p ACLRAWParameters) { a.Spec.ForProvider = p }

// GetInitProvider returns the InitProvider parameters.
func (a *ACLRAW) GetInitProvider() *ACLRAWInitParameters { return &a.Spec.InitProvider }

// GetAtProvider returns the current observed state.
func (a *ACLRAW) GetAtProvider() ACLRAWObservation { return a.Status.AtProvider }

// SetAtProvider sets the observed state.
func (a *ACLRAW) SetAtProvider(o ACLRAWObservation) { a.Status.AtProvider = o }
