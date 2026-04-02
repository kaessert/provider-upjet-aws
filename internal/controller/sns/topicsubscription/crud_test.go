// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package topicsubscription_test

import (
	"context"
	"errors"
	"testing"

	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native"
	"github.com/upbound/provider-aws/v2/internal/controller/sns/topicsubscription"
)

// ── mock SNS client ────────────────────────────────────────────────────────────

type mockSNSSubClient struct {
	subscribeOut   *awssns.SubscribeOutput
	subscribeErr   error
	getAttrsOut    *awssns.GetSubscriptionAttributesOutput
	getAttrsErr    error
	setAttrsErr    error
	unsubscribeErr error
	lastSetAttrs   []*awssns.SetSubscriptionAttributesInput
}

func (m *mockSNSSubClient) Subscribe(_ context.Context, _ *awssns.SubscribeInput, _ ...func(*awssns.Options)) (*awssns.SubscribeOutput, error) {
	return m.subscribeOut, m.subscribeErr
}

func (m *mockSNSSubClient) GetSubscriptionAttributes(_ context.Context, _ *awssns.GetSubscriptionAttributesInput, _ ...func(*awssns.Options)) (*awssns.GetSubscriptionAttributesOutput, error) {
	if m.getAttrsOut != nil {
		return m.getAttrsOut, m.getAttrsErr
	}
	return &awssns.GetSubscriptionAttributesOutput{Attributes: map[string]string{}}, m.getAttrsErr
}

func (m *mockSNSSubClient) SetSubscriptionAttributes(_ context.Context, params *awssns.SetSubscriptionAttributesInput, _ ...func(*awssns.Options)) (*awssns.SetSubscriptionAttributesOutput, error) {
	m.lastSetAttrs = append(m.lastSetAttrs, params)
	return &awssns.SetSubscriptionAttributesOutput{}, m.setAttrsErr
}

func (m *mockSNSSubClient) Unsubscribe(_ context.Context, _ *awssns.UnsubscribeInput, _ ...func(*awssns.Options)) (*awssns.UnsubscribeOutput, error) {
	return &awssns.UnsubscribeOutput{}, m.unsubscribeErr
}

// ── helpers ────────────────────────────────────────────────────────────────────

const (
	testSubARN   = "arn:aws:sns:us-east-1:123456789012:my-topic:sub-uuid-1234"
	testTopicARN = "arn:aws:sns:us-east-1:123456789012:my-topic"
	testEndpoint = "arn:aws:sqs:us-east-1:123456789012:my-queue"
	testProtocol = "sqs"
	testRegion   = "us-east-1"
)

func ptrStr(s string) *string { return &s }
func ptrBool(b bool) *bool    { return &b }

// makeCR creates a TopicSubscriptionRAW with the given external name.
// When extName equals "test-sub" (CR name) it simulates "not yet created".
func makeCR(extName string) *clusternative.TopicSubscriptionRAW {
	cr := &clusternative.TopicSubscriptionRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-sub",
		},
		Spec: clusternative.TopicSubscriptionRAWSpec{
			ForProvider: clusternative.TopicSubscriptionRAWParameters{
				Protocol: ptrStr(testProtocol),
				TopicArn: ptrStr(testTopicARN),
				Endpoint: ptrStr(testEndpoint),
				Region:   ptrStr(testRegion),
			},
		},
	}
	meta.SetExternalName(cr, extName)
	return cr
}

// baseAttrs returns the typical attributes returned by GetSubscriptionAttributes.
func baseAttrs() map[string]string {
	return map[string]string{
		"SubscriptionArn":              testSubARN,
		"TopicArn":                     testTopicARN,
		"Protocol":                     testProtocol,
		"Endpoint":                     testEndpoint,
		"Owner":                        "123456789012",
		"PendingConfirmation":          "false",
		"ConfirmationWasAuthenticated": "true",
		"RawMessageDelivery":           "false",
		"FilterPolicyScope":            "MessageAttributes",
	}
}

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_DefaultExternalName_ReturnsNotExists(t *testing.T) {
	// External name == CR name: means resource not yet created.
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{}}
	cr := makeCR("test-sub") // same as CR name → not created yet

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when external name equals CR name")
	}
}

func TestObserve_EmptyExternalName_ReturnsNotExists(t *testing.T) {
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{}}
	cr := makeCR("")

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty external name")
	}
}

func TestObserve_NotFound_ReturnsNotExists(t *testing.T) {
	notFoundErr := &snstypes.NotFoundException{Message: ptrStr("subscription not found")}
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		getAttrsErr: notFoundErr,
	}}
	cr := makeCR(testSubARN)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for NotFoundException")
	}
}

func TestObserve_PendingConfirmationARN_SetsUnavailable(t *testing.T) {
	// When Create() returns "pending confirmation" as the ARN, we store it.
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{}}
	cr := makeCR("pending confirmation")

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for pending confirmation")
	}
}

