// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusterv2native "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native"
)

// AuthenticationModeRAWParameters defines the authentication mode parameters for namespaced UserRAW v1beta2.
// Note: v1beta2 uses a pointer (*) while v1beta1 uses a slice ([]).
type AuthenticationModeRAWParameters struct {
	// Specifies the passwords to use for authentication if type is set to password.
	// +kubebuilder:validation:Optional
	PasswordsSecretRef *[]xpv1.SecretKeySelector `json:"passwordsSecretRef,omitempty"`

	// Specifies the authentication type. Possible options: password, no-password-required, iam.
	// +kubebuilder:validation:Optional
	Type *string `json:"type"`
}

// AuthenticationModeRAWInitParameters defines the init parameters for authentication mode (namespaced v1beta2).
type AuthenticationModeRAWInitParameters struct {
	// Specifies the authentication type. Possible options: password, no-password-required, iam.
	// +kubebuilder:validation:Optional
	Type *string `json:"type,omitempty"`
}

// UserRAWParameters defines the namespaced configuration parameters for a native ElastiCache User (v1beta2 hub).
// Fields are identical to the cluster-scoped v1beta2 type; defined locally so that angryjet
// generates namespaced resolvers.
type UserRAWParameters struct {
	// Access permissions string used for this user.
	// +kubebuilder:validation:Optional
	AccessString *string `json:"accessString,omitempty"`

	// Denotes the user's authentication properties.
	// Note: v1beta2 uses a pointer (object) while v1beta1 uses a slice (array).
	// +kubebuilder:validation:Optional
	AuthenticationMode *AuthenticationModeRAWParameters `json:"authenticationMode,omitempty"`

	// The current supported values are redis, valkey (case insensitive).
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// Indicates a password is not required for this user.
	// +kubebuilder:validation:Optional
	NoPasswordRequired *bool `json:"noPasswordRequired,omitempty"`

	// Passwords used for this user (top-level path, legacy).
	// +kubebuilder:validation:Optional
	PasswordsSecretRef *[]xpv1.SecretKeySelector `json:"passwordsSecretRef,omitempty"`

	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// The username of the user.
	// +kubebuilder:validation:Optional
	UserName *string `json:"userName,omitempty"`
}

// UserRAWInitParameters defines the init parameters for namespaced UserRAW (v1beta2 hub).
type UserRAWInitParameters struct {
	// Access permissions string.
	// +kubebuilder:validation:Optional
	AccessString *string `json:"accessString,omitempty"`

	// Denotes the user's authentication properties.
	// +kubebuilder:validation:Optional
	AuthenticationMode *AuthenticationModeRAWInitParameters `json:"authenticationMode,omitempty"`

	// The current supported values are redis, valkey.
	// +kubebuilder:validation:Optional
	Engine *string `json:"engine,omitempty"`

	// Indicates a password is not required.
	// +kubebuilder:validation:Optional
	NoPasswordRequired *bool `json:"noPasswordRequired,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// The username of the user.
	// +kubebuilder:validation:Optional
	UserName *string `json:"userName,omitempty"`
}

// UserRAWSpec defines the desired state of namespaced UserRAW (v1beta2 hub).
type UserRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider UserRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider UserRAWInitParameters `json:"initProvider,omitempty"`
}

// UserRAWStatus defines the observed state of namespaced UserRAW (v1beta2 hub).
type UserRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider clusterv2native.UserRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// UserRAW is the native (non-Terraform) Schema for AWS ElastiCache User (namespaced, v1beta2 hub).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}
type UserRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserRAWSpec   `json:"spec"`
	Status UserRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserRAWList contains a list of UserRAW resources (namespaced, v1beta2).
type UserRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserRAW `json:"items"`
}

// Repository type metadata for namespaced UserRAW v1beta2.
var (
	UserRAW_Kind             = "UserRAW"
	UserRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: UserRAW_Kind}.String()
	UserRAW_KindAPIVersion   = UserRAW_Kind + "." + CRDGroupVersion.String()
	UserRAW_GroupVersionKind = CRDGroupVersion.WithKind(UserRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&UserRAW{}, &UserRAWList{})
}

// Hub marks UserRAW v1beta2 as the hub version for conversion.
func (*UserRAW) Hub() {}

// GetForProvider returns a cluster-scoped UserRAWParameters populated from
// this namespaced resource's ForProvider fields.
func (u *UserRAW) GetForProvider() *clusterv2native.UserRAWParameters {
	var authMode *clusterv2native.AuthenticationModeRAWParameters
	if u.Spec.ForProvider.AuthenticationMode != nil {
		authMode = &clusterv2native.AuthenticationModeRAWParameters{
			PasswordsSecretRef: u.Spec.ForProvider.AuthenticationMode.PasswordsSecretRef,
			Type:               u.Spec.ForProvider.AuthenticationMode.Type,
		}
	}
	return &clusterv2native.UserRAWParameters{
		AccessString:       u.Spec.ForProvider.AccessString,
		AuthenticationMode: authMode,
		Engine:             u.Spec.ForProvider.Engine,
		NoPasswordRequired: u.Spec.ForProvider.NoPasswordRequired,
		PasswordsSecretRef: u.Spec.ForProvider.PasswordsSecretRef,
		Region:             u.Spec.ForProvider.Region,
		Tags:               u.Spec.ForProvider.Tags,
		UserName:           u.Spec.ForProvider.UserName,
	}
}

// GetInitProvider returns a cluster-scoped UserRAWInitParameters populated from
// this namespaced resource's InitProvider fields.
func (u *UserRAW) GetInitProvider() *clusterv2native.UserRAWInitParameters {
	var authMode *clusterv2native.AuthenticationModeRAWInitParameters
	if u.Spec.InitProvider.AuthenticationMode != nil {
		authMode = &clusterv2native.AuthenticationModeRAWInitParameters{
			Type: u.Spec.InitProvider.AuthenticationMode.Type,
		}
	}
	return &clusterv2native.UserRAWInitParameters{
		AccessString:       u.Spec.InitProvider.AccessString,
		AuthenticationMode: authMode,
		Engine:             u.Spec.InitProvider.Engine,
		NoPasswordRequired: u.Spec.InitProvider.NoPasswordRequired,
		Tags:               u.Spec.InitProvider.Tags,
		UserName:           u.Spec.InitProvider.UserName,
	}
}

// GetAtProvider returns the current observed state.
func (u *UserRAW) GetAtProvider() clusterv2native.UserRAWObservation { return u.Status.AtProvider }

// SetAtProvider sets the observed state.
func (u *UserRAW) SetAtProvider(o clusterv2native.UserRAWObservation) { u.Status.AtProvider = o }
