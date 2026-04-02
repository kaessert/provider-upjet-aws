// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package stream

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awskinesis "github.com/aws/aws-sdk-go-v2/service/kinesis"
	ktypes "github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1beta2native "github.com/upbound/provider-aws/v2/apis/cluster/kinesis/v1beta2/native"
	"github.com/upbound/provider-aws/v2/internal/native"
)

// ── Mock client ────────────────────────────────────────────────────────────────

// mockKinesisClient is a test double for KinesisStreamClient.
// Each field is a function that the corresponding method delegates to.
type mockKinesisClient struct {
	describeStreamSummaryFn     func(ctx context.Context, params *awskinesis.DescribeStreamSummaryInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error)
	createStreamFn              func(ctx context.Context, params *awskinesis.CreateStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.CreateStreamOutput, error)
	deleteStreamFn              func(ctx context.Context, params *awskinesis.DeleteStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DeleteStreamOutput, error)
	updateShardCountFn          func(ctx context.Context, params *awskinesis.UpdateShardCountInput, optFns ...func(*awskinesis.Options)) (*awskinesis.UpdateShardCountOutput, error)
	startStreamEncryptionFn     func(ctx context.Context, params *awskinesis.StartStreamEncryptionInput, optFns ...func(*awskinesis.Options)) (*awskinesis.StartStreamEncryptionOutput, error)
	stopStreamEncryptionFn      func(ctx context.Context, params *awskinesis.StopStreamEncryptionInput, optFns ...func(*awskinesis.Options)) (*awskinesis.StopStreamEncryptionOutput, error)
	increaseStreamRetentionFn   func(ctx context.Context, params *awskinesis.IncreaseStreamRetentionPeriodInput, optFns ...func(*awskinesis.Options)) (*awskinesis.IncreaseStreamRetentionPeriodOutput, error)
	decreaseStreamRetentionFn   func(ctx context.Context, params *awskinesis.DecreaseStreamRetentionPeriodInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DecreaseStreamRetentionPeriodOutput, error)
	updateStreamModeFn          func(ctx context.Context, params *awskinesis.UpdateStreamModeInput, optFns ...func(*awskinesis.Options)) (*awskinesis.UpdateStreamModeOutput, error)
	enableEnhancedMonitoringFn  func(ctx context.Context, params *awskinesis.EnableEnhancedMonitoringInput, optFns ...func(*awskinesis.Options)) (*awskinesis.EnableEnhancedMonitoringOutput, error)
	disableEnhancedMonitoringFn func(ctx context.Context, params *awskinesis.DisableEnhancedMonitoringInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DisableEnhancedMonitoringOutput, error)
	listTagsForStreamFn         func(ctx context.Context, params *awskinesis.ListTagsForStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.ListTagsForStreamOutput, error)
	addTagsToStreamFn           func(ctx context.Context, params *awskinesis.AddTagsToStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.AddTagsToStreamOutput, error)
	removeTagsFromStreamFn      func(ctx context.Context, params *awskinesis.RemoveTagsFromStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.RemoveTagsFromStreamOutput, error)
	updateMaxRecordSizeFn       func(ctx context.Context, params *awskinesis.UpdateMaxRecordSizeInput, optFns ...func(*awskinesis.Options)) (*awskinesis.UpdateMaxRecordSizeOutput, error)
}

