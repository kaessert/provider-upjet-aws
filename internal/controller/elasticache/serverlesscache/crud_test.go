// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package serverlesscache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	ectypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	smithy "github.com/aws/smithy-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native"
	"github.com/upbound/provider-aws/v2/internal/native"
)

// ── Mock client ────────────────────────────────────────────────────────────────

type mockElastiCacheClient struct {
	createServerlessCacheFn    func(ctx context.Context, params *awselasticache.CreateServerlessCacheInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateServerlessCacheOutput, error)
	describeServerlessCachesFn func(ctx context.Context, params *awselasticache.DescribeServerlessCachesInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error)
	modifyServerlessCacheFn    func(ctx context.Context, params *awselasticache.ModifyServerlessCacheInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyServerlessCacheOutput, error)
	deleteServerlessCacheFn    func(ctx context.Context, params *awselasticache.DeleteServerlessCacheInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteServerlessCacheOutput, error)
	listTagsForResourceFn      func(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	addTagsToResourceFn        func(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	removeTagsFromResourceFn   func(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
}

func (m *mockElastiCacheClient) CreateServerlessCache(ctx context.Context, params *awselasticache.CreateServerlessCacheInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateServerlessCacheOutput, error) {
	return m.createServerlessCacheFn(ctx, params, optFns...)
}
func (m *mockElastiCacheClient) DescribeServerlessCaches(ctx context.Context, params *awselasticache.DescribeServerlessCachesInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
	return m.describeServerlessCachesFn(ctx, params, optFns...)
}
func (m *mockElastiCacheClient) ModifyServerlessCache(ctx context.Context, params *awselasticache.ModifyServerlessCacheInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyServerlessCacheOutput, error) {
	return m.modifyServerlessCacheFn(ctx, params, optFns...)
}
func (m *mockElastiCacheClient) DeleteServerlessCache(ctx context.Context, params *awselasticache.DeleteServerlessCacheInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteServerlessCacheOutput, error) {
	return m.deleteServerlessCacheFn(ctx, params, optFns...)
}
func (m *mockElastiCacheClient) ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
	if m.listTagsForResourceFn != nil {
		return m.listTagsForResourceFn(ctx, params, optFns...)
	}
	return &awselasticache.ListTagsForResourceOutput{TagList: []ectypes.Tag{}}, nil
}
func (m *mockElastiCacheClient) AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error) {
	if m.addTagsToResourceFn != nil {
		return m.addTagsToResourceFn(ctx, params, optFns...)
	}
	return &awselasticache.AddTagsToResourceOutput{}, nil
}
func (m *mockElastiCacheClient) RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error) {
	if m.removeTagsFromResourceFn != nil {
		return m.removeTagsFromResourceFn(ctx, params, optFns...)
	}
	return &awselasticache.RemoveTagsFromResourceOutput{}, nil
}

// ── Helpers ────────────────────────────────────────────────────────────────────

const (
	testCacheName = "my-serverless-cache"
	testRegion    = "us-east-1"
	testARN       = "arn:aws:elasticache:us-east-1:123456789012:serverlesscache:my-serverless-cache"
)

// newTestCR builds a minimal cluster-scoped ServerlessCacheRAW for tests.
func newTestCR(name, extName string) *clusternative.ServerlessCacheRAW {
	cr := &clusternative.ServerlessCacheRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: clusternative.ServerlessCacheRAWSpec{
			ForProvider: clusternative.ServerlessCacheRAWParameters{
				Region: aws.String(testRegion),
				Engine: aws.String("redis"),
			},
		},
	}
	if extName != "" {
		native.SetExternalName(cr, extName)
	}
	return cr
}

// availableServerlessCache returns a minimal available serverless cache.
func availableServerlessCache(name, arn string) ectypes.ServerlessCache {
	return ectypes.ServerlessCache{
		ServerlessCacheName: aws.String(name),
		ARN:                 aws.String(arn),
		Status:              aws.String("available"),
		Engine:              aws.String("redis"),
	}
}

