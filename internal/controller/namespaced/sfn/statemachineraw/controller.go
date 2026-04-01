// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

//go:build sfn || all

// Package statemachineraw contains the namespaced-scope native controller for
// StateMachineRAW resources. It is a thin wrapper that delegates all CRUD
// logic to the shared ExternalClient in internal/controller/sfn/statemachine.
package statemachineraw

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssfn "github.com/aws/aws-sdk-go-v2/service/sfn"
	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nativev1beta2 "github.com/upbound/provider-aws/v2/apis/namespaced/sfn/v1beta2/native"
	namespacedcontroller "github.com/upbound/provider-aws/v2/internal/controller/namespaced"
	statemachine "github.com/upbound/provider-aws/v2/internal/controller/sfn/statemachine"
	native "github.com/upbound/provider-aws/v2/internal/native"
)

func init() {
	// Register this Setup function into the NativeSetupHook for the sfn service.
	// This runs automatically when the package is imported by the provider binary.
	namespacedcontroller.NativeSetupHook_sfn = Setup
}

// Setup adds a controller that reconciles StateMachineRAW managed resources (namespaced).
func Setup(mgr ctrl.Manager, o xpcontroller.Options) error {
	name := managed.ControllerName(nativev1beta2.StateMachineRAW_GroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&nativev1beta2.StateMachineRAW{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(nativev1beta2.StateMachineRAW_GroupVersionKind),
			managed.WithTypedExternalConnector[*nativev1beta2.StateMachineRAW](
				native.NewTypedConnector(
					mgr.GetClient(),
					func(cfg aws.Config) *awssfn.Client {
						return awssfn.NewFromConfig(cfg)
					},
					func(c *awssfn.Client, kube client.Client) managed.TypedExternalClient[*nativev1beta2.StateMachineRAW] {
						return &typedExternalClient{
							shared: &statemachine.ExternalClient{Client: c, Kube: kube},
						}
					},
				),
			),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithPollInterval(o.PollInterval),
		))
}

// typedExternalClient bridges managed.TypedExternalClient[*nativev1beta2.StateMachineRAW]
// to the scope-agnostic shared ExternalClient.
type typedExternalClient struct {
	shared *statemachine.ExternalClient
}

func (t *typedExternalClient) Observe(ctx context.Context, cr *nativev1beta2.StateMachineRAW) (managed.ExternalObservation, error) {
	return t.shared.Observe(ctx, cr)
}

func (t *typedExternalClient) Create(ctx context.Context, cr *nativev1beta2.StateMachineRAW) (managed.ExternalCreation, error) {
	return t.shared.Create(ctx, cr)
}

func (t *typedExternalClient) Update(ctx context.Context, cr *nativev1beta2.StateMachineRAW) (managed.ExternalUpdate, error) {
	return t.shared.Update(ctx, cr)
}

func (t *typedExternalClient) Delete(ctx context.Context, cr *nativev1beta2.StateMachineRAW) (managed.ExternalDelete, error) {
	return t.shared.Delete(ctx, cr)
}

func (t *typedExternalClient) Disconnect(_ context.Context) error { return nil }
