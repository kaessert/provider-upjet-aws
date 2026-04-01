// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package queueredrivepolicy implements the shared CRUD logic for QueueRedrivePolicyRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the QueueRedrivePolicyCR interface.
package queueredrivepolicy

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
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
	errCreate   = "cannot set SQS queue redrive policy"
	errUpdate   = "cannot update SQS queue redrive policy"
	errDelete   = "cannot remove SQS queue redrive policy"
)

// SQSClient is the interface for AWS SQS operations required by this controller.
type SQSClient interface {
	GetQueueAttributes(ctx context.Context, params *awssqs.GetQueueAttributesInput, optFns ...func(*awssqs.Options)) (*awssqs.GetQueueAttributesOutput, error)
	SetQueueAttributes(ctx context.Context, params *awssqs.SetQueueAttributesInput, optFns ...func(*awssqs.Options)) (*awssqs.SetQueueAttributesOutput, error)
}

// QueueRedrivePolicyCR abstracts over cluster-scoped and namespaced QueueRedrivePolicyRAW types.
// Both scope types implement this interface so that the shared ExternalClient
// can operate on either without scope-specific logic.
type QueueRedrivePolicyCR interface {
	resource.Managed
	GetForProvider() *clusternative.QueueRedrivePolicyRAWParameters
	GetInitProvider() *clusternative.QueueRedrivePolicyRAWInitParameters
	GetAtProvider() clusternative.QueueRedrivePolicyRAWObservation
	SetAtProvider(clusternative.QueueRedrivePolicyRAWObservation)
}

// ExternalClient implements the shared CRUD logic for QueueRedrivePolicy resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS SQS SDK client (interface for testability).
	Client SQSClient
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external QueueRedrivePolicy resource exists and is up-to-date.
//
// QueueRedrivePolicy is not a standalone AWS resource — it is stored as a queue attribute.
// Observation works by calling GetQueueAttributes with the RedrivePolicy attribute name.
// If the queue does not exist, or the RedrivePolicy attribute is absent or empty,
// we treat the resource as not existing.
func (e *ExternalClient) Observe(ctx context.Context, cr QueueRedrivePolicyCR) (managed.ExternalObservation, error) {
	queueURL := nativehelper.GetExternalName(cr)

	// External-name guard: before first creation the annotation holds the K8s
	// resource name, not a URL. Only proceed when we have a real queue URL.
	if !strings.HasPrefix(queueURL, "https://sqs.") {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	resp, err := e.Client.GetQueueAttributes(ctx, &awssqs.GetQueueAttributesInput{
		QueueUrl:       &queueURL,
		AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameRedrivePolicy},
	})
	if err != nil {
		if isQueueDoesNotExist(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errGetAttrs)
	}

	// If the RedrivePolicy attribute is absent or empty, treat the policy as not existing.
	redrivePolicy, hasPolicy := resp.Attributes["RedrivePolicy"]
	if !hasPolicy || redrivePolicy == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Populate atProvider with observed state.
	qURL := queueURL
	rdp := redrivePolicy
	cr.SetAtProvider(clusternative.QueueRedrivePolicyRAWObservation{
		ID:            &qURL,
		QueueURL:      &qURL,
		RedrivePolicy: &rdp,
	})

	cr.SetConditions(xpv1.Available())

	upToDate := isUpToDate(cr, redrivePolicy)

	// Set the "Test=True" condition when the resource is annotated as a test
	// resource and is fully up-to-date. This allows uptest's
	// --default-conditions="Test" assertion to pass for native controllers.
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

// Create creates the external QueueRedrivePolicy resource by setting the RedrivePolicy
// queue attribute. The external name is set to the queue URL, since the redrive policy
// is an attribute of the queue.
func (e *ExternalClient) Create(ctx context.Context, cr QueueRedrivePolicyCR) (managed.ExternalCreation, error) {
	cr.SetConditions(xpv1.Creating())

	spec := cr.GetForProvider()
	queueURL := resolveQueueURL(cr)
	redrivePolicy := ""
	if spec.RedrivePolicy != nil {
		redrivePolicy = *spec.RedrivePolicy
	}

	if _, err := e.Client.SetQueueAttributes(ctx, &awssqs.SetQueueAttributesInput{
		QueueUrl:   &queueURL,
		Attributes: map[string]string{"RedrivePolicy": redrivePolicy},
	}); err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	// The external name for a queue redrive policy is the queue URL (the policy is
	// identified by the queue it belongs to, not a separate AWS resource ID).
	nativehelper.SetExternalName(cr, queueURL)

	return managed.ExternalCreation{}, nil
}

// Update updates the external QueueRedrivePolicy resource by setting the new
// RedrivePolicy attribute on the queue.
func (e *ExternalClient) Update(ctx context.Context, cr QueueRedrivePolicyCR) (managed.ExternalUpdate, error) {
	spec := cr.GetForProvider()
	queueURL := nativehelper.GetExternalName(cr)
	redrivePolicy := ""
	if spec.RedrivePolicy != nil {
		redrivePolicy = *spec.RedrivePolicy
	}

	if _, err := e.Client.SetQueueAttributes(ctx, &awssqs.SetQueueAttributesInput{
		QueueUrl:   &queueURL,
		Attributes: map[string]string{"RedrivePolicy": redrivePolicy},
	}); err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
	}

	return managed.ExternalUpdate{}, nil
}

// Delete removes the QueueRedrivePolicy by setting an empty RedrivePolicy attribute
// on the queue. This is idempotent: if the queue does not exist, we return nil.
func (e *ExternalClient) Delete(ctx context.Context, cr QueueRedrivePolicyCR) (managed.ExternalDelete, error) {
	cr.SetConditions(xpv1.Deleting())

	queueURL := nativehelper.GetExternalName(cr)
	if !strings.HasPrefix(queueURL, "https://sqs.") {
		return managed.ExternalDelete{}, nil
	}

	if _, err := e.Client.SetQueueAttributes(ctx, &awssqs.SetQueueAttributesInput{
		QueueUrl:   &queueURL,
		Attributes: map[string]string{"RedrivePolicy": ""},
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
func resolveQueueURL(cr QueueRedrivePolicyCR) string {
	spec := cr.GetForProvider()
	if spec.QueueURL != nil && *spec.QueueURL != "" {
		return *spec.QueueURL
	}
	return nativehelper.GetExternalName(cr)
}

// isUpToDate compares the desired spec redrive policy against the observed AWS policy.
// It uses JSON-semantic comparison (json.Unmarshal + reflect.DeepEqual) to avoid
// false positives from whitespace or key-order differences in AWS JSON normalisation.
func isUpToDate(cr QueueRedrivePolicyCR, observedPolicy string) bool {
	spec := cr.GetForProvider()
	if spec.RedrivePolicy == nil || *spec.RedrivePolicy == "" {
		return observedPolicy == ""
	}
	return jsonEqual(*spec.RedrivePolicy, observedPolicy)
}

// jsonEqual returns true if the two JSON strings are semantically equivalent.
// It unmarshals both into map[string]interface{} and uses reflect.DeepEqual.
// Falls back to raw string comparison if either value fails to unmarshal.
func jsonEqual(a, b string) bool {
	if a == b {
		return true
	}
	var aObj, bObj map[string]interface{}
	if err := json.Unmarshal([]byte(a), &aObj); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(b), &bObj); err != nil {
		return false
	}
	return reflect.DeepEqual(aObj, bObj)
}
