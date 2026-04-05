// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package secretrotation implements the shared CRUD logic for SecretRotationRAW
// resources. It is scope-agnostic: both the cluster-scoped (v1beta2) and
// namespaced (v1beta1) controllers delegate to ExternalClient here via the
// SecretRotationCR interface.
package secretrotation

import (
	"context"
	"errors"

	awssecretsmanager "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta1native "github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native"
)

// SecretRotationCR is the scope-agnostic interface that all SecretRotationRAW
// types (cluster v1beta2 and namespaced v1beta1) must implement. The shared
// ExternalClient operates exclusively through this interface, enabling a single
// CRUD implementation for both scopes.
//
// All methods use cluster-scoped v1beta1 parameter types as the canonical shape
// for CRUD logic. Types that use a different internal representation
// (v1beta2 pointer-based RotationRules, or namespaced NamespacedReference
// fields) perform the necessary conversion in GetForProvider/SetForProvider.
type SecretRotationCR interface {
	resource.Managed
	// GetForProvider returns the ForProvider parameters in v1beta1 cluster shape.
	// Implementations that use a different internal representation (e.g.
	// v1beta2 pointer RotationRules) return a freshly allocated copy.
	GetForProvider() *v1beta1native.SecretRotationRAWParameters
	// SetForProvider writes back mutated parameters (needed for late-init).
	// Implementations with converted GetForProvider MUST implement this.
	SetForProvider(v1beta1native.SecretRotationRAWParameters)
	// GetInitProvider returns the InitProvider parameters in v1beta1 cluster shape.
	GetInitProvider() *v1beta1native.SecretRotationRAWInitParameters
	// GetAtProvider returns the current observed state.
	GetAtProvider() v1beta1native.SecretRotationRAWObservation
	// SetAtProvider updates the observed state.
	SetAtProvider(v1beta1native.SecretRotationRAWObservation)
}

// SecretsManagerRotationClient is the AWS Secrets Manager API client interface
// for rotation operations. Defined here for future use — currently all methods
// return not-implemented.
type SecretsManagerRotationClient interface{}

// ExternalClient implements managed.TypedExternalClient[SecretRotationCR] using
// the AWS Secrets Manager SDK. CRUD methods are stubs until phase:implement.
type ExternalClient struct {
	Client *awssecretsmanager.Client
	Kube   client.Client
}

// Observe checks whether the secret rotation exists in AWS.
func (e *ExternalClient) Observe(_ context.Context, _ SecretRotationCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented")
}

// Create creates the secret rotation in AWS.
func (e *ExternalClient) Create(_ context.Context, _ SecretRotationCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented")
}

// Update updates the secret rotation in AWS.
func (e *ExternalClient) Update(_ context.Context, _ SecretRotationCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented")
}

// Delete deletes the secret rotation from AWS.
func (e *ExternalClient) Delete(_ context.Context, _ SecretRotationCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented")
}
