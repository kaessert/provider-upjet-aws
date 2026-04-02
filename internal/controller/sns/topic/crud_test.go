// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package topic_test

import (
	"context"
	"errors"
	"testing"

	awssns "github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native"
	"github.com/upbound/provider-aws/v2/internal/controller/sns/topic"
)

// ── mock SNS client ────────────────────────────────────────────────────────────

type mockSNSClient struct {
	createTopicOut   *awssns.CreateTopicOutput
	createTopicErr   error
	getAttrsOut      *awssns.GetTopicAttributesOutput
	getAttrsErr      error
	setAttrsErr      error
	deleteTopicErr   error
	listTagsOut      *awssns.ListTagsForResourceOutput
	listTagsErr      error
	tagResourceErr   error
	untagResourceErr error
	lastSetAttrs     []*awssns.SetTopicAttributesInput
	lastTagInput     *awssns.TagResourceInput
	lastUntagInput   *awssns.UntagResourceInput
}

func (m *mockSNSClient) CreateTopic(_ context.Context, params *awssns.CreateTopicInput, _ ...func(*awssns.Options)) (*awssns.CreateTopicOutput, error) {
	return m.createTopicOut, m.createTopicErr
}

func (m *mockSNSClient) GetTopicAttributes(_ context.Context, _ *awssns.GetTopicAttributesInput, _ ...func(*awssns.Options)) (*awssns.GetTopicAttributesOutput, error) {
	if m.getAttrsOut != nil {
		return m.getAttrsOut, m.getAttrsErr
	}
	return &awssns.GetTopicAttributesOutput{Attributes: map[string]string{}}, m.getAttrsErr
}

func (m *mockSNSClient) SetTopicAttributes(_ context.Context, params *awssns.SetTopicAttributesInput, _ ...func(*awssns.Options)) (*awssns.SetTopicAttributesOutput, error) {
	m.lastSetAttrs = append(m.lastSetAttrs, params)
	return &awssns.SetTopicAttributesOutput{}, m.setAttrsErr
}

func (m *mockSNSClient) DeleteTopic(_ context.Context, _ *awssns.DeleteTopicInput, _ ...func(*awssns.Options)) (*awssns.DeleteTopicOutput, error) {
	return &awssns.DeleteTopicOutput{}, m.deleteTopicErr
}

func (m *mockSNSClient) ListTagsForResource(_ context.Context, _ *awssns.ListTagsForResourceInput, _ ...func(*awssns.Options)) (*awssns.ListTagsForResourceOutput, error) {
	if m.listTagsOut != nil {
		return m.listTagsOut, m.listTagsErr
	}
	return &awssns.ListTagsForResourceOutput{}, m.listTagsErr
}

func (m *mockSNSClient) TagResource(_ context.Context, params *awssns.TagResourceInput, _ ...func(*awssns.Options)) (*awssns.TagResourceOutput, error) {
	m.lastTagInput = params
	return &awssns.TagResourceOutput{}, m.tagResourceErr
}

func (m *mockSNSClient) UntagResource(_ context.Context, params *awssns.UntagResourceInput, _ ...func(*awssns.Options)) (*awssns.UntagResourceOutput, error) {
	m.lastUntagInput = params
	return &awssns.UntagResourceOutput{}, m.untagResourceErr
}

// ── helpers ────────────────────────────────────────────────────────────────────

const (
	testTopicName = "my-test-topic"
	testTopicARN  = "arn:aws:sns:us-east-1:123456789012:my-test-topic"
	testRegion    = "us-east-1"
)

func ptrStr(s string) *string { return &s }

// makeCR creates a test TopicRAW with the external name already set and an ARN
// pre-populated in atProvider (simulating a resource that has been created).
func makeCR(extName string) *clusternative.TopicRAW {
	cr := &clusternative.TopicRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-topic",
		},
		Spec: clusternative.TopicRAWSpec{
			ForProvider: clusternative.TopicRAWParameters{
				Region: ptrStr(testRegion),
			},
		},
	}
	meta.SetExternalName(cr, extName)
	return cr
}

