// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package kinesis registers all native namespaced-scoped Kinesis controllers by
// setting the NativeSetupHook_kinesis variable. This package is imported by the
// provider binary via cmd/provider/kinesis/native_imports.go.
package kinesis

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	namespacedcontroller "github.com/upbound/provider-aws/v2/internal/controller/namespaced"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/kinesis/streamconsumerraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/kinesis/streamraw"
)

func init() {
	// Register this SetupAll function into the NativeSetupHook for the kinesis service.
	// This runs automatically when the package is imported by the provider binary.
	namespacedcontroller.NativeSetupHook_kinesis = SetupAll
}

// SetupAll registers all native namespaced-scoped Kinesis controllers.
func SetupAll(mgr ctrl.Manager, o xpcontroller.Options) error {
	for _, setup := range []func(ctrl.Manager, xpcontroller.Options) error{
		streamraw.Setup,
		streamconsumerraw.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
