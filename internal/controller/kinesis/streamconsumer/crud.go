// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package streamconsumer implements the shared CRUD logic for StreamConsumerRAW
// resources. It is scope-agnostic: both the cluster-scoped and namespaced
// controllers delegate to ExternalClient here via the StreamConsumerCR interface.
package streamconsumer

import (
	"context"
	"errors"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta1native "github.com/upbound/provider-aws/v2/apis/cluster/kinesis/v1beta1/native"
)

// KinesisConsumerClient is the interface for AWS Kinesis operations required by
// the stream consumer controller. Methods will be added in the implement phase.
type KinesisConsumerClient interface {
	// placeholder — no methods yet; will be populated in the implement phase
}

// StreamConsumerCR abstracts over cluster-scoped and namespaced StreamConsumerRAW types.
// Both scope types implement this interface so that the shared ExternalClient
// can operate on either without scope-specific logic.
type StreamConsumerCR interface {
	resource.Managed
	GetForProvider() *v1beta1native.StreamConsumerRAWParameters
	GetInitProvider() *v1beta1native.StreamConsumerRAWInitParameters
	GetAtProvider() v1beta1native.StreamConsumerRAWObservation
	SetAtProvider(v1beta1native.StreamConsumerRAWObservation)
}

// ExternalClient implements the shared CRUD logic for StreamConsumerRAW resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS Kinesis SDK client (interface for testability).
	Client KinesisConsumerClient
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external StreamConsumer resource exists and is up-to-date.
// Not implemented yet — will be populated in the implement phase.
func (e *ExternalClient) Observe(_ context.Context, _ StreamConsumerCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented")
}

// Create creates the external StreamConsumer resource.
// Not implemented yet — will be populated in the implement phase.
func (e *ExternalClient) Create(_ context.Context, _ StreamConsumerCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented")
}

// Update updates the external StreamConsumer resource.
// Not implemented yet — will be populated in the implement phase.
func (e *ExternalClient) Update(_ context.Context, _ StreamConsumerCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented")
}

// Delete deletes the external StreamConsumer resource.
// Not implemented yet — will be populated in the implement phase.
func (e *ExternalClient) Delete(_ context.Context, _ StreamConsumerCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented")
}
