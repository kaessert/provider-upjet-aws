// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	namespacedv2native "github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta2/native"
)

// AuthenticationModeRAWParameters defines authentication mode parameters for namespaced UserRAW v1beta1.
// Note: v1beta1 uses a slice (array) while v1beta2 uses a pointer (object).
type AuthenticationModeRAWParameters struct {
	// Specifies the passwords to use for authentication if type is set to password.
	// +kubebuilder:validation:Optional
	PasswordsSecretRef *[]xpv1.SecretKeySelector `json:"passwordsSecretRef,omitempty"`

	// Specifies the authentication type. Possible options: password, no-password-required, iam.
	// +kubebuilder:validation:Optional
	Type *string `json:"type"`
}

// AuthenticationModeRAWInitParameters defines init parameters for authentication mode (namespaced v1beta1 spoke).
type AuthenticationModeRAWInitParameters struct {
	// Specifies the authentication type.
	// +kubebuilder:validation:Optional
	Type *string `json:"type,omitempty"`
}

// AuthenticationModeRAWObservation defines observed authentication mode state (namespaced v1beta1 spoke).
type AuthenticationModeRAWObservation struct {
	PasswordCount *float64 `json:"passwordCount,omitempty"`
	Type          *string  `json:"type,omitempty"`
}

// UserRAWParameters defines the namespaced v1beta1 spoke parameters for UserRAW.
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

// UserRAWInitParameters defines the namespaced v1beta1 spoke init parameters.
type UserRAWInitParameters struct {
	AccessString       *string                               `json:"accessString,omitempty"`
	AuthenticationMode []AuthenticationModeRAWInitParameters `json:"authenticationMode,omitempty"`
	Engine             *string                               `json:"engine,omitempty"`
	NoPasswordRequired *bool                                 `json:"noPasswordRequired,omitempty"`
	Tags               map[string]*string                    `json:"tags,omitempty"`
	UserName           *string                               `json:"userName,omitempty"`
}

// UserRAWObservation defines the observed state for the namespaced v1beta1 spoke.
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

// UserRAWSpec defines the desired state of namespaced UserRAW v1beta1 spoke.
type UserRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider UserRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider UserRAWInitParameters `json:"initProvider,omitempty"`
}

// UserRAWStatus defines the observed state of namespaced UserRAW v1beta1 spoke.
type UserRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider UserRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// UserRAW is the namespaced v1beta1 spoke for the native ElastiCache User resource.
// The hub version is UserRAW in namespaced v1beta2. Use v1beta2 for new resources.
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

// UserRAWList contains a list of UserRAW resources (namespaced v1beta1 spoke).
type UserRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []UserRAW `json:"items"`
}

// Repository type metadata for namespaced UserRAW v1beta1.
var (
	UserRAW_Kind             = "UserRAW"
	UserRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: UserRAW_Kind}.String()
	UserRAW_KindAPIVersion   = UserRAW_Kind + "." + CRDGroupVersion.String()
	UserRAW_GroupVersionKind = CRDGroupVersion.WithKind(UserRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&UserRAW{}, &UserRAWList{})
}

// ConvertTo converts namespaced UserRAW v1beta1 to the namespaced hub UserRAW v1beta2.
// The key structural change: AuthenticationMode []slice → *pointer.
func (src *UserRAW) ConvertTo(dstRaw conversion.Hub) error {
	dst, ok := dstRaw.(*namespacedv2native.UserRAW)
	if !ok {
		return fmt.Errorf("expected *namespacedv2native.UserRAW, got %T", dstRaw)
	}
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return err
	}
	// Convert AuthenticationMode from slice to pointer.
	// v1beta1 has []AuthenticationModeRAWParameters, v1beta2 has *AuthenticationModeRAWParameters.
	if len(src.Spec.ForProvider.AuthenticationMode) > 0 {
		am := src.Spec.ForProvider.AuthenticationMode[0]
		dst.Spec.ForProvider.AuthenticationMode = &namespacedv2native.AuthenticationModeRAWParameters{
			PasswordsSecretRef: am.PasswordsSecretRef,
			Type:               am.Type,
		}
	} else {
		dst.Spec.ForProvider.AuthenticationMode = nil
	}
	// Explicitly set the hub's TypeMeta.
	dst.TypeMeta = metav1.TypeMeta{
		APIVersion: namespacedv2native.CRDGroup + "/" + namespacedv2native.CRDVersion,
		Kind:       "UserRAW",
	}
	return nil
}

// ConvertFrom converts from the namespaced hub UserRAW v1beta2 to namespaced UserRAW v1beta1.
func (dst *UserRAW) ConvertFrom(srcRaw conversion.Hub) error {
	src, ok := srcRaw.(*namespacedv2native.UserRAW)
	if !ok {
		return fmt.Errorf("expected *namespacedv2native.UserRAW, got %T", srcRaw)
	}
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return err
	}
	// Convert AuthenticationMode from pointer to slice.
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
	// Explicitly set the spoke's TypeMeta to v1beta1.
	dst.TypeMeta = metav1.TypeMeta{
		APIVersion: CRDGroup + "/" + CRDVersion,
		Kind:       "UserRAW",
	}
	return nil
}

// GetForProvider returns the ForProvider parameters.
func (u *UserRAW) GetForProvider() *UserRAWParameters { return &u.Spec.ForProvider }

// GetInitProvider returns the InitProvider parameters.
func (u *UserRAW) GetInitProvider() *UserRAWInitParameters { return &u.Spec.InitProvider }

// GetAtProvider returns the current observed state.
func (u *UserRAW) GetAtProvider() UserRAWObservation { return u.Status.AtProvider }

// SetAtProvider sets the observed state.
func (u *UserRAW) SetAtProvider(o UserRAWObservation) { u.Status.AtProvider = o }
