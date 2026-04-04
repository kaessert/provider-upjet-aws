// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package subnetgroup

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

type mockElastiCacheSGClient struct {
	describeFn   func(ctx context.Context, params *awselasticache.DescribeCacheSubnetGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheSubnetGroupsOutput, error)
	createFn     func(ctx context.Context, params *awselasticache.CreateCacheSubnetGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateCacheSubnetGroupOutput, error)
	modifyFn     func(ctx context.Context, params *awselasticache.ModifyCacheSubnetGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheSubnetGroupOutput, error)
	deleteFn     func(ctx context.Context, params *awselasticache.DeleteCacheSubnetGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheSubnetGroupOutput, error)
	listTagsFn   func(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	addTagsFn    func(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	removeTagsFn func(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
}

func (m *mockElastiCacheSGClient) DescribeCacheSubnetGroups(ctx context.Context, params *awselasticache.DescribeCacheSubnetGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheSubnetGroupsOutput, error) {
	return m.describeFn(ctx, params, optFns...)
}
func (m *mockElastiCacheSGClient) CreateCacheSubnetGroup(ctx context.Context, params *awselasticache.CreateCacheSubnetGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateCacheSubnetGroupOutput, error) {
	return m.createFn(ctx, params, optFns...)
}
func (m *mockElastiCacheSGClient) ModifyCacheSubnetGroup(ctx context.Context, params *awselasticache.ModifyCacheSubnetGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheSubnetGroupOutput, error) {
	return m.modifyFn(ctx, params, optFns...)
}
func (m *mockElastiCacheSGClient) DeleteCacheSubnetGroup(ctx context.Context, params *awselasticache.DeleteCacheSubnetGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheSubnetGroupOutput, error) {
	return m.deleteFn(ctx, params, optFns...)
}
func (m *mockElastiCacheSGClient) ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
	return m.listTagsFn(ctx, params, optFns...)
}
func (m *mockElastiCacheSGClient) AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error) {
	return m.addTagsFn(ctx, params, optFns...)
}
func (m *mockElastiCacheSGClient) RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error) {
	return m.removeTagsFn(ctx, params, optFns...)
}

// ── Helpers ────────────────────────────────────────────────────────────────────

const (
	testSGName      = "my-subnet-group"
	testSGARN       = "arn:aws:elasticache:us-east-1:123456789012:subnetgroup:my-subnet-group"
	testDescription = "Test subnet group"
	testSubnetID1   = "subnet-aaaaaa01"
	testSubnetID2   = "subnet-bbbbbb02"
)

// newTestCR builds a minimal cluster-scoped SubnetGroupRAW for tests.
func newTestCR(name, extName string) *clusternative.SubnetGroupRAW {
	cr := &clusternative.SubnetGroupRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: clusternative.SubnetGroupRAWSpec{
			ForProvider: clusternative.SubnetGroupRAWParameters{
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
func noopListTags(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
	return &awselasticache.ListTagsForResourceOutput{TagList: []ectypes.Tag{}}, nil
}

// subnetGroupResponse returns a minimal DescribeCacheSubnetGroupsOutput.
func subnetGroupResponse(name, arn, description string, subnetIDs []string) *awselasticache.DescribeCacheSubnetGroupsOutput {
	subnets := make([]ectypes.Subnet, 0, len(subnetIDs))
	for _, id := range subnetIDs {
		id := id
		subnets = append(subnets, ectypes.Subnet{SubnetIdentifier: &id})
	}
	return &awselasticache.DescribeCacheSubnetGroupsOutput{
		CacheSubnetGroups: []ectypes.CacheSubnetGroup{
			{
				ARN:                         aws.String(arn),
				CacheSubnetGroupName:        aws.String(name),
				CacheSubnetGroupDescription: aws.String(description),
				Subnets:                     subnets,
				VpcId:                       aws.String("vpc-12345678"),
			},
		},
	}
}

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_EmptyExternalName(t *testing.T) {
	cr := newTestCR("my-sg", "")
	e := &ExternalClient{Client: &mockElastiCacheSGClient{}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty external name")
	}
}

func TestObserve_ResourceNotFound(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)

	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheSubnetGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheSubnetGroupsOutput, error) {
			return nil, &ectypes.CacheSubnetGroupNotFoundFault{Message: aws.String("not found")}
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for NotFound, got: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when subnet group not found")
	}
}

func TestObserve_ResourceExists_UpToDate(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)
	cr.Spec.ForProvider.Description = aws.String(testDescription)
	cr.Spec.ForProvider.SubnetIds = []*string{aws.String(testSubnetID1), aws.String(testSubnetID2)}

	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheSubnetGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheSubnetGroupsOutput, error) {
			return subnetGroupResponse(testSGName, testSGARN, testDescription, []string{testSubnetID1, testSubnetID2}), nil
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
		t.Error("expected ResourceUpToDate=true when spec matches AWS")
	}
	if cr.Status.AtProvider.Arn == nil || *cr.Status.AtProvider.Arn != testSGARN {
		t.Errorf("expected ARN %s in atProvider, got %v", testSGARN, cr.Status.AtProvider.Arn)
	}
}

