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

// EncryptionConfigurationRAWParameters defines the encryption configuration for a StateMachine.
type EncryptionConfigurationRAWParameters struct {
	// Maximum duration for which Step Functions will reuse data keys. When the
	// period expires, Step Functions will call GenerateDataKey. This setting only
	// applies to customer managed KMS key and does not apply when type is
	// AWS_OWNED_KEY.
	// +kubebuilder:validation:Optional
	KMSDataKeyReusePeriodSeconds *float64 `json:"kmsDataKeyReusePeriodSeconds,omitempty"`

	// The alias, alias ARN, key ID, or key ARN of the symmetric encryption KMS key
	// that encrypts the data key.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kms/v1beta1.Key
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

// EncryptionConfigurationRAWInitParameters defines the init parameters for encryption configuration.
type EncryptionConfigurationRAWInitParameters struct {
	// Maximum duration for which Step Functions will reuse data keys.
	// +kubebuilder:validation:Optional
	KMSDataKeyReusePeriodSeconds *float64 `json:"kmsDataKeyReusePeriodSeconds,omitempty"`

	// The alias, alias ARN, key ID, or key ARN of the symmetric encryption KMS key.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/kms/v1beta1.Key
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

// EncryptionConfigurationRAWObservation defines the observed state for encryption configuration.
type EncryptionConfigurationRAWObservation struct {
	// Maximum duration for which Step Functions will reuse data keys.
	KMSDataKeyReusePeriodSeconds *float64 `json:"kmsDataKeyReusePeriodSeconds,omitempty"`

	// The alias, alias ARN, key ID, or key ARN of the symmetric encryption KMS key.
	KMSKeyID *string `json:"kmsKeyId,omitempty"`

	// The encryption option specified for the state machine.
	Type *string `json:"type,omitempty"`
}

// LoggingConfigurationRAWParameters defines the logging configuration for a StateMachine.
type LoggingConfigurationRAWParameters struct {
	// Determines whether execution data is included in your log.
	// +kubebuilder:validation:Optional
	IncludeExecutionData *bool `json:"includeExecutionData,omitempty"`

	// Defines which category of execution history events are logged.
	// Valid values: ALL, ERROR, FATAL, OFF
	// +kubebuilder:validation:Optional
	Level *string `json:"level,omitempty"`

	// Amazon Resource Name (ARN) of a CloudWatch log group.
	// +kubebuilder:validation:Optional
	LogDestination *string `json:"logDestination,omitempty"`
}

// LoggingConfigurationRAWInitParameters defines the init parameters for logging configuration.
type LoggingConfigurationRAWInitParameters struct {
	// Determines whether execution data is included in your log.
	// +kubebuilder:validation:Optional
	IncludeExecutionData *bool `json:"includeExecutionData,omitempty"`

	// Defines which category of execution history events are logged.
	// +kubebuilder:validation:Optional
	Level *string `json:"level,omitempty"`

	// Amazon Resource Name (ARN) of a CloudWatch log group.
	// +kubebuilder:validation:Optional
	LogDestination *string `json:"logDestination,omitempty"`
}

// LoggingConfigurationRAWObservation defines the observed state for logging configuration.
type LoggingConfigurationRAWObservation struct {
	// Determines whether execution data is included in your log.
	IncludeExecutionData *bool `json:"includeExecutionData,omitempty"`

	// Defines which category of execution history events are logged.
	Level *string `json:"level,omitempty"`

	// Amazon Resource Name (ARN) of a CloudWatch log group.
	LogDestination *string `json:"logDestination,omitempty"`
}

// TracingConfigurationRAWParameters defines the tracing configuration for a StateMachine.
type TracingConfigurationRAWParameters struct {
	// When set to true, AWS X-Ray tracing is enabled.
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
}

// TracingConfigurationRAWInitParameters defines the init parameters for tracing configuration.
type TracingConfigurationRAWInitParameters struct {
	// When set to true, AWS X-Ray tracing is enabled.
	// +kubebuilder:validation:Optional
	Enabled *bool `json:"enabled,omitempty"`
}

// TracingConfigurationRAWObservation defines the observed state for tracing configuration.
type TracingConfigurationRAWObservation struct {
	// When set to true, AWS X-Ray tracing is enabled.
	Enabled *bool `json:"enabled,omitempty"`
}

// StateMachineRAWParameters defines the configuration parameters for a native StateMachine.
type StateMachineRAWParameters struct {
	// The Amazon States Language definition of the state machine.
	// +kubebuilder:validation:Optional
	Definition *string `json:"definition,omitempty"`

	// Defines what encryption configuration is used to encrypt data in the State Machine.
	// +kubebuilder:validation:Optional
	EncryptionConfiguration *EncryptionConfigurationRAWParameters `json:"encryptionConfiguration,omitempty"`

	// Defines what execution history events are logged and where they are logged.
	// +kubebuilder:validation:Optional
	LoggingConfiguration *LoggingConfigurationRAWParameters `json:"loggingConfiguration,omitempty"`

	// Set to true to publish a version of the state machine during creation.
	// +kubebuilder:validation:Optional
	Publish *bool `json:"publish,omitempty"`

	// Region where this resource will be managed. Required for credential resolution.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// The Amazon Resource Name (ARN) of the IAM role to use for this state machine.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
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
	TracingConfiguration *TracingConfigurationRAWParameters `json:"tracingConfiguration,omitempty"`

	// Determines whether a Standard or Express state machine is created.
	// Valid values: STANDARD, EXPRESS.
	// +kubebuilder:validation:Optional
	Type *string `json:"type,omitempty"`
}

// StateMachineRAWInitParameters defines the init parameters for a native StateMachine.
type StateMachineRAWInitParameters struct {
	// The Amazon States Language definition of the state machine.
	// +kubebuilder:validation:Optional
	Definition *string `json:"definition,omitempty"`

	// Defines what encryption configuration is used to encrypt data in the State Machine.
	// +kubebuilder:validation:Optional
	EncryptionConfiguration *EncryptionConfigurationRAWInitParameters `json:"encryptionConfiguration,omitempty"`

	// Defines what execution history events are logged and where they are logged.
	// +kubebuilder:validation:Optional
	LoggingConfiguration *LoggingConfigurationRAWInitParameters `json:"loggingConfiguration,omitempty"`

	// Set to true to publish a version of the state machine during creation.
	// +kubebuilder:validation:Optional
	Publish *bool `json:"publish,omitempty"`

	// The Amazon Resource Name (ARN) of the IAM role to use for this state machine.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/iam/v1beta1.Role
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
	TracingConfiguration *TracingConfigurationRAWInitParameters `json:"tracingConfiguration,omitempty"`

	// Determines whether a Standard or Express state machine is created.
	// +kubebuilder:validation:Optional
	Type *string `json:"type,omitempty"`
}

// StateMachineRAWObservation defines the observed state of a native StateMachine.
type StateMachineRAWObservation struct {
	// The ARN of the state machine.
	Arn *string `json:"arn,omitempty"`

	// The date the state machine was created.
	CreationDate *string `json:"creationDate,omitempty"`

	// The Amazon States Language definition of the state machine.
	Definition *string `json:"definition,omitempty"`

	// Description of the state machine version.
	Description *string `json:"description,omitempty"`

	// Encryption configuration observed state.
	EncryptionConfiguration *EncryptionConfigurationRAWObservation `json:"encryptionConfiguration,omitempty"`

	// The ARN of the state machine (also used as ID).
	ID *string `json:"id,omitempty"`

	// Logging configuration observed state.
	LoggingConfiguration *LoggingConfigurationRAWObservation `json:"loggingConfiguration,omitempty"`

	// Whether a version was published on creation.
	Publish *bool `json:"publish,omitempty"`

	// The ARN of the state machine version revision.
	RevisionID *string `json:"revisionId,omitempty"`

	// The Amazon Resource Name (ARN) of the IAM role.
	RoleArn *string `json:"roleArn,omitempty"`

	// The ARN of the state machine version.
	StateMachineVersionArn *string `json:"stateMachineVersionArn,omitempty"`

	// The current status of the state machine. Either ACTIVE or DELETING.
	Status *string `json:"status,omitempty"`

	// Key-value map of resource tags including provider defaults.
	// +mapType=granular
	TagsAll map[string]*string `json:"tagsAll,omitempty"`

	// Tracing configuration observed state.
	TracingConfiguration *TracingConfigurationRAWObservation `json:"tracingConfiguration,omitempty"`

	// The type of the state machine.
	Type *string `json:"type,omitempty"`

	// Version description.
	VersionDescription *string `json:"versionDescription,omitempty"`
}

// StateMachineRAWSpec defines the desired state of StateMachineRAW.
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
type StateMachineRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// AtProvider holds the provider-specific observation fields for the resource.
	AtProvider StateMachineRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// StateMachineRAW is the native (non-Terraform) Schema for AWS Step Functions
// State Machines API.
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
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
func (sm *StateMachineRAW) GetForProvider() *StateMachineRAWParameters {
	return &sm.Spec.ForProvider
}

// GetInitProvider returns the InitProvider parameters.
func (sm *StateMachineRAW) GetInitProvider() *StateMachineRAWInitParameters {
	return &sm.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (sm *StateMachineRAW) GetAtProvider() StateMachineRAWObservation {
	return sm.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (sm *StateMachineRAW) SetAtProvider(o StateMachineRAWObservation) {
	sm.Status.AtProvider = o
}

// SetForProviderType sets spec.forProvider.type. Used by late-initialization
// logic in the shared CRUD package to write AWS-defaulted type values back
// into the spec without requiring a separate interface for the cluster type.
func (sm *StateMachineRAW) SetForProviderType(t *string) {
	sm.Spec.ForProvider.Type = t
}
