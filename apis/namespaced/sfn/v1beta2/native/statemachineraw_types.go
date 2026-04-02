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

// EncryptionConfigurationRAWParameters defines the encryption configuration for
// a namespaced StateMachine. Fields are identical to the cluster-scoped type;
// only the +crossplane:generate:reference:type annotations differ (pointing to
// namespaced KMS Key instead of cluster KMS Key).
type EncryptionConfigurationRAWParameters struct {
	// Maximum duration for which Step Functions will reuse data keys. When the
	// period expires, Step Functions will call GenerateDataKey. This setting only
	// applies to customer managed KMS key and does not apply when type is
	// AWS_OWNED_KEY.
	// +kubebuilder:validation:Optional
	KMSDataKeyReusePeriodSeconds *float64 `json:"kmsDataKeyReusePeriodSeconds,omitempty"`

	// The alias, alias ARN, key ID, or key ARN of the symmetric encryption KMS key
	// that encrypts the data key.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/kms/v1beta1.Key
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("arn")
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Reference to a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDRef *xpv1.NamespacedReference `json:"kmsKeyIdRef,omitempty"`

	// Selector for a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDSelector *xpv1.NamespacedSelector `json:"kmsKeyIdSelector,omitempty"`

	// The encryption option specified for the state machine.
	// Valid values: AWS_OWNED_KEY, CUSTOMER_MANAGED_KMS_KEY
	// +kubebuilder:validation:Optional
	Type *string `json:"type,omitempty"`
}

// EncryptionConfigurationRAWInitParameters defines the init parameters for
// encryption configuration in namespaced scope.
type EncryptionConfigurationRAWInitParameters struct {
	// Maximum duration for which Step Functions will reuse data keys.
	// +kubebuilder:validation:Optional
	KMSDataKeyReusePeriodSeconds *float64 `json:"kmsDataKeyReusePeriodSeconds,omitempty"`

	// The alias, alias ARN, key ID, or key ARN of the symmetric encryption KMS key.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/kms/v1beta1.Key
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("arn")
	// +kubebuilder:validation:Optional
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// Reference to a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDRef *xpv1.NamespacedReference `json:"kmsKeyIdRef,omitempty"`

	// Selector for a Key in kms to populate kmsKeyId.
	// +kubebuilder:validation:Optional
	KMSKeyIDSelector *xpv1.NamespacedSelector `json:"kmsKeyIdSelector,omitempty"`

	// The encryption option specified for the state machine.
	// +kubebuilder:validation:Optional
	Type *string `json:"type,omitempty"`
}

