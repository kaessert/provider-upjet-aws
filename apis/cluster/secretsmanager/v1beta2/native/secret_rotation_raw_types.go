// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"

	v1beta1native "github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native"
)

// RotationRulesRAWParameters defines the rotation schedule configuration.
// In v1beta2, rotation_rules is a POINTER (the conversion-hub / storage shape).
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

// RotationRulesRAWObservation holds the observed rotation rule state (v1beta2).
type RotationRulesRAWObservation struct {
	// Specifies the number of days between automatic scheduled rotations.
	AutomaticallyAfterDays *float64 `json:"automaticallyAfterDays,omitempty"`

	// The length of the rotation window in hours.
	Duration *string `json:"duration,omitempty"`

	// A cron() or rate() expression that defines the rotation schedule.
	ScheduleExpression *string `json:"scheduleExpression,omitempty"`
}

// RotationRulesRAWInitParameters defines the init parameters for rotation rules
// (v1beta2).
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
// SecretRotationRAW resource (v1beta2 — conversion hub).
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
	// In v1beta2 this is a POINTER (conversion hub / storage shape).
	// +kubebuilder:validation:Optional
	RotationRules *RotationRulesRAWParameters `json:"rotationRules,omitempty"`

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

// SecretRotationRAWInitParameters defines the init parameters for a
// SecretRotationRAW (v1beta2).
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
	RotationRules *RotationRulesRAWInitParameters `json:"rotationRules,omitempty"`

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

// SecretRotationRAWObservation holds the observed state of a SecretRotationRAW
// (v1beta2).
type SecretRotationRAWObservation struct {
	// Amazon Resource Name (ARN) of the secret (used as external name / ID).
	ID *string `json:"id,omitempty"`

	// Specifies whether automatic rotation is enabled for this secret.
	RotationEnabled *bool `json:"rotationEnabled,omitempty"`

	// The ARN of the Lambda function that can rotate the secret.
	RotationLambdaArn *string `json:"rotationLambdaArn,omitempty"`

	// A structure that defines the rotation configuration for this secret.
	RotationRules *RotationRulesRAWObservation `json:"rotationRules,omitempty"`
}

// SecretRotationRAWSpec defines the desired state of SecretRotationRAW (v1beta2).
type SecretRotationRAWSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SecretRotationRAWParameters `json:"forProvider"`
	// InitProvider holds the same fields as ForProvider, with the exception
	// of Identifier and other resource reference fields.
	// +optional
	InitProvider SecretRotationRAWInitParameters `json:"initProvider,omitempty"`
}

// SecretRotationRAWStatus defines the observed state of SecretRotationRAW (v1beta2).
type SecretRotationRAWStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SecretRotationRAWObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// SecretRotationRAW is the Schema for the native Secrets Manager Secret
// Rotation API (v1beta2 — conversion hub / storage version).
//
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type SecretRotationRAW struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// +kubebuilder:validation:XValidation:rule="!('*' in self.managementPolicies || 'Create' in self.managementPolicies || 'Update' in self.managementPolicies) || has(self.forProvider.rotationRules) || (has(self.initProvider) && has(self.initProvider.rotationRules))",message="spec.forProvider.rotationRules is a required parameter"
	Spec   SecretRotationRAWSpec   `json:"spec"`
	Status SecretRotationRAWStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SecretRotationRAWList contains a list of SecretRotationRAW (v1beta2) resources.
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

// GetForProvider returns a v1beta1 SecretRotationRAWParameters converted from
// the v1beta2 storage format. The RotationRules pointer is converted to a
// slice (v1beta1 shape) for use by the shared CRUD implementation.
// Because this returns a copy, callers MUST use SetForProvider() to write back
// any modifications (e.g. late-initialized fields).
func (s *SecretRotationRAW) GetForProvider() *v1beta1native.SecretRotationRAWParameters {
	out := &v1beta1native.SecretRotationRAWParameters{
		Region:                    s.Spec.ForProvider.Region,
		RotateImmediately:         s.Spec.ForProvider.RotateImmediately,
		RotationLambdaArn:         s.Spec.ForProvider.RotationLambdaArn,
		RotationLambdaArnRef:      s.Spec.ForProvider.RotationLambdaArnRef,
		RotationLambdaArnSelector: s.Spec.ForProvider.RotationLambdaArnSelector,
		SecretID:                  s.Spec.ForProvider.SecretID,
		SecretIDRef:               s.Spec.ForProvider.SecretIDRef,
		SecretIDSelector:          s.Spec.ForProvider.SecretIDSelector,
	}
	if s.Spec.ForProvider.RotationRules != nil {
		out.RotationRules = []v1beta1native.RotationRulesRAWParameters{
			{
				AutomaticallyAfterDays: s.Spec.ForProvider.RotationRules.AutomaticallyAfterDays,
				Duration:               s.Spec.ForProvider.RotationRules.Duration,
				ScheduleExpression:     s.Spec.ForProvider.RotationRules.ScheduleExpression,
			},
		}
	}
	return out
}