// makeCRWithARN creates a CR that has already been through Create (has ARN stored).
func makeCRWithARN(extName, arn string) *clusternative.TopicRAW {
	cr := makeCR(extName)
	cr.Status.AtProvider = clusternative.TopicRAWObservation{
		Arn: ptrStr(arn),
	}
	return cr
}

func baseAttrs() map[string]string {
	return map[string]string{
		"TopicArn":         testTopicARN,
		"Owner":            "123456789012",
		"TracingConfig":    "PassThrough",
		"SignatureVersion": "1",
	}
}

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_NoARN_ReturnsNotExists(t *testing.T) {
	e := &topic.ExternalClient{Client: &mockSNSClient{}}
	// External name is the K8s name, no ARN stored yet
	cr := makeCR("test-topic")

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when no ARN is stored")
	}
}

func TestObserve_TopicNotFound_ReturnsNotExists(t *testing.T) {
	notFoundErr := &snstypes.NotFoundException{Message: ptrStr("topic not found")}
	e := &topic.ExternalClient{Client: &mockSNSClient{
		getAttrsErr: notFoundErr,
	}}
	cr := makeCRWithARN(testTopicName, testTopicARN)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for NotFoundException")
	}
}

func TestObserve_TopicExists_ReturnsExists(t *testing.T) {
	attrs := baseAttrs()
	e := &topic.ExternalClient{Client: &mockSNSClient{
		getAttrsOut: &awssns.GetTopicAttributesOutput{Attributes: attrs},
		listTagsOut: &awssns.ListTagsForResourceOutput{},
	}}
	cr := makeCRWithARN(testTopicName, testTopicARN)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
}

func TestObserve_SetsAtProvider(t *testing.T) {
	owner := "123456789012"
	attrs := map[string]string{
		"TopicArn":         testTopicARN,
		"Owner":            owner,
		"TracingConfig":    "PassThrough",
		"SignatureVersion": "1",
	}
	e := &topic.ExternalClient{Client: &mockSNSClient{
		getAttrsOut: &awssns.GetTopicAttributesOutput{Attributes: attrs},
		listTagsOut: &awssns.ListTagsForResourceOutput{},
	}}
	cr := makeCRWithARN(testTopicName, testTopicARN)

	_, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	atProvider := cr.GetAtProvider()
	if atProvider.Arn == nil || *atProvider.Arn != testTopicARN {
		t.Errorf("expected Arn=%q, got %v", testTopicARN, atProvider.Arn)
	}
	if atProvider.Owner == nil || *atProvider.Owner != owner {
		t.Errorf("expected Owner=%q, got %v", owner, atProvider.Owner)
	}
}

func TestObserve_UpToDate_WhenSpecMatchesAWS(t *testing.T) {
	attrs := baseAttrs()
	attrs["DisplayName"] = "My Topic"
	e := &topic.ExternalClient{Client: &mockSNSClient{
		getAttrsOut: &awssns.GetTopicAttributesOutput{Attributes: attrs},
		listTagsOut: &awssns.ListTagsForResourceOutput{},
	}}
	cr := makeCRWithARN(testTopicName, testTopicARN)
	cr.Spec.ForProvider.DisplayName = ptrStr("My Topic")

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when spec matches AWS")
	}
}

func TestObserve_NotUpToDate_WhenDisplayNameDiffers(t *testing.T) {
	attrs := baseAttrs()
	attrs["DisplayName"] = "Old Name"
	e := &topic.ExternalClient{Client: &mockSNSClient{
		getAttrsOut: &awssns.GetTopicAttributesOutput{Attributes: attrs},
		listTagsOut: &awssns.ListTagsForResourceOutput{},
	}}
	cr := makeCRWithARN(testTopicName, testTopicARN)
	cr.Spec.ForProvider.DisplayName = ptrStr("New Name")

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when DisplayName differs")
	}
}

