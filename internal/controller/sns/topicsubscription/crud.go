// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package topicsubscription implements the shared CRUD logic for TopicSubscriptionRAW resources.
// It is scope-agnostic: both the cluster-scoped and namespaced controllers
// delegate to ExternalClient here via the TopicSubscriptionCR interface.
package topicsubscription

import (
	"context"
	"errors"

	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native"
	nativehelper "github.com/upbound/provider-aws/v2/internal/native"
)

const (
	errGetAttrs = "cannot get SNS subscription attributes"
	errCreate   = "cannot subscribe to SNS topic"
	errUpdate   = "cannot set SNS subscription attribute"
	errDelete   = "cannot unsubscribe from SNS topic"

	// attrTrue / attrFalse are the string representations of booleans used
	// by the SNS attribute-based API.
	attrTrue  = "true"
	attrFalse = "false"

	// pendingConfirmationARN is the special ARN value AWS returns when a
	// subscription requires external confirmation (e.g. HTTP/HTTPS/email).
	pendingConfirmationARN = "pending confirmation"
)

// SNSSubscriptionClient is the interface for AWS SNS subscription operations
// required by this controller. Defining it as an interface enables mocking in
// unit tests.
type SNSSubscriptionClient interface {
	Subscribe(ctx context.Context, params *awssns.SubscribeInput, optFns ...func(*awssns.Options)) (*awssns.SubscribeOutput, error)
	GetSubscriptionAttributes(ctx context.Context, params *awssns.GetSubscriptionAttributesInput, optFns ...func(*awssns.Options)) (*awssns.GetSubscriptionAttributesOutput, error)
	SetSubscriptionAttributes(ctx context.Context, params *awssns.SetSubscriptionAttributesInput, optFns ...func(*awssns.Options)) (*awssns.SetSubscriptionAttributesOutput, error)
	Unsubscribe(ctx context.Context, params *awssns.UnsubscribeInput, optFns ...func(*awssns.Options)) (*awssns.UnsubscribeOutput, error)
}

// Compile-time assertion that *awssns.Client satisfies SNSSubscriptionClient.
var _ SNSSubscriptionClient = (*awssns.Client)(nil)

// TopicSubscriptionCR abstracts over cluster-scoped and namespaced TopicSubscriptionRAW types.
// Both scope types implement this interface so that the shared ExternalClient
// can operate on either without scope-specific logic.
type TopicSubscriptionCR interface {
	resource.Managed
	GetForProvider() *clusternative.TopicSubscriptionRAWParameters
	GetInitProvider() *clusternative.TopicSubscriptionRAWInitParameters
	GetAtProvider() clusternative.TopicSubscriptionRAWObservation
	SetAtProvider(clusternative.TopicSubscriptionRAWObservation)
	// SetForProviderFilterPolicyScope is needed for late-initialization of AWS-defaulted
	// fields. GetForProvider() returns a field-copied struct for namespaced scope,
	// so mutations through the returned pointer do not propagate back to the spec.
	SetForProviderFilterPolicyScope(*string)
}

// ExternalClient implements the shared CRUD logic for TopicSubscription resources.
// It is exported so that scope-specific wrapper packages can reference it.
type ExternalClient struct {
	// Client is the AWS SNS SDK client (interface for testability).
	Client SNSSubscriptionClient
	// Kube is the Kubernetes client for reading referenced resources.
	Kube client.Client
}

// Observe checks whether the external TopicSubscription resource exists and is up-to-date.
func (e *ExternalClient) Observe(ctx context.Context, cr TopicSubscriptionCR) (managed.ExternalObservation, error) {
	subARN := subscriptionARN(cr)
	if subARN == "" {
		// External name not set yet — resource hasn't been created.
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	// Special-case: subscription created but awaiting external confirmation.
	if subARN == pendingConfirmationARN {
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
		}, nil
	}

	resp, err := e.Client.GetSubscriptionAttributes(ctx, &awssns.GetSubscriptionAttributesInput{
		SubscriptionArn: &subARN,
	})
	if err != nil {
		if isSubscriptionNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, nativehelper.Wrap(err, errGetAttrs)
	}

	attrs := resp.Attributes

	// Populate status.atProvider from the attributes.
	cr.SetAtProvider(mapAttrsToObservation(attrs))

	// Late-initialize AWS-defaulted fields.
	lateInited := e.lateInitialize(cr, attrs)

	// If the subscription is still pending external confirmation, mark it
	// unavailable but report it as existing so the reconciler does not
	// attempt to re-create it.
	if attrs["PendingConfirmation"] == attrTrue {
		cr.SetConditions(xpv1.Unavailable())
		return managed.ExternalObservation{
			ResourceExists:          true,
			ResourceUpToDate:        true,
			ResourceLateInitialized: lateInited,
		}, nil
	}

	cr.SetConditions(xpv1.Available())

	upToDate := isUpToDate(cr, attrs)

	// Set the "Test=True" condition when the resource is annotated as a test
	// resource and is fully up-to-date. This allows uptest's
	// --default-conditions="Test" assertion to pass for native controllers.
	nativehelper.SetTestConditionIfAnnotated(cr, upToDate)

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		ResourceLateInitialized: lateInited,
	}, nil
}

