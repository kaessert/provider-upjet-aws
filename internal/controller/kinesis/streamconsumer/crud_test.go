// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package streamconsumer

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awskinesis "github.com/aws/aws-sdk-go-v2/service/kinesis"
	ktypes "github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"

	v1beta1native "github.com/upbound/provider-aws/v2/apis/cluster/kinesis/v1beta1/native"
	"github.com/upbound/provider-aws/v2/internal/native"
)

// ── Mock client ────────────────────────────────────────────────────────────────

// mockKinesisConsumerClient is a test double for KinesisConsumerClient.
type mockKinesisConsumerClient struct {
	registerStreamConsumerFn   func(ctx context.Context, params *awskinesis.RegisterStreamConsumerInput, optFns ...func(*awskinesis.Options)) (*awskinesis.RegisterStreamConsumerOutput, error)
	describeStreamConsumerFn   func(ctx context.Context, params *awskinesis.DescribeStreamConsumerInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamConsumerOutput, error)
	deregisterStreamConsumerFn func(ctx context.Context, params *awskinesis.DeregisterStreamConsumerInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DeregisterStreamConsumerOutput, error)
	listTagsForResourceFn      func(ctx context.Context, params *awskinesis.ListTagsForResourceInput, optFns ...func(*awskinesis.Options)) (*awskinesis.ListTagsForResourceOutput, error)
	tagResourceFn              func(ctx context.Context, params *awskinesis.TagResourceInput, optFns ...func(*awskinesis.Options)) (*awskinesis.TagResourceOutput, error)
	untagResourceFn            func(ctx context.Context, params *awskinesis.UntagResourceInput, optFns ...func(*awskinesis.Options)) (*awskinesis.UntagResourceOutput, error)
}

func (m *mockKinesisConsumerClient) RegisterStreamConsumer(ctx context.Context, params *awskinesis.RegisterStreamConsumerInput, optFns ...func(*awskinesis.Options)) (*awskinesis.RegisterStreamConsumerOutput, error) {
	return m.registerStreamConsumerFn(ctx, params, optFns...)
}
func (m *mockKinesisConsumerClient) DescribeStreamConsumer(ctx context.Context, params *awskinesis.DescribeStreamConsumerInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamConsumerOutput, error) {
	return m.describeStreamConsumerFn(ctx, params, optFns...)
}
func (m *mockKinesisConsumerClient) DeregisterStreamConsumer(ctx context.Context, params *awskinesis.DeregisterStreamConsumerInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DeregisterStreamConsumerOutput, error) {
	return m.deregisterStreamConsumerFn(ctx, params, optFns...)
}
func (m *mockKinesisConsumerClient) ListTagsForResource(ctx context.Context, params *awskinesis.ListTagsForResourceInput, optFns ...func(*awskinesis.Options)) (*awskinesis.ListTagsForResourceOutput, error) {
	return m.listTagsForResourceFn(ctx, params, optFns...)
}
func (m *mockKinesisConsumerClient) TagResource(ctx context.Context, params *awskinesis.TagResourceInput, optFns ...func(*awskinesis.Options)) (*awskinesis.TagResourceOutput, error) {
	return m.tagResourceFn(ctx, params, optFns...)
}
func (m *mockKinesisConsumerClient) UntagResource(ctx context.Context, params *awskinesis.UntagResourceInput, optFns ...func(*awskinesis.Options)) (*awskinesis.UntagResourceOutput, error) {
	return m.untagResourceFn(ctx, params, optFns...)
}

// ── helpers ────────────────────────────────────────────────────────────────────

const (
	testConsumerName = "my-consumer"
	testStreamARN    = "arn:aws:kinesis:us-east-1:123456789012:stream/my-stream"
	testConsumerARN  = "arn:aws:kinesis:us-east-1:123456789012:stream/my-stream/consumer/my-consumer:1234567890"
	testRegion       = "us-east-1"
)

// newTestCR builds a minimal cluster-scoped StreamConsumerRAW for tests.
func newTestCR(name, extName string) *v1beta1native.StreamConsumerRAW {
	cr := &v1beta1native.StreamConsumerRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: v1beta1native.StreamConsumerRAWSpec{
			ForProvider: v1beta1native.StreamConsumerRAWParameters{
				Name:      aws.String(testConsumerName),
				Region:    aws.String(testRegion),
				StreamArn: aws.String(testStreamARN),
			},
		},
	}
	if extName != "" {
		native.SetExternalName(cr, extName)
	}
	return cr
}

// noopListTagsForResource returns an empty tag list.
func noopListTagsForResource(_ context.Context, _ *awskinesis.ListTagsForResourceInput, _ ...func(*awskinesis.Options)) (*awskinesis.ListTagsForResourceOutput, error) {
	return &awskinesis.ListTagsForResourceOutput{Tags: []ktypes.Tag{}}, nil
}

