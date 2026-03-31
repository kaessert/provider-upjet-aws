// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// extractorFakeManaged is a minimal implementation of resource.Managed used
// to test the ExtractResourceID and ExtractAtProviderField functions.
// It implements all required interface methods explicitly.
type extractorFakeManaged struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              extractorFakeSpec   `json:"spec,omitempty"`
	Status            extractorFakeStatus `json:"status,omitempty"`
}

type extractorFakeSpec struct {
	ManagementPolicies xpv1.ManagementPolicies `json:"managementPolicies,omitempty"`
	DeletionPolicy     xpv1.DeletionPolicy     `json:"deletionPolicy,omitempty"`
}

type extractorFakeStatus struct {
	Conditions []xpv1.Condition         `json:"conditions,omitempty"`
	AtProvider extractorFakeObservation `json:"atProvider,omitempty"`
}

type extractorFakeObservation struct {
	ID  *string `json:"id,omitempty"`
	ARN *string `json:"arn,omitempty"`
}

// GetObjectKind satisfies runtime.Object — delegated to embedded TypeMeta.
func (f *extractorFakeManaged) GetObjectKind() schema.ObjectKind { return &f.TypeMeta }

// DeepCopyObject satisfies runtime.Object.
func (f *extractorFakeManaged) DeepCopyObject() runtime.Object { out := *f; return &out }

// GetCondition satisfies resource.Conditioned.
func (f *extractorFakeManaged) GetCondition(ct xpv1.ConditionType) xpv1.Condition {
	for _, c := range f.Status.Conditions {
		if c.Type == ct {
			return c
		}
	}
	return xpv1.Condition{}
}

// SetConditions satisfies resource.Conditioned.
func (f *extractorFakeManaged) SetConditions(c ...xpv1.Condition) {
	f.Status.Conditions = append(f.Status.Conditions, c...)
}

// GetManagementPolicies satisfies resource.Manageable.
func (f *extractorFakeManaged) GetManagementPolicies() xpv1.ManagementPolicies {
	return f.Spec.ManagementPolicies
}

// SetManagementPolicies satisfies resource.Manageable.
func (f *extractorFakeManaged) SetManagementPolicies(p xpv1.ManagementPolicies) {
	f.Spec.ManagementPolicies = p
}

func strPtr(s string) *string { return &s }

func TestExtractResourceID(t *testing.T) {
	tests := []struct {
		name string
		mr   *extractorFakeManaged
		want string
	}{
		{
			name: "returns id from status.atProvider.id",
			mr: &extractorFakeManaged{
				Status: extractorFakeStatus{
					AtProvider: extractorFakeObservation{
						ID: strPtr("my-resource-id"),
					},
				},
			},
			want: "my-resource-id",
		},
		{
			name: "returns empty string when id is not set",
			mr: &extractorFakeManaged{
				Status: extractorFakeStatus{
					AtProvider: extractorFakeObservation{},
				},
			},
			want: "",
		},
	}

	extractor := ExtractResourceID()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractor(tt.mr)
			if got != tt.want {
				t.Errorf("ExtractResourceID() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractAtProviderField(t *testing.T) {
	tests := []struct {
		name  string
		field string
		mr    *extractorFakeManaged
		want  string
	}{
		{
			name:  "extracts arn from status.atProvider.arn",
			field: "arn",
			mr: &extractorFakeManaged{
				Status: extractorFakeStatus{
					AtProvider: extractorFakeObservation{
						ARN: strPtr("arn:aws:s3:::my-bucket"),
					},
				},
			},
			want: "arn:aws:s3:::my-bucket",
		},
		{
			name:  "returns empty string when field is not set",
			field: "arn",
			mr: &extractorFakeManaged{
				Status: extractorFakeStatus{
					AtProvider: extractorFakeObservation{},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			extractor := ExtractAtProviderField(tt.field)
			got := extractor(tt.mr)
			if got != tt.want {
				t.Errorf("ExtractAtProviderField(%q) = %q, want %q", tt.field, got, tt.want)
			}
		})
	}
}
