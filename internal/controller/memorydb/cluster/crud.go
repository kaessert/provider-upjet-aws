// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package cluster implements the shared CRUD logic for ClusterRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the ClusterCR interface.
package cluster

import (
	"context"
	"errors"

	awsmemorydb "github.com/aws/aws-sdk-go-v2/service/memorydb"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/memorydb/v1beta1/native"
)

// MemoryDBClusterClient is the interface for AWS MemoryDB operations
// required by the Cluster controller. Defined as an interface to enable
// mocking in unit tests; *awsmemorydb.Client satisfies it.
type MemoryDBClusterClient interface {
	DescribeClusters(ctx context.Context, params *awsmemorydb.DescribeClustersInput, optFns ...func(*awsmemorydb.Options)) (*awsmemorydb.DescribeClustersOutput, error)
	CreateCluster(ctx context.Context, params *awsmemorydb.CreateClusterInput, optFns ...func(*awsmemorydb.Options)) (*awsmemorydb.CreateClusterOutput, error)
	UpdateCluster(ctx context.Context, params *awsmemorydb.UpdateClusterInput, optFns ...func(*awsmemorydb.Options)) (*awsmemorydb.UpdateClusterOutput, error)
	DeleteCluster(ctx context.Context, params *awsmemorydb.DeleteClusterInput, optFns ...func(*awsmemorydb.Options)) (*awsmemorydb.DeleteClusterOutput, error)
}

// ClusterCR abstracts over cluster-scoped and namespaced ClusterRAW types.
type ClusterCR interface {
	resource.Managed
	GetForProvider() *clusternative.ClusterRAWParameters
	SetForProvider(clusternative.ClusterRAWParameters)
	GetInitProvider() *clusternative.ClusterRAWInitParameters
	GetAtProvider() clusternative.ClusterRAWObservation
	SetAtProvider(clusternative.ClusterRAWObservation)
}

// ExternalClient implements the shared CRUD logic for ClusterRAW resources.
type ExternalClient struct {
	// Client is the AWS MemoryDB SDK client (interface for testability).
	Client MemoryDBClusterClient
	// Kube is the Kubernetes client (reserved for future use, e.g. secret reads).
	Kube client.Client
}

// Observe checks whether the external ClusterRAW resource exists and is up-to-date.
func (e *ExternalClient) Observe(_ context.Context, _ ClusterCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented")
}

// Create creates the external ClusterRAW resource.
func (e *ExternalClient) Create(_ context.Context, _ ClusterCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented")
}

// Update updates the external ClusterRAW resource.
func (e *ExternalClient) Update(_ context.Context, _ ClusterCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented")
}

// Delete deletes the external ClusterRAW resource.
func (e *ExternalClient) Delete(_ context.Context, _ ClusterCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented")
}
