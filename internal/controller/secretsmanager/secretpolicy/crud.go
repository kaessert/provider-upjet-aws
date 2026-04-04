// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package secretpolicy implements the shared CRUD logic for SecretPolicyRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the SecretPolicyCR interface.
package secretpolicy

import (
	"context"
	"errors"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta1native "github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native"
)

// SecretPolicyCR is the scope-agnostic interface that both the cluster-scoped and
// namespaced SecretPolicyRAW types must implement. The shared ExternalClient
// operates exclusively through this interface, enabling a single CRUD
// implementation for both scopes.
type SecretPolicyCR interface {
	resource.Managed
	// GetForProvider returns the ForProvider parameters (cluster-scoped type).
	// Namespaced implementations return a freshly allocated copy.
	GetForProvider() *v1beta1native.SecretPolicyRAWParameters
	// SetForProvider writes back mutated parameters (needed for late-init).
	// Namespaced implementations copy each field individually.
	SetForProvider(v1beta1native.SecretPolicyRAWParameters)
	// GetInitProvider returns the InitProvider parameters (cluster-scoped type).
	GetInitProvider() *v1beta1native.SecretPolicyRAWInitParameters
	// GetAtProvider returns the current observed state.
	GetAtProvider() v1beta1native.SecretPolicyRAWObservation
	// SetAtProvider updates the observed state.
	SetAtProvider(v1beta1native.SecretPolicyRAWObservation)
}

// SecretsManagerClient is the AWS Secrets Manager API client interface.
// Defined here for future use — currently all methods return not-implemented.
type SecretsManagerClient interface{}

// ExternalClient implements managed.TypedExternalClient[SecretPolicyCR] using
// the AWS Secrets Manager SDK. CRUD methods are stubs until phase:implement.
type ExternalClient struct {
	Client SecretsManagerClient
	Kube   client.Client
}

// Observe checks whether the secret policy exists in AWS.
func (e *ExternalClient) Observe(_ context.Context, _ SecretPolicyCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented")
}

// Create creates the secret policy in AWS.
func (e *ExternalClient) Create(_ context.Context, _ SecretPolicyCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented")
}

// Update updates the secret policy in AWS.
func (e *ExternalClient) Update(_ context.Context, _ SecretPolicyCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented")
}

// Delete deletes the secret policy from AWS.
func (e *ExternalClient) Delete(_ context.Context, _ SecretPolicyCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented")
}
