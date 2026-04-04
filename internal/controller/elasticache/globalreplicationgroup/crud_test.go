// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package globalreplicationgroup

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

type mockGRGClient struct {
	describeFn func(ctx context.Context, params *awselasticache.DescribeGlobalReplicationGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeGlobalReplicationGroupsOutput, error)
	createFn   func(ctx context.Context, params *awselasticache.CreateGlobalReplicationGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateGlobalReplicationGroupOutput, error)
	modifyFn   func(ctx context.Context, params *awselasticache.ModifyGlobalReplicationGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyGlobalReplicationGroupOutput, error)
	deleteFn   func(ctx context.Context, params *awselasticache.DeleteGlobalReplicationGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteGlobalReplicationGroupOutput, error)
}

func (m *mockGRGClient) DescribeGlobalReplicationGroups(ctx context.Context, params *awselasticache.DescribeGlobalReplicationGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeGlobalReplicationGroupsOutput, error) {
	return m.describeFn(ctx, params, optFns...)
}
func (m *mockGRGClient) CreateGlobalReplicationGroup(ctx context.Context, params *awselasticache.CreateGlobalReplicationGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateGlobalReplicationGroupOutput, error) {
	return m.createFn(ctx, params, optFns...)
}
func (m *mockGRGClient) ModifyGlobalReplicationGroup(ctx context.Context, params *awselasticache.ModifyGlobalReplicationGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyGlobalReplicationGroupOutput, error) {
	return m.modifyFn(ctx, params, optFns...)
}
func (m *mockGRGClient) DeleteGlobalReplicationGroup(ctx context.Context, params *awselasticache.DeleteGlobalReplicationGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteGlobalReplicationGroupOutput, error) {
	return m.deleteFn(ctx, params, optFns...)
}

// ── Constants ─────────────────────────────────────────────────────────────────

const (
	testGRGID         = "ldgnf-my-global-rg"
	testGRGSuffix     = "my-global-rg"
	testGRGARN        = "arn:aws:elasticache:us-east-1:609897127049:globalreplicationgroup:ldgnf-my-global-rg"
	testPrimaryRGID   = "primary-rg-id"
	testDescription   = "my global rg description"
	testCacheNodeType = "cache.r6g.large"
	testEngine        = "redis"
	testEngineVersion = "7.2"
)

// ── Helpers ────────────────────────────────────────────────────────────────────

// newTestCR builds a minimal cluster-scoped GlobalReplicationGroupRAW for tests.
func newTestCR(name, extName string) *clusternative.GlobalReplicationGroupRAW {
	cr := &clusternative.GlobalReplicationGroupRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: clusternative.GlobalReplicationGroupRAWSpec{
			ForProvider: clusternative.GlobalReplicationGroupRAWParameters{
				Region:                         aws.String("us-east-1"),
				GlobalReplicationGroupIDSuffix: aws.String(testGRGSuffix),
				PrimaryReplicationGroupID:      aws.String(testPrimaryRGID),
			},
		},
	}
	if extName != "" {
		native.SetExternalName(cr, extName)
	}
	return cr
}

// grgResponse returns a DescribeGlobalReplicationGroupsOutput with a single group.
func grgResponse(id, arn, status, engine, engineVersion, cacheNodeType string, description *string) *awselasticache.DescribeGlobalReplicationGroupsOutput {
	return &awselasticache.DescribeGlobalReplicationGroupsOutput{
		GlobalReplicationGroups: []ectypes.GlobalReplicationGroup{
			{
				GlobalReplicationGroupId:          aws.String(id),
				ARN:                               aws.String(arn),
				Status:                            aws.String(status),
				Engine:                            aws.String(engine),
				EngineVersion:                     aws.String(engineVersion),
				CacheNodeType:                     aws.String(cacheNodeType),
				GlobalReplicationGroupDescription: description,
			},
		},
	}
}

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_EmptyExternalName(t *testing.T) {
	cr := newTestCR("my-grg", "")
	e := &ExternalClient{Client: &mockGRGClient{}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty external name")
	}
}

func TestObserve_GlobalReplicationGroupNotFound(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)

	e := &ExternalClient{Client: &mockGRGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeGlobalReplicationGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeGlobalReplicationGroupsOutput, error) {
			return nil, &ectypes.GlobalReplicationGroupNotFoundFault{Message: aws.String("not found")}
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for NotFound, got: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when global replication group not found")
	}
}