// activeConsumerDescription returns an ACTIVE consumer description.
func activeConsumerDescription(name, consumerARN, streamARN string) *ktypes.ConsumerDescription {
	ts := time.Now()
	return &ktypes.ConsumerDescription{
		ConsumerARN:               aws.String(consumerARN),
		ConsumerName:              aws.String(name),
		ConsumerStatus:            ktypes.ConsumerStatusActive,
		ConsumerCreationTimestamp: &ts,
		StreamARN:                 aws.String(streamARN),
	}
}

// ── Observe tests ──────────────────────────────────────────────────────────────

// TestObserve_EmptyExternalName verifies that an empty external name annotation
// results in ResourceExists: false without any API call.
func TestObserve_EmptyExternalName(t *testing.T) {
	cr := newTestCR("my-consumer", "") // no external name
	e := &ExternalClient{Client: &mockKinesisConsumerClient{}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty external name")
	}
}

// TestObserve_ResourceNotFound verifies that ResourceNotFoundException is
// translated to ResourceExists=false (not an error).
func TestObserve_ResourceNotFound(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)

	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		describeStreamConsumerFn: func(_ context.Context, _ *awskinesis.DescribeStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamConsumerOutput, error) {
			return nil, &ktypes.ResourceNotFoundException{Message: aws.String("Consumer not found")}
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when consumer not found")
	}
}

// TestObserve_CREATING verifies that a CREATING consumer sets Unavailable
// condition and returns ResourceExists=true, ResourceUpToDate=true.
func TestObserve_CREATING(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)
	ts := time.Now()
	desc := &ktypes.ConsumerDescription{
		ConsumerARN:               aws.String(testConsumerARN),
		ConsumerName:              aws.String(testConsumerName),
		ConsumerStatus:            ktypes.ConsumerStatusCreating,
		ConsumerCreationTimestamp: &ts,
		StreamARN:                 aws.String(testStreamARN),
	}

	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		describeStreamConsumerFn: func(_ context.Context, _ *awskinesis.DescribeStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamConsumerOutput, error) {
			return &awskinesis.DescribeStreamConsumerOutput{ConsumerDescription: desc}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for CREATING consumer")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for CREATING consumer (no update should be triggered)")
	}
	// Check Unavailable condition is set.
	cond := cr.GetCondition(xpv1.TypeReady)
	if cond.Status != "False" {
		t.Errorf("expected Ready=False (Unavailable) for CREATING consumer, got %v", cond.Status)
	}
}

// TestObserve_DELETING verifies that a DELETING consumer sets Unavailable
// condition and returns ResourceExists=true, ResourceUpToDate=true.
func TestObserve_DELETING(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)
	ts := time.Now()
	desc := &ktypes.ConsumerDescription{
		ConsumerARN:               aws.String(testConsumerARN),
		ConsumerName:              aws.String(testConsumerName),
		ConsumerStatus:            ktypes.ConsumerStatusDeleting,
		ConsumerCreationTimestamp: &ts,
		StreamARN:                 aws.String(testStreamARN),
	}

	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		describeStreamConsumerFn: func(_ context.Context, _ *awskinesis.DescribeStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamConsumerOutput, error) {
			return &awskinesis.DescribeStreamConsumerOutput{ConsumerDescription: desc}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for DELETING consumer")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for DELETING consumer (no update should be triggered)")
	}
	// Check Unavailable condition is set.
	cond := cr.GetCondition(xpv1.TypeReady)
	if cond.Status != "False" {
		t.Errorf("expected Ready=False (Unavailable) for DELETING consumer, got %v", cond.Status)
	}
}

// TestObserve_ACTIVE_UpToDate verifies that an ACTIVE consumer with matching
// spec returns ResourceExists=true, ResourceUpToDate=true.
func TestObserve_ACTIVE_UpToDate(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)

	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		describeStreamConsumerFn: func(_ context.Context, _ *awskinesis.DescribeStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamConsumerOutput, error) {
			return &awskinesis.DescribeStreamConsumerOutput{
				ConsumerDescription: activeConsumerDescription(testConsumerName, testConsumerARN, testStreamARN),
			}, nil
		},
		listTagsForResourceFn: noopListTagsForResource,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for active consumer")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when spec matches AWS state")
	}
}

// TestObserve_ACTIVE_TagsDiffer verifies that an ACTIVE consumer with different
// tags returns ResourceUpToDate=false.
func TestObserve_ACTIVE_TagsDiffer(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)
	val := "desired-value"
	cr.Spec.ForProvider.Tags = map[string]*string{"env": &val}

	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		describeStreamConsumerFn: func(_ context.Context, _ *awskinesis.DescribeStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamConsumerOutput, error) {
			return &awskinesis.DescribeStreamConsumerOutput{
				ConsumerDescription: activeConsumerDescription(testConsumerName, testConsumerARN, testStreamARN),
			}, nil
		},
		listTagsForResourceFn: noopListTagsForResource, // no tags on AWS side
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when tags differ")
	}
}