// notFoundError returns an AWS-style "not found" error for serverless cache.
func notFoundError() error {
	return &smithy.GenericAPIError{
		Code:    "ServerlessCacheNotFoundFault",
		Message: "Serverless cache not found",
	}
}

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_EmptyExternalName(t *testing.T) {
	cr := newTestCR("my-cache", "") // no external name
	e := &ExternalClient{Client: &mockElastiCacheClient{}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty external name")
	}
}

func TestObserve_ResourceNotFound(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	e := &ExternalClient{Client: &mockElastiCacheClient{
		describeServerlessCachesFn: func(_ context.Context, _ *awselasticache.DescribeServerlessCachesInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
			return nil, notFoundError()
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when serverless cache not found")
	}
}

func TestObserve_Status_CREATING_ReturnsUnavailableUpToDate(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	e := &ExternalClient{Client: &mockElastiCacheClient{
		describeServerlessCachesFn: func(_ context.Context, _ *awselasticache.DescribeServerlessCachesInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
			return &awselasticache.DescribeServerlessCachesOutput{
				ServerlessCaches: []ectypes.ServerlessCache{
					{
						ServerlessCacheName: aws.String(testCacheName),
						ARN:                 aws.String(testARN),
						Status:              aws.String("creating"),
					},
				},
			}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for creating status")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for creating status (no spurious Update)")
	}
	// Endpoint is nil during creating — nil-guard test
	// No panic should occur
}

func TestObserve_Status_AVAILABLE_ClearsAsyncState(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	// Set async state as if a create was in progress
	native.SetAsyncState(cr, native.AsyncState{
		Operation: "creating",
		StartedAt: time.Now(),
		RequestID: "req-123",
	})

	e := &ExternalClient{Client: &mockElastiCacheClient{
		describeServerlessCachesFn: func(_ context.Context, _ *awselasticache.DescribeServerlessCachesInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
			sc := availableServerlessCache(testCacheName, testARN)
			return &awselasticache.DescribeServerlessCachesOutput{
				ServerlessCaches: []ectypes.ServerlessCache{sc},
			}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for available status")
	}

	// Async state should be cleared
	asyncState := native.GetAsyncState(cr)
	if asyncState != nil {
		t.Errorf("expected async state to be cleared when available, got: %+v", asyncState)
	}
}

func TestObserve_Status_MODIFYING_ReturnsUnavailableUpToDate(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	e := &ExternalClient{Client: &mockElastiCacheClient{
		describeServerlessCachesFn: func(_ context.Context, _ *awselasticache.DescribeServerlessCachesInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
			return &awselasticache.DescribeServerlessCachesOutput{
				ServerlessCaches: []ectypes.ServerlessCache{
					{
						ServerlessCacheName: aws.String(testCacheName),
						ARN:                 aws.String(testARN),
						Status:              aws.String("modifying"),
					},
				},
			}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for modifying status")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for modifying status (no spurious Update)")
	}
}

func TestObserve_Status_DELETING_ReturnsDeletingUpToDate(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	e := &ExternalClient{Client: &mockElastiCacheClient{
		describeServerlessCachesFn: func(_ context.Context, _ *awselasticache.DescribeServerlessCachesInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
			return &awselasticache.DescribeServerlessCachesOutput{
				ServerlessCaches: []ectypes.ServerlessCache{
					{
						ServerlessCacheName: aws.String(testCacheName),
						ARN:                 aws.String(testARN),
						Status:              aws.String("deleting"),
					},
				},
			}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for deleting status")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for deleting status")
	}
}

func TestObserve_Status_CREATEFAILED_AutoRecovery(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	// Set async state as if a create was in progress
	native.SetAsyncState(cr, native.AsyncState{
		Operation: "creating",
		StartedAt: time.Now(),
		RequestID: "req-123",
	})

	var deleteCalled bool
	e := &ExternalClient{Client: &mockElastiCacheClient{
		describeServerlessCachesFn: func(_ context.Context, _ *awselasticache.DescribeServerlessCachesInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
			return &awselasticache.DescribeServerlessCachesOutput{
				ServerlessCaches: []ectypes.ServerlessCache{
					{
						ServerlessCacheName: aws.String(testCacheName),
						ARN:                 aws.String(testARN),
						Status:              aws.String("create-failed"),
					},
				},
			}, nil
		},
		deleteServerlessCacheFn: func(_ context.Context, params *awselasticache.DeleteServerlessCacheInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteServerlessCacheOutput, error) {
			deleteCalled = true
			return &awselasticache.DeleteServerlessCacheOutput{}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Errorf("expected no error for create-failed (auto-recovery), got: %v", err)
	}
	// Auto-recovery: delete was initiated, treat as still-existing until deleted
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true during create-failed auto-recovery")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true during create-failed auto-recovery (waiting for delete)")
	}
	if !deleteCalled {
		t.Error("expected DeleteServerlessCache to be called for auto-recovery")
	}

	// Async state should be cleared
	asyncState := native.GetAsyncState(cr)
	if asyncState != nil {
		t.Errorf("expected async state to be cleared after create-failed, got: %+v", asyncState)
	}
}

func TestObserve_NotFound_DeletingAsyncState_ReturnsResourceNotExists(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	// Simulate a delete that has completed (async state says "deleting", but resource is gone)
	native.SetAsyncState(cr, native.AsyncState{
		Operation: "deleting",
		StartedAt: time.Now(),
		RequestID: "req-456",
	})

	e := &ExternalClient{Client: &mockElastiCacheClient{
		describeServerlessCachesFn: func(_ context.Context, _ *awselasticache.DescribeServerlessCachesInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
			return nil, notFoundError()
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when NotFound with deleting asyncState")
	}

	// Async state should be cleared
	asyncState := native.GetAsyncState(cr)
	if asyncState != nil {
		t.Errorf("expected async state to be cleared after delete completed, got: %+v", asyncState)
	}
}

func TestObserve_ConnectionDetails_WithEndpoints(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)

	e := &ExternalClient{Client: &mockElastiCacheClient{
		describeServerlessCachesFn: func(_ context.Context, _ *awselasticache.DescribeServerlessCachesInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
			return &awselasticache.DescribeServerlessCachesOutput{
				ServerlessCaches: []ectypes.ServerlessCache{
					{
						ServerlessCacheName: aws.String(testCacheName),
						ARN:                 aws.String(testARN),
						Status:              aws.String("available"),
						Engine:              aws.String("redis"),
						Endpoint: &ectypes.Endpoint{
							Address: aws.String("my-cache.serverless.use1.cache.amazonaws.com"),
							Port:    aws.Int32(6379),
						},
						ReaderEndpoint: &ectypes.Endpoint{
							Address: aws.String("my-cache.ro.serverless.use1.cache.amazonaws.com"),
							Port:    aws.Int32(6379),
						},
					},
				},
			}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for available status")
	}

	// Verify connection details key names (indexed format for backward compat)
	cd := obs.ConnectionDetails
	if cd == nil {
		t.Fatal("expected non-nil connection details")
	}

	wantKeys := []string{
		"endpoint_0_address",
		"endpoint_0_port",
		"reader_endpoint_0_address",
		"reader_endpoint_0_port",
	}
	for _, k := range wantKeys {
		if _, ok := cd[k]; !ok {
			t.Errorf("expected connection detail key %q, not found; got keys: %v", k, keysOf(cd))
		}
	}

	// Verify exact values
	if got := string(cd["endpoint_0_address"]); got != "my-cache.serverless.use1.cache.amazonaws.com" {
		t.Errorf("endpoint_0_address: want %q, got %q", "my-cache.serverless.use1.cache.amazonaws.com", got)
	}
	if got := string(cd["endpoint_0_port"]); got != "6379" {
		t.Errorf("endpoint_0_port: want %q, got %q", "6379", got)
	}
	if got := string(cd["reader_endpoint_0_address"]); got != "my-cache.ro.serverless.use1.cache.amazonaws.com" {
		t.Errorf("reader_endpoint_0_address: want %q, got %q", "my-cache.ro.serverless.use1.cache.amazonaws.com", got)
	}
	if got := string(cd["reader_endpoint_0_port"]); got != "6379" {
		t.Errorf("reader_endpoint_0_port: want %q, got %q", "6379", got)
	}
}

func TestObserve_ConnectionDetails_NilEndpoint_NoNilPanic(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	// Set async state to "creating" to simulate creating state; endpoint should be nil.
	native.SetAsyncState(cr, native.AsyncState{
		Operation: "creating",
		StartedAt: time.Now(),
		RequestID: "req-789",
	})

	e := &ExternalClient{Client: &mockElastiCacheClient{
		describeServerlessCachesFn: func(_ context.Context, _ *awselasticache.DescribeServerlessCachesInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
			return &awselasticache.DescribeServerlessCachesOutput{
				ServerlessCaches: []ectypes.ServerlessCache{
					{
						ServerlessCacheName: aws.String(testCacheName),
						ARN:                 aws.String(testARN),
						Status:              aws.String("creating"),
						Endpoint:            nil, // nil endpoint during creating
						ReaderEndpoint:      nil, // nil reader endpoint during creating
					},
				},
			}, nil
		},
	}}

	// This must not panic
	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for creating status")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for creating status")
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

func TestCreate_Success_SetsAsyncState(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)

	var createCalled bool
	e := &ExternalClient{Client: &mockElastiCacheClient{
		createServerlessCacheFn: func(_ context.Context, params *awselasticache.CreateServerlessCacheInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateServerlessCacheOutput, error) {
			createCalled = true
			if aws.ToString(params.ServerlessCacheName) != testCacheName {
				return nil, fmt.Errorf("unexpected cache name: %s", aws.ToString(params.ServerlessCacheName))
			}
			return &awselasticache.CreateServerlessCacheOutput{
				ServerlessCache: &ectypes.ServerlessCache{
					ServerlessCacheName: aws.String(testCacheName),
					ARN:                 aws.String(testARN),
					Status:              aws.String("creating"),
				},
			}, nil
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !createCalled {
		t.Error("expected CreateServerlessCache to be called")
	}

	// Async state should be set with Operation="creating"
	asyncState := native.GetAsyncState(cr)
	if asyncState == nil {
		t.Fatal("expected async state to be set after Create")
	}
	if asyncState.Operation != "creating" {
		t.Errorf("expected async state Operation=creating, got %q", asyncState.Operation)
	}
}

func TestCreate_AlreadyExists_SetsAsyncStateNoError(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)

	e := &ExternalClient{Client: &mockElastiCacheClient{
		createServerlessCacheFn: func(_ context.Context, _ *awselasticache.CreateServerlessCacheInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateServerlessCacheOutput, error) {
			return nil, &smithy.GenericAPIError{
				Code:    "ServerlessCacheAlreadyExistsFault",
				Message: "Serverless Cache already exists",
			}
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Errorf("expected no error for AlreadyExistsFault (name reservation period), got: %v", err)
	}

	// Async state should be set to "creating" so Observe backs off
	asyncState := native.GetAsyncState(cr)
	if asyncState == nil {
		t.Fatal("expected async state to be set after AlreadyExistsFault")
	}
	if asyncState.Operation != "creating" {
		t.Errorf("expected async state Operation=creating, got %q", asyncState.Operation)
	}
}

func TestObserve_NotFound_WithRecentCreatingState_BacksOff(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	// Simulate the case where Create was called and got AlreadyExistsFault,
	// setting a recent "creating" async state.
	native.SetAsyncState(cr, native.AsyncState{
		Operation: "creating",
		StartedAt: time.Now(), // very recent
		RequestID: "",
	})

	e := &ExternalClient{Client: &mockElastiCacheClient{
		describeServerlessCachesFn: func(_ context.Context, _ *awselasticache.DescribeServerlessCachesInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
			return nil, notFoundError()
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should back off — resource "exists" from controller's perspective (waiting for AWS)
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true during name-reservation backoff")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true during name-reservation backoff")
	}
}

func TestObserve_NotFound_WithOldCreatingState_ReturnsNotExists(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	// Simulate the case where a "creating" async state is old (> 5 minutes)
	native.SetAsyncState(cr, native.AsyncState{
		Operation: "creating",
		StartedAt: time.Now().Add(-10 * time.Minute), // 10 minutes ago (expired)
		RequestID: "",
	})

	e := &ExternalClient{Client: &mockElastiCacheClient{
		describeServerlessCachesFn: func(_ context.Context, _ *awselasticache.DescribeServerlessCachesInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
			return nil, notFoundError()
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Old state should not prevent recreation
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when creating state is expired and resource not found")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_Success_SetsAsyncState(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	newDesc := "updated description"
	cr.Spec.ForProvider.Description = &newDesc

	var modifyCalled bool
	e := &ExternalClient{Client: &mockElastiCacheClient{
		modifyServerlessCacheFn: func(_ context.Context, params *awselasticache.ModifyServerlessCacheInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyServerlessCacheOutput, error) {
			modifyCalled = true
			if aws.ToString(params.ServerlessCacheName) != testCacheName {
				return nil, fmt.Errorf("unexpected cache name: %s", aws.ToString(params.ServerlessCacheName))
			}
			return &awselasticache.ModifyServerlessCacheOutput{
				ServerlessCache: &ectypes.ServerlessCache{
					ServerlessCacheName: aws.String(testCacheName),
					ARN:                 aws.String(testARN),
					Status:              aws.String("modifying"),
				},
			}, nil
		},
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !modifyCalled {
		t.Error("expected ModifyServerlessCache to be called")
	}

	// Async state should be set with Operation="updating"
	asyncState := native.GetAsyncState(cr)
	if asyncState == nil {
		t.Fatal("expected async state to be set after Update")
	}
	if asyncState.Operation != "updating" {
		t.Errorf("expected async state Operation=updating, got %q", asyncState.Operation)
	}
}

// ── Delete tests ───────────────────────────────────────────────────────────────

func TestDelete_Success_SetsAsyncState(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)

	var deleteCalled bool
	e := &ExternalClient{Client: &mockElastiCacheClient{
		deleteServerlessCacheFn: func(_ context.Context, params *awselasticache.DeleteServerlessCacheInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteServerlessCacheOutput, error) {
			deleteCalled = true
			if aws.ToString(params.ServerlessCacheName) != testCacheName {
				return nil, fmt.Errorf("unexpected cache name: %s", aws.ToString(params.ServerlessCacheName))
			}
			return &awselasticache.DeleteServerlessCacheOutput{
				ServerlessCache: &ectypes.ServerlessCache{
					ServerlessCacheName: aws.String(testCacheName),
					ARN:                 aws.String(testARN),
					Status:              aws.String("deleting"),
				},
			}, nil
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleteCalled {
		t.Error("expected DeleteServerlessCache to be called")
	}

	// Async state should be set with Operation="deleting"
	asyncState := native.GetAsyncState(cr)
	if asyncState == nil {
		t.Fatal("expected async state to be set after Delete")
	}
	if asyncState.Operation != "deleting" {
		t.Errorf("expected async state Operation=deleting, got %q", asyncState.Operation)
	}
}

func TestDelete_NotFound_Idempotent(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)

	e := &ExternalClient{Client: &mockElastiCacheClient{
		deleteServerlessCacheFn: func(_ context.Context, _ *awselasticache.DeleteServerlessCacheInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteServerlessCacheOutput, error) {
			return nil, notFoundError()
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected no error for not-found delete (idempotent), got: %v", err)
	}
}

// ── isUpToDate tests ───────────────────────────────────────────────────────────

func TestIsUpToDate_UpToDate(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	desc := "my description"
	cr.Spec.ForProvider.Description = &desc

	sc := availableServerlessCache(testCacheName, testARN)
	sc.Description = &desc

	if !isUpToDate(cr, sc, nil) {
		t.Error("expected isUpToDate=true when spec matches observed state")
	}
}

func TestIsUpToDate_DescriptionChanged(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	newDesc := "new description"
	cr.Spec.ForProvider.Description = &newDesc

	sc := availableServerlessCache(testCacheName, testARN)
	oldDesc := "old description"
	sc.Description = &oldDesc

	if isUpToDate(cr, sc, nil) {
		t.Error("expected isUpToDate=false when description differs")
	}
}

func TestIsUpToDate_NilSpecFieldsAreIgnored(t *testing.T) {
	// When spec fields are nil (not set), they should be ignored in comparison
	// even if AWS has defaults set (e.g., description=" ", dailySnapshotTime="11:30").
	cr := newTestCR("my-cache", testCacheName)
	// No description, dailySnapshotTime, etc. set in spec

	sc := availableServerlessCache(testCacheName, testARN)
	space := " "
	sc.Description = &space // AWS default
	snapshotTime := "11:30"
	sc.DailySnapshotTime = &snapshotTime // AWS default
	majorVersion := "7"
	sc.MajorEngineVersion = &majorVersion
	defaultSg := "sg-default"
	sc.SecurityGroupIds = []string{defaultSg} // AWS-assigned default SG
	retention := int32(0)
	sc.SnapshotRetentionLimit = &retention

	// isUpToDate must return true (nil spec = don't care about these fields)
	if !isUpToDate(cr, sc, nil) {
		t.Error("expected isUpToDate=true when spec fields are nil (AWS defaults should be ignored)")
	}
}

// ── helpers ────────────────────────────────────────────────────────────────────

// keysOf returns the keys of a ConnectionDetails map for error messages.
func keysOf(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ── Late-initialization tests ──────────────────────────────────────────────────

// TestObserve_LateInit_NilMajorEngineVersion verifies that when spec.MajorEngineVersion
// is nil and AWS returns a value, Observe returns ResourceLateInitialized=true and
// populates the spec field.
func TestObserve_LateInit_NilMajorEngineVersion(t *testing.T) {
	cr := newTestCR("my-cache", testCacheName)
	cr.Spec.ForProvider.MajorEngineVersion = nil // intentionally nil

	sc := availableServerlessCache(testCacheName, testARN)
	sc.MajorEngineVersion = aws.String("7")

	e := &ExternalClient{Client: &mockElastiCacheClient{
		describeServerlessCachesFn: func(_ context.Context, _ *awselasticache.DescribeServerlessCachesInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeServerlessCachesOutput, error) {
			return &awselasticache.DescribeServerlessCachesOutput{
				ServerlessCaches: []ectypes.ServerlessCache{sc},
			}, nil
		},
		listTagsForResourceFn: func(_ context.Context, _ *awselasticache.ListTagsForResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
			return &awselasticache.ListTagsForResourceOutput{TagList: []ectypes.Tag{}}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true when MajorEngineVersion is nil and AWS returns a value")
	}
	if cr.Spec.ForProvider.MajorEngineVersion == nil {
		t.Fatal("expected spec.MajorEngineVersion to be populated after late initialization")
	}
	if got, want := *cr.Spec.ForProvider.MajorEngineVersion, "7"; got != want {
		t.Errorf("spec.MajorEngineVersion: got %q, want %q", got, want)
	}
}