func (m *mockKinesisClient) DescribeStreamSummary(ctx context.Context, params *awskinesis.DescribeStreamSummaryInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
	return m.describeStreamSummaryFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) CreateStream(ctx context.Context, params *awskinesis.CreateStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.CreateStreamOutput, error) {
	return m.createStreamFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) DeleteStream(ctx context.Context, params *awskinesis.DeleteStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DeleteStreamOutput, error) {
	return m.deleteStreamFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) UpdateShardCount(ctx context.Context, params *awskinesis.UpdateShardCountInput, optFns ...func(*awskinesis.Options)) (*awskinesis.UpdateShardCountOutput, error) {
	return m.updateShardCountFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) StartStreamEncryption(ctx context.Context, params *awskinesis.StartStreamEncryptionInput, optFns ...func(*awskinesis.Options)) (*awskinesis.StartStreamEncryptionOutput, error) {
	return m.startStreamEncryptionFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) StopStreamEncryption(ctx context.Context, params *awskinesis.StopStreamEncryptionInput, optFns ...func(*awskinesis.Options)) (*awskinesis.StopStreamEncryptionOutput, error) {
	return m.stopStreamEncryptionFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) IncreaseStreamRetentionPeriod(ctx context.Context, params *awskinesis.IncreaseStreamRetentionPeriodInput, optFns ...func(*awskinesis.Options)) (*awskinesis.IncreaseStreamRetentionPeriodOutput, error) {
	return m.increaseStreamRetentionFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) DecreaseStreamRetentionPeriod(ctx context.Context, params *awskinesis.DecreaseStreamRetentionPeriodInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DecreaseStreamRetentionPeriodOutput, error) {
	return m.decreaseStreamRetentionFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) UpdateStreamMode(ctx context.Context, params *awskinesis.UpdateStreamModeInput, optFns ...func(*awskinesis.Options)) (*awskinesis.UpdateStreamModeOutput, error) {
	return m.updateStreamModeFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) EnableEnhancedMonitoring(ctx context.Context, params *awskinesis.EnableEnhancedMonitoringInput, optFns ...func(*awskinesis.Options)) (*awskinesis.EnableEnhancedMonitoringOutput, error) {
	return m.enableEnhancedMonitoringFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) DisableEnhancedMonitoring(ctx context.Context, params *awskinesis.DisableEnhancedMonitoringInput, optFns ...func(*awskinesis.Options)) (*awskinesis.DisableEnhancedMonitoringOutput, error) {
	return m.disableEnhancedMonitoringFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) ListTagsForStream(ctx context.Context, params *awskinesis.ListTagsForStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.ListTagsForStreamOutput, error) {
	return m.listTagsForStreamFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) AddTagsToStream(ctx context.Context, params *awskinesis.AddTagsToStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.AddTagsToStreamOutput, error) {
	return m.addTagsToStreamFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) RemoveTagsFromStream(ctx context.Context, params *awskinesis.RemoveTagsFromStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.RemoveTagsFromStreamOutput, error) {
	return m.removeTagsFromStreamFn(ctx, params, optFns...)
}
func (m *mockKinesisClient) UpdateMaxRecordSize(ctx context.Context, params *awskinesis.UpdateMaxRecordSizeInput, optFns ...func(*awskinesis.Options)) (*awskinesis.UpdateMaxRecordSizeOutput, error) {
	return m.updateMaxRecordSizeFn(ctx, params, optFns...)
}

// ── helpers ────────────────────────────────────────────────────────────────────

const testStreamName = "my-test-stream"
const testStreamARN = "arn:aws:kinesis:us-east-1:123456789012:stream/my-test-stream"
const testRegion = "us-east-1"

// newTestCR builds a minimal cluster-scoped StreamRAW for tests.
func newTestCR(name string, extName string) *v1beta2native.StreamRAW {
	cr := &v1beta2native.StreamRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: v1beta2native.StreamRAWSpec{
			ForProvider: v1beta2native.StreamRAWParameters{
				Region: aws.String(testRegion),
			},
		},
	}
	if extName != "" {
		native.SetExternalName(cr, extName)
	}
	return cr
}

// noopListTags returns an empty tag list.
func noopListTags(ctx context.Context, params *awskinesis.ListTagsForStreamInput, optFns ...func(*awskinesis.Options)) (*awskinesis.ListTagsForStreamOutput, error) {
	return &awskinesis.ListTagsForStreamOutput{Tags: []ktypes.Tag{}, HasMoreTags: aws.Bool(false)}, nil
}

// activeStreamSummary returns a minimal ACTIVE stream summary.
func activeStreamSummary(name, arn string, shards int32, retention int32) *ktypes.StreamDescriptionSummary {
	mode := ktypes.StreamModeProvisioned
	return &ktypes.StreamDescriptionSummary{
		StreamARN:            aws.String(arn),
		StreamName:           aws.String(name),
		StreamStatus:         ktypes.StreamStatusActive,
		OpenShardCount:       aws.Int32(shards),
		RetentionPeriodHours: aws.Int32(retention),
		EncryptionType:       ktypes.EncryptionTypeNone,
		StreamModeDetails:    &ktypes.StreamModeDetails{StreamMode: mode},
		EnhancedMonitoring:   []ktypes.EnhancedMetrics{{ShardLevelMetrics: []ktypes.MetricsName{}}},
	}
}