// Create subscribes an endpoint to the SNS topic.
func (e *ExternalClient) Create(ctx context.Context, cr TopicSubscriptionCR) (managed.ExternalCreation, error) {
	cr.SetConditions(xpv1.Creating())

	spec := cr.GetForProvider()

	input := &awssns.SubscribeInput{
		Protocol:   spec.Protocol,
		TopicArn:   spec.TopicArn,
		Endpoint:   spec.Endpoint,
		Attributes: buildSubscribeAttributes(spec),
	}

	resp, err := e.Client.Subscribe(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, nativehelper.Wrap(err, errCreate)
	}

	subARN := ""
	if resp.SubscriptionArn != nil {
		subARN = *resp.SubscriptionArn
	}

	// Store the subscription ARN (or "pending confirmation") as the external name.
	// The external name is the primary identifier for Observe/Update/Delete.
	meta.SetExternalName(cr, subARN)

	return managed.ExternalCreation{}, nil
}

// Update updates mutable attributes of the SNS subscription.
func (e *ExternalClient) Update(ctx context.Context, cr TopicSubscriptionCR) (managed.ExternalUpdate, error) {
	subARN := subscriptionARN(cr)
	if subARN == "" || subARN == pendingConfirmationARN {
		// Nothing to update until the subscription has a real ARN.
		return managed.ExternalUpdate{}, nil
	}

	// Fetch current attributes to compare against spec.
	resp, err := e.Client.GetSubscriptionAttributes(ctx, &awssns.GetSubscriptionAttributesInput{
		SubscriptionArn: &subARN,
	})
	if err != nil {
		return managed.ExternalUpdate{}, nativehelper.Wrap(err, errGetAttrs)
	}

	updates := buildUpdateAttributes(cr.GetForProvider(), resp.Attributes)
	for attrName, attrValue := range updates {
		attrName, attrValue := attrName, attrValue
		if _, err := e.Client.SetSubscriptionAttributes(ctx, &awssns.SetSubscriptionAttributesInput{
			SubscriptionArn: &subARN,
			AttributeName:   &attrName,
			AttributeValue:  &attrValue,
		}); err != nil {
			return managed.ExternalUpdate{}, nativehelper.Wrap(err, errUpdate)
		}
	}

	return managed.ExternalUpdate{}, nil
}

// Delete unsubscribes the endpoint from the SNS topic.
func (e *ExternalClient) Delete(ctx context.Context, cr TopicSubscriptionCR) (managed.ExternalDelete, error) {
	subARN := subscriptionARN(cr)
	if subARN == "" || subARN == pendingConfirmationARN {
		// Nothing to delete — resource was never fully created.
		return managed.ExternalDelete{}, nil
	}

	_, err := e.Client.Unsubscribe(ctx, &awssns.UnsubscribeInput{
		SubscriptionArn: &subARN,
	})
	if err != nil {
		if isSubscriptionNotFound(err) {
			return managed.ExternalDelete{}, nil
		}
		return managed.ExternalDelete{}, nativehelper.Wrap(err, errDelete)
	}

	return managed.ExternalDelete{}, nil
}

// ── identifier helpers ──────────────────────────────────────────────────────────

// subscriptionARN returns the subscription ARN stored in the external-name
// annotation. Returns "" when the annotation is absent or equals the CR name
// (the crossplane-runtime default before Create sets it to the real ARN).
func subscriptionARN(cr TopicSubscriptionCR) string {
	arn := meta.GetExternalName(cr)
	if arn == "" || arn == cr.GetName() {
		return ""
	}
	return arn
}

