// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package replicationgroup implements the shared CRUD logic for ReplicationGroupRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the ReplicationGroupCR interface.
package replicationgroup

import (
	"context"
	"errors"

	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternativev2 "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native"
)

// Client interface for AWS operations.
type Client interface {
	// Placeholder — implement in implement phase.
}

// ReplicationGroupCR abstracts over cluster-scoped and namespaced ReplicationGroupRAW types.
type ReplicationGroupCR interface {
	resource.Managed
	GetForProvider() *clusternativev2.ReplicationGroupRAWParameters
	GetInitProvider() *clusternativev2.ReplicationGroupRAWInitParameters
	GetAtProvider() clusternativev2.ReplicationGroupRAWObservation
	SetAtProvider(clusternativev2.ReplicationGroupRAWObservation)
}

// ExternalClient implements the shared CRUD logic for ReplicationGroupRAW resources.
type ExternalClient struct {
	Client *awselasticache.Client
	Kube   client.Client
}

// Observe checks whether the external ReplicationGroupRAW resource exists and is up-to-date.
func (e *ExternalClient) Observe(ctx context.Context, cr ReplicationGroupCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented — implement in implement phase")
}

// Create creates the external ReplicationGroupRAW resource.
func (e *ExternalClient) Create(ctx context.Context, cr ReplicationGroupCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented — implement in implement phase")
}

// Update updates the external ReplicationGroupRAW resource.
func (e *ExternalClient) Update(ctx context.Context, cr ReplicationGroupCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented — implement in implement phase")
}

// Delete deletes the external ReplicationGroupRAW resource.
func (e *ExternalClient) Delete(ctx context.Context, cr ReplicationGroupCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented — implement in implement phase")
}