// ── Observe tests ──────────────────────────────────────────────────────────────

// TestObserve_EmptyExternalName verifies that an empty external name annotation
// results in ResourceExists: false without any API call.
func TestObserve_EmptyExternalName(t *testing.T) {
	cr := newTestCR("my-stream", "") // no external name
	e := &ExternalClient{Client: &mockKinesisClient{}}

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
	cr := newTestCR("my-stream", testStreamName)

	e := &ExternalClient{Client: &mockKinesisClient{
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return nil, &ktypes.ResourceNotFoundException{Message: aws.String("Stream not found")}
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when stream not found")
	}
}

// TestObserve_ResourceExists_UpToDate verifies that an ACTIVE stream with
// matching spec returns ResourceExists=true, ResourceUpToDate=true.
func TestObserve_ResourceExists_UpToDate(t *testing.T) {
	shardCount := float64(1)
	retention := float64(24)
	encType := "NONE"
	streamMode := "PROVISIONED"

	cr := newTestCR("my-stream", testStreamName)
	cr.Spec.ForProvider.ShardCount = &shardCount
	cr.Spec.ForProvider.RetentionPeriod = &retention
	cr.Spec.ForProvider.EncryptionType = &encType
	cr.Spec.ForProvider.StreamModeDetails = &v1beta2native.StreamModeDetailsRAWParameters{
		StreamMode: &streamMode,
	}

	summary := activeStreamSummary(testStreamName, testStreamARN, 1, 24)

	e := &ExternalClient{Client: &mockKinesisClient{
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return &awskinesis.DescribeStreamSummaryOutput{StreamDescriptionSummary: summary}, nil
		},
		listTagsForStreamFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when spec matches AWS")
	}
}

// TestObserve_ResourceExists_NotUpToDate_ShardCount verifies that a shard count
// mismatch causes ResourceUpToDate=false.
func TestObserve_ResourceExists_NotUpToDate_ShardCount(t *testing.T) {
	shardCount := float64(4) // spec wants 4
	cr := newTestCR("my-stream", testStreamName)
	cr.Spec.ForProvider.ShardCount = &shardCount

	// AWS currently has 1 shard.
	summary := activeStreamSummary(testStreamName, testStreamARN, 1, 24)

	e := &ExternalClient{Client: &mockKinesisClient{
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return &awskinesis.DescribeStreamSummaryOutput{StreamDescriptionSummary: summary}, nil
		},
		listTagsForStreamFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when shard count differs")
	}
}

// TestObserve_Transitioning_ReturnsUpToDate verifies that a stream in CREATING
// state returns ResourceUpToDate=true to prevent spurious Update calls.
func TestObserve_Transitioning_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-stream", testStreamName)

	summary := &ktypes.StreamDescriptionSummary{
		StreamARN:            aws.String(testStreamARN),
		StreamName:           aws.String(testStreamName),
		StreamStatus:         ktypes.StreamStatusCreating,
		OpenShardCount:       aws.Int32(1),
		RetentionPeriodHours: aws.Int32(24),
		EncryptionType:       ktypes.EncryptionTypeNone,
	}

	e := &ExternalClient{Client: &mockKinesisClient{
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return &awskinesis.DescribeStreamSummaryOutput{StreamDescriptionSummary: summary}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for transitioning stream (prevents spurious Updates)")
	}
}

// TestObserve_LateInitialize verifies that nil spec fields are populated from
// AWS response during late-initialization.
func TestObserve_LateInitialize(t *testing.T) {
	cr := newTestCR("my-stream", testStreamName)
	// No spec fields set — late-init should populate them.

	summary := activeStreamSummary(testStreamName, testStreamARN, 2, 48)
	summary.MaxRecordSizeInKiB = aws.Int32(1024)

	e := &ExternalClient{Client: &mockKinesisClient{
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return &awskinesis.DescribeStreamSummaryOutput{StreamDescriptionSummary: summary}, nil
		},
		listTagsForStreamFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true when spec fields are nil")
	}
	// Verify individual fields were populated.
	if cr.Spec.ForProvider.ShardCount == nil || *cr.Spec.ForProvider.ShardCount != 2 {
		t.Errorf("expected shardCount=2, got %v", cr.Spec.ForProvider.ShardCount)
	}
	if cr.Spec.ForProvider.RetentionPeriod == nil || *cr.Spec.ForProvider.RetentionPeriod != 48 {
		t.Errorf("expected retentionPeriod=48, got %v", cr.Spec.ForProvider.RetentionPeriod)
	}
	if cr.Spec.ForProvider.EncryptionType == nil || *cr.Spec.ForProvider.EncryptionType != "NONE" {
		t.Errorf("expected encryptionType=NONE, got %v", cr.Spec.ForProvider.EncryptionType)
	}
	if cr.Spec.ForProvider.StreamModeDetails == nil {
		t.Error("expected streamModeDetails to be late-initialized")
	}
	if cr.Spec.ForProvider.MaxRecordSizeInKib == nil || *cr.Spec.ForProvider.MaxRecordSizeInKib != 1024 {
		t.Errorf("expected maxRecordSizeInKib=1024, got %v", cr.Spec.ForProvider.MaxRecordSizeInKib)
	}
}

// TestObserve_SetsARNAsExternalName verifies that the external name is set to
// the stream ARN after Observe populates status.atProvider.arn.
func TestObserve_SetsARNAsExternalName(t *testing.T) {
	cr := newTestCR("my-stream", testStreamName)

	summary := activeStreamSummary(testStreamName, testStreamARN, 1, 24)

	e := &ExternalClient{Client: &mockKinesisClient{
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return &awskinesis.DescribeStreamSummaryOutput{StreamDescriptionSummary: summary}, nil
		},
		listTagsForStreamFn: noopListTags,
	}}

	_, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After Observe, external name should be the ARN.
	got := native.GetExternalName(cr)
	if got != testStreamARN {
		t.Errorf("expected external name=%s, got=%s", testStreamARN, got)
	}
	// Also check status.atProvider.arn.
	if cr.Status.AtProvider.Arn == nil || *cr.Status.AtProvider.Arn != testStreamARN {
		t.Errorf("expected atProvider.arn=%s", testStreamARN)
	}
}

