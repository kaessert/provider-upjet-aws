// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
)

// GetExternalName returns the external name annotation value for the given
// managed resource. Returns an empty string if the annotation is not set.
// This is a thin wrapper around meta.GetExternalName for use in native
// controllers — it avoids importing the meta package in every controller file.
func GetExternalName(cr resource.Managed) string {
	return meta.GetExternalName(cr)
}

// SetExternalName sets the external name annotation on the given managed
// resource. This is a thin wrapper around meta.SetExternalName for use in
// native controllers.
func SetExternalName(cr resource.Managed, name string) {
	meta.SetExternalName(cr, name)
}