// SetForProvider copies v1beta1-shaped parameters back into the v1beta2 spec.
// The RotationRules slice is converted to a pointer (v1beta2 shape).
// Required by the SecretRotationCR interface for late-initialization write-back.
func (s *SecretRotationRAW) SetForProvider(p v1beta1native.SecretRotationRAWParameters) {
	s.Spec.ForProvider.Region = p.Region
	s.Spec.ForProvider.RotateImmediately = p.RotateImmediately
	s.Spec.ForProvider.RotationLambdaArn = p.RotationLambdaArn
	s.Spec.ForProvider.RotationLambdaArnRef = p.RotationLambdaArnRef
	s.Spec.ForProvider.RotationLambdaArnSelector = p.RotationLambdaArnSelector
	s.Spec.ForProvider.SecretID = p.SecretID
	s.Spec.ForProvider.SecretIDRef = p.SecretIDRef
	s.Spec.ForProvider.SecretIDSelector = p.SecretIDSelector
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

// GetInitProvider returns v1beta1-shaped InitProvider parameters converted from
// the v1beta2 storage format.
func (s *SecretRotationRAW) GetInitProvider() *v1beta1native.SecretRotationRAWInitParameters {
	out := &v1beta1native.SecretRotationRAWInitParameters{
		RotateImmediately:         s.Spec.InitProvider.RotateImmediately,
		RotationLambdaArn:         s.Spec.InitProvider.RotationLambdaArn,
		RotationLambdaArnRef:      s.Spec.InitProvider.RotationLambdaArnRef,
		RotationLambdaArnSelector: s.Spec.InitProvider.RotationLambdaArnSelector,
		SecretID:                  s.Spec.InitProvider.SecretID,
		SecretIDRef:               s.Spec.InitProvider.SecretIDRef,
		SecretIDSelector:          s.Spec.InitProvider.SecretIDSelector,
	}
	if s.Spec.InitProvider.RotationRules != nil {
		out.RotationRules = []v1beta1native.RotationRulesRAWInitParameters{
			{
				AutomaticallyAfterDays: s.Spec.InitProvider.RotationRules.AutomaticallyAfterDays,
				Duration:               s.Spec.InitProvider.RotationRules.Duration,
				ScheduleExpression:     s.Spec.InitProvider.RotationRules.ScheduleExpression,
			},
		}
	}
	return out
}

// GetAtProvider returns the observed state as v1beta1 observation type.
// The RotationRules pointer is converted to a slice for v1beta1 compatibility.
func (s *SecretRotationRAW) GetAtProvider() v1beta1native.SecretRotationRAWObservation {
	out := v1beta1native.SecretRotationRAWObservation{
		ID:                s.Status.AtProvider.ID,
		RotationEnabled:   s.Status.AtProvider.RotationEnabled,
		RotationLambdaArn: s.Status.AtProvider.RotationLambdaArn,
	}
	if s.Status.AtProvider.RotationRules != nil {
		out.RotationRules = []v1beta1native.RotationRulesRAWObservation{
			{
				AutomaticallyAfterDays: s.Status.AtProvider.RotationRules.AutomaticallyAfterDays,
				Duration:               s.Status.AtProvider.RotationRules.Duration,
				ScheduleExpression:     s.Status.AtProvider.RotationRules.ScheduleExpression,
			},
		}
	}
	return out
}

// SetAtProvider stores the observed state from the v1beta1 observation type.
// The RotationRules slice is converted to a pointer for v1beta2 storage.
func (s *SecretRotationRAW) SetAtProvider(o v1beta1native.SecretRotationRAWObservation) {
	s.Status.AtProvider.ID = o.ID
	s.Status.AtProvider.RotationEnabled = o.RotationEnabled
	s.Status.AtProvider.RotationLambdaArn = o.RotationLambdaArn
	if len(o.RotationRules) > 0 {
		s.Status.AtProvider.RotationRules = &RotationRulesRAWObservation{
			AutomaticallyAfterDays: o.RotationRules[0].AutomaticallyAfterDays,
			Duration:               o.RotationRules[0].Duration,
			ScheduleExpression:     o.RotationRules[0].ScheduleExpression,
		}
	} else {
		s.Status.AtProvider.RotationRules = nil
	}
}
