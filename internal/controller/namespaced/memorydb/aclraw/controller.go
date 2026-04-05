// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package aclraw contains the namespaced-scope native controller for
// ACLRAW resources.
package aclraw

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsmemorydb "github.com/aws/aws-sdk-go-v2/service/memorydb"
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nativev1beta1 "github.com/upbound/provider-aws/v2/apis/namespaced/memorydb/v1beta1/native"
	memaclshared "github.com/upbound/provider-aws/v2/internal/controller/memorydb/acl"
	native "github.com/upbound/provider-aws/v2/internal/native"
)

// Setup adds a controller that reconciles ACLRAW managed resources (namespaced).
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	name := managed.ControllerName(nativev1beta1.ACLRAW_GroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&nativev1beta1.ACLRAW{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(nativev1beta1.ACLRAW_GroupVersionKind),
			managed.WithTypedExternalConnector[*nativev1beta1.ACLRAW](
				native.NewTypedConnector(
					mgr.GetClient(),
					func(cfg aws.Config) *awsmemorydb.Client {
						return awsmemorydb.NewFromConfig(cfg)
					},
					func(c *awsmemorydb.Client, kube client.Client) managed.TypedExternalClient[*nativev1beta1.ACLRAW] {
						return &typedExternalClient{
							shared: &memaclshared.ExternalClient{Client: c, Kube: kube},
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
	shared *memaclshared.ExternalClient
}

func (t *typedExternalClient) Observe(ctx context.Context, cr *nativev1beta1.ACLRAW) (managed.ExternalObservation, error) {
	return t.shared.Observe(ctx, cr)
}

func (t *typedExternalClient) Create(ctx context.Context, cr *nativev1beta1.ACLRAW) (managed.ExternalCreation, error) {
	return t.shared.Create(ctx, cr)
}

func (t *typedExternalClient) Update(ctx context.Context, cr *nativev1beta1.ACLRAW) (managed.ExternalUpdate, error) {
	return t.shared.Update(ctx, cr)
}

func (t *typedExternalClient) Delete(ctx context.Context, cr *nativev1beta1.ACLRAW) (managed.ExternalDelete, error) {
	return t.shared.Delete(ctx, cr)
}

func (t *typedExternalClient) Disconnect(_ context.Context) error { return nil }