// TestObserve_StreamModeNil_AcceptsAWSDefaults verifies that when spec
// StreamModeDetails is nil, AWS defaults are accepted (no update loop).
func TestObserve_StreamModeNil_AcceptsAWSDefaults(t *testing.T) {
	cr := newTestCR("my-stream", testStreamName)
	// spec.StreamModeDetails is nil

	summary := activeStreamSummary(testStreamName, testStreamARN, 1, 24)
	// AWS returns PROVISIONED mode (its default).

	e := &ExternalClient{Client: &mockKinesisClient{
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return &awskinesis.DescribeStreamSummaryOutput{StreamDescriptionSummary: summary}, nil
		},
		listTagsForStreamFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Late-init should populate StreamModeDetails, so ResourceLateInitialized=true.
	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true (StreamModeDetails late-inited)")
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

// TestCreate_SetsExternalNameToARN verifies that Create calls CreateStream and
// sets the external name to the ARN (via DescribeStreamSummary after create).
func TestCreate_SetsExternalNameToARN(t *testing.T) {
	cr := newTestCR("my-stream", testStreamName)
	shardCount := float64(1)
	cr.Spec.ForProvider.ShardCount = &shardCount

	var createdName string
	e := &ExternalClient{Client: &mockKinesisClient{
		createStreamFn: func(_ context.Context, params *awskinesis.CreateStreamInput, _ ...func(*awskinesis.Options)) (*awskinesis.CreateStreamOutput, error) {
			createdName = aws.ToString(params.StreamName)
			return &awskinesis.CreateStreamOutput{}, nil
		},
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return &awskinesis.DescribeStreamSummaryOutput{
				StreamDescriptionSummary: &ktypes.StreamDescriptionSummary{
					StreamARN:            aws.String(testStreamARN),
					StreamName:           aws.String(testStreamName),
					StreamStatus:         ktypes.StreamStatusCreating,
					OpenShardCount:       aws.Int32(1),
					RetentionPeriodHours: aws.Int32(24),
					EncryptionType:       ktypes.EncryptionTypeNone,
				},
			}, nil
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if createdName != testStreamName {
		t.Errorf("expected CreateStream called with name=%s, got=%s", testStreamName, createdName)
	}

	// External name must be the ARN after Create.
	gotExt := native.GetExternalName(cr)
	if gotExt != testStreamARN {
		t.Errorf("expected external name=%s after Create, got=%s", testStreamARN, gotExt)
	}
}

// TestCreate_ShardCountPassedToAWS verifies that spec.ShardCount is forwarded
// to the CreateStream call.
func TestCreate_ShardCountPassedToAWS(t *testing.T) {
	cr := newTestCR("my-stream", testStreamName)
	shardCount := float64(4)
	cr.Spec.ForProvider.ShardCount = &shardCount

	var gotShards *int32
	e := &ExternalClient{Client: &mockKinesisClient{
		createStreamFn: func(_ context.Context, params *awskinesis.CreateStreamInput, _ ...func(*awskinesis.Options)) (*awskinesis.CreateStreamOutput, error) {
			gotShards = params.ShardCount
			return &awskinesis.CreateStreamOutput{}, nil
		},
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return &awskinesis.DescribeStreamSummaryOutput{
				StreamDescriptionSummary: &ktypes.StreamDescriptionSummary{
					StreamARN:            aws.String(testStreamARN),
					StreamName:           aws.String(testStreamName),
					StreamStatus:         ktypes.StreamStatusCreating,
					OpenShardCount:       aws.Int32(4),
					RetentionPeriodHours: aws.Int32(24),
					EncryptionType:       ktypes.EncryptionTypeNone,
				},
			}, nil
		},
	}}

	if _, err := e.Create(context.Background(), cr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotShards == nil || *gotShards != 4 {
		t.Errorf("expected ShardCount=4 in CreateStream call, got %v", gotShards)
	}
}

// ── Delete tests ───────────────────────────────────────────────────────────────

// TestDelete_CallsDeleteStream verifies that Delete calls the AWS DeleteStream API.
func TestDelete_CallsDeleteStream(t *testing.T) {
	cr := newTestCR("my-stream", testStreamARN)

	var deleteCalled bool
	e := &ExternalClient{Client: &mockKinesisClient{
		deleteStreamFn: func(_ context.Context, params *awskinesis.DeleteStreamInput, _ ...func(*awskinesis.Options)) (*awskinesis.DeleteStreamOutput, error) {
			deleteCalled = true
			if aws.ToString(params.StreamARN) != testStreamARN {
				t.Errorf("expected StreamARN=%s, got=%s", testStreamARN, aws.ToString(params.StreamARN))
			}
			return &awskinesis.DeleteStreamOutput{}, nil
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DeleteStream to be called")
	}
}

// TestDelete_Idempotent verifies that Delete returns nil when the stream
// is already gone (ResourceNotFoundException → nil error).
func TestDelete_Idempotent(t *testing.T) {
	cr := newTestCR("my-stream", testStreamARN)

	e := &ExternalClient{Client: &mockKinesisClient{
		deleteStreamFn: func(_ context.Context, _ *awskinesis.DeleteStreamInput, _ ...func(*awskinesis.Options)) (*awskinesis.DeleteStreamOutput, error) {
			return nil, &ktypes.ResourceNotFoundException{Message: aws.String("Stream not found")}
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for already-deleted stream, got: %v", err)
	}
}

// TestDelete_EnforceConsumerDeletion verifies that EnforceConsumerDeletion is
// set when the spec requests it.
func TestDelete_EnforceConsumerDeletion(t *testing.T) {
	cr := newTestCR("my-stream", testStreamARN)
	enforceConsumerDeletion := true
	cr.Spec.ForProvider.EnforceConsumerDeletion = &enforceConsumerDeletion

	var gotEnforce bool
	e := &ExternalClient{Client: &mockKinesisClient{
		deleteStreamFn: func(_ context.Context, params *awskinesis.DeleteStreamInput, _ ...func(*awskinesis.Options)) (*awskinesis.DeleteStreamOutput, error) {
			gotEnforce = aws.ToBool(params.EnforceConsumerDeletion)
			return &awskinesis.DeleteStreamOutput{}, nil
		},
	}}

	if _, err := e.Delete(context.Background(), cr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotEnforce {
		t.Error("expected EnforceConsumerDeletion=true in DeleteStream call")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

// TestUpdate_ShardCount verifies that UpdateShardCount is called when shard
// count differs.
func TestUpdate_ShardCount(t *testing.T) {
	cr := newTestCR("my-stream", testStreamARN)
	shardCount := float64(4)
	cr.Spec.ForProvider.ShardCount = &shardCount

	// Put ARN in status so buildDescribeInput uses ARN.
	cr.Status.AtProvider.Arn = aws.String(testStreamARN)

	// Current state: 1 shard.
	summary := activeStreamSummary(testStreamName, testStreamARN, 1, 24)

	var gotTargetShards *int32
	e := &ExternalClient{Client: &mockKinesisClient{
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return &awskinesis.DescribeStreamSummaryOutput{StreamDescriptionSummary: summary}, nil
		},
		updateShardCountFn: func(_ context.Context, params *awskinesis.UpdateShardCountInput, _ ...func(*awskinesis.Options)) (*awskinesis.UpdateShardCountOutput, error) {
			gotTargetShards = params.TargetShardCount
			return &awskinesis.UpdateShardCountOutput{}, nil
		},
		listTagsForStreamFn: noopListTags,
	}}

	if _, err := e.Update(context.Background(), cr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotTargetShards == nil || *gotTargetShards != 4 {
		t.Errorf("expected TargetShardCount=4, got %v", gotTargetShards)
	}
}

// TestUpdate_Encryption_Enable verifies that StartStreamEncryption is called
// when encryption is enabled.
func TestUpdate_Encryption_Enable(t *testing.T) {
	cr := newTestCR("my-stream", testStreamARN)
	encType := "KMS"
	keyID := "arn:aws:kms:us-east-1:123:key/abc"
	cr.Spec.ForProvider.EncryptionType = &encType
	cr.Spec.ForProvider.KMSKeyID = &keyID
	cr.Status.AtProvider.Arn = aws.String(testStreamARN)

	// Current: no encryption.
	summary := activeStreamSummary(testStreamName, testStreamARN, 1, 24)

	var startEncCalled bool
	e := &ExternalClient{Client: &mockKinesisClient{
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return &awskinesis.DescribeStreamSummaryOutput{StreamDescriptionSummary: summary}, nil
		},
		startStreamEncryptionFn: func(_ context.Context, _ *awskinesis.StartStreamEncryptionInput, _ ...func(*awskinesis.Options)) (*awskinesis.StartStreamEncryptionOutput, error) {
			startEncCalled = true
			return &awskinesis.StartStreamEncryptionOutput{}, nil
		},
		listTagsForStreamFn: noopListTags,
	}}

	if _, err := e.Update(context.Background(), cr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !startEncCalled {
		t.Error("expected StartStreamEncryption to be called")
	}
}

// TestUpdate_RetentionPeriod_Increase verifies that IncreaseStreamRetentionPeriod
// is called when retention period is increased.
func TestUpdate_RetentionPeriod_Increase(t *testing.T) {
	cr := newTestCR("my-stream", testStreamARN)
	retention := float64(48)
	cr.Spec.ForProvider.RetentionPeriod = &retention
	cr.Status.AtProvider.Arn = aws.String(testStreamARN)

	// Current: 24 hours.
	summary := activeStreamSummary(testStreamName, testStreamARN, 1, 24)

	var gotHours *int32
	e := &ExternalClient{Client: &mockKinesisClient{
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return &awskinesis.DescribeStreamSummaryOutput{StreamDescriptionSummary: summary}, nil
		},
		increaseStreamRetentionFn: func(_ context.Context, params *awskinesis.IncreaseStreamRetentionPeriodInput, _ ...func(*awskinesis.Options)) (*awskinesis.IncreaseStreamRetentionPeriodOutput, error) {
			gotHours = params.RetentionPeriodHours
			return &awskinesis.IncreaseStreamRetentionPeriodOutput{}, nil
		},
		listTagsForStreamFn: noopListTags,
	}}

	if _, err := e.Update(context.Background(), cr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotHours == nil || *gotHours != 48 {
		t.Errorf("expected RetentionPeriodHours=48, got %v", gotHours)
	}
}

// TestUpdate_Tags verifies that AddTagsToStream is called for new tags.
func TestUpdate_Tags(t *testing.T) {
	cr := newTestCR("my-stream", testStreamARN)
	tagVal := "bar"
	cr.Spec.ForProvider.Tags = map[string]*string{"foo": &tagVal}
	cr.Status.AtProvider.Arn = aws.String(testStreamARN)

	summary := activeStreamSummary(testStreamName, testStreamARN, 1, 24)

	var addedTags map[string]string
	e := &ExternalClient{Client: &mockKinesisClient{
		describeStreamSummaryFn: func(_ context.Context, _ *awskinesis.DescribeStreamSummaryInput, _ ...func(*awskinesis.Options)) (*awskinesis.DescribeStreamSummaryOutput, error) {
			return &awskinesis.DescribeStreamSummaryOutput{StreamDescriptionSummary: summary}, nil
		},
		listTagsForStreamFn: noopListTags,
		addTagsToStreamFn: func(_ context.Context, params *awskinesis.AddTagsToStreamInput, _ ...func(*awskinesis.Options)) (*awskinesis.AddTagsToStreamOutput, error) {
			addedTags = params.Tags
			return &awskinesis.AddTagsToStreamOutput{}, nil
		},
	}}

	if _, err := e.Update(context.Background(), cr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if diff := cmp.Diff(map[string]string{"foo": "bar"}, addedTags); diff != "" {
		t.Errorf("unexpected tags added: %s", diff)
	}
}

// ── helper function tests ──────────────────────────────────────────────────────

// TestStreamARN verifies that streamARN returns the correct ARN from various
// sources.
func TestStreamARN(t *testing.T) {
	tests := []struct {
		name    string
		cr      *v1beta2native.StreamRAW
		wantARN string
	}{
		{
			name: "from_status_arn",
			cr: func() *v1beta2native.StreamRAW {
				c := newTestCR("s", "my-stream")
				c.Status.AtProvider.Arn = aws.String(testStreamARN)
				return c
			}(),
			wantARN: testStreamARN,
		},
		{
			name:    "from_external_name_arn",
			cr:      newTestCR("s", testStreamARN),
			wantARN: testStreamARN,
		},
		{
			name:    "stream_name_returns_empty",
			cr:      newTestCR("s", "my-stream"),
			wantARN: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := streamARN(tt.cr)
			if got != tt.wantARN {
				t.Errorf("streamARN()=%q, want %q", got, tt.wantARN)
			}
		})
	}
}

// TestStreamName verifies that streamName extracts the stream name from ARN and
// plain name correctly.
func TestStreamName(t *testing.T) {
	tests := []struct {
		name     string
		extName  string
		wantName string
	}{
		{"from_arn", testStreamARN, testStreamName},
		{"from_plain_name", testStreamName, testStreamName},
		{"empty_uses_cr_name", "", "my-stream"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cr := newTestCR("my-stream", tt.extName)
			got := streamName(cr)
			if got != tt.wantName {
				t.Errorf("streamName()=%q, want %q", got, tt.wantName)
			}
		})
	}
}

// TestShardLevelMetricsUpToDate verifies set-based comparison.
func TestShardLevelMetricsUpToDate(t *testing.T) {
	incomingBytes := "IncomingBytes"
	outgoingBytes := "OutgoingBytes"

	tests := []struct {
		name     string
		spec     []*string
		enhanced []ktypes.EnhancedMetrics
		want     bool
	}{
		{
			name:     "both_empty",
			spec:     []*string{},
			enhanced: []ktypes.EnhancedMetrics{},
			want:     true,
		},
		{
			name:     "spec_set_aws_empty",
			spec:     []*string{&incomingBytes},
			enhanced: []ktypes.EnhancedMetrics{},
			want:     false,
		},
		{
			name: "matching_sets",
			spec: []*string{&incomingBytes, &outgoingBytes},
			enhanced: []ktypes.EnhancedMetrics{
				{ShardLevelMetrics: []ktypes.MetricsName{"IncomingBytes", "OutgoingBytes"}},
			},
			want: true,
		},
		{
			name: "different_sets",
			spec: []*string{&incomingBytes},
			enhanced: []ktypes.EnhancedMetrics{
				{ShardLevelMetrics: []ktypes.MetricsName{"OutgoingBytes"}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shardLevelMetricsUpToDate(tt.spec, tt.enhanced)
			if got != tt.want {
				t.Errorf("shardLevelMetricsUpToDate()=%v, want %v", got, tt.want)
			}
		})
	}
}
