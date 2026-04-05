// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
)

// RotationRulesRAWParameters defines the rotation schedule configuration.
// In v1beta1, rotation_rules is a SLICE (matching TF v1beta1 shape).
type RotationRulesRAWParameters struct {
	// Specifies the number of days between automatic scheduled rotations of the
	// secret. Either automatically_after_days or schedule_expression must be
	// specified.
	// +kubebuilder:validation:Optional
	AutomaticallyAfterDays *float64 `json:"automaticallyAfterDays,omitempty"`

	// The length of the rotation window in hours. For example, 3h for a three
	// hour window.
	// +kubebuilder:validation:Optional
	Duration *string `json:"duration,omitempty"`

	// A cron() or rate() expression that defines the schedule for rotating your
	// secret. Either automatically_after_days or schedule_expression must be
	// specified.
	// +kubebuilder:validation:Optional
	ScheduleExpression *string `json:"scheduleExpression,omitempty"`
}

// RotationRulesRAWObservation holds the observed rotation rule state.
type RotationRulesRAWObservation struct {
	// Specifies the number of days between automatic scheduled rotations.
	AutomaticallyAfterDays *float64 `json:"automaticallyAfterDays,omitempty"`

	// The length of the rotation window in hours.
	Duration *string `json:"duration,omitempty"`

	// A cron() or rate() expression that defines the rotation schedule.
	ScheduleExpression *string `json:"scheduleExpression,omitempty"`
}

// RotationRulesRAWInitParameters defines the init parameters for rotation rules.
type RotationRulesRAWInitParameters struct {
	// Specifies the number of days between automatic scheduled rotations.
	// +kubebuilder:validation:Optional
	AutomaticallyAfterDays *float64 `json:"automaticallyAfterDays,omitempty"`

	// The length of the rotation window in hours.
	// +kubebuilder:validation:Optional
	Duration *string `json:"duration,omitempty"`

	// A cron() or rate() expression that defines the rotation schedule.
	// +kubebuilder:validation:Optional
	ScheduleExpression *string `json:"scheduleExpression,omitempty"`
}

// SecretRotationRAWParameters defines the configuration parameters for a
// SecretRotationRAW resource.
type SecretRotationRAWParameters struct {
	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Specifies whether to rotate the secret immediately or wait until the next
	// scheduled rotation window. The rotation schedule is defined in
	// rotation_rules. Defaults to true.
	// +kubebuilder:validation:Optional
	RotateImmediately *bool `json:"rotateImmediately,omitempty"`

	// Specifies the ARN of the Lambda function that can rotate the secret. Must
	// be supplied if the secret is not managed by AWS.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/lambda/v1beta1.Function
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("arn")
	// +kubebuilder:validation:Optional
	RotationLambdaArn *string `json:"rotationLambdaArn,omitempty"`

	// Reference to a Function in lambda to populate rotationLambdaArn.
	// +kubebuilder:validation:Optional
	RotationLambdaArnRef *xpv1.Reference `json:"rotationLambdaArnRef,omitempty"`

	// Selector for a Function in lambda to populate rotationLambdaArn.
	// +kubebuilder:validation:Optional
	RotationLambdaArnSelector *xpv1.Selector `json:"rotationLambdaArnSelector,omitempty"`

	// A structure that defines the rotation configuration for this secret.
	// In v1beta1 this is a slice (matching the Terraform v1beta1 shape).
	// +kubebuilder:validation:Optional
	RotationRules []RotationRulesRAWParameters `json:"rotationRules,omitempty"`

	// Specifies the secret to which you want to add rotation configuration.
	// You can specify either the Amazon Resource Name (ARN) or the friendly name
	// of the secret.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native.SecretRAW
	// +kubebuilder:validation:Optional
	SecretID *string `json:"secretId,omitempty"`

	// Reference to a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDRef *xpv1.Reference `json:"secretIdRef,omitempty"`

	// Selector for a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDSelector *xpv1.Selector `json:"secretIdSelector,omitempty"`
}

