// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package stream implements the shared CRUD logic for StreamRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the StreamCR interface.
package stream

import (
	"context"
	"errors"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta2native "github.com/upbound/provider-aws/v2/apis/cluster/kinesis/v1beta2/native"
)

// KinesisStreamClient is the interface for AWS Kinesis operations required by
// the stream controller. Methods will be added in the implement phase.
type KinesisStreamClient interface {
	// placeholder — no methods yet; will be populated in the implement phase
}

// StreamCR abstracts over cluster-scoped and namespaced StreamRAW types.
// Both scope types implement this interface so that the shared ExternalClient
// can operate on either without scope-specific logic.
type StreamCR interface {
	resource.Managed
	GetForProvider() *v1beta2native.StreamRAWParameters
	GetInitProvider() *v1beta2native.StreamRAWInitParameters
	GetAtProvider() v1beta2native.StreamRAWObservation
	SetAtProvider(v1beta2native.StreamRAWObservation)
}

// ExternalClient implements the shared CRUD logic for StreamRAW resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS Kinesis SDK client (interface for testability).
	Client KinesisStreamClient
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external Stream resource exists and is up-to-date.
// Not implemented yet — will be populated in the implement phase.
func (e *ExternalClient) Observe(_ context.Context, _ StreamCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented")
}

// Create creates the external Stream resource.
// Not implemented yet — will be populated in the implement phase.
func (e *ExternalClient) Create(_ context.Context, _ StreamCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented")
}

// Update updates the external Stream resource.
// Not implemented yet — will be populated in the implement phase.
func (e *ExternalClient) Update(_ context.Context, _ StreamCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented")
}

// Delete deletes the external Stream resource.
// Not implemented yet — will be populated in the implement phase.
func (e *ExternalClient) Delete(_ context.Context, _ StreamCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented")
}