func TestObserve_ResourceExists_NotUpToDate_Description(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)
	cr.Spec.ForProvider.Description = aws.String("new description") // differs
	cr.Spec.ForProvider.SubnetIds = []*string{aws.String(testSubnetID1)}

	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheSubnetGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheSubnetGroupsOutput, error) {
			return subnetGroupResponse(testSGName, testSGARN, testDescription, []string{testSubnetID1}), nil
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
		t.Error("expected ResourceUpToDate=false when description differs")
	}
}

func TestObserve_ResourceExists_NotUpToDate_SubnetIds(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)
	cr.Spec.ForProvider.Description = aws.String(testDescription)
	cr.Spec.ForProvider.SubnetIds = []*string{aws.String(testSubnetID1), aws.String(testSubnetID2)} // wants 2

	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheSubnetGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheSubnetGroupsOutput, error) {
			return subnetGroupResponse(testSGName, testSGARN, testDescription, []string{testSubnetID1}), nil // only 1
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
		t.Error("expected ResourceUpToDate=false when subnet IDs differ")
	}
}

func TestObserve_DescribeError(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)
	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheSubnetGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheSubnetGroupsOutput, error) {
			return nil, errors.New("unexpected AWS error")
		},
	}}

	_, err := e.Observe(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Observe when describe fails")
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

func TestCreate_Success(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)
	cr.Spec.ForProvider.Description = aws.String(testDescription)
	cr.Spec.ForProvider.SubnetIds = []*string{aws.String(testSubnetID1)}

	var gotName string
	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		createFn: func(_ context.Context, params *awselasticache.CreateCacheSubnetGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateCacheSubnetGroupOutput, error) {
			gotName = aws.ToString(params.CacheSubnetGroupName)
			return &awselasticache.CreateCacheSubnetGroupOutput{
				CacheSubnetGroup: &ectypes.CacheSubnetGroup{
					CacheSubnetGroupName:        params.CacheSubnetGroupName,
					CacheSubnetGroupDescription: params.CacheSubnetGroupDescription,
					ARN:                         aws.String(testSGARN),
				},
			}, nil
		},
	}}

	creation, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotName != testSGName {
		t.Errorf("expected group name %s, got %s", testSGName, gotName)
	}
	_ = creation // no connection details
}

func TestCreate_Error(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)
	cr.Spec.ForProvider.Description = aws.String(testDescription)
	cr.Spec.ForProvider.SubnetIds = []*string{aws.String(testSubnetID1)}

	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		createFn: func(_ context.Context, _ *awselasticache.CreateCacheSubnetGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateCacheSubnetGroupOutput, error) {
			return nil, errors.New("create failed")
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Create")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_DescriptionChange(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)
	cr.Spec.ForProvider.Description = aws.String("updated description")
	cr.Spec.ForProvider.SubnetIds = []*string{aws.String(testSubnetID1)}
	cr.Status.AtProvider.Arn = aws.String(testSGARN)

	var modifyCalled bool
	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		modifyFn: func(_ context.Context, params *awselasticache.ModifyCacheSubnetGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheSubnetGroupOutput, error) {
			modifyCalled = true
			return &awselasticache.ModifyCacheSubnetGroupOutput{
				CacheSubnetGroup: &ectypes.CacheSubnetGroup{
					CacheSubnetGroupName: aws.String(testSGName),
				},
			}, nil
		},
		listTagsFn: noopListTags,
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !modifyCalled {
		t.Error("expected ModifyCacheSubnetGroup to be called")
	}
}

func TestUpdate_TagSync(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)
	cr.Spec.ForProvider.Description = aws.String(testDescription)
	cr.Spec.ForProvider.SubnetIds = []*string{aws.String(testSubnetID1)}
	cr.Spec.ForProvider.Tags = map[string]*string{"env": aws.String("prod")}
	cr.Status.AtProvider.Arn = aws.String(testSGARN)

	var addTagsCalled bool
	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		modifyFn: func(_ context.Context, _ *awselasticache.ModifyCacheSubnetGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyCacheSubnetGroupOutput, error) {
			return &awselasticache.ModifyCacheSubnetGroupOutput{}, nil
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

// ── Delete tests ───────────────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)

	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		deleteFn: func(_ context.Context, params *awselasticache.DeleteCacheSubnetGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheSubnetGroupOutput, error) {
			return &awselasticache.DeleteCacheSubnetGroupOutput{}, nil
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDelete_NotFound_Idempotent(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)

	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteCacheSubnetGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheSubnetGroupOutput, error) {
			return nil, &ectypes.CacheSubnetGroupNotFoundFault{Message: aws.String("not found")}
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for CacheSubnetGroupNotFoundFault (idempotent delete), got: %v", err)
	}
}

