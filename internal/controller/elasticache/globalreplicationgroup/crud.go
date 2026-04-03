// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package globalreplicationgroup implements the shared CRUD logic for GlobalReplicationGroupRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the GlobalReplicationGroupCR interface.
package globalreplicationgroup

import (
	"context"
	"errors"

	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native"
)

// Client interface for AWS operations.
type Client interface {
	// Placeholder — implement in implement phase.
}

// GlobalReplicationGroupCR abstracts over cluster-scoped and namespaced GlobalReplicationGroupRAW types.
type GlobalReplicationGroupCR interface {
	resource.Managed
	GetForProvider() *clusternative.GlobalReplicationGroupRAWParameters
	GetInitProvider() *clusternative.GlobalReplicationGroupRAWInitParameters
	GetAtProvider() clusternative.GlobalReplicationGroupRAWObservation
	SetAtProvider(clusternative.GlobalReplicationGroupRAWObservation)
}

// ExternalClient implements the shared CRUD logic for GlobalReplicationGroupRAW resources.
type ExternalClient struct {
	Client *awselasticache.Client
	Kube   client.Client
}

// Observe checks whether the external GlobalReplicationGroupRAW resource exists and is up-to-date.
func (e *ExternalClient) Observe(ctx context.Context, cr GlobalReplicationGroupCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented — implement in implement phase")
}

// Create creates the external GlobalReplicationGroupRAW resource.
func (e *ExternalClient) Create(ctx context.Context, cr GlobalReplicationGroupCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented — implement in implement phase")
}

// Update updates the external GlobalReplicationGroupRAW resource.
func (e *ExternalClient) Update(ctx context.Context, cr GlobalReplicationGroupCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented — implement in implement phase")
}

// Delete deletes the external GlobalReplicationGroupRAW resource.
func (e *ExternalClient) Delete(ctx context.Context, cr GlobalReplicationGroupCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented — implement in implement phase")
}