// isSubscriptionNotFound returns true when err is an SNS NotFoundException.
func isSubscriptionNotFound(err error) bool {
	var notFound *snstypes.NotFoundException
	return errors.As(err, &notFound)
}

// ── observation ────────────────────────────────────────────────────────────────

// mapAttrsToObservation converts the flat attribute map returned by
// GetSubscriptionAttributes to a TopicSubscriptionRAWObservation.
func mapAttrsToObservation(attrs map[string]string) clusternative.TopicSubscriptionRAWObservation {
	obs := clusternative.TopicSubscriptionRAWObservation{}

	setStrObs(&obs.Arn, attrs, "SubscriptionArn")
	setStrObs(&obs.ID, attrs, "SubscriptionArn")
	setStrObs(&obs.TopicArn, attrs, "TopicArn")
	setStrObs(&obs.Protocol, attrs, "Protocol")
	setStrObs(&obs.Endpoint, attrs, "Endpoint")
	setStrObs(&obs.OwnerID, attrs, "Owner")
	setStrObs(&obs.DeliveryPolicy, attrs, "DeliveryPolicy")
	setStrObs(&obs.FilterPolicy, attrs, "FilterPolicy")
	setStrObs(&obs.FilterPolicyScope, attrs, "FilterPolicyScope")
	setStrObs(&obs.RedrivePolicy, attrs, "RedrivePolicy")
	setStrObs(&obs.ReplayPolicy, attrs, "ReplayPolicy")
	setStrObs(&obs.SubscriptionRoleArn, attrs, "SubscriptionRoleArn")

	setBoolObs(&obs.ConfirmationWasAuthenticated, attrs, "ConfirmationWasAuthenticated")
	setBoolObs(&obs.PendingConfirmation, attrs, "PendingConfirmation")
	setBoolObs(&obs.RawMessageDelivery, attrs, "RawMessageDelivery")

	return obs
}

// setStrObs sets *dst from attrs[key] when the key exists in the map.
func setStrObs(dst **string, attrs map[string]string, key string) {
	if v, ok := attrs[key]; ok {
		v := v
		*dst = &v
	}
}

// setBoolObs sets *dst from attrs[key] parsed as a bool ("true"/"false").
func setBoolObs(dst **bool, attrs map[string]string, key string) {
	if v, ok := attrs[key]; ok {
		b := v == attrTrue
		*dst = &b
	}
}

// ── late initialization ─────────────────────────────────────────────────────────

// lateInitialize copies AWS-defaulted field values into the spec when those
// fields were not explicitly set. Returns true if any field was populated.
//
// Uses SetForProvider* setters rather than mutating through GetForProvider()
// because the namespaced GetForProvider() returns a field-copied struct.
func (e *ExternalClient) lateInitialize(cr TopicSubscriptionCR, attrs map[string]string) bool {
	spec := cr.GetForProvider()
	changed := false

	if v, ok := lateInitStringVal(spec.FilterPolicyScope, attrs, "FilterPolicyScope"); ok {
		cr.SetForProviderFilterPolicyScope(v)
		changed = true
	}

	return changed
}

// lateInitStringVal returns the string value to late-initialise and true when
// the field is unset and the attribute exists and is non-empty.
func lateInitStringVal(current *string, attrs map[string]string, key string) (*string, bool) {
	v, ok := attrs[key]
	if !ok || v == "" || current != nil {
		return nil, false
	}
	copy := v
	return &copy, true
}

// ── up-to-date comparison ───────────────────────────────────────────────────────

// isUpToDate compares the desired spec against the observed AWS state for the
// mutable attributes of an SNS subscription. Protocol, Endpoint, and TopicArn
// are immutable after creation and are therefore excluded from the comparison.
// Returns true when all mutable fields match.
func isUpToDate(cr TopicSubscriptionCR, attrs map[string]string) bool {
	spec := cr.GetForProvider()

	return strAttrUpToDate(spec.DeliveryPolicy, attrs, "DeliveryPolicy") &&
		strAttrUpToDate(spec.FilterPolicy, attrs, "FilterPolicy") &&
		strAttrUpToDate(spec.FilterPolicyScope, attrs, "FilterPolicyScope") &&
		strAttrUpToDate(spec.RedrivePolicy, attrs, "RedrivePolicy") &&
		strAttrUpToDate(spec.ReplayPolicy, attrs, "ReplayPolicy") &&
		strAttrUpToDate(spec.SubscriptionRoleArn, attrs, "SubscriptionRoleArn") &&
		boolAttrUpToDate(spec.RawMessageDelivery, attrs, "RawMessageDelivery")
}