func TestObserve_NotUpToDate_WhenPolicyDiffers(t *testing.T) {
	attrs := baseAttrs()
	attrs["Policy"] = `{"Version":"2012-10-17","Statement":[]}`
	e := &topic.ExternalClient{Client: &mockSNSClient{
		getAttrsOut: &awssns.GetTopicAttributesOutput{Attributes: attrs},
		listTagsOut: &awssns.ListTagsForResourceOutput{},
	}}
	cr := makeCRWithARN(testTopicName, testTopicARN)
	cr.Spec.ForProvider.Policy = ptrStr(`{"Statement":[{"Effect":"Allow","Principal":"*","Action":"sns:Publish","Resource":"*"}],"Version":"2012-10-17"}`)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when policy semantics differ")
	}
}

func TestObserve_UpToDate_WhenPolicyEquivalent(t *testing.T) {
	policy1 := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"sns:Publish","Resource":"*"}]}`
	policy2 := `{"Statement":[{"Action":"sns:Publish","Effect":"Allow","Principal":"*","Resource":"*"}],"Version":"2012-10-17"}`
	attrs := baseAttrs()
	attrs["Policy"] = policy2
	e := &topic.ExternalClient{Client: &mockSNSClient{
		getAttrsOut: &awssns.GetTopicAttributesOutput{Attributes: attrs},
		listTagsOut: &awssns.ListTagsForResourceOutput{},
	}}
	cr := makeCRWithARN(testTopicName, testTopicARN)
	cr.Spec.ForProvider.Policy = ptrStr(policy1)

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when policies are semantically equivalent")
	}
}

func TestObserve_NotUpToDate_WhenTagsDiffer(t *testing.T) {
	attrs := baseAttrs()
	e := &topic.ExternalClient{Client: &mockSNSClient{
		getAttrsOut: &awssns.GetTopicAttributesOutput{Attributes: attrs},
		listTagsOut: &awssns.ListTagsForResourceOutput{
			Tags: []snstypes.Tag{{Key: ptrStr("env"), Value: ptrStr("old")}},
		},
	}}
	cr := makeCRWithARN(testTopicName, testTopicARN)
	cr.Spec.ForProvider.Tags = map[string]*string{"env": ptrStr("new")}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when tags differ")
	}
}

func TestObserve_LateInitializesTracingConfig(t *testing.T) {
	attrs := baseAttrs()
	attrs["TracingConfig"] = "PassThrough"
	e := &topic.ExternalClient{Client: &mockSNSClient{
		getAttrsOut: &awssns.GetTopicAttributesOutput{Attributes: attrs},
		listTagsOut: &awssns.ListTagsForResourceOutput{},
	}}
	cr := makeCRWithARN(testTopicName, testTopicARN)
	// TracingConfig not set in spec

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true when TracingConfig late-inited")
	}
	if cr.Spec.ForProvider.TracingConfig == nil || *cr.Spec.ForProvider.TracingConfig != "PassThrough" {
		t.Errorf("expected TracingConfig=PassThrough after late-init, got %v", cr.Spec.ForProvider.TracingConfig)
	}
}

func TestObserve_LateInitializesSignatureVersion(t *testing.T) {
	attrs := baseAttrs()
	// SignatureVersion already in baseAttrs as "1"
	e := &topic.ExternalClient{Client: &mockSNSClient{
		getAttrsOut: &awssns.GetTopicAttributesOutput{Attributes: attrs},
		listTagsOut: &awssns.ListTagsForResourceOutput{},
	}}
	cr := makeCRWithARN(testTopicName, testTopicARN)
	// SignatureVersion not set in spec

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true when SignatureVersion late-inited")
	}
	if cr.Spec.ForProvider.SignatureVersion == nil || *cr.Spec.ForProvider.SignatureVersion != 1.0 {
		t.Errorf("expected SignatureVersion=1.0 after late-init, got %v", cr.Spec.ForProvider.SignatureVersion)
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

func TestCreate_SetsExternalNameAndARN(t *testing.T) {
	e := &topic.ExternalClient{Client: &mockSNSClient{
		createTopicOut: &awssns.CreateTopicOutput{
			TopicArn: ptrStr(testTopicARN),
		},
	}}
	cr := makeCR("test-topic")

	creation, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// External name should be set to the full ARN (persisted atomically by
	// the managed reconciler so Observe can find the topic on the next cycle
	// even before status.atProvider is flushed).
	extName := meta.GetExternalName(cr)
	if extName != testTopicARN {
		t.Errorf("expected external name=%q (full ARN), got %q", testTopicARN, extName)
	}

	// atProvider.Arn should be set
	if cr.Status.AtProvider.Arn == nil || *cr.Status.AtProvider.Arn != testTopicARN {
		t.Errorf("expected atProvider.Arn=%q, got %v", testTopicARN, cr.Status.AtProvider.Arn)
	}

	// Connection details should include "arn"
	if arn, ok := creation.ConnectionDetails["arn"]; !ok {
		t.Error("expected connection details to contain 'arn' key")
	} else if string(arn) != testTopicARN {
		t.Errorf("expected connection detail arn=%q, got %q", testTopicARN, string(arn))
	}
}

func TestCreate_ReturnsError_WhenCreateFails(t *testing.T) {
	e := &topic.ExternalClient{Client: &mockSNSClient{
		createTopicErr: errors.New("API error"),
	}}
	cr := makeCR("test-topic")

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_SetsDisplayName(t *testing.T) {
	mock := &mockSNSClient{
		listTagsOut: &awssns.ListTagsForResourceOutput{},
	}
	e := &topic.ExternalClient{Client: mock}
	cr := makeCRWithARN(testTopicName, testTopicARN)
	cr.Spec.ForProvider.DisplayName = ptrStr("Updated Name")

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// SetTopicAttributes should have been called with DisplayName
	found := false
	for _, call := range mock.lastSetAttrs {
		if call.AttributeName != nil && *call.AttributeName == "DisplayName" {
			if call.AttributeValue != nil && *call.AttributeValue == "Updated Name" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("expected SetTopicAttributes called with DisplayName=Updated Name, got %v", mock.lastSetAttrs)
	}
}

func TestUpdate_ReconcilesTags(t *testing.T) {
	mock := &mockSNSClient{
		listTagsOut: &awssns.ListTagsForResourceOutput{
			Tags: []snstypes.Tag{{Key: ptrStr("old-key"), Value: ptrStr("old-val")}},
		},
	}
	e := &topic.ExternalClient{Client: mock}
	cr := makeCRWithARN(testTopicName, testTopicARN)
	cr.Spec.ForProvider.Tags = map[string]*string{"new-key": ptrStr("new-val")}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Should have called TagResource to add new tag
	if mock.lastTagInput == nil {
		t.Error("expected TagResource to be called")
	}
	// Should have called UntagResource to remove old tag
	if mock.lastUntagInput == nil {
		t.Error("expected UntagResource to be called")
	}
}

// ── Delete tests ───────────────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	e := &topic.ExternalClient{Client: &mockSNSClient{}}
	cr := makeCRWithARN(testTopicName, testTopicARN)

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDelete_Idempotent_WhenTopicNotFound(t *testing.T) {
	notFoundErr := &snstypes.NotFoundException{Message: ptrStr("topic not found")}
	e := &topic.ExternalClient{Client: &mockSNSClient{
		deleteTopicErr: notFoundErr,
	}}
	cr := makeCRWithARN(testTopicName, testTopicARN)

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for not-found during delete, got %v", err)
	}
}

func TestDelete_NoARN_ReturnsNil(t *testing.T) {
	e := &topic.ExternalClient{Client: &mockSNSClient{}}
	// CR has no ARN stored — nothing to delete
	cr := makeCR("test-topic")

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error when no ARN, got %v", err)
	}
}

func TestDelete_ReturnsError_WhenDeleteFails(t *testing.T) {
	e := &topic.ExternalClient{Client: &mockSNSClient{
		deleteTopicErr: errors.New("internal server error"),
	}}
	cr := makeCRWithARN(testTopicName, testTopicARN)

	_, err := e.Delete(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
