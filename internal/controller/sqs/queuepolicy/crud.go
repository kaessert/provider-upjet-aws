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
	"strings"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
	nativehelper "github.com/upbound/provider-aws/v2/internal/native"
)

const (
	errGetAttrs = "cannot get SQS queue attributes"
	errCreate   = "cannot set SQS queue policy"
	errUpdate   = "cannot update SQS queue policy"
	errDelete   = "cannot remove SQS queue policy"
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
//
// QueuePolicy is not a standalone AWS resource — it is stored as a queue attribute.
// Observation works by calling GetQueueAttributes with the Policy attribute name.
// If the queue does not exist, or the Policy attribute is absent (empty policy),
// we treat the resource as not existing.
func (e *ExternalClient) Observe(ctx context.Context, cr QueuePolicyCR) (managed.ExternalObservation, error) {
	queueURL := nativehelper.GetExternalName(cr)

	// External-name guard: before first creation the annotation holds the K8s
	// resource name, not a URL. Only proceed when we have a real queue URL.
	if !strings.HasPrefix(queueURL, "https://sqs.") {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	resp, err := e.Client.GetQueueAttributes(ctx, &awssqs.GetQueueAttributesInput{
		QueueUrl:       &queueURL,
		AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNamePolicy},
	})
	if err != nil {
		if isQueueDoesNotExist(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errGetAttrs)
	}

	// If the Policy attribute is absent or empty, treat the policy as not existing.
	policy, hasPolicy := resp.Attributes["Policy"]
	if !hasPolicy || policy == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Populate atProvider with observed state.
	qURL := queueURL
	cr.SetAtProvider(clusternative.QueuePolicyRAWObservation{
		ID:       &qURL,
		QueueURL: &qURL,
	})

	cr.SetConditions(xpv1.Available())

	upToDate := isUpToDate(cr, policy)

	// Set the "Test=True" condition when the resource is annotated as a test
	// resource and is fully up-to-date. This allows uptest's
	// --default-conditions="Test" assertion to pass for native controllers.
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

// Create creates the external QueuePolicy resource by setting the Policy queue attribute.
// The external name is set to the queue URL, since the policy is an attribute of the queue.
func (e *ExternalClient) Create(ctx context.Context, cr QueuePolicyCR) (managed.ExternalCreation, error) {
	cr.SetConditions(xpv1.Creating())

	spec := cr.GetForProvider()
	queueURL := resolveQueueURL(cr)

	// Guard: if the queue URL hasn't been resolved yet (reference to a QueueRAW
	// that is not yet Ready), return early rather than calling AWS with an invalid
	// address. The reconciler will retry after the poll interval.
	if !strings.HasPrefix(queueURL, "https://sqs.") {
		return managed.ExternalCreation{}, nativehelper.Wrap(
			errors.New("queue URL not yet resolved; waiting for referenced QueueRAW to become ready"),
			errCreate,
		)
	}

	policy := ""
	if spec.Policy != nil {
		policy = *spec.Policy
	}

	if _, err := e.Client.SetQueueAttributes(ctx, &awssqs.SetQueueAttributesInput{
		QueueUrl:   &queueURL,
		Attributes: map[string]string{"Policy": policy},
	}); err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// The external name for a queue policy is the queue URL (the policy is
	// identified by the queue it belongs to, not a separate AWS resource ID).
	nativehelper.SetExternalName(cr, queueURL)

	return managed.ExternalCreation{}, nil
}

// Update updates the external QueuePolicy resource by setting the new Policy attribute.
func (e *ExternalClient) Update(ctx context.Context, cr QueuePolicyCR) (managed.ExternalUpdate, error) {
	spec := cr.GetForProvider()
	queueURL := nativehelper.GetExternalName(cr)
	policy := ""
	if spec.Policy != nil {
		policy = *spec.Policy
	}

	if _, err := e.Client.SetQueueAttributes(ctx, &awssqs.SetQueueAttributesInput{
		QueueUrl:   &queueURL,
		Attributes: map[string]string{"Policy": policy},
	}); err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
	}

	return managed.ExternalUpdate{}, nil
}

// Delete removes the QueuePolicy by setting an empty Policy attribute on the queue.
// This is idempotent: if the queue does not exist, we return nil.
func (e *ExternalClient) Delete(ctx context.Context, cr QueuePolicyCR) (managed.ExternalDelete, error) {
	cr.SetConditions(xpv1.Deleting())

	queueURL := nativehelper.GetExternalName(cr)
	if !strings.HasPrefix(queueURL, "https://sqs.") {
		return managed.ExternalDelete{}, nil
	}

	if _, err := e.Client.SetQueueAttributes(ctx, &awssqs.SetQueueAttributesInput{
		QueueUrl:   &queueURL,
		Attributes: map[string]string{"Policy": ""},
	}); err != nil {
		if isQueueDoesNotExist(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, nativehelper.Wrap(err, errDelete)
	}

	return managed.ExternalDelete{}, nil
}

// ── helpers ────────────────────────────────────────────────────────────────────

// isQueueDoesNotExist returns true if err is a QueueDoesNotExist AWS error.
func isQueueDoesNotExist(err error) bool {
	var notFound *sqstypes.QueueDoesNotExist
	return errors.As(err, &notFound)
}

// resolveQueueURL returns the queue URL from the spec ForProvider or from the
// external name annotation (whichever is populated). ForProvider.QueueURL is
// populated by the reference resolver before CRUD methods are called.
func resolveQueueURL(cr QueuePolicyCR) string {
	spec := cr.GetForProvider()
	if spec.QueueURL != nil && *spec.QueueURL != "" {
		return *spec.QueueURL
	}
	return nativehelper.GetExternalName(cr)
}

// isUpToDate compares the desired spec policy against the observed AWS policy.
// It uses semantic IAM policy comparison to avoid false positives from
// whitespace or key-order differences introduced by AWS JSON normalisation.
func isUpToDate(cr QueuePolicyCR, observedPolicy string) bool {
	spec := cr.GetForProvider()
	if spec.Policy == nil || *spec.Policy == "" {
		// No desired policy — if AWS has one, it's not up to date.
		return observedPolicy == ""
	}
	// Use semantic IAM policy comparison — AWS may return differently formatted JSON.
	return !nativehelper.PolicyNeedsUpdate(*spec.Policy, observedPolicy)
}
