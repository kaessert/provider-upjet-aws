// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package acl implements the shared CRUD logic for ACLRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the ACLCR interface.
package acl

import (
	"context"
	"errors"

	awsmemorydb "github.com/aws/aws-sdk-go-v2/service/memorydb"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/memorydb/v1beta1/native"
)

// MemoryDBACLClient is the interface for AWS MemoryDB operations
// required by the ACL controller. Defined as an interface to enable
// mocking in unit tests; *awsmemorydb.Client satisfies it.
type MemoryDBACLClient interface {
	DescribeACLs(ctx context.Context, params *awsmemorydb.DescribeACLsInput, optFns ...func(*awsmemorydb.Options)) (*awsmemorydb.DescribeACLsOutput, error)
	CreateACL(ctx context.Context, params *awsmemorydb.CreateACLInput, optFns ...func(*awsmemorydb.Options)) (*awsmemorydb.CreateACLOutput, error)
	UpdateACL(ctx context.Context, params *awsmemorydb.UpdateACLInput, optFns ...func(*awsmemorydb.Options)) (*awsmemorydb.UpdateACLOutput, error)
	DeleteACL(ctx context.Context, params *awsmemorydb.DeleteACLInput, optFns ...func(*awsmemorydb.Options)) (*awsmemorydb.DeleteACLOutput, error)
}

// ACLCR abstracts over cluster-scoped and namespaced ACLRAW types.
type ACLCR interface {
	resource.Managed
	GetForProvider() *clusternative.ACLRAWParameters
	SetForProvider(clusternative.ACLRAWParameters)
	GetInitProvider() *clusternative.ACLRAWInitParameters
	GetAtProvider() clusternative.ACLRAWObservation
	SetAtProvider(clusternative.ACLRAWObservation)
}

// ExternalClient implements the shared CRUD logic for ACLRAW resources.
type ExternalClient struct {
	// Client is the AWS MemoryDB SDK client (interface for testability).
	Client MemoryDBACLClient
	// Kube is the Kubernetes client (reserved for future use, e.g. secret reads).
	Kube client.Client
}

// Observe checks whether the external ACLRAW resource exists and is up-to-date.
func (e *ExternalClient) Observe(_ context.Context, _ ACLCR) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, errors.New("not implemented")
}

// Create creates the external ACLRAW resource.
func (e *ExternalClient) Create(_ context.Context, _ ACLCR) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, errors.New("not implemented")
}

// Update updates the external ACLRAW resource.
func (e *ExternalClient) Update(_ context.Context, _ ACLCR) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, errors.New("not implemented")
}

// Delete deletes the external ACLRAW resource.
func (e *ExternalClient) Delete(_ context.Context, _ ACLCR) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, errors.New("not implemented")
}
