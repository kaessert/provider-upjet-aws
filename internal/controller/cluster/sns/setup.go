// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package sns registers all native cluster-scoped SNS controllers by setting
// the NativeSetupHook_sns variable. This package is imported by the provider
// binary via cmd/provider/sns/native_imports.go.
package sns

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	clustercontroller "github.com/upbound/provider-aws/v2/internal/controller/cluster"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/sns/topicraw"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/sns/topicsubscriptionraw"
)

func init() {
	// Register this SetupAll function into the NativeSetupHook for the sns service.
	// This runs automatically when the package is imported by the provider binary.
	clustercontroller.NativeSetupHook_sns = SetupAll
}

// SetupAll registers all native cluster-scoped SNS controllers.
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
