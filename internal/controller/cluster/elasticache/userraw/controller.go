// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package userraw contains the cluster-scoped native controller for
// UserRAW resources.
package userraw

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

	nativev1beta2 "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native"
	elasticacheuser "github.com/upbound/provider-aws/v2/internal/controller/elasticache/user"
	native "github.com/upbound/provider-aws/v2/internal/native"
)

// Setup adds a controller that reconciles UserRAW managed resources.
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	name := managed.ControllerName(nativev1beta2.UserRAW_GroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&nativev1beta2.UserRAW{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(nativev1beta2.UserRAW_GroupVersionKind),
			managed.WithTypedExternalConnector[*nativev1beta2.UserRAW](
				native.NewTypedConnector(
					mgr.GetClient(),
					func(cfg aws.Config) *awselasticache.Client {
						return awselasticache.NewFromConfig(cfg)
					},
					func(c *awselasticache.Client, kube client.Client) managed.TypedExternalClient[*nativev1beta2.UserRAW] {
						return &typedExternalClient{
							shared: &elasticacheuser.ExternalClient{Client: c, Kube: kube},
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
	shared *elasticacheuser.ExternalClient
}

func (t *typedExternalClient) Observe(ctx context.Context, cr *nativev1beta2.UserRAW) (managed.ExternalObservation, error) {
	return t.shared.Observe(ctx, cr)
}

func (t *typedExternalClient) Create(ctx context.Context, cr *nativev1beta2.UserRAW) (managed.ExternalCreation, error) {
	return t.shared.Create(ctx, cr)
}

func (t *typedExternalClient) Update(ctx context.Context, cr *nativev1beta2.UserRAW) (managed.ExternalUpdate, error) {
	return t.shared.Update(ctx, cr)
}

func (t *typedExternalClient) Delete(ctx context.Context, cr *nativev1beta2.UserRAW) (managed.ExternalDelete, error) {
	return t.shared.Delete(ctx, cr)
}

func (t *typedExternalClient) Disconnect(_ context.Context) error { return nil }
