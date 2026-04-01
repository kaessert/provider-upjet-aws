// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package queue implements the shared CRUD logic for QueueRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the QueueCR interface.
package queue

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
	CreateQueue(ctx context.Context, params *awssqs.CreateQueueInput, optFns ...func(*awssqs.Options)) (*awssqs.CreateQueueOutput, error)
	DeleteQueue(ctx context.Context, params *awssqs.DeleteQueueInput, optFns ...func(*awssqs.Options)) (*awssqs.DeleteQueueOutput, error)
	GetQueueAttributes(ctx context.Context, params *awssqs.GetQueueAttributesInput, optFns ...func(*awssqs.Options)) (*awssqs.GetQueueAttributesOutput, error)
	GetQueueUrl(ctx context.Context, params *awssqs.GetQueueUrlInput, optFns ...func(*awssqs.Options)) (*awssqs.GetQueueUrlOutput, error)
	SetQueueAttributes(ctx context.Context, params *awssqs.SetQueueAttributesInput, optFns ...func(*awssqs.Options)) (*awssqs.SetQueueAttributesOutput, error)
	ListQueueTags(ctx context.Context, params *awssqs.ListQueueTagsInput, optFns ...func(*awssqs.Options)) (*awssqs.ListQueueTagsOutput, error)
	TagQueue(ctx context.Context, params *awssqs.TagQueueInput, optFns ...func(*awssqs.Options)) (*awssqs.TagQueueOutput, error)
	UntagQueue(ctx context.Context, params *awssqs.UntagQueueInput, optFns ...func(*awssqs.Options)) (*awssqs.UntagQueueOutput, error)
}

// QueueCR abstracts over cluster-scoped and namespaced QueueRAW types.
// Both scope types implement this interface so that the shared ExternalClient
// can operate on either without scope-specific logic.
type QueueCR interface {
	resource.Managed
	GetForProvider() *clusternative.QueueRAWParameters
	GetInitProvider() *clusternative.QueueRAWInitParameters
	GetAtProvider() clusternative.QueueRAWObservation
	SetAtProvider(clusternative.QueueRAWObservation)
}

// ExternalClient implements the shared CRUD logic for Queue resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS SQS SDK client (interface for testability).
	Client SQSClient
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external Queue resource exists and is up-to-date.
func (e *ExternalClient) Observe(_ context.Context, _ QueueCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented")
}

// Create creates the external Queue resource.
func (e *ExternalClient) Create(_ context.Context, _ QueueCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented")
}

// Update updates the external Queue resource.
func (e *ExternalClient) Update(_ context.Context, _ QueueCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented")
}

// Delete deletes the external Queue resource.
func (e *ExternalClient) Delete(_ context.Context, _ QueueCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented")
}
