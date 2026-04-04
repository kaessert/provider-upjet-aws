// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package secretversionraw contains the namespaced-scope native controller for
// SecretVersionRAW resources. It is a thin wrapper that delegates all CRUD
// logic to the shared ExternalClient in internal/controller/secretsmanager/secretversion.
package secretversionraw

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssecretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nativev1beta1 "github.com/upbound/provider-aws/v2/apis/namespaced/secretsmanager/v1beta1/native"
	secretversioncontroller "github.com/upbound/provider-aws/v2/internal/controller/secretsmanager/secretversion"
	native "github.com/upbound/provider-aws/v2/internal/native"
)

// Setup adds a controller that reconciles SecretVersionRAW managed resources (namespaced).
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	name := managed.ControllerName(nativev1beta1.SecretVersionRAW_GroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&nativev1beta1.SecretVersionRAW{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(nativev1beta1.SecretVersionRAW_GroupVersionKind),
			managed.WithTypedExternalConnector[*nativev1beta1.SecretVersionRAW](
				native.NewTypedConnector(
					mgr.GetClient(),
					func(cfg aws.Config) *awssecretsmanager.Client {
						return awssecretsmanager.NewFromConfig(cfg)
					},
					func(c *awssecretsmanager.Client, kube client.Client) managed.TypedExternalClient[*nativev1beta1.SecretVersionRAW] {
						return &typedExternalClient{
							shared: &secretversioncontroller.ExternalClient{Client: c, Kube: kube},
						}
					},
				),
			),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithPollInterval(o.PollInterval),
		))
}

// typedExternalClient bridges managed.TypedExternalClient[*nativev1beta1.SecretVersionRAW]
// to the scope-agnostic shared ExternalClient.
type typedExternalClient struct {
	shared *secretversioncontroller.ExternalClient
}

func (t *typedExternalClient) Observe(ctx context.Context, cr *nativev1beta1.SecretVersionRAW) (managed.ExternalObservation, error) {
	return t.shared.Observe(ctx, cr)
}

func (t *typedExternalClient) Create(ctx context.Context, cr *nativev1beta1.SecretVersionRAW) (managed.ExternalCreation, error) {
	return t.shared.Create(ctx, cr)
}

func (t *typedExternalClient) Update(ctx context.Context, cr *nativev1beta1.SecretVersionRAW) (managed.ExternalUpdate, error) {
	return t.shared.Update(ctx, cr)
}

func (t *typedExternalClient) Delete(ctx context.Context, cr *nativev1beta1.SecretVersionRAW) (managed.ExternalDelete, error) {
	return t.shared.Delete(ctx, cr)
}

func (t *typedExternalClient) Disconnect(_ context.Context) error { return nil }