// SecretRotationRAWInitParameters defines the init parameters for a SecretRotationRAW.
type SecretRotationRAWInitParameters struct {
	// Specifies whether to rotate the secret immediately or wait until the next
	// scheduled rotation window.
	// +kubebuilder:validation:Optional
	RotateImmediately *bool `json:"rotateImmediately,omitempty"`

	// Specifies the ARN of the Lambda function that can rotate the secret.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/lambda/v1beta1.Function
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("arn")
	// +kubebuilder:validation:Optional
	RotationLambdaArn *string `json:"rotationLambdaArn,omitempty"`

	// Reference to a Function in lambda to populate rotationLambdaArn.
	// +kubebuilder:validation:Optional
	RotationLambdaArnRef *xpv1.Reference `json:"rotationLambdaArnRef,omitempty"`

	// Selector for a Function in lambda to populate rotationLambdaArn.
	// +kubebuilder:validation:Optional
	RotationLambdaArnSelector *xpv1.Selector `json:"rotationLambdaArnSelector,omitempty"`

	// A structure that defines the rotation configuration for this secret.
	// +kubebuilder:validation:Optional
	RotationRules []RotationRulesRAWInitParameters `json:"rotationRules,omitempty"`

	// Specifies the secret to which you want to add rotation configuration.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native.SecretRAW
	// +kubebuilder:validation:Optional
	SecretID *string `json:"secretId,omitempty"`

	// Reference to a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDRef *xpv1.Reference `json:"secretIdRef,omitempty"`

	// Selector for a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDSelector *xpv1.Selector `json:"secretIdSelector,omitempty"`
}

// SecretRotationRAWObservation holds the observed state of a SecretRotationRAW.
type SecretRotationRAWObservation struct {
	// Amazon Resource Name (ARN) of the secret (used as external name / ID).
	ID *string `json:"id,omitempty"`

	// Specifies whether automatic rotation is enabled for this secret.
	RotationEnabled *bool `json:"rotationEnabled,omitempty"`

	// The ARN of the Lambda function that can rotate the secret.
	RotationLambdaArn *string `json:"rotationLambdaArn,omitempty"`

	// A structure that defines the rotation configuration for this secret.
	RotationRules []RotationRulesRAWObservation `json:"rotationRules,omitempty"`
}

// SecretRotationRAWSpec defines the desired state of SecretRotationRAW.
type SecretRotationRAWSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SecretRotationRAWParameters `json:"forProvider"`
	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields. The fields that are
	// in InitProvider are merged into ForProvider when the resource is created.
	// +optional
	InitProvider SecretRotationRAWInitParameters `json:"initProvider,omitempty"`
}

// SecretRotationRAWStatus defines the observed state of SecretRotationRAW.
type SecretRotationRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SecretRotationRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}

// SecretRotationRAW is the Schema for the native Secrets Manager Secret Rotation API.
type SecretRotationRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.rotationRules) || (has(self.initProvider) && has(self.initProvider.rotationRules))",message="spec.forProvider.rotationRules is a required parameter"
	Spec   SecretRotationRAWSpec   `json:"spec"`
	Status SecretRotationRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretRotationRAWList contains a list of SecretRotationRAW resources.
type SecretRotationRAWList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecretRotationRAW `json:"items"`
}

// Repository type metadata.
var (
	SecretRotationRAW_Kind             = "SecretRotationRAW"
	SecretRotationRAW_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: SecretRotationRAW_Kind}.String()
	SecretRotationRAW_KindAPIVersion   = SecretRotationRAW_Kind + "." + CRDGroupVersion.String()
	SecretRotationRAW_GroupVersionKind = CRDGroupVersion.WithKind(SecretRotationRAW_Kind)
)

func init() {
	SchemeBuilder.Register(&SecretRotationRAW{}, &SecretRotationRAWList{})
}

// GetForProvider returns a pointer to the ForProvider parameters.
func (s *SecretRotationRAW) GetForProvider() *SecretRotationRAWParameters {
	return &s.Spec.ForProvider
}

// SetForProvider sets spec.forProvider to the given parameters.
// Required by the SecretRotationCR interface for late-initialization write-back.
func (s *SecretRotationRAW) SetForProvider(p SecretRotationRAWParameters) {
	s.Spec.ForProvider = p
}

// GetInitProvider returns the InitProvider parameters.
func (s *SecretRotationRAW) GetInitProvider() *SecretRotationRAWInitParameters {
	return &s.Spec.InitProvider
}

// GetAtProvider returns the current observed state.
func (s *SecretRotationRAW) GetAtProvider() SecretRotationRAWObservation {
	return s.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (s *SecretRotationRAW) SetAtProvider(o SecretRotationRAWObservation) {
	s.Status.AtProvider = o
}
