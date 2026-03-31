// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"testing"
)

func TestPoliciesAreEquivalent(t *testing.T) {
	// A well-formed IAM policy with one statement.
	policyA := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:GetObject",
				"Principal": {"AWS": "arn:aws:iam::123456789012:root"},
				"Resource": "arn:aws:s3:::my-bucket/*"
			}
		]
	}`

	// Same policy with keys in different order and no whitespace.
	policyASameContent := `{"Statement":[{"Action":"s3:GetObject","Effect":"Allow","Principal":{"AWS":"arn:aws:iam::123456789012:root"},"Resource":"arn:aws:s3:::my-bucket/*"}],"Version":"2012-10-17"}`

	// A policy with a different action.
	policyB := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": "s3:PutObject",
				"Principal": {"AWS": "arn:aws:iam::123456789012:root"},
				"Resource": "arn:aws:s3:::my-bucket/*"
			}
		]
	}`

	tests := []struct {
		name      string
		policy1   string
		policy2   string
		wantEqual bool
		wantErr   bool
	}{
		{
			name:      "identical strings are equivalent",
			policy1:   policyA,
			policy2:   policyA,
			wantEqual: true,
			wantErr:   false,
		},
		{
			name:      "same policy different key order and whitespace are equivalent",
			policy1:   policyA,
			policy2:   policyASameContent,
			wantEqual: true,
			wantErr:   false,
		},
		{
			name:      "policies with different statements are not equivalent",
			policy1:   policyA,
			policy2:   policyB,
			wantEqual: false,
			wantErr:   false,
		},
		{
			name:      "invalid JSON returns error",
			policy1:   `not-valid-json`,
			policy2:   policyA,
			wantEqual: false,
			wantErr:   true,
		},
		{
			name:      "both invalid JSON returns error",
			policy1:   `{bad}`,
			policy2:   `{also bad}`,
			wantEqual: false,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := PoliciesAreEquivalent(tc.policy1, tc.policy2)
			if tc.wantErr && err == nil {
				t.Errorf("PoliciesAreEquivalent() expected error but got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("PoliciesAreEquivalent() unexpected error: %v", err)
			}
			if got != tc.wantEqual {
				t.Errorf("PoliciesAreEquivalent() = %v, want %v", got, tc.wantEqual)
			}
		})
	}
}

func TestPolicyNeedsUpdate(t *testing.T) {
	policyA := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:GetObject","Principal":{"AWS":"arn:aws:iam::123456789012:root"},"Resource":"arn:aws:s3:::my-bucket/*"}]}`
	policyASameContent := `{
		"Version": "2012-10-17",
		"Statement": [{
			"Resource": "arn:aws:s3:::my-bucket/*",
			"Principal": {"AWS": "arn:aws:iam::123456789012:root"},
			"Effect": "Allow",
			"Action": "s3:GetObject"
		}]
	}`
	policyB := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"s3:PutObject","Principal":{"AWS":"arn:aws:iam::123456789012:root"},"Resource":"arn:aws:s3:::my-bucket/*"}]}`

	tests := []struct {
		name     string
		desired  string
		observed string
		want     bool
	}{
		{
			name:     "equivalent policies do not need update",
			desired:  policyA,
			observed: policyASameContent,
			want:     false,
		},
		{
			name:     "identical strings do not need update",
			desired:  policyA,
			observed: policyA,
			want:     false,
		},
		{
			name:     "different policies need update",
			desired:  policyA,
			observed: policyB,
			want:     true,
		},
		{
			name:     "invalid desired policy conservatively needs update",
			desired:  `not-json`,
			observed: policyA,
			want:     true,
		},
		{
			name:     "invalid observed policy conservatively needs update",
			desired:  policyA,
			observed: `not-json`,
			want:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := PolicyNeedsUpdate(tc.desired, tc.observed)
			if got != tc.want {
				t.Errorf("PolicyNeedsUpdate() = %v, want %v", got, tc.want)
			}
		})
	}
}
