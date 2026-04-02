// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package sns registers all native namespaced-scoped SNS controllers by setting
// the NativeSetupHook_sns variable. This package is imported by the provider
// binary via cmd/provider/sns/native_imports.go.
package sns

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	namespacedcontroller "github.com/upbound/provider-aws/v2/internal/controller/namespaced"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/sns/topicraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/sns/topicsubscriptionraw"
)

func init() {
	// Register this SetupAll function into the NativeSetupHook for the sns service.
	// This runs automatically when the package is imported by the provider binary.
	namespacedcontroller.NativeSetupHook_sns = SetupAll
}

// SetupAll registers all native namespaced-scoped SNS controllers.
func SetupAll(mgr ctrl.Manager, o xpcontroller.Options) error {
	for _, setup := range []func(ctrl.Manager, xpcontroller.Options) error{
		topicraw.Setup,
		topicsubscriptionraw.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
