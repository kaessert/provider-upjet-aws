// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package queueredriveallowpolicy_test

import (
	"context"
	"fmt"
	"testing"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
	"github.com/upbound/provider-aws/v2/internal/controller/sqs/queueredriveallowpolicy"
)

// ── mock client ────────────────────────────────────────────────────────────────

type mockSQSClient struct {
	getAttrsOut  *awssqs.GetQueueAttributesOutput
	getAttrsErr  error
	setAttrsErr  error
	lastSetAttrs *awssqs.SetQueueAttributesInput
}

func (m *mockSQSClient) GetQueueAttributes(_ context.Context, _ *awssqs.GetQueueAttributesInput, _ ...func(*awssqs.Options)) (*awssqs.GetQueueAttributesOutput, error) {
	return m.getAttrsOut, m.getAttrsErr
}

func (m *mockSQSClient) SetQueueAttributes(_ context.Context, params *awssqs.SetQueueAttributesInput, _ ...func(*awssqs.Options)) (*awssqs.SetQueueAttributesOutput, error) {
	m.lastSetAttrs = params
	return &awssqs.SetQueueAttributesOutput{}, m.setAttrsErr
}

// Compile-time assertion: mockSQSClient implements the SQSClient interface.
var _ queueredriveallowpolicy.SQSClient = (*mockSQSClient)(nil)

// ── helpers ────────────────────────────────────────────────────────────────────

const (
	testQueueURL            = "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue"
	testRedriveAllowPolicy  = `{"redrivePermission":"byQueue","sourceQueueArns":["arn:aws:sqs:us-east-1:123456789012:src-queue"]}`
	testRedriveAllowPolicyB = `{"redrivePermission":"denyAll"}`
)

func ptrStr(s string) *string { return &s }

func makeCR(extName string) *clusternative.QueueRedriveAllowPolicyRAW {
	cr := &clusternative.QueueRedriveAllowPolicyRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-queue-redrive-allow-policy",
		},
		Spec: clusternative.QueueRedriveAllowPolicyRAWSpec{
			ForProvider: clusternative.QueueRedriveAllowPolicyRAWParameters{
				Region:             ptrStr("us-east-1"),
				QueueURL:           ptrStr(testQueueURL),
				RedriveAllowPolicy: ptrStr(testRedriveAllowPolicy),
			},
		},
	}
	meta.SetExternalName(cr, extName)
	return cr
}

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_NonURLExternalName_ReturnsNotExists(t *testing.T) {
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{}}
	cr := makeCR("test-queue-redrive-allow-policy") // K8s name, not a URL

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when external name is not a queue URL")
	}
}

func TestObserve_QueueDoesNotExist_ReturnsNotExists(t *testing.T) {
	notFoundErr := &sqstypes.QueueDoesNotExist{Message: ptrStr("queue does not exist")}
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsErr: notFoundErr,
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for QueueDoesNotExist")
	}
}

func TestObserve_NoRedriveAllowPolicyAttribute_ReturnsNotExists(t *testing.T) {
	// Queue exists but has no redrive allow policy — attribute absent from response.
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{}, // no RedriveAllowPolicy key
		},
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when RedriveAllowPolicy attribute absent")
	}
}

func TestObserve_EmptyRedriveAllowPolicyAttribute_ReturnsNotExists(t *testing.T) {
	// Queue exists but redrive allow policy is empty string.
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedriveAllowPolicy": ""},
		},
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when RedriveAllowPolicy is empty string")
	}
}

func TestObserve_RedriveAllowPolicyExists_ReturnsExistsAndUpToDate(t *testing.T) {
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedriveAllowPolicy": testRedriveAllowPolicy},
		},
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true when redrive allow policy exists")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when policies match")
	}
}

func TestObserve_RedriveAllowPolicyDiffers_ReturnsNotUpToDate(t *testing.T) {
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedriveAllowPolicy": testRedriveAllowPolicyB},
		},
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when policies differ")
	}
}

