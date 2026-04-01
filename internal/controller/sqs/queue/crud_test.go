// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package queue_test

import (
	"context"
	"fmt"
	"testing"

	awssqs "github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
	"github.com/upbound/provider-aws/v2/internal/controller/sqs/queue"
)

// ── mock client ────────────────────────────────────────────────────────────────

type mockSQSClient struct {
	getAttrsOut   *awssqs.GetQueueAttributesOutput
	getAttrsErr   error
	createOut     *awssqs.CreateQueueOutput
	createErr     error
	setAttrsErr   error
	deleteErr     error
	listTagsOut   *awssqs.ListQueueTagsOutput
	listTagsErr   error
	tagQueueErr   error
	untagQueueErr error
	// Capture call args for assertions
	lastSetAttrs   *awssqs.SetQueueAttributesInput
	lastTagInput   *awssqs.TagQueueInput
	lastUntagInput *awssqs.UntagQueueInput
}

func (m *mockSQSClient) GetQueueAttributes(_ context.Context, params *awssqs.GetQueueAttributesInput, _ ...func(*awssqs.Options)) (*awssqs.GetQueueAttributesOutput, error) {
	return m.getAttrsOut, m.getAttrsErr
}

func (m *mockSQSClient) CreateQueue(_ context.Context, _ *awssqs.CreateQueueInput, _ ...func(*awssqs.Options)) (*awssqs.CreateQueueOutput, error) {
	return m.createOut, m.createErr
}

func (m *mockSQSClient) SetQueueAttributes(_ context.Context, params *awssqs.SetQueueAttributesInput, _ ...func(*awssqs.Options)) (*awssqs.SetQueueAttributesOutput, error) {
	m.lastSetAttrs = params
	return &awssqs.SetQueueAttributesOutput{}, m.setAttrsErr
}

func (m *mockSQSClient) DeleteQueue(_ context.Context, _ *awssqs.DeleteQueueInput, _ ...func(*awssqs.Options)) (*awssqs.DeleteQueueOutput, error) {
	return &awssqs.DeleteQueueOutput{}, m.deleteErr
}

func (m *mockSQSClient) ListQueueTags(_ context.Context, _ *awssqs.ListQueueTagsInput, _ ...func(*awssqs.Options)) (*awssqs.ListQueueTagsOutput, error) {
	if m.listTagsOut != nil {
		return m.listTagsOut, m.listTagsErr
	}
	return &awssqs.ListQueueTagsOutput{}, m.listTagsErr
}

func (m *mockSQSClient) TagQueue(_ context.Context, params *awssqs.TagQueueInput, _ ...func(*awssqs.Options)) (*awssqs.TagQueueOutput, error) {
	m.lastTagInput = params
	return &awssqs.TagQueueOutput{}, m.tagQueueErr
}

func (m *mockSQSClient) UntagQueue(_ context.Context, params *awssqs.UntagQueueInput, _ ...func(*awssqs.Options)) (*awssqs.UntagQueueOutput, error) {
	m.lastUntagInput = params
	return &awssqs.UntagQueueOutput{}, m.untagQueueErr
}

func (m *mockSQSClient) GetQueueUrl(_ context.Context, _ *awssqs.GetQueueUrlInput, _ ...func(*awssqs.Options)) (*awssqs.GetQueueUrlOutput, error) {
	return nil, nil
}

// ── helpers ────────────────────────────────────────────────────────────────────

const testQueueURL = "https://sqs.us-east-1.amazonaws.com/123456789012/test-queue"
const testQueueARN = "arn:aws:sqs:us-east-1:123456789012:test-queue"

func ptrStr(s string) *string   { return &s }
func ptrF64(f float64) *float64 { return &f }
func ptrBool(b bool) *bool      { return &b }

func makeCR(extName string) *clusternative.QueueRAW {
	cr := &clusternative.QueueRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-queue",
		},
		Spec: clusternative.QueueRAWSpec{
			ForProvider: clusternative.QueueRAWParameters{
				Region: ptrStr("us-east-1"),
			},
		},
	}
	meta.SetExternalName(cr, extName)
	return cr
}

