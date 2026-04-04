// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package parametergroup

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	ectypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native"
	"github.com/upbound/provider-aws/v2/internal/native"
)

// ── Mock client ────────────────────────────────────────────────────────────────

type mockElastiCachePGClient struct {
	describePGFn    func(ctx context.Context, params *awselasticache.DescribeCacheParameterGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParameterGroupsOutput, error)
	describeParamFn func(ctx context.Context, params *awselasticache.DescribeCacheParametersInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParametersOutput, error)
	createFn        func(ctx context.Context, params *awselasticache.CreateCacheParameterGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateCacheParameterGroupOutput, error)
	modifyFn        func(ctx context.Context, params *awselasticache.ModifyCacheParameterGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheParameterGroupOutput, error)
	deleteFn        func(ctx context.Context, params *awselasticache.DeleteCacheParameterGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheParameterGroupOutput, error)
	listTagsFn      func(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	addTagsFn       func(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	removeTagsFn    func(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
	resetParamFn    func(ctx context.Context, params *awselasticache.ResetCacheParameterGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ResetCacheParameterGroupOutput, error)
}

func (m *mockElastiCachePGClient) DescribeCacheParameterGroups(ctx context.Context, params *awselasticache.DescribeCacheParameterGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParameterGroupsOutput, error) {
	return m.describePGFn(ctx, params, optFns...)
}
func (m *mockElastiCachePGClient) DescribeCacheParameters(ctx context.Context, params *awselasticache.DescribeCacheParametersInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParametersOutput, error) {
	return m.describeParamFn(ctx, params, optFns...)
}
func (m *mockElastiCachePGClient) CreateCacheParameterGroup(ctx context.Context, params *awselasticache.CreateCacheParameterGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateCacheParameterGroupOutput, error) {
	return m.createFn(ctx, params, optFns...)
}
func (m *mockElastiCachePGClient) ModifyCacheParameterGroup(ctx context.Context, params *awselasticache.ModifyCacheParameterGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheParameterGroupOutput, error) {
	return m.modifyFn(ctx, params, optFns...)
}
func (m *mockElastiCachePGClient) DeleteCacheParameterGroup(ctx context.Context, params *awselasticache.DeleteCacheParameterGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheParameterGroupOutput, error) {
	return m.deleteFn(ctx, params, optFns...)
}
func (m *mockElastiCachePGClient) ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
	return m.listTagsFn(ctx, params, optFns...)
}
func (m *mockElastiCachePGClient) AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error) {
	return m.addTagsFn(ctx, params, optFns...)
}
func (m *mockElastiCachePGClient) RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error) {
	return m.removeTagsFn(ctx, params, optFns...)
}
func (m *mockElastiCachePGClient) ResetCacheParameterGroup(ctx context.Context, params *awselasticache.ResetCacheParameterGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ResetCacheParameterGroupOutput, error) {
	if m.resetParamFn != nil {
		return m.resetParamFn(ctx, params, optFns...)
	}
	return &awselasticache.ResetCacheParameterGroupOutput{}, nil
}

// ── Helpers ────────────────────────────────────────────────────────────────────

const (
	testPGName      = "my-parameter-group"
	testPGARN       = "arn:aws:elasticache:us-east-1:609897127049:parametergroup:my-parameter-group"
	testPGFamily    = "redis7"
	testPGDesc      = "Test parameter group"
	testParamName1  = "maxmemory-policy"
	testParamValue1 = "allkeys-lru"
	testParamName2  = "timeout"
	testParamValue2 = "300"
)

// newTestCR builds a minimal cluster-scoped ParameterGroupRAW for tests.
func newTestCR(name, extName string) *clusternative.ParameterGroupRAW {
	cr := &clusternative.ParameterGroupRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: clusternative.ParameterGroupRAWSpec{
			ForProvider: clusternative.ParameterGroupRAWParameters{
				Region: aws.String("us-east-1"),
			},
		},
	}
	if extName != "" {
		native.SetExternalName(cr, extName)
	}
	return cr
}

