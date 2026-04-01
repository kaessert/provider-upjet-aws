// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package queuepolicy_test

import (
	"context"
	"fmt"
	"testing"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
	"github.com/upbound/provider-aws/v2/internal/controller/sqs/queuepolicy"
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
var _ queuepolicy.SQSClient = (*mockSQSClient)(nil)

// ── helpers ────────────────────────────────────────────────────────────────────

const (
	testQueueURL   = "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue"
	testPolicyJSON = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":"arn:aws:iam::123456789012:root"},"Action":"sqs:SendMessage","Resource":"*"}]}`
	testPolicyAlt  = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":"arn:aws:iam::999999999999:root"},"Action":"sqs:SendMessage","Resource":"*"}]}`
)

func ptrStr(s string) *string { return &s }

func makeCR(extName string) *clusternative.QueuePolicyRAW {
	cr := &clusternative.QueuePolicyRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-queue-policy",
		},
		Spec: clusternative.QueuePolicyRAWSpec{
			ForProvider: clusternative.QueuePolicyRAWParameters{
				Region:   ptrStr("us-east-1"),
				QueueURL: ptrStr(testQueueURL),
				Policy:   ptrStr(testPolicyJSON),
			},
		},
	}
	meta.SetExternalName(cr, extName)
	return cr
}

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_NonURLExternalName_ReturnsNotExists(t *testing.T) {
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{}}
	cr := makeCR("test-queue-policy") // K8s name, not a URL

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
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{
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

func TestObserve_NoPolicyAttribute_ReturnsNotExists(t *testing.T) {
	// Queue exists but has no policy set — Policy attribute absent from response.
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{}, // no Policy key
		},
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when Policy attribute absent")
	}
}

func TestObserve_PolicyExists_ReturnsExistsAndUpToDate(t *testing.T) {
	// Same policy — semantically equivalent.
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"Policy": testPolicyJSON},
		},
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true when policy exists")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when policy matches")
	}
}

func TestObserve_PolicyDiffers_ReturnsNotUpToDate(t *testing.T) {
	// Different policy — not up to date.
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"Policy": testPolicyAlt},
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
		t.Error("expected ResourceUpToDate=false when policy differs")
	}
}

func TestObserve_SetsAtProviderID(t *testing.T) {
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"Policy": testPolicyJSON},
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
}

func TestObserve_SetsAvailableCondition(t *testing.T) {
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"Policy": testPolicyJSON},
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
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsErr: fmt.Errorf("unexpected AWS error"),
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Observe(context.Background(), cr)
	if err == nil {
		t.Error("expected error for unexpected AWS error, got nil")
	}
}

func TestObserve_SemanticPolicyEquivalence_ReturnsUpToDate(t *testing.T) {
	// AWS may return different JSON formatting but same semantics.
	awsPolicy := `{"Version": "2012-10-17", "Statement": [{"Effect": "Allow", "Principal": {"AWS": "arn:aws:iam::123456789012:root"}, "Action": "sqs:SendMessage", "Resource": "*"}]}`
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{
			Attributes: map[string]string{"Policy": awsPolicy},
		},
	}}
	cr := makeCR(testQueueURL)
	// Spec has same policy without spaces
	cr.Spec.ForProvider.Policy = ptrStr(testPolicyJSON)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when policies are semantically equivalent")
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

func TestCreate_SetsExternalNameToQueueURL(t *testing.T) {
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{}}
	cr := makeCR("test-queue-policy") // Not yet a URL

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got := meta.GetExternalName(cr); got != testQueueURL {
		t.Errorf("expected external name=%q, got %q", testQueueURL, got)
	}
}

func TestCreate_CallsSetQueueAttributesWithPolicy(t *testing.T) {
	mock := &mockSQSClient{}
	e := &queuepolicy.ExternalClient{Client: mock}
	cr := makeCR("test-queue-policy")

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if mock.lastSetAttrs == nil {
		t.Fatal("expected SetQueueAttributes to be called")
	}
	if got := mock.lastSetAttrs.Attributes["Policy"]; got != testPolicyJSON {
		t.Errorf("expected Policy=%q, got %q", testPolicyJSON, got)
	}
	if mock.lastSetAttrs.QueueUrl == nil || *mock.lastSetAttrs.QueueUrl != testQueueURL {
		t.Errorf("expected QueueUrl=%q, got %v", testQueueURL, mock.lastSetAttrs.QueueUrl)
	}
}

func TestCreate_Error_ReturnsError(t *testing.T) {
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{
		setAttrsErr: fmt.Errorf("set attrs error"),
	}}
	cr := makeCR("test-queue-policy")

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_CallsSetQueueAttributesWithUpdatedPolicy(t *testing.T) {
	mock := &mockSQSClient{}
	e := &queuepolicy.ExternalClient{Client: mock}
	cr := makeCR(testQueueURL)
	newPolicy := `{"Version":"2012-10-17","Statement":[]}`
	cr.Spec.ForProvider.Policy = ptrStr(newPolicy)

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastSetAttrs == nil {
		t.Fatal("expected SetQueueAttributes to be called")
	}
	if got := mock.lastSetAttrs.Attributes["Policy"]; got != newPolicy {
		t.Errorf("expected Policy=%q, got %q", newPolicy, got)
	}
}

func TestUpdate_Error_ReturnsError(t *testing.T) {
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{
		setAttrsErr: fmt.Errorf("set attrs error"),
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Update(context.Background(), cr)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ── Delete tests ───────────────────────────────────────────────────────────────

func TestDelete_SetsEmptyPolicyToRemove(t *testing.T) {
	mock := &mockSQSClient{}
	e := &queuepolicy.ExternalClient{Client: mock}
	cr := makeCR(testQueueURL)

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastSetAttrs == nil {
		t.Fatal("expected SetQueueAttributes to be called")
	}
	if got := mock.lastSetAttrs.Attributes["Policy"]; got != "" {
		t.Errorf("expected empty Policy on delete, got %q", got)
	}
}

func TestDelete_IdempotentWhenQueueDoesNotExist(t *testing.T) {
	notFoundErr := &sqstypes.QueueDoesNotExist{Message: ptrStr("not found")}
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{
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
	e := &queuepolicy.ExternalClient{Client: mock}
	cr := makeCR("not-a-url")

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Errorf("expected no error for non-URL external name, got %v", err)
	}
	// Should not call AWS
	if mock.lastSetAttrs != nil {
		t.Error("expected no SetQueueAttributes call for non-URL external name")
	}
}

func TestDelete_Error_ReturnsError(t *testing.T) {
	e := &queuepolicy.ExternalClient{Client: &mockSQSClient{
		setAttrsErr: fmt.Errorf("delete error"),
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Delete(context.Background(), cr)
	if err == nil {
		t.Error("expected error, got nil")
	}
}