func baseAttrs() map[string]string {
	return map[string]string{
		"QueueArn":                      testQueueARN,
		"VisibilityTimeout":             "30",
		"MaximumMessageSize":            "262144",
		"MessageRetentionPeriod":        "345600",
		"DelaySeconds":                  "0",
		"ReceiveMessageWaitTimeSeconds": "0",
		"SqsManagedSseEnabled":          "false",
	}
}

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_NotAURL_ReturnsNotExists(t *testing.T) {
	e := &queue.ExternalClient{Client: &mockSQSClient{}}
	cr := makeCR("test-queue") // K8s name, not a URL

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for non-URL external name")
	}
}

func TestObserve_QueueDoesNotExist_ReturnsNotExists(t *testing.T) {
	notFoundErr := &sqstypes.QueueDoesNotExist{Message: ptrStr("queue does not exist")}
	e := &queue.ExternalClient{Client: &mockSQSClient{
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

func TestObserve_QueueExists_ReturnsExists(t *testing.T) {
	attrs := baseAttrs()
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
}

func TestObserve_SetsARNAndURL(t *testing.T) {
	attrs := baseAttrs()
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	atProvider := cr.GetAtProvider()
	if atProvider.Arn == nil || *atProvider.Arn != testQueueARN {
		t.Errorf("expected ARN=%q, got %v", testQueueARN, atProvider.Arn)
	}
	if atProvider.URL == nil || *atProvider.URL != testQueueURL {
		t.Errorf("expected URL=%q, got %v", testQueueURL, atProvider.URL)
	}
	if atProvider.ID == nil || *atProvider.ID != testQueueURL {
		t.Errorf("expected ID=%q, got %v", testQueueURL, atProvider.ID)
	}
}

func TestObserve_UpToDate_WhenSpecMatchesAWS(t *testing.T) {
	attrs := baseAttrs()
	attrs["DelaySeconds"] = "60"
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)
	cr.Spec.ForProvider.DelaySeconds = ptrF64(60)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when spec matches AWS")
	}
}

func TestObserve_NotUpToDate_WhenDelaySecondsDiffer(t *testing.T) {
	attrs := baseAttrs()
	attrs["DelaySeconds"] = "0"
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)
	cr.Spec.ForProvider.DelaySeconds = ptrF64(30)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when DelaySeconds differs")
	}
}

