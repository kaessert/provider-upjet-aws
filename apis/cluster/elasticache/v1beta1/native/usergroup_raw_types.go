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

// UserGroupRAWParameters defines the configuration parameters for a native ElastiCache User Group.
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native.UserRAW
	// +crossplane:generate:reference:refFieldName=UserIDRefs
	// +crossplane:generate:reference:selectorFieldName=UserIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	UserIds []*string `json:"userIds,omitempty"`
}

// UserGroupRAWInitParameters defines the init parameters for UserGroupRAW.
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
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native.UserRAW
	// +crossplane:generate:reference:refFieldName=UserIDRefs
	// +crossplane:generate:reference:selectorFieldName=UserIDSelector
	// +kubebuilder:validation:Optional
	// +listType=set
	UserIds []*string `json:"userIds,omitempty"`
}

// UserGroupRAWObservation defines the observed state of UserGroupRAW.
type UserGroupRAWObservation struct {
	// The ARN that identifies the user group.
	Arn *string `json:"arn,omitempty"`

	// The current supported values are redis, valkey.
	Engine *string `json:"engine,omitempty"`

	// The user group identifier.
	ID *string `json:"id,omitempty"`

	// The status of the user group.
	Status *string `json:"status,omitempty"`

	// Tags assigned to the resource.
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// The list of user IDs that belong to the user group.
	// +listType=set
	UserIds []*string `json:"userIds,omitempty"`
}

// UserGroupRAWSpec defines the desired state of UserGroupRAW.
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

// UserGroupRAWStatus defines the observed state of UserGroupRAW.
type UserGroupRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider UserGroupRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// UserGroupRAW is the native (non-Terraform) Schema for AWS ElastiCache User Group.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type UserGroupRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserGroupRAWSpec   `json:"spec"`
	Status UserGroupRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserGroupRAWList contains a list of UserGroupRAW resources.
type UserGroupRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserGroupRAW `json:"items"`
}

// Repository type metadata for UserGroupRAW.
var (
	UserGroupRAW_Kind             = "UserGroupRAW"
	UserGroupRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: UserGroupRAW_Kind}.String()
	UserGroupRAW_KindAPIVersion   = UserGroupRAW_Kind + "." + CRDGroupVersion.String()
	UserGroupRAW_GroupVersionKind = CRDGroupVersion.WithKind(UserGroupRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&UserGroupRAW{}, &UserGroupRAWList{})
}

// GetForProvider returns the ForProvider parameters.
func (u *UserGroupRAW) GetForProvider() *UserGroupRAWParameters { return &u.Spec.ForProvider }

// GetInitProvider returns the InitProvider parameters.
func (u *UserGroupRAW) GetInitProvider() *UserGroupRAWInitParameters { return &u.Spec.InitProvider }

// GetAtProvider returns the current observed state.
func (u *UserGroupRAW) GetAtProvider() UserGroupRAWObservation { return u.Status.AtProvider }

// SetAtProvider sets the observed state.
func (u *UserGroupRAW) SetAtProvider(o UserGroupRAWObservation) { u.Status.AtProvider = o }