// StateMachineRAWParameters defines the configuration parameters for a native
// StateMachine in namespaced scope. Fields are identical to the cluster-scoped
// type; only the +crossplane:generate:reference:type annotations differ
// (pointing to namespaced IAM Role / KMS Key instead of cluster-scoped types).
// +kubebuilder:object:generate=true
type StateMachineRAWParameters struct {
	// The Amazon States Language definition of the state machine.
	// +kubebuilder:validation:Optional
	Definition *string `json:"definition,omitempty"`

	// Defines what encryption configuration is used to encrypt data in the State Machine.
	// +kubebuilder:validation:Optional
	EncryptionConfiguration *EncryptionConfigurationRAWParameters `json:"encryptionConfiguration,omitempty"`

	// Defines what execution history events are logged and where they are logged.
	// +kubebuilder:validation:Optional
	LoggingConfiguration *clusternative.LoggingConfigurationRAWParameters `json:"loggingConfiguration,omitempty"`

	// Set to true to publish a version of the state machine during creation.
	// +kubebuilder:validation:Optional
	Publish *bool `json:"publish,omitempty"`

	// Region where this resource will be managed. Required for credential resolution.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// The Amazon Resource Name (ARN) of the IAM role to use for this state machine.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	// +kubebuilder:validation:Optional
	RoleArn *string `json:"roleArn,omitempty"`

	// Reference to a Role in iam to populate roleArn.
	// +kubebuilder:validation:Optional
	RoleArnRef *xpv1.NamespacedReference `json:"roleArnRef,omitempty"`

	// Selector for a Role in iam to populate roleArn.
	// +kubebuilder:validation:Optional
	RoleArnSelector *xpv1.NamespacedSelector `json:"roleArnSelector,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// Selects whether AWS X-Ray tracing is enabled.
	// +kubebuilder:validation:Optional
	TracingConfiguration *clusternative.TracingConfigurationRAWParameters `json:"tracingConfiguration,omitempty"`

	// Determines whether a Standard or Express state machine is created.
	// Valid values: STANDARD, EXPRESS.
	// +kubebuilder:validation:Optional
	Type *string `json:"type,omitempty"`
}

// StateMachineRAWInitParameters defines the init parameters for a native
// StateMachine in namespaced scope.
// +kubebuilder:object:generate=true
type StateMachineRAWInitParameters struct {
	// The Amazon States Language definition of the state machine.
	// +kubebuilder:validation:Optional
	Definition *string `json:"definition,omitempty"`

	// Defines what encryption configuration is used to encrypt data in the State Machine.
	// +kubebuilder:validation:Optional
	EncryptionConfiguration *EncryptionConfigurationRAWInitParameters `json:"encryptionConfiguration,omitempty"`

	// Defines what execution history events are logged and where they are logged.
	// +kubebuilder:validation:Optional
	LoggingConfiguration *clusternative.LoggingConfigurationRAWInitParameters `json:"loggingConfiguration,omitempty"`

	// Set to true to publish a version of the state machine during creation.
	// +kubebuilder:validation:Optional
	Publish *bool `json:"publish,omitempty"`

	// The Amazon Resource Name (ARN) of the IAM role to use for this state machine.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/iam/v1beta1.Role
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/config/cluster/common.ARNExtractor()
	// +kubebuilder:validation:Optional
	RoleArn *string `json:"roleArn,omitempty"`

	// Reference to a Role in iam to populate roleArn.
	// +kubebuilder:validation:Optional
	RoleArnRef *xpv1.NamespacedReference `json:"roleArnRef,omitempty"`

	// Selector for a Role in iam to populate roleArn.
	// +kubebuilder:validation:Optional
	RoleArnSelector *xpv1.NamespacedSelector `json:"roleArnSelector,omitempty"`

	// Key-value map of resource tags.
	// +kubebuilder:validation:Optional
	// +mapType=granular
	Tags map[string]*string `json:"tags,omitempty"`

	// Selects whether AWS X-Ray tracing is enabled.
	// +kubebuilder:validation:Optional
	TracingConfiguration *clusternative.TracingConfigurationRAWInitParameters `json:"tracingConfiguration,omitempty"`

	// Determines whether a Standard or Express state machine is created.
	// +kubebuilder:validation:Optional
	Type *string `json:"type,omitempty"`
}

// StateMachineRAWSpec defines the desired state of StateMachineRAW (namespaced scope).
type StateMachineRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider StateMachineRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider StateMachineRAWInitParameters `json:"initProvider,omitempty"`
}

// StateMachineRAWStatus defines the observed state of StateMachineRAW.
// Note: using xpv1.ResourceStatus (not ConditionedStatus) so that angryjet's
// ManagedV2() matcher recognizes this as a v2-style managed resource and
// generates the zz_generated.resolvers.go file.
type StateMachineRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

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

// GetForProvider returns a cluster-scoped StateMachineRAWParameters populated
// from this namespaced resource's ForProvider fields.
//
// A field-by-field copy is necessary because the namespaced package defines its
// own StateMachineRAWParameters struct (with scope-correct reference annotations
// for angryjet resolver generation) while the shared CRUD interface expects the
// cluster-scoped parameter type. Fields are value-copied; the returned pointer
// is a new allocation and mutations to it do NOT propagate back to the spec.
// For late initialization use SetForProviderType instead of mutating via this
// method.
func (sm *StateMachineRAW) GetForProvider() *clusternative.StateMachineRAWParameters {
	p := &clusternative.StateMachineRAWParameters{
		Definition:           sm.Spec.ForProvider.Definition,
		LoggingConfiguration: sm.Spec.ForProvider.LoggingConfiguration,
		Publish:              sm.Spec.ForProvider.Publish,
		Region:               sm.Spec.ForProvider.Region,
		RoleArn:              sm.Spec.ForProvider.RoleArn,
		Tags:                 sm.Spec.ForProvider.Tags,
		TracingConfiguration: sm.Spec.ForProvider.TracingConfiguration,
		Type:                 sm.Spec.ForProvider.Type,
	}
	if enc := sm.Spec.ForProvider.EncryptionConfiguration; enc != nil {
		p.EncryptionConfiguration = &clusternative.EncryptionConfigurationRAWParameters{
			KMSDataKeyReusePeriodSeconds: enc.KMSDataKeyReusePeriodSeconds,
			KMSKeyID:                     enc.KMSKeyID,
			Type:                         enc.Type,
		}
	}
	return p
}

// GetInitProvider returns a cluster-scoped StateMachineRAWInitParameters
// populated from this namespaced resource's InitProvider fields.
// See GetForProvider for rationale on field-by-field copy.
func (sm *StateMachineRAW) GetInitProvider() *clusternative.StateMachineRAWInitParameters {
	p := &clusternative.StateMachineRAWInitParameters{
		Definition:           sm.Spec.InitProvider.Definition,
		LoggingConfiguration: sm.Spec.InitProvider.LoggingConfiguration,
		Publish:              sm.Spec.InitProvider.Publish,
		RoleArn:              sm.Spec.InitProvider.RoleArn,
		Tags:                 sm.Spec.InitProvider.Tags,
		TracingConfiguration: sm.Spec.InitProvider.TracingConfiguration,
		Type:                 sm.Spec.InitProvider.Type,
	}
	if enc := sm.Spec.InitProvider.EncryptionConfiguration; enc != nil {
		p.EncryptionConfiguration = &clusternative.EncryptionConfigurationRAWInitParameters{
			KMSDataKeyReusePeriodSeconds: enc.KMSDataKeyReusePeriodSeconds,
			KMSKeyID:                     enc.KMSKeyID,
			Type:                         enc.Type,
		}
	}
	return p
}

// GetAtProvider returns the current observed state.
func (sm *StateMachineRAW) GetAtProvider() clusternative.StateMachineRAWObservation {
	return sm.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (sm *StateMachineRAW) SetAtProvider(o clusternative.StateMachineRAWObservation) {
	sm.Status.AtProvider = o
}

// SetForProviderType sets spec.forProvider.type. This is called by the shared
// CRUD late-initialization logic, which cannot mutate the spec through
// GetForProvider() (since that returns a field-copied struct, not a direct
// pointer to spec.forProvider). Implementing this method here ensures that
// late-initialized values are persisted in the namespaced resource's spec.
func (sm *StateMachineRAW) SetForProviderType(t *string) {
	sm.Spec.ForProvider.Type = t
}