func TestObserve_NotUpToDate_WhenPolicyDiffers(t *testing.T) {
	attrs := baseAttrs()
	attrs["Policy"] = `{"Version":"2012-10-17","Statement":[]}`
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)
	// Different policy semantically
	cr.Spec.ForProvider.Policy = ptrStr(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"sqs:SendMessage","Resource":"*"}]}`)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when policy differs")
	}
}

func TestObserve_LateInitializesKMSDataKeyReusePeriodSeconds(t *testing.T) {
	attrs := baseAttrs()
	attrs["KmsDataKeyReusePeriodSeconds"] = "300"
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)
	// KMSDataKeyReusePeriodSeconds not set in spec

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true when KMSDataKeyReusePeriodSeconds was late-initialized")
	}
	if cr.Spec.ForProvider.KMSDataKeyReusePeriodSeconds == nil {
		t.Error("expected KMSDataKeyReusePeriodSeconds to be set after late init")
	} else if *cr.Spec.ForProvider.KMSDataKeyReusePeriodSeconds != 300 {
		t.Errorf("expected KMSDataKeyReusePeriodSeconds=300, got %v", *cr.Spec.ForProvider.KMSDataKeyReusePeriodSeconds)
	}
}

func TestObserve_LateInitializesSqsManagedSseEnabled(t *testing.T) {
	attrs := baseAttrs()
	attrs["SqsManagedSseEnabled"] = "true"
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true")
	}
	if cr.Spec.ForProvider.SqsManagedSseEnabled == nil || !*cr.Spec.ForProvider.SqsManagedSseEnabled {
		t.Error("expected SqsManagedSseEnabled to be true after late init")
	}
}

func TestObserve_NotUpToDate_WhenTagsDiffer(t *testing.T) {
	attrs := baseAttrs()
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{Tags: map[string]string{"env": "old"}},
	}}
	cr := makeCR(testQueueURL)
	cr.Spec.ForProvider.Tags = map[string]*string{"env": ptrStr("new")}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when tags differ")
	}
}

func TestObserve_SetsAvailableCondition(t *testing.T) {
	attrs := baseAttrs()
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	conditions := cr.GetCondition("Ready")
	if conditions.Status != "True" {
		t.Errorf("expected Ready=True, got %s", conditions.Status)
	}
}

func TestObserve_AWSerror_ReturnsError(t *testing.T) {
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsErr: fmt.Errorf("unexpected AWS error"),
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Observe(context.Background(), cr)
	if err == nil {
		t.Error("expected error for unexpected AWS error, got nil")
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

func TestCreate_SetsExternalName(t *testing.T) {
	e := &queue.ExternalClient{Client: &mockSQSClient{
		createOut: &awssqs.CreateQueueOutput{QueueUrl: ptrStr(testQueueURL)},
	}}
	cr := makeCR("test-queue") // Not yet a URL

	creation, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	_ = creation

	if got := meta.GetExternalName(cr); got != testQueueURL {
		t.Errorf("expected external name=%q, got %q", testQueueURL, got)
	}
}

func TestCreate_PublishesConnectionDetail(t *testing.T) {
	e := &queue.ExternalClient{Client: &mockSQSClient{
		createOut: &awssqs.CreateQueueOutput{QueueUrl: ptrStr(testQueueURL)},
	}}
	cr := makeCR("test-queue")

	creation, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	urlDetail, ok := creation.ConnectionDetails["url"]
	if !ok {
		t.Fatal("expected 'url' in connection details")
	}
	if string(urlDetail) != testQueueURL {
		t.Errorf("expected url=%q, got %q", testQueueURL, string(urlDetail))
	}
}

func TestCreate_UsesNameFieldWhenSet(t *testing.T) {
	var capturedInput *awssqs.CreateQueueInput
	mock := &mockSQSClient{
		createOut: &awssqs.CreateQueueOutput{QueueUrl: ptrStr(testQueueURL)},
	}
	// Wrap to capture
	captureCreate := func(ctx context.Context, params *awssqs.CreateQueueInput, optFns ...func(*awssqs.Options)) (*awssqs.CreateQueueOutput, error) {
		capturedInput = params
		return mock.createOut, nil
	}
	_ = captureCreate

	// Use a different approach: use mock directly and check the name via test double
	e := &queue.ExternalClient{Client: &captureClient{inner: mock, capturedCreate: &capturedInput}}
	cr := makeCR("test-queue")
	cr.Spec.ForProvider.Name = ptrStr("my-custom-queue")

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if capturedInput == nil {
		t.Fatal("expected CreateQueue to be called")
	}
	if capturedInput.QueueName == nil || *capturedInput.QueueName != "my-custom-queue" {
		t.Errorf("expected queue name 'my-custom-queue', got %v", capturedInput.QueueName)
	}
}

func TestCreate_FallsBackToK8sName(t *testing.T) {
	var capturedInput *awssqs.CreateQueueInput
	mock := &mockSQSClient{
		createOut: &awssqs.CreateQueueOutput{QueueUrl: ptrStr(testQueueURL)},
	}
	e := &queue.ExternalClient{Client: &captureClient{inner: mock, capturedCreate: &capturedInput}}
	cr := makeCR("test-queue")
	// No Name set

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if capturedInput == nil {
		t.Fatal("expected CreateQueue to be called")
	}
	if capturedInput.QueueName == nil || *capturedInput.QueueName != "test-queue" {
		t.Errorf("expected queue name 'test-queue', got %v", capturedInput.QueueName)
	}
}

func TestCreate_Error_ReturnsError(t *testing.T) {
	e := &queue.ExternalClient{Client: &mockSQSClient{
		createErr: fmt.Errorf("create error"),
	}}
	cr := makeCR("test-queue")

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_CallsSetQueueAttributes(t *testing.T) {
	mock := &mockSQSClient{
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}
	e := &queue.ExternalClient{Client: mock}
	cr := makeCR(testQueueURL)
	cr.Spec.ForProvider.DelaySeconds = ptrF64(60)

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastSetAttrs == nil {
		t.Fatal("expected SetQueueAttributes to be called")
	}
	if got := mock.lastSetAttrs.Attributes["DelaySeconds"]; got != "60" {
		t.Errorf("expected DelaySeconds=60, got %q", got)
	}
}

func TestUpdate_DoesNotIncludeFifoQueue(t *testing.T) {
	mock := &mockSQSClient{
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}
	e := &queue.ExternalClient{Client: mock}
	cr := makeCR(testQueueURL)
	cr.Spec.ForProvider.FifoQueue = ptrBool(true)

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.lastSetAttrs != nil {
		if _, ok := mock.lastSetAttrs.Attributes["FifoQueue"]; ok {
			t.Error("FifoQueue must not be included in SetQueueAttributes (immutable)")
		}
	}
}

func TestUpdate_ReconcilesTags(t *testing.T) {
	mock := &mockSQSClient{
		listTagsOut: &awssqs.ListQueueTagsOutput{Tags: map[string]string{"old": "tag"}},
	}
	e := &queue.ExternalClient{Client: mock}
	cr := makeCR(testQueueURL)
	cr.Spec.ForProvider.Tags = map[string]*string{"new": ptrStr("tag")}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Should tag the new tag and untag the old tag
	if mock.lastTagInput == nil {
		t.Error("expected TagQueue to be called for new tag")
	}
	if mock.lastUntagInput == nil {
		t.Error("expected UntagQueue to be called for old tag")
	}
}

// ── Delete tests ───────────────────────────────────────────────────────────────

func TestDelete_CallsDeleteQueue(t *testing.T) {
	mock := &mockSQSClient{}
	e := &queue.ExternalClient{Client: mock}
	cr := makeCR(testQueueURL)

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDelete_IdempotentWhenQueueDoesNotExist(t *testing.T) {
	e := &queue.ExternalClient{Client: &mockSQSClient{
		deleteErr: &sqstypes.QueueDoesNotExist{Message: ptrStr("not found")},
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Errorf("expected no error for QueueDoesNotExist, got %v", err)
	}
}

func TestDelete_NonURL_ReturnsNoError(t *testing.T) {
	e := &queue.ExternalClient{Client: &mockSQSClient{
		deleteErr: fmt.Errorf("should not be called"),
	}}
	cr := makeCR("not-a-url")

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Errorf("expected no error for non-URL external name, got %v", err)
	}
}

func TestDelete_Error_ReturnsError(t *testing.T) {
	e := &queue.ExternalClient{Client: &mockSQSClient{
		deleteErr: fmt.Errorf("delete error"),
	}}
	cr := makeCR(testQueueURL)

	_, err := e.Delete(context.Background(), cr)
	if err == nil {
		t.Error("expected error, got nil")
	}
}

// ── Observe: UpToDate for all mutable fields ───────────────────────────────────

func TestObserve_NotUpToDate_WhenVisibilityTimeoutDiffers(t *testing.T) {
	attrs := baseAttrs()
	attrs["VisibilityTimeout"] = "30"
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)
	cr.Spec.ForProvider.VisibilityTimeoutSeconds = ptrF64(60)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when VisibilityTimeout differs")
	}
}

func TestObserve_NotUpToDate_WhenMaxMessageSizeDiffers(t *testing.T) {
	attrs := baseAttrs()
	attrs["MaximumMessageSize"] = "262144"
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)
	cr.Spec.ForProvider.MaxMessageSize = ptrF64(1024)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when MaximumMessageSize differs")
	}
}

func TestObserve_NotUpToDate_WhenKMSMasterKeyIDDiffers(t *testing.T) {
	attrs := baseAttrs()
	attrs["KmsMasterKeyId"] = "old-key"
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)
	cr.Spec.ForProvider.KMSMasterKeyID = ptrStr("new-key")

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when KmsMasterKeyId differs")
	}
}

func TestObserve_NotUpToDate_WhenRedriveAllowPolicyDiffers(t *testing.T) {
	attrs := baseAttrs()
	attrs["RedriveAllowPolicy"] = `{"redrivePermission":"allowAll"}`
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)
	cr.Spec.ForProvider.RedriveAllowPolicy = ptrStr(`{"redrivePermission":"denyAll"}`)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when RedriveAllowPolicy differs")
	}
}

func TestObserve_LateInitializesDeduplicationScope(t *testing.T) {
	attrs := baseAttrs()
	attrs["DeduplicationScope"] = "messageGroup"
	attrs["FifoQueue"] = "true"
	e := &queue.ExternalClient{Client: &mockSQSClient{
		getAttrsOut: &awssqs.GetQueueAttributesOutput{Attributes: attrs},
		listTagsOut: &awssqs.ListQueueTagsOutput{},
	}}
	cr := makeCR(testQueueURL)
	cr.Spec.ForProvider.FifoQueue = ptrBool(true)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true for DeduplicationScope")
	}
	if cr.Spec.ForProvider.DeduplicationScope == nil || *cr.Spec.ForProvider.DeduplicationScope != "messageGroup" {
		t.Errorf("expected DeduplicationScope=messageGroup, got %v", cr.Spec.ForProvider.DeduplicationScope)
	}
}

// ── captureClient ──────────────────────────────────────────────────────────────

// captureClient wraps mockSQSClient and captures CreateQueue input.
type captureClient struct {
	inner          *mockSQSClient
	capturedCreate **awssqs.CreateQueueInput
}

func (c *captureClient) GetQueueAttributes(ctx context.Context, params *awssqs.GetQueueAttributesInput, optFns ...func(*awssqs.Options)) (*awssqs.GetQueueAttributesOutput, error) {
	return c.inner.GetQueueAttributes(ctx, params, optFns...)
}

func (c *captureClient) CreateQueue(_ context.Context, params *awssqs.CreateQueueInput, _ ...func(*awssqs.Options)) (*awssqs.CreateQueueOutput, error) {
	*c.capturedCreate = params
	return c.inner.createOut, c.inner.createErr
}

func (c *captureClient) SetQueueAttributes(ctx context.Context, params *awssqs.SetQueueAttributesInput, optFns ...func(*awssqs.Options)) (*awssqs.SetQueueAttributesOutput, error) {
	return c.inner.SetQueueAttributes(ctx, params, optFns...)
}

func (c *captureClient) DeleteQueue(ctx context.Context, params *awssqs.DeleteQueueInput, optFns ...func(*awssqs.Options)) (*awssqs.DeleteQueueOutput, error) {
	return c.inner.DeleteQueue(ctx, params, optFns...)
}

func (c *captureClient) ListQueueTags(ctx context.Context, params *awssqs.ListQueueTagsInput, optFns ...func(*awssqs.Options)) (*awssqs.ListQueueTagsOutput, error) {
	return c.inner.ListQueueTags(ctx, params, optFns...)
}

func (c *captureClient) TagQueue(ctx context.Context, params *awssqs.TagQueueInput, optFns ...func(*awssqs.Options)) (*awssqs.TagQueueOutput, error) {
	return c.inner.TagQueue(ctx, params, optFns...)
}

func (c *captureClient) UntagQueue(ctx context.Context, params *awssqs.UntagQueueInput, optFns ...func(*awssqs.Options)) (*awssqs.UntagQueueOutput, error) {
	return c.inner.UntagQueue(ctx, params, optFns...)
}

func (c *captureClient) GetQueueUrl(ctx context.Context, params *awssqs.GetQueueUrlInput, optFns ...func(*awssqs.Options)) (*awssqs.GetQueueUrlOutput, error) {
	return c.inner.GetQueueUrl(ctx, params, optFns...)
}

// Verify ExternalClient fields are accessible from tests.
var _ queue.SQSClient = (*mockSQSClient)(nil)
var _ managed.TypedExternalClient[*clusternative.QueueRAW] = (*stubTypedClient)(nil)

type stubTypedClient struct{}

func (*stubTypedClient) Observe(_ context.Context, _ *clusternative.QueueRAW) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, nil
}
func (*stubTypedClient) Create(_ context.Context, _ *clusternative.QueueRAW) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, nil
}
func (*stubTypedClient) Update(_ context.Context, _ *clusternative.QueueRAW) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}
func (*stubTypedClient) Delete(_ context.Context, _ *clusternative.QueueRAW) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, nil
}
func (*stubTypedClient) Disconnect(_ context.Context) error { return nil }