func TestObserve_EmptyList(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)

	e := &ExternalClient{Client: &mockGRGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeGlobalReplicationGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeGlobalReplicationGroupsOutput, error) {
			return &awselasticache.DescribeGlobalReplicationGroupsOutput{
				GlobalReplicationGroups: []ectypes.GlobalReplicationGroup{},
			}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty GlobalReplicationGroups list")
	}
}

// TestObserve_Available_UpToDate verifies that an available and in-sync resource
// returns ResourceExists=true, ResourceUpToDate=true.
func TestObserve_Available_UpToDate(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)
	cr.Spec.ForProvider.GlobalReplicationGroupDescription = aws.String(testDescription)
	cr.Spec.ForProvider.CacheNodeType = aws.String(testCacheNodeType)
	cr.Spec.ForProvider.EngineVersion = aws.String(testEngineVersion)
	cr.Spec.ForProvider.Engine = aws.String(testEngine)

	e := &ExternalClient{Client: &mockGRGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeGlobalReplicationGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeGlobalReplicationGroupsOutput, error) {
			return grgResponse(testGRGID, testGRGARN, "available", testEngine, testEngineVersion, testCacheNodeType, aws.String(testDescription)), nil
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
		t.Error("expected ResourceUpToDate=true when spec matches AWS state")
	}
	if cr.Status.AtProvider.GlobalReplicationGroupID == nil || *cr.Status.AtProvider.GlobalReplicationGroupID != testGRGID {
		t.Errorf("expected atProvider.GlobalReplicationGroupId=%s, got %v", testGRGID, cr.Status.AtProvider.GlobalReplicationGroupID)
	}
	if cr.Status.AtProvider.Arn == nil || *cr.Status.AtProvider.Arn != testGRGARN {
		t.Errorf("expected atProvider.Arn=%s, got %v", testGRGARN, cr.Status.AtProvider.Arn)
	}
}

// TestObserve_Available_Drifted verifies that a drifted (description changed) resource
// returns ResourceUpToDate=false.
func TestObserve_Available_Drifted(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)
	cr.Spec.ForProvider.GlobalReplicationGroupDescription = aws.String("new description")

	e := &ExternalClient{Client: &mockGRGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeGlobalReplicationGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeGlobalReplicationGroupsOutput, error) {
			return grgResponse(testGRGID, testGRGARN, "available", testEngine, testEngineVersion, testCacheNodeType, aws.String("old description")), nil
		},
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

// TestObserve_Creating_ReturnsUpToDate verifies that transitional state "creating"
// skips isUpToDate and returns UpToDate=true to prevent spurious Updates that
// would cause InvalidGlobalReplicationGroupState errors.
func TestObserve_Creating_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)
	// Spec has a different description than AWS, but since we're "creating",
	// isUpToDate should NOT be called and UpToDate=true must be returned.
	cr.Spec.ForProvider.GlobalReplicationGroupDescription = aws.String("different description")

	var describeCalled bool
	e := &ExternalClient{Client: &mockGRGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeGlobalReplicationGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeGlobalReplicationGroupsOutput, error) {
			describeCalled = true
			return grgResponse(testGRGID, testGRGARN, "creating", testEngine, testEngineVersion, testCacheNodeType, aws.String("old description")), nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for creating state")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for creating state (prevent spurious Update)")
	}
	if !describeCalled {
		t.Error("expected DescribeGlobalReplicationGroups to be called")
	}
}

// TestObserve_Modifying_ReturnsUpToDate verifies that "modifying" state returns
// UpToDate=true without calling isUpToDate.
func TestObserve_Modifying_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)
	cr.Spec.ForProvider.GlobalReplicationGroupDescription = aws.String("different description")

	e := &ExternalClient{Client: &mockGRGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeGlobalReplicationGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeGlobalReplicationGroupsOutput, error) {
			return grgResponse(testGRGID, testGRGARN, "modifying", testEngine, testEngineVersion, testCacheNodeType, aws.String("old description")), nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for modifying state")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for modifying state")
	}
}

