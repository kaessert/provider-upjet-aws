// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package topic implements the shared CRUD logic for TopicRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the TopicCR interface.
package topic

import (
	"context"
	"errors"

	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native"
)

// TopicCR abstracts over cluster-scoped and namespaced TopicRAW types.
// Both scope types implement this interface so that the shared ExternalClient
// can operate on either without scope-specific logic.
type TopicCR interface {
	resource.Managed
	GetForProvider() *clusternative.TopicRAWParameters
	GetInitProvider() *clusternative.TopicRAWInitParameters
	GetAtProvider() clusternative.TopicRAWObservation
	SetAtProvider(clusternative.TopicRAWObservation)
}

// ExternalClient implements the shared CRUD logic for Topic resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS SNS SDK client.
	Client *awssns.Client
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external Topic resource exists and is up-to-date.
func (e *ExternalClient) Observe(_ context.Context, _ TopicCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented")
}

// Create creates the external Topic resource.
func (e *ExternalClient) Create(_ context.Context, _ TopicCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented")
}

// Update updates the external Topic resource.
func (e *ExternalClient) Update(_ context.Context, _ TopicCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented")
}

// Delete deletes the external Topic resource.
func (e *ExternalClient) Delete(_ context.Context, _ TopicCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented")
}
