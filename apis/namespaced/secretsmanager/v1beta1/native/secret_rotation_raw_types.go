// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpv2 "github.com/crossplane/crossplane-runtime/v2/apis/common/v2"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native"
)

// RotationRulesRAWParameters defines the rotation schedule configuration for
// the namespaced SecretRotationRAW.
// Fields are identical to the cluster-scoped type except Ref/Selector fields
// use NamespacedReference/NamespacedSelector.
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

// RotationRulesRAWInitParameters defines the init parameters for namespaced
// rotation rules.
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

// SecretRotationRAWParameters defines the namespaced configuration parameters
// for a SecretRotationRAW resource.
// Fields are identical to the cluster-scoped v1beta1 shape except Ref/Selector
// fields use NamespacedReference/NamespacedSelector (as required by angryjet).
type SecretRotationRAWParameters struct {
	// Region where this resource will be managed.
	// +kubebuilder:validation:Required
	Region *string `json:"region"`

	// Specifies whether to rotate the secret immediately or wait until the next
	// scheduled rotation window. Defaults to true.
	// +kubebuilder:validation:Optional
	RotateImmediately *bool `json:"rotateImmediately,omitempty"`

	// Specifies the ARN of the Lambda function that can rotate the secret. Must
	// be supplied if the secret is not managed by AWS.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/lambda/v1beta1.Function
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("arn")
	// +kubebuilder:validation:Optional
	RotationLambdaArn *string `json:"rotationLambdaArn,omitempty"`

	// Reference to a Function in lambda to populate rotationLambdaArn.
	// +kubebuilder:validation:Optional
	RotationLambdaArnRef *xpv1.NamespacedReference `json:"rotationLambdaArnRef,omitempty"`

	// Selector for a Function in lambda to populate rotationLambdaArn.
	// +kubebuilder:validation:Optional
	RotationLambdaArnSelector *xpv1.NamespacedSelector `json:"rotationLambdaArnSelector,omitempty"`

	// A structure that defines the rotation configuration for this secret.
	// Namespaced v1beta1 uses a POINTER (matching the namespaced TF v1beta1 shape).
	// +kubebuilder:validation:Optional
	RotationRules *RotationRulesRAWParameters `json:"rotationRules,omitempty"`

	// Specifies the secret to which you want to add rotation configuration.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/secretsmanager/v1beta1/native.SecretRAW
	// +kubebuilder:validation:Optional
	SecretID *string `json:"secretId,omitempty"`

	// Reference to a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDRef *xpv1.NamespacedReference `json:"secretIdRef,omitempty"`

	// Selector for a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDSelector *xpv1.NamespacedSelector `json:"secretIdSelector,omitempty"`
}

// SecretRotationRAWInitParameters defines the namespaced init parameters for a
// SecretRotationRAW.
type SecretRotationRAWInitParameters struct {
	// Specifies whether to rotate the secret immediately or wait until the next
	// scheduled rotation window.
	// +kubebuilder:validation:Optional
	RotateImmediately *bool `json:"rotateImmediately,omitempty"`

	// Specifies the ARN of the Lambda function that can rotate the secret.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/lambda/v1beta1.Function
	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("arn")
	// +kubebuilder:validation:Optional
	RotationLambdaArn *string `json:"rotationLambdaArn,omitempty"`

	// Reference to a Function in lambda to populate rotationLambdaArn.
	// +kubebuilder:validation:Optional
	RotationLambdaArnRef *xpv1.NamespacedReference `json:"rotationLambdaArnRef,omitempty"`

	// Selector for a Function in lambda to populate rotationLambdaArn.
	// +kubebuilder:validation:Optional
	RotationLambdaArnSelector *xpv1.NamespacedSelector `json:"rotationLambdaArnSelector,omitempty"`

	// A structure that defines the rotation configuration for this secret.
	// +kubebuilder:validation:Optional
	RotationRules *RotationRulesRAWInitParameters `json:"rotationRules,omitempty"`

	// Specifies the secret to which you want to add rotation configuration.
	// +crossplane:generate:reference:type=github.com/upbound/provider-aws/v2/apis/namespaced/secretsmanager/v1beta1/native.SecretRAW
	// +kubebuilder:validation:Optional
	SecretID *string `json:"secretId,omitempty"`

	// Reference to a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDRef *xpv1.NamespacedReference `json:"secretIdRef,omitempty"`

	// Selector for a SecretRAW in secretsmanager to populate secretId.
	// +kubebuilder:validation:Optional
	SecretIDSelector *xpv1.NamespacedSelector `json:"secretIdSelector,omitempty"`
}

// SecretRotationRAWSpec defines the desired state of the namespaced SecretRotationRAW.
type SecretRotationRAWSpec struct {
	xpv2.ManagedResourceSpec `json:",inline"`

	// ForProvider holds the provider-specific configuration for the resource.
	ForProvider SecretRotationRAWParameters `json:"forProvider"`

	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider SecretRotationRAWInitParameters `json:"initProvider,omitempty"`
}