// TestObserve_SetsObservation verifies that Observe populates atProvider fields.
func TestObserve_SetsObservation(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)

	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		describeStreamConsumerFn: func(_ context.Context, _ *awskinesis.DescribeStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamConsumerOutput, error) {
			return &awskinesis.DescribeStreamConsumerOutput{
				ConsumerDescription: activeConsumerDescription(testConsumerName, testConsumerARN, testStreamARN),
			}, nil
		},
		listTagsForResourceFn: noopListTagsForResource,
	}}

	_, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff := cmp.Diff(aws.String(testConsumerARN), cr.Status.AtProvider.Arn); diff != "" {
		t.Errorf("atProvider.arn mismatch (-want +got):\n%s", diff)
	}
	if cr.Status.AtProvider.ID == nil || *cr.Status.AtProvider.ID != testConsumerARN {
		t.Errorf("expected atProvider.id to be %s, got %v", testConsumerARN, cr.Status.AtProvider.ID)
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

// TestCreate_SetsExternalName verifies that Create calls RegisterStreamConsumer
// and sets the consumer ARN as the external name.
func TestCreate_SetsExternalName(t *testing.T) {
	cr := newTestCR("my-consumer", "")

	var calledWithName, calledWithStreamARN string
	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		registerStreamConsumerFn: func(_ context.Context, params *awskinesis.RegisterStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.RegisterStreamConsumerOutput, error) {
			calledWithName = aws.ToString(params.ConsumerName)
			calledWithStreamARN = aws.ToString(params.StreamARN)
			ts := time.Now()
			return &awskinesis.RegisterStreamConsumerOutput{
				Consumer: &ktypes.Consumer{
					ConsumerARN:               aws.String(testConsumerARN),
					ConsumerName:              aws.String(testConsumerName),
					ConsumerStatus:            ktypes.ConsumerStatusCreating,
					ConsumerCreationTimestamp: &ts,
				},
			}, nil
		},
	}}

	creation, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = creation

	// External name must be set to the consumer ARN.
	if extName := native.GetExternalName(cr); extName != testConsumerARN {
		t.Errorf("expected external name=%q, got %q", testConsumerARN, extName)
	}
	if calledWithName != testConsumerName {
		t.Errorf("expected ConsumerName=%q, got %q", testConsumerName, calledWithName)
	}
	if calledWithStreamARN != testStreamARN {
		t.Errorf("expected StreamARN=%q, got %q", testStreamARN, calledWithStreamARN)
	}
}

// TestCreate_WithTags verifies that Create passes tags to RegisterStreamConsumer.
func TestCreate_WithTags(t *testing.T) {
	cr := newTestCR("my-consumer", "")
	val := "production"
	cr.Spec.ForProvider.Tags = map[string]*string{"env": &val}

	var capturedTags map[string]string
	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		registerStreamConsumerFn: func(_ context.Context, params *awskinesis.RegisterStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.RegisterStreamConsumerOutput, error) {
			capturedTags = params.Tags
			ts := time.Now()
			return &awskinesis.RegisterStreamConsumerOutput{
				Consumer: &ktypes.Consumer{
					ConsumerARN:               aws.String(testConsumerARN),
					ConsumerName:              aws.String(testConsumerName),
					ConsumerStatus:            ktypes.ConsumerStatusCreating,
					ConsumerCreationTimestamp: &ts,
				},
			}, nil
		},
	}}

	if _, err := e.Create(context.Background(), cr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedTags["env"] != "production" {
		t.Errorf("expected tag env=production in RegisterStreamConsumer call, got %v", capturedTags)
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

// TestUpdate_TagsOnly verifies that Update reconciles tags when only tags differ.
func TestUpdate_TagsOnly(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)
	val := "new-value"
	cr.Spec.ForProvider.Tags = map[string]*string{"env": &val}

	var tagsCalled bool
	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		describeStreamConsumerFn: func(_ context.Context, _ *awskinesis.DescribeStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamConsumerOutput, error) {
			return &awskinesis.DescribeStreamConsumerOutput{
				ConsumerDescription: activeConsumerDescription(testConsumerName, testConsumerARN, testStreamARN),
			}, nil
		},
		listTagsForResourceFn: noopListTagsForResource,
		tagResourceFn: func(_ context.Context, _ *awskinesis.TagResourceInput, _ ...func(*awskinesis.Options)) (*awskinesis.TagResourceOutput, error) {
			tagsCalled = true
			return &awskinesis.TagResourceOutput{}, nil
		},
		untagResourceFn: func(_ context.Context, _ *awskinesis.UntagResourceInput, _ ...func(*awskinesis.Options)) (*awskinesis.UntagResourceOutput, error) {
			return &awskinesis.UntagResourceOutput{}, nil
		},
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tagsCalled {
		t.Error("expected TagResource to be called for tag reconciliation")
	}
}

// TestUpdate_ImmutableNameChange verifies that Update returns an error when
// the consumer name has changed (immutable field).
func TestUpdate_ImmutableNameChange(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)
	newName := "different-name"
	cr.Spec.ForProvider.Name = &newName // changed from testConsumerName

	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		describeStreamConsumerFn: func(_ context.Context, _ *awskinesis.DescribeStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamConsumerOutput, error) {
			return &awskinesis.DescribeStreamConsumerOutput{
				ConsumerDescription: activeConsumerDescription(testConsumerName, testConsumerARN, testStreamARN),
			}, nil
		},
	}}

	_, err := e.Update(context.Background(), cr)
	if err == nil {
		t.Error("expected error when consumer name is changed (immutable field)")
	}
}

