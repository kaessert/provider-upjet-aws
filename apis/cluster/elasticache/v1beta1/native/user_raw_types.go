// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	v1beta2native "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native"
)

// AuthenticationModeRAWParameters defines authentication mode parameters for UserRAW v1beta1.
// Note: v1beta1 uses a slice (array) while v1beta2 uses a pointer (object).
type AuthenticationModeRAWParameters struct {
	// Specifies the passwords to use for authentication if type is set to password.
	// +kubebuilder:validation:Optional
	PasswordsSecretRef *[]xpv1.SecretKeySelector `json:"passwordsSecretRef,omitempty"`

	// Specifies the authentication type. Possible options: password, no-password-required, iam.
	// +kubebuilder:validation:Optional
	Type *string `json:"type"`
}

// AuthenticationModeRAWInitParameters defines init parameters for authentication mode (v1beta1 spoke).
type AuthenticationModeRAWInitParameters struct {
	// Specifies the authentication type.
	// +kubebuilder:validation:Optional
	Type *string `json:"type,omitempty"`
}

// AuthenticationModeRAWObservation defines observed authentication mode state (v1beta1 spoke).
type AuthenticationModeRAWObservation struct {
	PasswordCount *float64 `json:"passwordCount,omitempty"`
	Type          *string  `json:"type,omitempty"`
}

// UserRAWParameters defines the v1beta1 spoke parameters for UserRAW.
type UserRAWParameters struct {
	// Access permissions string used for this user.
	// +kubebuilder:validation:Optional
	AccessString *string `json:"accessString,omitempty"`

	// Denotes the user's authentication properties.
	// Note: v1beta1 uses a slice (array) while v1beta2 uses a pointer (object).
	// +kubebuilder:validation:Optional
	AuthenticationMode []AuthenticationModeRAWParameters `json:"authenticationMode,omitempty"`

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

// UserRAWInitParameters defines the v1beta1 spoke init parameters.
type UserRAWInitParameters struct {
	AccessString       *string                               `json:"accessString,omitempty"`
	AuthenticationMode []AuthenticationModeRAWInitParameters `json:"authenticationMode,omitempty"`
	Engine             *string                               `json:"engine,omitempty"`
	NoPasswordRequired *bool                                 `json:"noPasswordRequired,omitempty"`
	Tags               map[string]*string                    `json:"tags,omitempty"`
	UserName           *string                               `json:"userName,omitempty"`
}

// UserRAWObservation defines the observed state for the v1beta1 spoke.
type UserRAWObservation struct {
	AccessString       *string                            `json:"accessString,omitempty"`
	Arn                *string                            `json:"arn,omitempty"`
	AuthenticationMode []AuthenticationModeRAWObservation `json:"authenticationMode,omitempty"`
	Engine             *string                            `json:"engine,omitempty"`
	ID                 *string                            `json:"id,omitempty"`
	NoPasswordRequired *bool                              `json:"noPasswordRequired,omitempty"`
	Status             *string                            `json:"status,omitempty"`
	Tags               map[string]*string                 `json:"tags,omitempty"`
	UserName           *string                            `json:"userName,omitempty"`
}

// UserRAWSpec defines the desired state of UserRAW v1beta1 spoke.
type UserRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider UserRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider UserRAWInitParameters `json:"initProvider,omitempty"`
}

// UserRAWStatus defines the observed state of UserRAW v1beta1 spoke.
type UserRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider UserRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// UserRAW is the v1beta1 spoke for the native ElastiCache User resource.
// The hub version is UserRAW in v1beta2. Use v1beta2 for new resources.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type UserRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   UserRAWSpec   `json:"spec"`
	Status UserRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// UserRAWList contains a list of UserRAW resources (v1beta1 spoke).
type UserRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserRAW `json:"items"`
}

// Repository type metadata for UserRAW (v1beta1 spoke).
var (
	UserRAW_Kind             = "UserRAW"
	UserRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: UserRAW_Kind}.String()
	UserRAW_KindAPIVersion   = UserRAW_Kind + "." + CRDGroupVersion.String()
	UserRAW_GroupVersionKind = CRDGroupVersion.WithKind(UserRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&UserRAW{}, &UserRAWList{})
}