// SecretRotationRAWStatus defines the observed state of the namespaced SecretRotationRAW.
type SecretRotationRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          clusternative.SecretRotationRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Namespaced,categories={crossplane,managed,aws}

// SecretRotationRAW is the Schema for the native Secrets Manager Secret
// Rotation API (namespaced scope).
type SecretRotationRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.rotationRules) || (has(self.initProvider) && has(self.initProvider.rotationRules))",message="spec.forProvider.rotationRules is a required parameter"
	Spec   SecretRotationRAWSpec   `json:"spec"`
	Status SecretRotationRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretRotationRAWList contains a list of namespaced SecretRotationRAW resources.
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

// GetForProvider returns a cluster-scoped v1beta1 SecretRotationRAWParameters
// populated from this namespaced resource's ForProvider fields. The RotationRules
// pointer is converted to a slice for the v1beta1 shape. Because this returns a
// copy, callers MUST use SetForProvider() to write back any modifications.
func (s *SecretRotationRAW) GetForProvider() *clusternative.SecretRotationRAWParameters {
	out := &clusternative.SecretRotationRAWParameters{
		Region:            s.Spec.ForProvider.Region,
		RotateImmediately: s.Spec.ForProvider.RotateImmediately,
		RotationLambdaArn: s.Spec.ForProvider.RotationLambdaArn,
		// RotationLambdaArnRef and RotationLambdaArnSelector are intentionally
		// not copied — they are NamespacedReference/NamespacedSelector types
		// that live only in the namespaced spec.
		SecretID: s.Spec.ForProvider.SecretID,
		// SecretIDRef and SecretIDSelector: same as above.
	}
	if s.Spec.ForProvider.RotationRules != nil {
		out.RotationRules = []clusternative.RotationRulesRAWParameters{
			{
				AutomaticallyAfterDays: s.Spec.ForProvider.RotationRules.AutomaticallyAfterDays,
				Duration:               s.Spec.ForProvider.RotationRules.Duration,
				ScheduleExpression:     s.Spec.ForProvider.RotationRules.ScheduleExpression,
			},
		}
	}
	return out
}

// SetForProvider copies cluster-scoped v1beta1 SecretRotationRAWParameters back
// to the namespaced spec. Required by the SecretRotationCR interface for
// late-initialization write-back. Namespaced GetForProvider() returns a
// freshly-allocated copy, so any mutations made by the shared CRUD late-init
// logic must be written back via this method to persist in the spec.
// Note: Ref/Selector fields are NOT overwritten — they live only in the
// namespaced type and are managed by angryjet resolvers.
func (s *SecretRotationRAW) SetForProvider(p clusternative.SecretRotationRAWParameters) {
	s.Spec.ForProvider.Region = p.Region
	s.Spec.ForProvider.RotateImmediately = p.RotateImmediately
	s.Spec.ForProvider.RotationLambdaArn = p.RotationLambdaArn
	s.Spec.ForProvider.SecretID = p.SecretID
	if len(p.RotationRules) > 0 {
		s.Spec.ForProvider.RotationRules = &RotationRulesRAWParameters{
			AutomaticallyAfterDays: p.RotationRules[0].AutomaticallyAfterDays,
			Duration:               p.RotationRules[0].Duration,
			ScheduleExpression:     p.RotationRules[0].ScheduleExpression,
		}
	} else {
		s.Spec.ForProvider.RotationRules = nil
	}
}

// GetInitProvider returns cluster-scoped v1beta1 SecretRotationRAWInitParameters
// populated from this namespaced resource's InitProvider fields.
func (s *SecretRotationRAW) GetInitProvider() *clusternative.SecretRotationRAWInitParameters {
	out := &clusternative.SecretRotationRAWInitParameters{
		RotateImmediately: s.Spec.InitProvider.RotateImmediately,
		RotationLambdaArn: s.Spec.InitProvider.RotationLambdaArn,
		SecretID:          s.Spec.InitProvider.SecretID,
	}
	if s.Spec.InitProvider.RotationRules != nil {
		out.RotationRules = []clusternative.RotationRulesRAWInitParameters{
			{
				AutomaticallyAfterDays: s.Spec.InitProvider.RotationRules.AutomaticallyAfterDays,
				Duration:               s.Spec.InitProvider.RotationRules.Duration,
				ScheduleExpression:     s.Spec.InitProvider.RotationRules.ScheduleExpression,
			},
		}
	}
	return out
}

// GetAtProvider returns the current observed state.
func (s *SecretRotationRAW) GetAtProvider() clusternative.SecretRotationRAWObservation {
	return s.Status.AtProvider
}

// SetAtProvider sets the observed state.
func (s *SecretRotationRAW) SetAtProvider(o clusternative.SecretRotationRAWObservation) {
	s.Status.AtProvider = o
}