// strAttrUpToDate returns false (indicating drift) when the spec value is non-nil
// and does not match the observed AWS attribute value.
func strAttrUpToDate(spec *string, attrs map[string]string, key string) bool {
	if spec == nil {
		return true
	}
	return *spec == attrs[key]
}

// boolAttrUpToDate returns false (indicating drift) when the spec value is non-nil
// and does not match the observed AWS attribute bool string.
func boolAttrUpToDate(spec *bool, attrs map[string]string, key string) bool {
	if spec == nil {
		return true
	}
	obsVal := attrs[key] == attrTrue
	return *spec == obsVal
}

// ── create / update helpers ─────────────────────────────────────────────────────

// buildSubscribeAttributes assembles the attribute map for the Subscribe call.
// Note: ConfirmationTimeoutInMinutes and EndpointAutoConfirms are client-side
// parameters only and are NOT sent as subscription attributes to AWS.
func buildSubscribeAttributes(spec *clusternative.TopicSubscriptionRAWParameters) map[string]string {
	attrs := make(map[string]string)
	setStrAttr(attrs, "DeliveryPolicy", spec.DeliveryPolicy)
	setStrAttr(attrs, "FilterPolicy", spec.FilterPolicy)
	setStrAttr(attrs, "FilterPolicyScope", spec.FilterPolicyScope)
	setStrAttr(attrs, "RedrivePolicy", spec.RedrivePolicy)
	setStrAttr(attrs, "ReplayPolicy", spec.ReplayPolicy)
	setStrAttr(attrs, "SubscriptionRoleArn", spec.SubscriptionRoleArn)
	setBoolAttr(attrs, "RawMessageDelivery", spec.RawMessageDelivery)
	return attrs
}

// buildUpdateAttributes returns a map of attribute name → desired value for
// mutable attributes that differ from the observed AWS state.
func buildUpdateAttributes(spec *clusternative.TopicSubscriptionRAWParameters, attrs map[string]string) map[string]string {
	updates := make(map[string]string)
	maybeSetStr(updates, "DeliveryPolicy", spec.DeliveryPolicy, attrs["DeliveryPolicy"])
	maybeSetStr(updates, "FilterPolicy", spec.FilterPolicy, attrs["FilterPolicy"])
	maybeSetStr(updates, "FilterPolicyScope", spec.FilterPolicyScope, attrs["FilterPolicyScope"])
	maybeSetStr(updates, "RedrivePolicy", spec.RedrivePolicy, attrs["RedrivePolicy"])
	maybeSetStr(updates, "ReplayPolicy", spec.ReplayPolicy, attrs["ReplayPolicy"])
	maybeSetStr(updates, "SubscriptionRoleArn", spec.SubscriptionRoleArn, attrs["SubscriptionRoleArn"])
	maybeSetBool(updates, "RawMessageDelivery", spec.RawMessageDelivery, attrs["RawMessageDelivery"])
	return updates
}

// setStrAttr adds key→value to attrs when value is non-nil and non-empty.
func setStrAttr(attrs map[string]string, key string, value *string) {
	if value != nil && *value != "" {
		attrs[key] = *value
	}
}

// setBoolAttr adds key→"true"/"false" to attrs when value is non-nil.
func setBoolAttr(attrs map[string]string, key string, value *bool) {
	if value == nil {
		return
	}
	if *value {
		attrs[key] = attrTrue
	} else {
		attrs[key] = attrFalse
	}
}

// maybeSetStr adds an attribute to updates when the spec value is non-nil and
// differs from the observed value.
func maybeSetStr(updates map[string]string, key string, spec *string, observed string) {
	if spec == nil {
		return
	}
	if *spec != observed {
		updates[key] = *spec
	}
}

// maybeSetBool adds an attribute to updates when the spec bool (as "true"/"false"
// string) differs from the observed value.
func maybeSetBool(updates map[string]string, key string, spec *bool, observed string) {
	if spec == nil {
		return
	}
	want := attrFalse
	if *spec {
		want = attrTrue
	}
	if want != observed {
		updates[key] = want
	}
}