// ConvertTo converts UserRAW v1beta1 to the hub UserRAW v1beta2.
// The key structural change: AuthenticationMode []slice → *pointer.
//
// Uses explicit field copies (no ujconversion.RoundTrip) because JSON
// round-tripping fails: v1beta1 marshals AuthenticationMode as an array "[...]"
// while v1beta2 expects an object "{...}".
func (src *UserRAW) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*v1beta2native.UserRAW)
	if !ok {
		return fmt.Errorf("expected *v1beta2native.UserRAW, got %T", dstRaw)
	}

	// Copy ObjectMeta and TypeMeta.
	dst.ObjectMeta = src.ObjectMeta
	dst.TypeMeta = metav1.TypeMeta{
		APIVersion: v1beta2native.CRDGroup + "/" + v1beta2native.CRDVersion,
		Kind:       "UserRAW",
	}

	// Copy ManagedResourceSpec (all non-AuthenticationMode fields).
	dst.Spec.ManagedResourceSpec = src.Spec.ManagedResourceSpec

	// Copy ForProvider fields explicitly.
	dst.Spec.ForProvider.AccessString = src.Spec.ForProvider.AccessString
	dst.Spec.ForProvider.Engine = src.Spec.ForProvider.Engine
	dst.Spec.ForProvider.NoPasswordRequired = src.Spec.ForProvider.NoPasswordRequired
	dst.Spec.ForProvider.PasswordsSecretRef = src.Spec.ForProvider.PasswordsSecretRef
	dst.Spec.ForProvider.Region = src.Spec.ForProvider.Region
	dst.Spec.ForProvider.Tags = src.Spec.ForProvider.Tags
	dst.Spec.ForProvider.UserName = src.Spec.ForProvider.UserName

	// Convert AuthenticationMode: slice (v1beta1) → pointer (v1beta2).
	if len(src.Spec.ForProvider.AuthenticationMode) > 0 {
		am := src.Spec.ForProvider.AuthenticationMode[0]
		dst.Spec.ForProvider.AuthenticationMode = &v1beta2native.AuthenticationModeRAWParameters{
			PasswordsSecretRef: am.PasswordsSecretRef,
			Type:               am.Type,
		}
	} else {
		dst.Spec.ForProvider.AuthenticationMode = nil
	}

	// Copy InitProvider fields explicitly.
	dst.Spec.InitProvider.AccessString = src.Spec.InitProvider.AccessString
	dst.Spec.InitProvider.Engine = src.Spec.InitProvider.Engine
	dst.Spec.InitProvider.NoPasswordRequired = src.Spec.InitProvider.NoPasswordRequired
	dst.Spec.InitProvider.Tags = src.Spec.InitProvider.Tags
	dst.Spec.InitProvider.UserName = src.Spec.InitProvider.UserName
	if len(src.Spec.InitProvider.AuthenticationMode) > 0 {
		am := src.Spec.InitProvider.AuthenticationMode[0]
		dst.Spec.InitProvider.AuthenticationMode = &v1beta2native.AuthenticationModeRAWInitParameters{
			Type: am.Type,
		}
	} else {
		dst.Spec.InitProvider.AuthenticationMode = nil
	}

	// Copy Status.
	dst.Status.ResourceStatus = src.Status.ResourceStatus
	dst.Status.AtProvider.AccessString = src.Status.AtProvider.AccessString
	dst.Status.AtProvider.Arn = src.Status.AtProvider.Arn
	dst.Status.AtProvider.Engine = src.Status.AtProvider.Engine
	dst.Status.AtProvider.ID = src.Status.AtProvider.ID
	dst.Status.AtProvider.NoPasswordRequired = src.Status.AtProvider.NoPasswordRequired
	dst.Status.AtProvider.Status = src.Status.AtProvider.Status
	dst.Status.AtProvider.Tags = src.Status.AtProvider.Tags
	dst.Status.AtProvider.UserName = src.Status.AtProvider.UserName
	if len(src.Status.AtProvider.AuthenticationMode) > 0 {
		am := src.Status.AtProvider.AuthenticationMode[0]
		dst.Status.AtProvider.AuthenticationMode = &v1beta2native.AuthenticationModeRAWObservation{
			PasswordCount: am.PasswordCount,
			Type:          am.Type,
		}
	} else {
		dst.Status.AtProvider.AuthenticationMode = nil
	}

	return nil
}