// TestObserve_Deleting_ReturnsUpToDate verifies that "deleting" state returns
// UpToDate=true and sets Deleting condition without calling isUpToDate.
func TestObserve_Deleting_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)

	e := &ExternalClient{Client: &mockGRGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeGlobalReplicationGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeGlobalReplicationGroupsOutput, error) {
			return grgResponse(testGRGID, testGRGARN, "deleting", testEngine, testEngineVersion, testCacheNodeType, nil), nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("expected ResourceExists=true for deleting state")
	}
	if !obs.ResourceUpToDate {
		t.Error("expected ResourceUpToDate=true for deleting state")
	}
}

func TestObserve_DescribeError(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)

	e := &ExternalClient{Client: &mockGRGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeGlobalReplicationGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeGlobalReplicationGroupsOutput, error) {
			return nil, errors.New("unexpected AWS error")
		},
	}}

	_, err := e.Observe(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Observe when describe fails")
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

// TestCreate_Success_SetsExternalName verifies that Create stores the AWS-assigned
// GlobalReplicationGroupId as the external name annotation (IdentifierFromProvider §11b).
// Without this, the external name is lost and the next Observe cannot find the resource.
func TestCreate_Success_SetsExternalName(t *testing.T) {
	cr := newTestCR("my-grg", "")
	cr.Spec.ForProvider.GlobalReplicationGroupDescription = aws.String(testDescription)

	var gotSuffix, gotPrimaryRGID string
	e := &ExternalClient{Client: &mockGRGClient{
		createFn: func(_ context.Context, params *awselasticache.CreateGlobalReplicationGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateGlobalReplicationGroupOutput, error) {
			gotSuffix = aws.ToString(params.GlobalReplicationGroupIdSuffix)
			gotPrimaryRGID = aws.ToString(params.PrimaryReplicationGroupId)
			return &awselasticache.CreateGlobalReplicationGroupOutput{
				GlobalReplicationGroup: &ectypes.GlobalReplicationGroup{
					GlobalReplicationGroupId: aws.String(testGRGID),
					ARN:                      aws.String(testGRGARN),
					Status:                   aws.String("creating"),
				},
			}, nil
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the AWS-assigned ID is stored as the external name.
	gotExtName := native.GetExternalName(cr)
	if gotExtName != testGRGID {
		t.Errorf("expected external name=%s after Create, got %q (meta.SetExternalName not called correctly)", testGRGID, gotExtName)
	}

	// Verify inputs.
	if gotSuffix != testGRGSuffix {
		t.Errorf("expected GlobalReplicationGroupIdSuffix=%s, got %s", testGRGSuffix, gotSuffix)
	}
	if gotPrimaryRGID != testPrimaryRGID {
		t.Errorf("expected PrimaryReplicationGroupId=%s, got %s", testPrimaryRGID, gotPrimaryRGID)
	}
}

func TestCreate_Error(t *testing.T) {
	cr := newTestCR("my-grg", "")

	e := &ExternalClient{Client: &mockGRGClient{
		createFn: func(_ context.Context, _ *awselasticache.CreateGlobalReplicationGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateGlobalReplicationGroupOutput, error) {
			return nil, errors.New("create failed")
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Create")
	}
	// Verify external name is NOT set on failure.
	if gotExtName := native.GetExternalName(cr); gotExtName != "" {
		t.Errorf("expected external name to remain empty on Create failure, got %q", gotExtName)
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_DescriptionChange(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)
	cr.Spec.ForProvider.GlobalReplicationGroupDescription = aws.String("updated description")
	cr.Spec.ForProvider.CacheNodeType = aws.String(testCacheNodeType)
	cr.Spec.ForProvider.Engine = aws.String(testEngine)
	cr.Spec.ForProvider.EngineVersion = aws.String(testEngineVersion)

	var gotID, gotDescription string
	var gotApplyImmediately bool
	e := &ExternalClient{Client: &mockGRGClient{
		modifyFn: func(_ context.Context, params *awselasticache.ModifyGlobalReplicationGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyGlobalReplicationGroupOutput, error) {
			gotID = aws.ToString(params.GlobalReplicationGroupId)
			gotDescription = aws.ToString(params.GlobalReplicationGroupDescription)
			gotApplyImmediately = aws.ToBool(params.ApplyImmediately)
			return &awselasticache.ModifyGlobalReplicationGroupOutput{
				GlobalReplicationGroup: &ectypes.GlobalReplicationGroup{
					GlobalReplicationGroupId: aws.String(testGRGID),
					Status:                   aws.String("modifying"),
				},
			}, nil
		},
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotID != testGRGID {
		t.Errorf("expected GlobalReplicationGroupId=%s, got %s", testGRGID, gotID)
	}
	if gotDescription != "updated description" {
		t.Errorf("expected GlobalReplicationGroupDescription=%q, got %q", "updated description", gotDescription)
	}
	if !gotApplyImmediately {
		t.Error("expected ApplyImmediately=true")
	}
}

func TestUpdate_Error(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)

	e := &ExternalClient{Client: &mockGRGClient{
		modifyFn: func(_ context.Context, _ *awselasticache.ModifyGlobalReplicationGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyGlobalReplicationGroupOutput, error) {
			return nil, errors.New("modify failed")
		},
	}}

	_, err := e.Update(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Update")
	}
}

// ── Delete tests ───────────────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)

	var gotID string
	var gotRetain bool
	e := &ExternalClient{Client: &mockGRGClient{
		deleteFn: func(_ context.Context, params *awselasticache.DeleteGlobalReplicationGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteGlobalReplicationGroupOutput, error) {
			gotID = aws.ToString(params.GlobalReplicationGroupId)
			gotRetain = aws.ToBool(params.RetainPrimaryReplicationGroup)
			return &awselasticache.DeleteGlobalReplicationGroupOutput{}, nil
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotID != testGRGID {
		t.Errorf("expected GlobalReplicationGroupId=%s, got %s", testGRGID, gotID)
	}
	if !gotRetain {
		t.Error("expected RetainPrimaryReplicationGroup=true")
	}
}

// TestDelete_NotFound_Idempotent verifies that GlobalReplicationGroupNotFoundFault
// is treated as success (idempotent delete).
func TestDelete_NotFound_Idempotent(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)

	e := &ExternalClient{Client: &mockGRGClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteGlobalReplicationGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteGlobalReplicationGroupOutput, error) {
			return nil, &ectypes.GlobalReplicationGroupNotFoundFault{Message: aws.String("not found")}
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for GlobalReplicationGroupNotFoundFault (idempotent delete), got: %v", err)
	}
}

func TestDelete_OtherError(t *testing.T) {
	cr := newTestCR("my-grg", testGRGID)

	e := &ExternalClient{Client: &mockGRGClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteGlobalReplicationGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteGlobalReplicationGroupOutput, error) {
			return nil, errors.New("some AWS error")
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Delete when AWS returns unexpected error")
	}
}

// ── isUpToDate / versionMatchesSpec tests ─────────────────────────────────────

func TestVersionMatchesSpec(t *testing.T) {
	tests := []struct {
		specVer string
		awsVer  string
		want    bool
	}{
		{"7.2", "7.2", true},
		{"7.2", "7.2.4", true},
		{"7.2", "7.20", false},
		{"7.2", "7.3", false},
		{"6.x", "6.x", true},
		{"7.2.4", "7.2.4", true},
		{"7.2", "8.0", false},
	}
	for _, tc := range tests {
		got := versionMatchesSpec(tc.specVer, tc.awsVer)
		if got != tc.want {
			t.Errorf("versionMatchesSpec(%q, %q) = %v, want %v", tc.specVer, tc.awsVer, got, tc.want)
		}
	}
}

func TestIsUpToDate_EngineVersionPatch(t *testing.T) {
	spec := &clusternative.GlobalReplicationGroupRAWParameters{
		EngineVersion: aws.String("7.2"),
	}
	grg := ectypes.GlobalReplicationGroup{
		EngineVersion: aws.String("7.2.4"), // AWS returns patch version
	}
	if !isUpToDate(spec, grg) {
		t.Error("expected isUpToDate=true when spec 7.2 matches AWS 7.2.4")
	}
}

func TestIsUpToDate_DescriptionDrift(t *testing.T) {
	spec := &clusternative.GlobalReplicationGroupRAWParameters{
		GlobalReplicationGroupDescription: aws.String("new description"),
	}
	grg := ectypes.GlobalReplicationGroup{
		GlobalReplicationGroupDescription: aws.String("old description"),
	}
	if isUpToDate(spec, grg) {
		t.Error("expected isUpToDate=false when description differs")
	}
}
