// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package secret implements the shared CRUD logic for SecretRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the SecretCR interface.
package secret

import (
	"context"
	"errors"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta1native "github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native"
)

// SecretCR is the scope-agnostic interface that both the cluster-scoped and
// namespaced SecretRAW types must implement. The shared ExternalClient
// operates exclusively through this interface, enabling a single CRUD
// implementation for both scopes.
type SecretCR interface {
	resource.Managed
	// GetForProvider returns the ForProvider parameters (cluster-scoped type).
	// Namespaced implementations return a freshly allocated copy.
	GetForProvider() *v1beta1native.SecretRAWParameters
	// SetForProvider writes back mutated parameters (needed for late-init).
	// Namespaced implementations copy each field individually.
	SetForProvider(v1beta1native.SecretRAWParameters)
	// GetInitProvider returns the InitProvider parameters (cluster-scoped type).
	GetInitProvider() *v1beta1native.SecretRAWInitParameters
	// GetAtProvider returns the current observed state.
	GetAtProvider() v1beta1native.SecretRAWObservation
	// SetAtProvider updates the observed state.
	SetAtProvider(v1beta1native.SecretRAWObservation)
}

// SecretsManagerClient is the AWS Secrets Manager API client interface.
// Defined here for future use — currently all methods return not-implemented.
type SecretsManagerClient interface{}

// ExternalClient implements managed.TypedExternalClient[SecretCR] using
// the AWS Secrets Manager SDK. CRUD methods are stubs until phase:implement.
type ExternalClient struct {
	Client SecretsManagerClient
	Kube   client.Client
}

// Observe checks whether the secret exists in AWS.
func (e *ExternalClient) Observe(_ context.Context, _ SecretCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented")
}

// Create creates the secret in AWS.
func (e *ExternalClient) Create(_ context.Context, _ SecretCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented")
}

// Update updates the secret in AWS.
func (e *ExternalClient) Update(_ context.Context, _ SecretCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented")
}

// Delete deletes the secret from AWS.
func (e *ExternalClient) Delete(_ context.Context, _ SecretCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented")
}