// ConvertFrom converts from the hub UserRAW v1beta2 to UserRAW v1beta1.
//
// Uses explicit field copies (no ujconversion.RoundTrip) because JSON
// round-tripping fails: v1beta2 marshals AuthenticationMode as an object "{...}"
// while v1beta1 expects an array "[...]".
func (dst *UserRAW) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*v1beta2native.UserRAW)
	if !ok {
		return fmt.Errorf("expected *v1beta2native.UserRAW, got %T", srcRaw)
	}

	// Copy ObjectMeta and TypeMeta.
	dst.ObjectMeta = src.ObjectMeta
	dst.TypeMeta = metav1.TypeMeta{
		APIVersion: CRDGroup + "/" + CRDVersion,
		Kind:       "UserRAW",
	}

	// Copy ManagedResourceSpec.
	dst.Spec.ManagedResourceSpec = src.Spec.ManagedResourceSpec

	// Copy ForProvider fields explicitly.
	dst.Spec.ForProvider.AccessString = src.Spec.ForProvider.AccessString
	dst.Spec.ForProvider.Engine = src.Spec.ForProvider.Engine
	dst.Spec.ForProvider.NoPasswordRequired = src.Spec.ForProvider.NoPasswordRequired
	dst.Spec.ForProvider.PasswordsSecretRef = src.Spec.ForProvider.PasswordsSecretRef
	dst.Spec.ForProvider.Region = src.Spec.ForProvider.Region
	dst.Spec.ForProvider.Tags = src.Spec.ForProvider.Tags
	dst.Spec.ForProvider.UserName = src.Spec.ForProvider.UserName

	// Convert AuthenticationMode: pointer (v1beta2) → slice (v1beta1).
	if src.Spec.ForProvider.AuthenticationMode != nil {
		am := src.Spec.ForProvider.AuthenticationMode
		dst.Spec.ForProvider.AuthenticationMode = []AuthenticationModeRAWParameters{
			{
				PasswordsSecretRef: am.PasswordsSecretRef,
				Type:               am.Type,
			},
		}
	} else {
		dst.Spec.ForProvider.AuthenticationMode = nil
	}

	// Copy InitProvider fields explicitly.
	dst.Spec.InitProvider.AccessString = src.Spec.InitProvider.AccessString
	dst.Spec.InitProvider.Engine = src.Spec.InitProvider.Engine
	dst.Spec.InitProvider.NoPasswordRequired = src.Spec.InitProvider.NoPasswordRequired
	dst.Spec.InitProvider.Tags = src.Spec.InitProvider.Tags
	dst.Spec.InitProvider.UserName = src.Spec.InitProvider.UserName
	if src.Spec.InitProvider.AuthenticationMode != nil {
		am := src.Spec.InitProvider.AuthenticationMode
		dst.Spec.InitProvider.AuthenticationMode = []AuthenticationModeRAWInitParameters{
			{Type: am.Type},
		}
	} else {
		dst.Spec.InitProvider.AuthenticationMode = nil
	}

	// Copy Status.
	dst.Status.ResourceStatus = src.Status.ResourceStatus
	dst.Status.AtProvider.AccessString = src.Status.AtProvider.AccessString
	dst.Status.AtProvider.Arn = src.Status.AtProvider.Arn
	dst.Status.AtProvider.Engine = src.Status.AtProvider.Engine
	dst.Status.AtProvider.ID = src.Status.AtProvider.ID
	dst.Status.AtProvider.NoPasswordRequired = src.Status.AtProvider.NoPasswordRequired
	dst.Status.AtProvider.Status = src.Status.AtProvider.Status
	dst.Status.AtProvider.Tags = src.Status.AtProvider.Tags
	dst.Status.AtProvider.UserName = src.Status.AtProvider.UserName
	if src.Status.AtProvider.AuthenticationMode != nil {
		am := src.Status.AtProvider.AuthenticationMode
		dst.Status.AtProvider.AuthenticationMode = []AuthenticationModeRAWObservation{
			{
				PasswordCount: am.PasswordCount,
				Type:          am.Type,
			},
		}
	} else {
		dst.Status.AtProvider.AuthenticationMode = nil
	}

	return nil
}

// GetForProvider returns the ForProvider parameters.
func (u *UserRAW) GetForProvider() *UserRAWParameters { return &u.Spec.ForProvider }

// SetForProvider writes back the ForProvider parameters.
func (u *UserRAW) SetForProvider(p UserRAWParameters) { u.Spec.ForProvider = p }

// GetInitProvider returns the InitProvider parameters.
func (u *UserRAW) GetInitProvider() *UserRAWInitParameters { return &u.Spec.InitProvider }

// GetAtProvider returns the current observed state.
func (u *UserRAW) GetAtProvider() UserRAWObservation { return u.Status.AtProvider }

// SetAtProvider sets the observed state.
func (u *UserRAW) SetAtProvider(o UserRAWObservation) { u.Status.AtProvider = o }