func TestObserve_SetsAtProvider(t *testing.T) {
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedriveAllowPolicy": testRedriveAllowPolicy},
		},
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	atProvider := cr.GetAtProvider()
	if atProvider.ID == nil || *atProvider.ID != testQueueURL {
		t.Errorf("expected AtProvider.ID=%q, got %v", testQueueURL, atProvider.ID)
	}
	if atProvider.QueueURL == nil || *atProvider.QueueURL != testQueueURL {
		t.Errorf("expected AtProvider.QueueURL=%q, got %v", testQueueURL, atProvider.QueueURL)
	}
	if atProvider.RedriveAllowPolicy == nil || *atProvider.RedriveAllowPolicy != testRedriveAllowPolicy {
		t.Errorf("expected AtProvider.RedriveAllowPolicy=%q, got %v", testRedriveAllowPolicy, atProvider.RedriveAllowPolicy)
	}
}

func TestObserve_SetsAvailableCondition(t *testing.T) {
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedriveAllowPolicy": testRedriveAllowPolicy},
		},
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cond := cr.GetCondition("Ready")
	if cond.Status != "True" {
		t.Errorf("expected Ready=True, got %s", cond.Status)
	}
}

func TestObserve_AWSerror_ReturnsError(t *testing.T) {
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsErr: fmt.Errorf("unexpected AWS error"),
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Observe(context.Background(), cr)
	if err == nil {
		t.Error("expected error for unexpected AWS error, got nil")
	}
}

func TestObserve_JSONSemanticEquivalence_ReturnsUpToDate(t *testing.T) {
	// AWS may return JSON with different key ordering or whitespace.
	awsPolicy := `{"sourceQueueArns":["arn:aws:sqs:us-east-1:123456789012:src-queue"],"redrivePermission":"byQueue"}`
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedriveAllowPolicy": awsPolicy},
		},
	}}
	cr := makeCR(testQueueURL)
	// Spec has same content with different key ordering.
	cr.Spec.ForProvider.RedriveAllowPolicy = ptrStr(testRedriveAllowPolicy)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when policies are semantically equivalent (different key order)")
	}
}

func TestObserve_JSONWithWhitespaceDiff_ReturnsUpToDate(t *testing.T) {
	// Same content with extra whitespace.
	awsPolicy := `{ "redrivePermission" : "byQueue" , "sourceQueueArns" : ["arn:aws:sqs:us-east-1:123456789012:src-queue"] }`
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedriveAllowPolicy": awsPolicy},
		},
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when policies are semantically equivalent (whitespace diff)")
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

func TestCreate_SetsExternalNameToQueueURL(t *testing.T) {
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{}}
	cr := makeCR("test-queue-redrive-allow-policy") // Not yet a URL

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := meta.GetExternalName(cr); got != testQueueURL {
		t.Errorf("expected external name=%q, got %q", testQueueURL, got)
	}
}

func TestCreate_CallsSetQueueAttributesWithRedriveAllowPolicy(t *testing.T) {
	mock := &mockSQSClient{}
	e := &queueredriveallowpolicy.ExternalClient{Client: mock}
	cr := makeCR("test-queue-redrive-allow-policy")

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mock.lastSetAttrs == nil {
		t.Fatal("expected SetQueueAttributes to be called")
	}
	if got := mock.lastSetAttrs.Attributes["RedriveAllowPolicy"]; got != testRedriveAllowPolicy {
		t.Errorf("expected RedriveAllowPolicy=%q, got %q", testRedriveAllowPolicy, got)
	}
	if mock.lastSetAttrs.QueueUrl == nil || *mock.lastSetAttrs.QueueUrl != testQueueURL {
		t.Errorf("expected QueueUrl=%q, got %v", testQueueURL, mock.lastSetAttrs.QueueUrl)
	}
}