func TestObserve_PendingConfirmationAttr_SetsUnavailable(t *testing.T) {
	// Subscription exists but PendingConfirmation attribute is "true".
	attrs := baseAttrs()
	attrs["PendingConfirmation"] = "true"
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		getAttrsOut: &awssns.GetSubscriptionAttributesOutput{Attributes: attrs},
	}}
	cr := makeCR(testSubARN)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true when subscription is pending confirmation")
	}
	// Should set Unavailable condition (Ready=False).
	found := false
	for _, c := range cr.Status.Conditions {
		if c.Type == "Ready" && string(c.Status) == "False" {
			found = true
		}
	}
	if !found {
		t.Error("expected Unavailable (Ready=False) condition for pending confirmation")
	}
}

func TestObserve_ExistsAndUpToDate_ReturnsUpToDate(t *testing.T) {
	attrs := baseAttrs()
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		getAttrsOut: &awssns.GetSubscriptionAttributesOutput{Attributes: attrs},
	}}
	cr := makeCR(testSubARN)
	cr.Spec.ForProvider.FilterPolicyScope = ptrStr("MessageAttributes") // matches AWS default

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when spec matches AWS state")
	}
}

func TestObserve_SetsAtProvider(t *testing.T) {
	attrs := baseAttrs()
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		getAttrsOut: &awssns.GetSubscriptionAttributesOutput{Attributes: attrs},
	}}
	cr := makeCR(testSubARN)

	_, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	atProvider := cr.GetAtProvider()
	if atProvider.Arn == nil || *atProvider.Arn != testSubARN {
		t.Errorf("expected Arn=%q, got %v", testSubARN, atProvider.Arn)
	}
	if atProvider.TopicArn == nil || *atProvider.TopicArn != testTopicARN {
		t.Errorf("expected TopicArn=%q, got %v", testTopicARN, atProvider.TopicArn)
	}
	if atProvider.Protocol == nil || *atProvider.Protocol != testProtocol {
		t.Errorf("expected Protocol=%q, got %v", testProtocol, atProvider.Protocol)
	}
	if atProvider.OwnerID == nil || *atProvider.OwnerID != "123456789012" {
		t.Errorf("expected OwnerID=%q, got %v", "123456789012", atProvider.OwnerID)
	}
}

func TestObserve_LateInitializesFilterPolicyScope(t *testing.T) {
	attrs := baseAttrs()
	// FilterPolicyScope is present in AWS response but not set in spec.
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		getAttrsOut: &awssns.GetSubscriptionAttributesOutput{Attributes: attrs},
	}}
	cr := makeCR(testSubARN)
	// FilterPolicyScope not set in spec.

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true when FilterPolicyScope is late-inited")
	}
	if cr.Spec.ForProvider.FilterPolicyScope == nil || *cr.Spec.ForProvider.FilterPolicyScope != "MessageAttributes" {
		t.Errorf("expected FilterPolicyScope=MessageAttributes after late-init, got %v", cr.Spec.ForProvider.FilterPolicyScope)
	}
}

func TestObserve_NotUpToDate_WhenDeliveryPolicyDiffers(t *testing.T) {
	attrs := baseAttrs()
	attrs["DeliveryPolicy"] = `{"healthyRetryPolicy":{"numRetries":5}}`
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		getAttrsOut: &awssns.GetSubscriptionAttributesOutput{Attributes: attrs},
	}}
	cr := makeCR(testSubARN)
	cr.Spec.ForProvider.DeliveryPolicy = ptrStr(`{"healthyRetryPolicy":{"numRetries":3}}`)
	cr.Spec.ForProvider.FilterPolicyScope = ptrStr("MessageAttributes") // avoid late-init triggering

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when DeliveryPolicy differs")
	}
}

func TestObserve_NotUpToDate_WhenRawMessageDeliveryDiffers(t *testing.T) {
	attrs := baseAttrs()
	attrs["RawMessageDelivery"] = "false"
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		getAttrsOut: &awssns.GetSubscriptionAttributesOutput{Attributes: attrs},
	}}
	cr := makeCR(testSubARN)
	cr.Spec.ForProvider.RawMessageDelivery = ptrBool(true) // spec wants true, AWS has false
	cr.Spec.ForProvider.FilterPolicyScope = ptrStr("MessageAttributes")

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when RawMessageDelivery differs")
	}
}

func TestObserve_NotUpToDate_WhenFilterPolicyScopeDiffers(t *testing.T) {
	attrs := baseAttrs()
	attrs["FilterPolicyScope"] = "MessageAttributes"
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		getAttrsOut: &awssns.GetSubscriptionAttributesOutput{Attributes: attrs},
	}}
	cr := makeCR(testSubARN)
	cr.Spec.ForProvider.FilterPolicyScope = ptrStr("MessageBody") // differs from AWS

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when FilterPolicyScope differs")
	}
}

func TestObserve_GetAttrsError_ReturnsError(t *testing.T) {
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		getAttrsErr: errors.New("internal error"),
	}}
	cr := makeCR(testSubARN)

	_, err := e.Observe(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

func TestCreate_SetsExternalName(t *testing.T) {
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		subscribeOut: &awssns.SubscribeOutput{
			SubscriptionArn: ptrStr(testSubARN),
		},
	}}
	cr := makeCR("test-sub") // CR name as external name (not yet created)

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	extName := meta.GetExternalName(cr)
	if extName != testSubARN {
		t.Errorf("expected external name=%q, got %q", testSubARN, extName)
	}
}

