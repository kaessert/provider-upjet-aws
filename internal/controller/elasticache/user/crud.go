// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package user implements the shared CRUD logic for UserRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the UserCR interface.
package user

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

// UserCR abstracts over cluster-scoped and namespaced UserRAW types.
type UserCR interface {
	resource.Managed
	GetForProvider() *clusternativev2.UserRAWParameters
	GetInitProvider() *clusternativev2.UserRAWInitParameters
	GetAtProvider() clusternativev2.UserRAWObservation
	SetAtProvider(clusternativev2.UserRAWObservation)
}

// ExternalClient implements the shared CRUD logic for UserRAW resources.
type ExternalClient struct {
	Client *awselasticache.Client
	Kube   client.Client
}

// Observe checks whether the external UserRAW resource exists and is up-to-date.
func (e *ExternalClient) Observe(ctx context.Context, cr UserCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented — implement in implement phase")
}

// Create creates the external UserRAW resource.
func (e *ExternalClient) Create(ctx context.Context, cr UserCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented — implement in implement phase")
}

// Update updates the external UserRAW resource.
func (e *ExternalClient) Update(ctx context.Context, cr UserCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented — implement in implement phase")
}

// Delete deletes the external UserRAW resource.
func (e *ExternalClient) Delete(ctx context.Context, cr UserCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented — implement in implement phase")
}
