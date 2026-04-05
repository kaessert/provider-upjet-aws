// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package memorydb registers all native namespaced-scoped MemoryDB controllers.
package memorydb

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	namespacedcontroller "github.com/upbound/provider-aws/v2/internal/controller/namespaced"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/memorydb/aclraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/memorydb/clusterraw"
)

func init() {
	namespacedcontroller.NativeSetupHook_memorydb = SetupAll
}

// SetupAll registers all native namespaced-scoped MemoryDB controllers.
func SetupAll(mgr ctrl.Manager, o xpcontroller.Options) error {
	for _, setup := range []func(ctrl.Manager, xpcontroller.Options) error{
		aclraw.Setup,
		clusterraw.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
