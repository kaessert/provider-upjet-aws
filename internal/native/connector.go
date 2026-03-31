// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/upbound/provider-aws/v2/internal/clients"
)

// ConfigResolverFn is the function signature for resolving an aws.Config from a
// managed resource. It is abstracted to allow test mocking without real k8s or
// AWS credentials.
type ConfigResolverFn func(ctx context.Context, c client.Client, mg resource.Managed) (*aws.Config, error)

// TypedConnector implements managed.TypedExternalConnector[T].
// It resolves AWS credentials via a ProviderConfig and constructs a typed
// external client that operates against a specific AWS service client C.
//
// T is the concrete managed resource type (e.g. *BucketRAW).
// C is the AWS service client type (e.g. *s3.Client).
type TypedConnector[T resource.Managed, C any] struct {
	// kube is forwarded to GetAWSConfigWithTracking and to the newExternal
	// factory so that controllers can perform k8s reads during reconciliation.
	kube client.Client

	// clientFactory constructs the AWS service client from the resolved config.
	clientFactory func(cfg aws.Config) C

	// newExternal constructs the typed external client from the AWS service
	// client and the kube client.
	newExternal func(svcClient C, kube client.Client) managed.TypedExternalClient[T]

	// configFn resolves the AWS config. Defaults to
	// clients.GetAWSConfigWithTracking; overridable for testing.
	configFn ConfigResolverFn
}

// NewTypedConnector creates a TypedConnector that uses the standard
// clients.GetAWSConfigWithTracking resolver. This is the production constructor.
//
//   connector := native.NewTypedConnector(mgr.GetClient(), s3.NewFromConfig, newExternal)
func NewTypedConnector[T resource.Managed, C any](
	kube client.Client,
	clientFactory func(aws.Config) C,
	newExternal func(C, client.Client) managed.TypedExternalClient[T],
) *TypedConnector[T, C] {
	return &TypedConnector[T, C]{
		kube:          kube,
		clientFactory: clientFactory,
		newExternal:   newExternal,
		configFn:      clients.GetAWSConfigWithTracking,
	}
}

// Connect implements managed.TypedExternalConnector[T].
// It resolves the AWS config for the managed resource, builds the service
// client, and returns a TypedExternalClient ready for CRUD operations.
func (c *TypedConnector[T, C]) Connect(ctx context.Context, mg T) (managed.TypedExternalClient[T], error) {
	cfg, err := c.configFn(ctx, c.kube, mg)
	if err != nil {
		return nil, Wrap(err, "cannot get aws config")
	}
	svcClient := c.clientFactory(*cfg)
	return c.newExternal(svcClient, c.kube), nil
}