// noopListTags returns an empty tag list.
func noopListTags(_ context.Context, _ *awselasticache.ListTagsForResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
	return &awselasticache.ListTagsForResourceOutput{TagList: []ectypes.Tag{}}, nil
}

// parameterGroupResponse returns a minimal DescribeCacheParameterGroupsOutput.
func parameterGroupResponse(name, arn, family, description string) *awselasticache.DescribeCacheParameterGroupsOutput {
	return &awselasticache.DescribeCacheParameterGroupsOutput{
		CacheParameterGroups: []ectypes.CacheParameterGroup{
			{
				ARN:                       aws.String(arn),
				CacheParameterGroupName:   aws.String(name),
				CacheParameterGroupFamily: aws.String(family),
				Description:               aws.String(description),
			},
		},
	}
}

// paramResponse returns a DescribeCacheParametersOutput with the given params.
func paramResponse(params ...ectypes.Parameter) *awselasticache.DescribeCacheParametersOutput {
	return &awselasticache.DescribeCacheParametersOutput{Parameters: params}
}

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_EmptyExternalName(t *testing.T) {
	cr := newTestCR("my-pg", "")
	e := &ExternalClient{Client: &mockElastiCachePGClient{}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty external name")
	}
}

func TestObserve_NotFound(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)

	e := &ExternalClient{Client: &mockElastiCachePGClient{
		describePGFn: func(_ context.Context, _ *awselasticache.DescribeCacheParameterGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParameterGroupsOutput, error) {
			return nil, &ectypes.CacheParameterGroupNotFoundFault{Message: aws.String("not found")}
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for NotFound, got: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when parameter group not found")
	}
}

func TestObserve_FoundWithParametersMatching(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)
	cr.Spec.ForProvider.Family = aws.String(testPGFamily)
	cr.Spec.ForProvider.Description = aws.String(testPGDesc)
	cr.Spec.ForProvider.Parameter = []clusternative.ParameterRAWParameters{
		{Name: aws.String(testParamName1), Value: aws.String(testParamValue1)},
	}

	e := &ExternalClient{Client: &mockElastiCachePGClient{
		describePGFn: func(_ context.Context, _ *awselasticache.DescribeCacheParameterGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParameterGroupsOutput, error) {
			return parameterGroupResponse(testPGName, testPGARN, testPGFamily, testPGDesc), nil
		},
		describeParamFn: func(_ context.Context, _ *awselasticache.DescribeCacheParametersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParametersOutput, error) {
			return paramResponse(ectypes.Parameter{
				ParameterName:  aws.String(testParamName1),
				ParameterValue: aws.String(testParamValue1),
			}), nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true when parameters match")
	}
	if cr.Status.AtProvider.Arn == nil || *cr.Status.AtProvider.Arn != testPGARN {
		t.Errorf("expected ARN %s in atProvider, got %v", testPGARN, cr.Status.AtProvider.Arn)
	}
}

func TestObserve_FoundWithParameterDrift_ValueChanged(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)
	cr.Spec.ForProvider.Family = aws.String(testPGFamily)
	cr.Spec.ForProvider.Description = aws.String(testPGDesc)
	cr.Spec.ForProvider.Parameter = []clusternative.ParameterRAWParameters{
		{Name: aws.String(testParamName1), Value: aws.String("volatile-lru")}, // differs from AWS
	}

	e := &ExternalClient{Client: &mockElastiCachePGClient{
		describePGFn: func(_ context.Context, _ *awselasticache.DescribeCacheParameterGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParameterGroupsOutput, error) {
			return parameterGroupResponse(testPGName, testPGARN, testPGFamily, testPGDesc), nil
		},
		describeParamFn: func(_ context.Context, _ *awselasticache.DescribeCacheParametersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParametersOutput, error) {
			return paramResponse(ectypes.Parameter{
				ParameterName:  aws.String(testParamName1),
				ParameterValue: aws.String(testParamValue1), // "allkeys-lru"
			}), nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when parameter value differs")
	}
}

func TestObserve_FoundWithParameterDrift_MissingParam(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)
	cr.Spec.ForProvider.Family = aws.String(testPGFamily)
	cr.Spec.ForProvider.Description = aws.String(testPGDesc)
	cr.Spec.ForProvider.Parameter = []clusternative.ParameterRAWParameters{
		{Name: aws.String(testParamName1), Value: aws.String(testParamValue1)},
		{Name: aws.String(testParamName2), Value: aws.String(testParamValue2)}, // not in AWS yet
	}

	e := &ExternalClient{Client: &mockElastiCachePGClient{
		describePGFn: func(_ context.Context, _ *awselasticache.DescribeCacheParameterGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParameterGroupsOutput, error) {
			return parameterGroupResponse(testPGName, testPGARN, testPGFamily, testPGDesc), nil
		},
		describeParamFn: func(_ context.Context, _ *awselasticache.DescribeCacheParametersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParametersOutput, error) {
			return paramResponse(ectypes.Parameter{
				ParameterName:  aws.String(testParamName1),
				ParameterValue: aws.String(testParamValue1),
				// testParamName2 not in user-set params
			}), nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when a desired parameter is not set in AWS")
	}
}

func TestObserve_FoundWithExtraUserParam(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)
	cr.Spec.ForProvider.Family = aws.String(testPGFamily)
	cr.Spec.ForProvider.Description = aws.String(testPGDesc)
	cr.Spec.ForProvider.Parameter = []clusternative.ParameterRAWParameters{
		// testParamName2 is in AWS user-set but NOT in desired spec
	}

	e := &ExternalClient{Client: &mockElastiCachePGClient{
		describePGFn: func(_ context.Context, _ *awselasticache.DescribeCacheParameterGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParameterGroupsOutput, error) {
			return parameterGroupResponse(testPGName, testPGARN, testPGFamily, testPGDesc), nil
		},
		describeParamFn: func(_ context.Context, _ *awselasticache.DescribeCacheParametersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParametersOutput, error) {
			return paramResponse(ectypes.Parameter{
				ParameterName:  aws.String(testParamName2),
				ParameterValue: aws.String(testParamValue2),
			}), nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=false when AWS has extra user-set params not in desired")
	}
}

func TestObserve_DescribeError(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)
	e := &ExternalClient{Client: &mockElastiCachePGClient{
		describePGFn: func(_ context.Context, _ *awselasticache.DescribeCacheParameterGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParameterGroupsOutput, error) {
			return nil, errors.New("unexpected AWS error")
		},
	}}

	_, err := e.Observe(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Observe when describe fails")
	}
}

func TestObserve_EmptyResponse(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)
	e := &ExternalClient{Client: &mockElastiCachePGClient{
		describePGFn: func(_ context.Context, _ *awselasticache.DescribeCacheParameterGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParameterGroupsOutput, error) {
			return &awselasticache.DescribeCacheParameterGroupsOutput{
				CacheParameterGroups: []ectypes.CacheParameterGroup{},
			}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty response")
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

func TestCreate_Success_SetsExternalName(t *testing.T) {
	cr := newTestCR("my-pg", "my-pg") // crossplane sets extName to CR name initially
	cr.Spec.ForProvider.Family = aws.String(testPGFamily)
	cr.Spec.ForProvider.Description = aws.String(testPGDesc)

	var createCallCount int
	e := &ExternalClient{Client: &mockElastiCachePGClient{
		createFn: func(_ context.Context, params *awselasticache.CreateCacheParameterGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateCacheParameterGroupOutput, error) {
			createCallCount++
			return &awselasticache.CreateCacheParameterGroupOutput{
				CacheParameterGroup: &ectypes.CacheParameterGroup{
					CacheParameterGroupName: params.CacheParameterGroupName,
					ARN:                     aws.String(testPGARN),
				},
			}, nil
		},
	}}

	creation, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = creation

	// Verify external name is set exactly once in the happy path.
	if createCallCount != 1 {
		t.Errorf("expected Create to be called once, got %d", createCallCount)
	}

	// External name should be set to the AWS-returned group name.
	if got := native.GetExternalName(cr); got != "my-pg" {
		t.Errorf("expected external name 'my-pg', got %q", got)
	}
}

func TestCreate_SetsExternalNameFromSpec(t *testing.T) {
	cr := newTestCR("my-pg", "initial-extname")
	cr.Spec.ForProvider.Name = aws.String(testPGName) // spec.name overrides
	cr.Spec.ForProvider.Family = aws.String(testPGFamily)
	cr.Spec.ForProvider.Description = aws.String(testPGDesc)

	var gotName string
	e := &ExternalClient{Client: &mockElastiCachePGClient{
		createFn: func(_ context.Context, params *awselasticache.CreateCacheParameterGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateCacheParameterGroupOutput, error) {
			gotName = aws.ToString(params.CacheParameterGroupName)
			return &awselasticache.CreateCacheParameterGroupOutput{
				CacheParameterGroup: &ectypes.CacheParameterGroup{
					CacheParameterGroupName: params.CacheParameterGroupName,
					ARN:                     aws.String(testPGARN),
				},
			}, nil
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotName != testPGName {
		t.Errorf("expected create to use spec.name %s, got %s", testPGName, gotName)
	}
	// External name should be updated from the response.
	if got := native.GetExternalName(cr); got != testPGName {
		t.Errorf("expected external name %s, got %q", testPGName, got)
	}
}

func TestCreate_Error(t *testing.T) {
	cr := newTestCR("my-pg", "my-pg")
	cr.Spec.ForProvider.Family = aws.String(testPGFamily)
	cr.Spec.ForProvider.Description = aws.String(testPGDesc)

	e := &ExternalClient{Client: &mockElastiCachePGClient{
		createFn: func(_ context.Context, _ *awselasticache.CreateCacheParameterGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateCacheParameterGroupOutput, error) {
			return nil, errors.New("create failed")
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Create")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_ParameterChange_CallsModify(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)
	cr.Spec.ForProvider.Parameter = []clusternative.ParameterRAWParameters{
		{Name: aws.String(testParamName1), Value: aws.String("volatile-lru")},
	}
	cr.Status.AtProvider.Arn = aws.String(testPGARN)

	var modifyCalled bool
	var gotParams []ectypes.ParameterNameValue
	e := &ExternalClient{Client: &mockElastiCachePGClient{
		modifyFn: func(_ context.Context, params *awselasticache.ModifyCacheParameterGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheParameterGroupOutput, error) {
			modifyCalled = true
			gotParams = params.ParameterNameValues
			return &awselasticache.ModifyCacheParameterGroupOutput{
				CacheParameterGroupName: params.CacheParameterGroupName,
			}, nil
		},
		listTagsFn: noopListTags,
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !modifyCalled {
		t.Error("expected ModifyCacheParameterGroup to be called")
	}
	if len(gotParams) != 1 {
		t.Errorf("expected 1 parameter name-value, got %d", len(gotParams))
	}
	if aws.ToString(gotParams[0].ParameterName) != testParamName1 {
		t.Errorf("expected param name %s, got %s", testParamName1, aws.ToString(gotParams[0].ParameterName))
	}
}

func TestUpdate_NoParameters_SkipsModify(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)
	cr.Spec.ForProvider.Parameter = []clusternative.ParameterRAWParameters{} // empty
	cr.Status.AtProvider.Arn = aws.String(testPGARN)

	var modifyCalled bool
	e := &ExternalClient{Client: &mockElastiCachePGClient{
		modifyFn: func(_ context.Context, _ *awselasticache.ModifyCacheParameterGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheParameterGroupOutput, error) {
			modifyCalled = true
			return &awselasticache.ModifyCacheParameterGroupOutput{}, nil
		},
		listTagsFn: noopListTags,
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if modifyCalled {
		t.Error("expected ModifyCacheParameterGroup NOT to be called when no parameters")
	}
}

func TestUpdate_TagSync(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)
	cr.Spec.ForProvider.Parameter = []clusternative.ParameterRAWParameters{}
	cr.Spec.ForProvider.Tags = map[string]*string{"env": aws.String("prod")}
	cr.Status.AtProvider.Arn = aws.String(testPGARN)

	var addTagsCalled bool
	e := &ExternalClient{Client: &mockElastiCachePGClient{
		modifyFn: func(_ context.Context, _ *awselasticache.ModifyCacheParameterGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheParameterGroupOutput, error) {
			return &awselasticache.ModifyCacheParameterGroupOutput{}, nil
		},
		listTagsFn: noopListTags, // returns empty tags
		addTagsFn: func(_ context.Context, _ *awselasticache.AddTagsToResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error) {
			addTagsCalled = true
			return &awselasticache.AddTagsToResourceOutput{}, nil
		},
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !addTagsCalled {
		t.Error("expected AddTagsToResource to be called when tags differ")
	}
}

func TestUpdate_Error_ModifyFails(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)
	cr.Spec.ForProvider.Parameter = []clusternative.ParameterRAWParameters{
		{Name: aws.String(testParamName1), Value: aws.String(testParamValue1)},
	}

	e := &ExternalClient{Client: &mockElastiCachePGClient{
		modifyFn: func(_ context.Context, _ *awselasticache.ModifyCacheParameterGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheParameterGroupOutput, error) {
			return nil, errors.New("modify failed")
		},
	}}

	_, err := e.Update(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Update when modify fails")
	}
}

// ── Delete tests ───────────────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)

	e := &ExternalClient{Client: &mockElastiCachePGClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteCacheParameterGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheParameterGroupOutput, error) {
			return &awselasticache.DeleteCacheParameterGroupOutput{}, nil
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDelete_NotFound_Idempotent(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)

	e := &ExternalClient{Client: &mockElastiCachePGClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteCacheParameterGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheParameterGroupOutput, error) {
			return nil, &ectypes.CacheParameterGroupNotFoundFault{Message: aws.String("not found")}
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for CacheParameterGroupNotFoundFault (idempotent delete), got: %v", err)
	}
}

func TestDelete_OtherError(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)

	e := &ExternalClient{Client: &mockElastiCachePGClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteCacheParameterGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheParameterGroupOutput, error) {
			return nil, errors.New("some AWS error")
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Delete when AWS returns unexpected error")
	}
}

// ── Pagination tests ───────────────────────────────────────────────────────────

func TestObserve_PaginatesParameters(t *testing.T) {
	cr := newTestCR("my-pg", testPGName)
	cr.Spec.ForProvider.Parameter = []clusternative.ParameterRAWParameters{
		{Name: aws.String(testParamName1), Value: aws.String(testParamValue1)},
		{Name: aws.String(testParamName2), Value: aws.String(testParamValue2)},
	}

	callCount := 0
	e := &ExternalClient{Client: &mockElastiCachePGClient{
		describePGFn: func(_ context.Context, _ *awselasticache.DescribeCacheParameterGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParameterGroupsOutput, error) {
			return parameterGroupResponse(testPGName, testPGARN, testPGFamily, testPGDesc), nil
		},
		describeParamFn: func(_ context.Context, _ *awselasticache.DescribeCacheParametersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheParametersOutput, error) {
			callCount++
			if callCount == 1 {
				// First page: return testParamName1 with a Marker.
				return &awselasticache.DescribeCacheParametersOutput{
					Parameters: []ectypes.Parameter{
						{ParameterName: aws.String(testParamName1), ParameterValue: aws.String(testParamValue1)},
					},
					Marker: aws.String("next-marker"),
				}, nil
			}
			// Second page: return testParamName2.
			return &awselasticache.DescribeCacheParametersOutput{
				Parameters: []ectypes.Parameter{
					{ParameterName: aws.String(testParamName2), ParameterValue: aws.String(testParamValue2)},
				},
			}, nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true")
	}
	if !obs.ResourceUpToDate {
		t.Errorf("expected ResourceUpToDate=true when both paginated params match spec, got false")
	}
	if callCount != 2 {
		t.Errorf("expected 2 DescribeCacheParameters calls for pagination, got %d", callCount)
	}
}
