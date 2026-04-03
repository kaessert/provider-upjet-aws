// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package elasticache registers all native cluster-scoped ElastiCache controllers.
package elasticache

import (
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	clustercontroller "github.com/upbound/provider-aws/v2/internal/controller/cluster"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/elasticache/clusterraw"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/elasticache/globalreplicationgroupraw"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/elasticache/parametergroupraw"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/elasticache/replicationgroupraw"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/elasticache/serverlesscacheraw"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/elasticache/subnetgroupraw"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/elasticache/userraw"
	"github.com/upbound/provider-aws/v2/internal/controller/cluster/elasticache/usergroupraw"
)

func init() {
	clustercontroller.NativeSetupHook_elasticache = SetupAll
}

// SetupAll registers all native cluster-scoped ElastiCache controllers.
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
