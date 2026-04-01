// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package sqs registers all native namespaced-scoped SQS controllers by setting
// the NativeSetupHook_sqs variable. This package is imported by the provider
// binary via cmd/provider/sqs/native_imports.go.
package sqs

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	namespacedcontroller "github.com/upbound/provider-aws/v2/internal/controller/namespaced"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/sqs/queuepolicyraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/sqs/queueraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/sqs/queueredriveallowpolicyraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/sqs/queueredrivepolicyraw"
)

func init() {
	// Register this SetupAll function into the NativeSetupHook for the sqs service.
	// This runs automatically when the package is imported by the provider binary.
	namespacedcontroller.NativeSetupHook_sqs = SetupAll
}

// SetupAll registers all native namespaced-scoped SQS controllers.
func SetupAll(mgr ctrl.Manager, o xpcontroller.Options) error {
	for _, setup := range []func(ctrl.Manager, xpcontroller.Options) error{
		queueraw.Setup,
		queuepolicyraw.Setup,
		queueredrivepolicyraw.Setup,
		queueredriveallowpolicyraw.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