func TestCreate_Error_ReturnsError(t *testing.T) {
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		setAttrsErr: fmt.Errorf("set attrs error"),
	}}
	cr := makeCR("test-queue-redrive-allow-policy")

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCreate_NonURLQueueURL_ReturnsError(t *testing.T) {
	// Simulate an unresolved reference: ForProvider.QueueURL holds the K8s resource
	// name instead of a real SQS queue URL (the referenced QueueRAW is not yet Ready).
	mock := &mockSQSClient{}
	e := &queueredriveallowpolicy.ExternalClient{Client: mock}
	cr := makeCR("test-queue-redrive-allow-policy")
	cr.Spec.ForProvider.QueueURL = ptrStr("my-queue") // K8s name, not an SQS URL

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Error("expected error when queue URL is not resolved (not an SQS URL), got nil")
	}
	// Must not call AWS when the URL is not resolved.
	if mock.lastSetAttrs != nil {
		t.Error("expected no SetQueueAttributes call when queue URL is not resolved")
	}
}

func TestCreate_ValidQueueURL_Succeeds(t *testing.T) {
	// Simulate a successfully resolved reference: ForProvider.QueueURL is a real SQS URL.
	mock := &mockSQSClient{}
	e := &queueredriveallowpolicy.ExternalClient{Client: mock}
	cr := makeCR("test-queue-redrive-allow-policy")
	cr.Spec.ForProvider.QueueURL = ptrStr(testQueueURL) // valid SQS URL

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error for valid queue URL, got %v", err)
	}
	if mock.lastSetAttrs == nil {
		t.Error("expected SetQueueAttributes to be called for valid queue URL")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_CallsSetQueueAttributesWithUpdatedPolicy(t *testing.T) {
	mock := &mockSQSClient{}
	e := &queueredriveallowpolicy.ExternalClient{Client: mock}
	cr := makeCR(testQueueURL)
	newPolicy := `{"redrivePermission":"allowAll"}`
	cr.Spec.ForProvider.RedriveAllowPolicy = ptrStr(newPolicy)

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastSetAttrs == nil {
		t.Fatal("expected SetQueueAttributes to be called")
	}
	if got := mock.lastSetAttrs.Attributes["RedriveAllowPolicy"]; got != newPolicy {
		t.Errorf("expected RedriveAllowPolicy=%q, got %q", newPolicy, got)
	}
}

func TestUpdate_Error_ReturnsError(t *testing.T) {
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		setAttrsErr: fmt.Errorf("set attrs error"),
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Update(context.Background(), cr)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ── Delete tests ───────────────────────────────────────────────────────────────

func TestDelete_SetsEmptyRedriveAllowPolicyToRemove(t *testing.T) {
	mock := &mockSQSClient{}
	e := &queueredriveallowpolicy.ExternalClient{Client: mock}
	cr := makeCR(testQueueURL)

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastSetAttrs == nil {
		t.Fatal("expected SetQueueAttributes to be called")
	}
	if got := mock.lastSetAttrs.Attributes["RedriveAllowPolicy"]; got != "" {
		t.Errorf("expected empty RedriveAllowPolicy on delete, got %q", got)
	}
}

func TestDelete_IdempotentWhenQueueDoesNotExist(t *testing.T) {
	notFoundErr := &sqstypes.QueueDoesNotExist{Message: ptrStr("not found")}
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		setAttrsErr: notFoundErr,
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Errorf("expected no error for QueueDoesNotExist on delete, got %v", err)
	}
}

func TestDelete_NonURL_ReturnsNoError(t *testing.T) {
	mock := &mockSQSClient{}
	e := &queueredriveallowpolicy.ExternalClient{Client: mock}
	cr := makeCR("not-a-url")

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Errorf("expected no error for non-URL external name, got %v", err)
	}
	// Should not call AWS.
	if mock.lastSetAttrs != nil {
		t.Error("expected no SetQueueAttributes call for non-URL external name")
	}
}

func TestDelete_Error_ReturnsError(t *testing.T) {
	e := &queueredriveallowpolicy.ExternalClient{Client: &mockSQSClient{
		setAttrsErr: fmt.Errorf("delete error"),
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Delete(context.Background(), cr)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