func TestDelete_OtherError(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)

	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteCacheSubnetGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteCacheSubnetGroupOutput, error) {
			return nil, errors.New("some AWS error")
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Delete when AWS returns unexpected error")
	}
}

// ── ExternalName tests ─────────────────────────────────────────────────────────

func TestObserve_ExternalNamePassedToDescribe(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)
	var gotName string

	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		describeFn: func(_ context.Context, params *awselasticache.DescribeCacheSubnetGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheSubnetGroupsOutput, error) {
			gotName = aws.ToString(params.CacheSubnetGroupName)
			return nil, &ectypes.CacheSubnetGroupNotFoundFault{Message: aws.String("not found")}
		},
	}}

	_, _ = e.Observe(context.Background(), cr)
	if gotName != testSGName {
		t.Errorf("expected describe to use external name %s, got %s", testSGName, gotName)
	}
}

func TestCreate_UsesExternalNameAsGroupName(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)
	cr.Spec.ForProvider.Description = aws.String(testDescription)
	cr.Spec.ForProvider.SubnetIds = []*string{aws.String(testSubnetID1)}

	var gotName string
	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		createFn: func(_ context.Context, params *awselasticache.CreateCacheSubnetGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateCacheSubnetGroupOutput, error) {
			gotName = aws.ToString(params.CacheSubnetGroupName)
			return &awselasticache.CreateCacheSubnetGroupOutput{
				CacheSubnetGroup: &ectypes.CacheSubnetGroup{
					CacheSubnetGroupName: params.CacheSubnetGroupName,
					ARN:                  aws.String(testSGARN),
				},
			}, nil
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotName != testSGName {
		t.Errorf("expected create to use external name %s as group name, got %s", testSGName, gotName)
	}
}

// ── Late-initialization tests ──────────────────────────────────────────────────

// TestObserve_LateInit_NilDescription verifies that when spec.Description is nil
// and AWS returns a non-empty, non-whitespace description, Observe returns
// ResourceLateInitialized=true and populates the spec field.
//
// IMPORTANT: Whitespace-only descriptions (like " " which AWS uses for empty
// descriptions) are intentionally NOT late-initialized to avoid infinite
// reconciliation loops for namespaced types. isUpToDate uses TrimSpace to
// treat nil/"" and " " as equivalent.
func TestObserve_LateInit_NilDescription(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)
	cr.Spec.ForProvider.Description = nil // intentionally nil
	cr.Spec.ForProvider.SubnetIds = []*string{aws.String(testSubnetID1)}

	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheSubnetGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheSubnetGroupsOutput, error) {
			return subnetGroupResponse(testSGName, testSGARN, testDescription, []string{testSubnetID1}), nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must signal late initialization so the controller saves the spec
	// (testDescription = "Test subnet group" is non-empty and non-whitespace).
	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true when Description is nil and AWS returns a non-empty value")
	}

	// spec.Description must be populated from the AWS response.
	if cr.Spec.ForProvider.Description == nil {
		t.Fatal("expected spec.Description to be populated after late initialization")
	}
	if got, want := *cr.Spec.ForProvider.Description, testDescription; got != want {
		t.Errorf("spec.Description: got %q, want %q", got, want)
	}
}

// TestObserve_LateInit_WhitespaceDescription verifies that when spec.Description
// is nil and AWS returns a whitespace-only description (i.e., " "), Observe does
// NOT late-initialize the Description field and considers the resource up to date.
// This prevents an infinite reconciliation loop for namespaced types.
func TestObserve_LateInit_WhitespaceDescription(t *testing.T) {
	cr := newTestCR("my-sg", testSGName)
	cr.Spec.ForProvider.Description = nil // intentionally nil
	cr.Spec.ForProvider.SubnetIds = []*string{aws.String(testSubnetID1)}

	e := &ExternalClient{Client: &mockElastiCacheSGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeCacheSubnetGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeCacheSubnetGroupsOutput, error) {
			// AWS stores empty-string description as " " (single space)
			return subnetGroupResponse(testSGName, testSGARN, " ", []string{testSubnetID1}), nil
		},
		listTagsFn: noopListTags,
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must NOT late-initialize when AWS description is whitespace-only.
	if obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=false: whitespace description should be treated as empty, no LateInit needed")
	}

	// spec.Description must remain nil.
	if cr.Spec.ForProvider.Description != nil {
		t.Errorf("expected spec.Description to remain nil, got %q", *cr.Spec.ForProvider.Description)
	}

	// Resource MUST be considered up-to-date: nil spec description == " " AWS description.
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true: nil description and \" \" are equivalent after TrimSpace")
	}
}
