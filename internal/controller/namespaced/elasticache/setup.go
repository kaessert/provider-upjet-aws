// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package elasticache registers all native namespaced-scoped ElastiCache controllers.
package elasticache

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	namespacedcontroller "github.com/upbound/provider-aws/v2/internal/controller/namespaced"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/elasticache/clusterraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/elasticache/globalreplicationgroupraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/elasticache/parametergroupraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/elasticache/replicationgroupraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/elasticache/serverlesscacheraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/elasticache/subnetgroupraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/elasticache/userraw"
	"github.com/upbound/provider-aws/v2/internal/controller/namespaced/elasticache/usergroupraw"
)

func init() {
	namespacedcontroller.NativeSetupHook_elasticache = SetupAll
}

// SetupAll registers all native namespaced-scoped ElastiCache controllers.
func SetupAll(mgr ctrl.Manager, o xpcontroller.Options) error {
	for _, setup := range []func(ctrl.Manager, xpcontroller.Options) error{
		subnetgroupraw.Setup,
		parametergroupraw.Setup,
		usergroupraw.Setup,
		userraw.Setup,
		clusterraw.Setup,
		globalreplicationgroupraw.Setup,
		serverlesscacheraw.Setup,
		replicationgroupraw.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
