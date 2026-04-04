// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package secretsmanager registers all native cluster-scoped secretsmanager controllers
// by setting the NativeSetupHook_secretsmanager variable. This package is imported by
// the provider binary via cmd/provider/secretsmanager/native_imports.go.
package secretsmanager

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	clustercontroller "github.com/upbound/provider-aws/v2/internal/controller/cluster"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/secretsmanager/secretpolicyraw"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/secretsmanager/secretraw"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/secretsmanager/secretversionraw"
)

func init() {
	// Register this SetupAll function into the NativeSetupHook for the secretsmanager service.
	// This runs automatically when the package is imported by the provider binary.
	clustercontroller.NativeSetupHook_secretsmanager = SetupAll
}

// SetupAll registers all native cluster-scoped secretsmanager controllers.
func SetupAll(mgr ctrl.Manager, o xpcontroller.Options) error {
	for _, setup := range []func(ctrl.Manager, xpcontroller.Options) error{
		secretpolicyraw.Setup,
		secretraw.Setup,
		secretversionraw.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
