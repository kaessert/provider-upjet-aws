// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package secretsmanager registers all native namespaced-scope secretsmanager controllers
// by setting the NativeSetupHook_secretsmanager variable. This package is imported by
// the provider binary via cmd/provider/secretsmanager/native_imports.go.
package secretsmanager

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	namespacedcontroller "github.com/upbound/provider-aws/v2/internal/controller/namespaced"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/secretsmanager/secretpolicyraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/secretsmanager/secretraw"
)

func init() {
	// Register this SetupAll function into the NativeSetupHook for the secretsmanager service.
	// This runs automatically when the package is imported by the provider binary.
	namespacedcontroller.NativeSetupHook_secretsmanager = SetupAll
}

// SetupAll registers all native namespaced-scope secretsmanager controllers.
func SetupAll(mgr ctrl.Manager, o xpcontroller.Options) error {
	for _, setup := range []func(ctrl.Manager, xpcontroller.Options) error{
		secretpolicyraw.Setup,
		secretraw.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