// TestUpdate_ImmutableStreamARNChange verifies that Update returns an error
// when stream_arn has changed (immutable field).
func TestUpdate_ImmutableStreamARNChange(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)
	newStreamARN := "arn:aws:kinesis:us-east-1:123456789012:stream/different-stream"
	cr.Spec.ForProvider.StreamArn = &newStreamARN

	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		describeStreamConsumerFn: func(_ context.Context, _ *awskinesis.DescribeStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamConsumerOutput, error) {
			return &awskinesis.DescribeStreamConsumerOutput{
				ConsumerDescription: activeConsumerDescription(testConsumerName, testConsumerARN, testStreamARN),
			}, nil
		},
	}}

	_, err := e.Update(context.Background(), cr)
	if err == nil {
		t.Error("expected error when stream_arn is changed (immutable field)")
	}
}

// ── Delete tests ───────────────────────────────────────────────────────────────

// TestDelete_Success verifies that Delete calls DeregisterStreamConsumer.
func TestDelete_Success(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)

	var deregisterCalled bool
	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		deregisterStreamConsumerFn: func(_ context.Context, params *awskinesis.DeregisterStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.DeregisterStreamConsumerOutput, error) {
			deregisterCalled = true
			if aws.ToString(params.ConsumerARN) != testConsumerARN {
				t.Errorf("expected ConsumerARN=%q, got %q", testConsumerARN, aws.ToString(params.ConsumerARN))
			}
			return &awskinesis.DeregisterStreamConsumerOutput{}, nil
		},
	}}

	del, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = del
	if !deregisterCalled {
		t.Error("expected DeregisterStreamConsumer to be called")
	}
}

// TestDelete_Idempotent verifies that Delete returns nil when the consumer
// is already gone (ResourceNotFoundException is swallowed).
func TestDelete_Idempotent(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)

	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		deregisterStreamConsumerFn: func(_ context.Context, _ *awskinesis.DeregisterStreamConsumerInput, _ ...func(*awskinesis.Options)) (*awskinesis.DeregisterStreamConsumerOutput, error) {
			return nil, &ktypes.ResourceNotFoundException{Message: aws.String("Consumer not found")}
		},
	}}

	del, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for idempotent delete, got %v", err)
	}
	_ = del
}

// TestDelete_EmptyExternalName verifies that Delete returns nil when external
// name is empty (no-op).
func TestDelete_EmptyExternalName(t *testing.T) {
	cr := newTestCR("my-consumer", "") // no external name

	e := &ExternalClient{Client: &mockKinesisConsumerClient{}}

	del, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = del
}

// ── isUpToDate tests ──────────────────────────────────────────────────────────

// TestIsUpToDate_MatchingSpec verifies isUpToDate returns true for matching spec.
func TestIsUpToDate_MatchingSpec(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)

	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		listTagsForResourceFn: noopListTagsForResource,
	}}

	desc := activeConsumerDescription(testConsumerName, testConsumerARN, testStreamARN)
	upToDate, err := e.isUpToDate(context.Background(), cr, desc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !upToDate {
		t.Error("expected isUpToDate=true when spec matches AWS state")
	}
}

// TestIsUpToDate_TagsDiffer verifies isUpToDate returns false when tags differ.
func TestIsUpToDate_TagsDiffer(t *testing.T) {
	cr := newTestCR("my-consumer", testConsumerARN)
	val := "new-env"
	cr.Spec.ForProvider.Tags = map[string]*string{"env": &val}

	e := &ExternalClient{Client: &mockKinesisConsumerClient{
		listTagsForResourceFn: noopListTagsForResource,
	}}

	desc := activeConsumerDescription(testConsumerName, testConsumerARN, testStreamARN)
	upToDate, err := e.isUpToDate(context.Background(), cr, desc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if upToDate {
		t.Error("expected isUpToDate=false when tags differ")
	}
}
