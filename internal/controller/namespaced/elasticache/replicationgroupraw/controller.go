// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package replicationgroupraw contains the namespaced-scope native controller for
// ReplicationGroupRAW resources.
package replicationgroupraw

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nativev1beta2 "github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta2/native"
	elasticachereplgroup "github.com/upbound/provider-aws/v2/internal/controller/elasticache/replicationgroup"
	native "github.com/upbound/provider-aws/v2/internal/native"
)

// Setup adds a controller that reconciles ReplicationGroupRAW managed resources (namespaced).
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	name := managed.ControllerName(nativev1beta2.ReplicationGroupRAW_GroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&nativev1beta2.ReplicationGroupRAW{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(nativev1beta2.ReplicationGroupRAW_GroupVersionKind),
			managed.WithTypedExternalConnector[*nativev1beta2.ReplicationGroupRAW](
				native.NewTypedConnector(
					mgr.GetClient(),
					func(cfg aws.Config) *awselasticache.Client {
						return awselasticache.NewFromConfig(cfg)
					},
					func(c *awselasticache.Client, kube client.Client) managed.TypedExternalClient[*nativev1beta2.ReplicationGroupRAW] {
						return &typedExternalClient{
							shared: &elasticachereplgroup.ExternalClient{Client: c, Kube: kube},
						}
					},
				),
			),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithPollInterval(o.PollInterval),
		))
}

type typedExternalClient struct {
	shared *elasticachereplgroup.ExternalClient
}

func (t *typedExternalClient) Observe(ctx context.Context, cr *nativev1beta2.ReplicationGroupRAW) (managed.ExternalObservation, error) {
	return t.shared.Observe(ctx, cr)
}

func (t *typedExternalClient) Create(ctx context.Context, cr *nativev1beta2.ReplicationGroupRAW) (managed.ExternalCreation, error) {
	return t.shared.Create(ctx, cr)
}

func (t *typedExternalClient) Update(ctx context.Context, cr *nativev1beta2.ReplicationGroupRAW) (managed.ExternalUpdate, error) {
	return t.shared.Update(ctx, cr)
}

func (t *typedExternalClient) Delete(ctx context.Context, cr *nativev1beta2.ReplicationGroupRAW) (managed.ExternalDelete, error) {
	return t.shared.Delete(ctx, cr)
}

func (t *typedExternalClient) Disconnect(_ context.Context) error { return nil }