func TestCreate_PendingConfirmation_SetsExternalName(t *testing.T) {
	// For HTTP/HTTPS protocols, Subscribe may return "pending confirmation".
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		subscribeOut: &awssns.SubscribeOutput{
			SubscriptionArn: ptrStr("pending confirmation"),
		},
	}}
	cr := makeCR("test-sub")

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	extName := meta.GetExternalName(cr)
	if extName != "pending confirmation" {
		t.Errorf("expected external name=%q, got %q", "pending confirmation", extName)
	}
}

func TestCreate_Error_ReturnsError(t *testing.T) {
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		subscribeErr: errors.New("API error"),
	}}
	cr := makeCR("test-sub")

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_SetsDeliveryPolicy(t *testing.T) {
	newPolicy := `{"healthyRetryPolicy":{"numRetries":10}}`
	attrs := baseAttrs()
	attrs["FilterPolicyScope"] = "MessageAttributes" // already matches spec
	mock := &mockSNSSubClient{
		getAttrsOut: &awssns.GetSubscriptionAttributesOutput{Attributes: attrs},
	}
	e := &topicsubscription.ExternalClient{Client: mock}
	cr := makeCR(testSubARN)
	cr.Spec.ForProvider.DeliveryPolicy = ptrStr(newPolicy)
	cr.Spec.ForProvider.FilterPolicyScope = ptrStr("MessageAttributes")

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	found := false
	for _, call := range mock.lastSetAttrs {
		if call.AttributeName != nil && *call.AttributeName == "DeliveryPolicy" {
			if call.AttributeValue != nil && *call.AttributeValue == newPolicy {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected SetSubscriptionAttributes called with DeliveryPolicy=%s, got %v", newPolicy, mock.lastSetAttrs)
	}
}

func TestUpdate_SetsRawMessageDelivery(t *testing.T) {
	attrs := baseAttrs()
	attrs["RawMessageDelivery"] = "false"
	mock := &mockSNSSubClient{
		getAttrsOut: &awssns.GetSubscriptionAttributesOutput{Attributes: attrs},
	}
	e := &topicsubscription.ExternalClient{Client: mock}
	cr := makeCR(testSubARN)
	cr.Spec.ForProvider.RawMessageDelivery = ptrBool(true)
	cr.Spec.ForProvider.FilterPolicyScope = ptrStr("MessageAttributes")

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	found := false
	for _, call := range mock.lastSetAttrs {
		if call.AttributeName != nil && *call.AttributeName == "RawMessageDelivery" {
			if call.AttributeValue != nil && *call.AttributeValue == "true" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected SetSubscriptionAttributes called with RawMessageDelivery=true")
	}
}

func TestUpdate_PendingConfirmation_NoOp(t *testing.T) {
	mock := &mockSNSSubClient{}
	e := &topicsubscription.ExternalClient{Client: mock}
	cr := makeCR("pending confirmation")

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(mock.lastSetAttrs) > 0 {
		t.Error("expected no SetSubscriptionAttributes calls for pending confirmation")
	}
}

func TestUpdate_SetAttrsError_ReturnsError(t *testing.T) {
	attrs := baseAttrs()
	attrs["DeliveryPolicy"] = "old"
	mock := &mockSNSSubClient{
		getAttrsOut: &awssns.GetSubscriptionAttributesOutput{Attributes: attrs},
		setAttrsErr: errors.New("API error"),
	}
	e := &topicsubscription.ExternalClient{Client: mock}
	cr := makeCR(testSubARN)
	cr.Spec.ForProvider.DeliveryPolicy = ptrStr("new-policy")
	cr.Spec.ForProvider.FilterPolicyScope = ptrStr("MessageAttributes")

	_, err := e.Update(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── Delete tests ───────────────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{}}
	cr := makeCR(testSubARN)

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDelete_Idempotent_WhenNotFound(t *testing.T) {
	notFoundErr := &snstypes.NotFoundException{Message: ptrStr("subscription not found")}
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		unsubscribeErr: notFoundErr,
	}}
	cr := makeCR(testSubARN)

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for not-found during delete, got %v", err)
	}
}

func TestDelete_PendingConfirmation_NoOp(t *testing.T) {
	// Deleting a subscription whose ARN is "pending confirmation" should be a no-op.
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{}}
	cr := makeCR("pending confirmation")

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDelete_NoARN_NoOp(t *testing.T) {
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{}}
	cr := makeCR("test-sub") // CR name — not yet created

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error when no subscription ARN, got %v", err)
	}
}

func TestDelete_Error_ReturnsError(t *testing.T) {
	e := &topicsubscription.ExternalClient{Client: &mockSNSSubClient{
		unsubscribeErr: errors.New("internal server error"),
	}}
	cr := makeCR(testSubARN)

	_, err := e.Delete(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
