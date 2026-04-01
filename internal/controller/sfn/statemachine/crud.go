// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package statemachine implements the shared CRUD logic for StateMachineRAW
// resources. It is scope-agnostic: both the cluster-scoped and namespaced
// controllers delegate to ExternalClient here via the StateMachineCR interface.
package statemachine

import (
	"context"
	"errors"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	awssfn "github.com/aws/aws-sdk-go-v2/service/sfn"

	v1beta2native "github.com/upbound/provider-aws/v2/apis/cluster/sfn/v1beta2/native"
)

// StateMachineCR abstracts over cluster-scoped and namespaced StateMachineRAW
// types. Both scope types implement this interface so that the shared
// ExternalClient can operate on either without scope-specific logic.
type StateMachineCR interface {
	resource.Managed
	GetForProvider() *v1beta2native.StateMachineRAWParameters
	GetInitProvider() *v1beta2native.StateMachineRAWInitParameters
	GetAtProvider() v1beta2native.StateMachineRAWObservation
	SetAtProvider(v1beta2native.StateMachineRAWObservation)
}

// ExternalClient implements the shared CRUD logic for StateMachine resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS Step Functions SDK client.
	Client *awssfn.Client
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external StateMachine resource exists and is
// up-to-date with the desired state.
func (e *ExternalClient) Observe(_ context.Context, _ StateMachineCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented")
}

// Create creates the external StateMachine resource.
func (e *ExternalClient) Create(_ context.Context, _ StateMachineCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented")
}

// Update updates the external StateMachine resource.
func (e *ExternalClient) Update(_ context.Context, _ StateMachineCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented")
}

// Delete deletes the external StateMachine resource.
func (e *ExternalClient) Delete(_ context.Context, _ StateMachineCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented")
}
