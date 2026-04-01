// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package queuepolicy implements the shared CRUD logic for QueuePolicyRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the QueuePolicyCR interface.
package queuepolicy

import (
	"context"
	"errors"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
)

// SQSClient is the interface for AWS SQS operations required by this controller.
type SQSClient interface {
	GetQueueAttributes(ctx context.Context, params *awssqs.GetQueueAttributesInput, optFns ...func(*awssqs.Options)) (*awssqs.GetQueueAttributesOutput, error)
	SetQueueAttributes(ctx context.Context, params *awssqs.SetQueueAttributesInput, optFns ...func(*awssqs.Options)) (*awssqs.SetQueueAttributesOutput, error)
}

// QueuePolicyCR abstracts over cluster-scoped and namespaced QueuePolicyRAW types.
// Both scope types implement this interface so that the shared ExternalClient
// can operate on either without scope-specific logic.
type QueuePolicyCR interface {
	resource.Managed
	GetForProvider() *clusternative.QueuePolicyRAWParameters
	GetInitProvider() *clusternative.QueuePolicyRAWInitParameters
	GetAtProvider() clusternative.QueuePolicyRAWObservation
	SetAtProvider(clusternative.QueuePolicyRAWObservation)
}

// ExternalClient implements the shared CRUD logic for QueuePolicy resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS SQS SDK client (interface for testability).
	Client SQSClient
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external QueuePolicy resource exists and is up-to-date.
func (e *ExternalClient) Observe(_ context.Context, _ QueuePolicyCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented")
}

// Create creates the external QueuePolicy resource.
func (e *ExternalClient) Create(_ context.Context, _ QueuePolicyCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented")
}

// Update updates the external QueuePolicy resource.
func (e *ExternalClient) Update(_ context.Context, _ QueuePolicyCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented")
}

// Delete deletes the external QueuePolicy resource.
func (e *ExternalClient) Delete(_ context.Context, _ QueuePolicyCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented")
}
