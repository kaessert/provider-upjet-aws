// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native"
)

// UserGroupRAWParameters defines the namespaced configuration parameters for a native ElastiCache User Group.
// Fields are identical to the cluster-scoped type; defined locally so that angryjet
// generates namespaced resolvers.
type UserGroupRAWParameters struct {
	// The current supported values are redis, valkey (case insensitive).
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// References to UserRAW in elasticache to populate userIds.
	// +kubebuilder:validation:Optional
	UserIDRefs []xpv1.NamespacedReference `json:"userIdRefs,omitempty"`

	// Selector for a list of UserRAW in elasticache to populate userIds.
	// +kubebuilder:validation:Optional
	UserIDSelector *xpv1.NamespacedSelector `json:"userIdSelector,omitempty"`

	// The list of user IDs that belong to the user group.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta2/native.UserRAW
	// +crossplane:generate:reference:refFieldName=UserIDRefs
	// +crossplane:generate:reference:selectorFieldName=UserIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	UserIds []*string `json:"userIds,omitempty"`
}

// UserGroupRAWInitParameters defines the namespaced init parameters for UserGroupRAW.
type UserGroupRAWInitParameters struct {
	// The current supported values are redis, valkey (case insensitive).
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// References to UserRAW in elasticache to populate userIds.
	// +kubebuilder:validation:Optional
	UserIDRefs []xpv1.NamespacedReference `json:"userIdRefs,omitempty"`

	// Selector for a list of UserRAW in elasticache to populate userIds.
	// +kubebuilder:validation:Optional
	UserIDSelector *xpv1.NamespacedSelector `json:"userIdSelector,omitempty"`

	// The list of user IDs that belong to the user group.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta2/native.UserRAW
	// +crossplane:generate:reference:refFieldName=UserIDRefs
	// +crossplane:generate:reference:selectorFieldName=UserIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	UserIds []*string `json:"userIds,omitempty"`
}

// UserGroupRAWSpec defines the desired state of namespaced UserGroupRAW.
type UserGroupRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider UserGroupRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider UserGroupRAWInitParameters `json:"initProvider,omitempty"`
}

// UserGroupRAWStatus defines the observed state of namespaced UserGroupRAW.
type UserGroupRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusternative.UserGroupRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// UserGroupRAW is the native (non-Terraform) Schema for AWS ElastiCache User Group (namespaced scope).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type UserGroupRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserGroupRAWSpec   `json:"spec"`
	Status UserGroupRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserGroupRAWList contains a list of UserGroupRAW resources (namespaced scope).
type UserGroupRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserGroupRAW `json:"items"`
}

// Repository type metadata for namespaced UserGroupRAW.
var (
	UserGroupRAW_Kind             = "UserGroupRAW"
	UserGroupRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: UserGroupRAW_Kind}.String()
	UserGroupRAW_KindAPIVersion   = UserGroupRAW_Kind + "." + CRDGroupVersion.String()
	UserGroupRAW_GroupVersionKind = CRDGroupVersion.WithKind(UserGroupRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&UserGroupRAW{}, &UserGroupRAWList{})
}

// GetForProvider returns a cluster-scoped UserGroupRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (u *UserGroupRAW) GetForProvider() *clusternative.UserGroupRAWParameters {
	return &clusternative.UserGroupRAWParameters{
		Engine:  u.Spec.ForProvider.Engine,
		Region:  u.Spec.ForProvider.Region,
		Tags:    u.Spec.ForProvider.Tags,
		UserIds: u.Spec.ForProvider.UserIds,
		// Note: UserIDRefs and UserIDSelector use NamespacedReference in namespaced scope.
		// We don't copy reference fields — they're used for resolution but not passed to AWS API.
	}
}

// GetInitProvider returns a cluster-scoped UserGroupRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (u *UserGroupRAW) GetInitProvider() *clusternative.UserGroupRAWInitParameters {
	return &clusternative.UserGroupRAWInitParameters{
		Engine:  u.Spec.InitProvider.Engine,
		Tags:    u.Spec.InitProvider.Tags,
		UserIds: u.Spec.InitProvider.UserIds,
	}
}

// GetAtProvider returns the current observed state.
func (u *UserGroupRAW) GetAtProvider() clusternative.UserGroupRAWObservation {
	return u.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (u *UserGroupRAW) SetAtProvider(o clusternative.UserGroupRAWObservation) {
	u.Status.AtProvider = o
}
