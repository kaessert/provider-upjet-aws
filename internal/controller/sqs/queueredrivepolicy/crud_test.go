// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package queueredrivepolicy_test

import (
	"context"
	"fmt"
	"testing"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
	"github.com/upbound/provider-aws/v2/internal/controller/sqs/queueredrivepolicy"
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
var _ queueredrivepolicy.SQSClient = (*mockSQSClient)(nil)

// ── helpers ────────────────────────────────────────────────────────────────────

const (
	testQueueURL       = "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue"
	testRedrivePolicy  = `{"deadLetterTargetArn":"arn:aws:sqs:us-east-1:123456789012:dlq","maxReceiveCount":5}`
	testRedrivePolicyB = `{"deadLetterTargetArn":"arn:aws:sqs:us-east-1:999999999999:other-dlq","maxReceiveCount":3}`
)

func ptrStr(s string) *string { return &s }

func makeCR(extName string) *clusternative.QueueRedrivePolicyRAW {
	cr := &clusternative.QueueRedrivePolicyRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-queue-redrive-policy",
		},
		Spec: clusternative.QueueRedrivePolicyRAWSpec{
			ForProvider: clusternative.QueueRedrivePolicyRAWParameters{
				Region:        ptrStr("us-east-1"),
				QueueURL:      ptrStr(testQueueURL),
				RedrivePolicy: ptrStr(testRedrivePolicy),
			},
		},
	}
	meta.SetExternalName(cr, extName)
	return cr
}

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_NonURLExternalName_ReturnsNotExists(t *testing.T) {
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{}}
	cr := makeCR("test-queue-redrive-policy") // K8s name, not a URL

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
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
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

func TestObserve_NoRedrivePolicyAttribute_ReturnsNotExists(t *testing.T) {
	// Queue exists but has no redrive policy — attribute absent from response.
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{}, // no RedrivePolicy key
		},
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when RedrivePolicy attribute absent")
	}
}

func TestObserve_EmptyRedrivePolicyAttribute_ReturnsNotExists(t *testing.T) {
	// Queue exists but redrive policy is empty string.
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedrivePolicy": ""},
		},
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when RedrivePolicy is empty string")
	}
}

func TestObserve_RedrivePolicyExists_ReturnsExistsAndUpToDate(t *testing.T) {
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedrivePolicy": testRedrivePolicy},
		},
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true when redrive policy exists")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when policies match")
	}
}

func TestObserve_RedrivePolicyDiffers_ReturnsNotUpToDate(t *testing.T) {
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedrivePolicy": testRedrivePolicyB},
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
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedrivePolicy": testRedrivePolicy},
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
	if atProvider.RedrivePolicy == nil || *atProvider.RedrivePolicy != testRedrivePolicy {
		t.Errorf("expected AtProvider.RedrivePolicy=%q, got %v", testRedrivePolicy, atProvider.RedrivePolicy)
	}
}

func TestObserve_SetsAvailableCondition(t *testing.T) {
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedrivePolicy": testRedrivePolicy},
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
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
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
	awsPolicy := `{"maxReceiveCount": 5, "deadLetterTargetArn": "arn:aws:sqs:us-east-1:123456789012:dlq"}`
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedrivePolicy": awsPolicy},
		},
	}}
	cr := makeCR(testQueueURL)
	// Spec has same content with different key ordering.
	cr.Spec.ForProvider.RedrivePolicy = ptrStr(testRedrivePolicy)

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
	awsPolicy := `{ "deadLetterTargetArn" : "arn:aws:sqs:us-east-1:123456789012:dlq" , "maxReceiveCount" : 5 }`
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"RedrivePolicy": awsPolicy},
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
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{}}
	cr := makeCR("test-queue-redrive-policy") // Not yet a URL

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := meta.GetExternalName(cr); got != testQueueURL {
		t.Errorf("expected external name=%q, got %q", testQueueURL, got)
	}
}

func TestCreate_CallsSetQueueAttributesWithRedrivePolicy(t *testing.T) {
	mock := &mockSQSClient{}
	e := &queueredrivepolicy.ExternalClient{Client: mock}
	cr := makeCR("test-queue-redrive-policy")

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mock.lastSetAttrs == nil {
		t.Fatal("expected SetQueueAttributes to be called")
	}
	if got := mock.lastSetAttrs.Attributes["RedrivePolicy"]; got != testRedrivePolicy {
		t.Errorf("expected RedrivePolicy=%q, got %q", testRedrivePolicy, got)
	}
	if mock.lastSetAttrs.QueueUrl == nil || *mock.lastSetAttrs.QueueUrl != testQueueURL {
		t.Errorf("expected QueueUrl=%q, got %v", testQueueURL, mock.lastSetAttrs.QueueUrl)
	}
}

func TestCreate_Error_ReturnsError(t *testing.T) {
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
		setAttrsErr: fmt.Errorf("set attrs error"),
	}}
	cr := makeCR("test-queue-redrive-policy")

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_CallsSetQueueAttributesWithUpdatedPolicy(t *testing.T) {
	mock := &mockSQSClient{}
	e := &queueredrivepolicy.ExternalClient{Client: mock}
	cr := makeCR(testQueueURL)
	newPolicy := `{"deadLetterTargetArn":"arn:aws:sqs:us-east-1:123456789012:dlq","maxReceiveCount":10}`
	cr.Spec.ForProvider.RedrivePolicy = ptrStr(newPolicy)

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastSetAttrs == nil {
		t.Fatal("expected SetQueueAttributes to be called")
	}
	if got := mock.lastSetAttrs.Attributes["RedrivePolicy"]; got != newPolicy {
		t.Errorf("expected RedrivePolicy=%q, got %q", newPolicy, got)
	}
}

func TestUpdate_Error_ReturnsError(t *testing.T) {
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
		setAttrsErr: fmt.Errorf("set attrs error"),
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Update(context.Background(), cr)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ── Delete tests ───────────────────────────────────────────────────────────────

func TestDelete_SetsEmptyRedrivePolicyToRemove(t *testing.T) {
	mock := &mockSQSClient{}
	e := &queueredrivepolicy.ExternalClient{Client: mock}
	cr := makeCR(testQueueURL)

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastSetAttrs == nil {
		t.Fatal("expected SetQueueAttributes to be called")
	}
	if got := mock.lastSetAttrs.Attributes["RedrivePolicy"]; got != "" {
		t.Errorf("expected empty RedrivePolicy on delete, got %q", got)
	}
}

func TestDelete_IdempotentWhenQueueDoesNotExist(t *testing.T) {
	notFoundErr := &sqstypes.QueueDoesNotExist{Message: ptrStr("not found")}
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
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
	e := &queueredrivepolicy.ExternalClient{Client: mock}
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
	e := &queueredrivepolicy.ExternalClient{Client: &mockSQSClient{
		setAttrsErr: fmt.Errorf("delete error"),
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Delete(context.Background(), cr)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
