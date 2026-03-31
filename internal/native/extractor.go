// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"github.com/crossplane/crossplane-runtime/v2/pkg/fieldpath"
	xpref "github.com/crossplane/crossplane-runtime/v2/pkg/reference"
	xpresource "github.com/crossplane/crossplane-runtime/v2/pkg/resource"
)

// ExtractResourceID extracts the value of status.atProvider.id from a managed
// resource using fieldpath traversal. This is the native (non-Terraform)
// equivalent of upjet's resource.TerraformID() — it reads the same field path
// but does NOT require the resource to implement the upjet Terraformed
// interface.
//
// RAW (native) types must use this extractor (or resource.ExternalName()) in
// their +crossplane:generate:reference:extractor annotations instead of
// TerraformID(), which panics for non-Terraformed resources.
//
// Usage in type annotations:
//
//	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractResourceID()
func ExtractResourceID() xpref.ExtractValueFn {
	return func(mr xpresource.Managed) string {
		paved, err := fieldpath.PaveObject(mr)
		if err != nil {
			return ""
		}
		r, err := paved.GetString("status.atProvider.id")
		if err != nil {
			return ""
		}
		return r
	}
}

// ExtractAtProviderField returns an extractor that reads an arbitrary field
// from status.atProvider using fieldpath. For example, pass "arn" to extract
// status.atProvider.arn without casting to any upjet interface.
//
// Usage in type annotations:
//
//	// +crossplane:generate:reference:extractor=github.com/upbound/provider-aws/v2/internal/native.ExtractAtProviderField("arn")
func ExtractAtProviderField(field string) xpref.ExtractValueFn {
	return func(mr xpresource.Managed) string {
		paved, err := fieldpath.PaveObject(mr)
		if err != nil {
			return ""
		}
		r, err := paved.GetString("status.atProvider." + field)
		if err != nil {
			return ""
		}
		return r
	}
}
