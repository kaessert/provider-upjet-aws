// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package usergroup

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

type mockUGClient struct {
	describeFn   func(ctx context.Context, params *awselasticache.DescribeUserGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeUserGroupsOutput, error)
	createFn     func(ctx context.Context, params *awselasticache.CreateUserGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateUserGroupOutput, error)
	modifyFn     func(ctx context.Context, params *awselasticache.ModifyUserGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyUserGroupOutput, error)
	deleteFn     func(ctx context.Context, params *awselasticache.DeleteUserGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteUserGroupOutput, error)
	listTagsFn   func(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	addTagsFn    func(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	removeTagsFn func(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
}

func (m *mockUGClient) DescribeUserGroups(ctx context.Context, params *awselasticache.DescribeUserGroupsInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeUserGroupsOutput, error) {
	return m.describeFn(ctx, params, optFns...)
}
func (m *mockUGClient) CreateUserGroup(ctx context.Context, params *awselasticache.CreateUserGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateUserGroupOutput, error) {
	return m.createFn(ctx, params, optFns...)
}
func (m *mockUGClient) ModifyUserGroup(ctx context.Context, params *awselasticache.ModifyUserGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyUserGroupOutput, error) {
	return m.modifyFn(ctx, params, optFns...)
}
func (m *mockUGClient) DeleteUserGroup(ctx context.Context, params *awselasticache.DeleteUserGroupInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteUserGroupOutput, error) {
	return m.deleteFn(ctx, params, optFns...)
}
func (m *mockUGClient) ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
	if m.listTagsFn != nil {
		return m.listTagsFn(ctx, params, optFns...)
	}
	return &awselasticache.ListTagsForResourceOutput{TagList: []ectypes.Tag{}}, nil
}
func (m *mockUGClient) AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error) {
	if m.addTagsFn != nil {
		return m.addTagsFn(ctx, params, optFns...)
	}
	return &awselasticache.AddTagsToResourceOutput{}, nil
}
func (m *mockUGClient) RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error) {
	if m.removeTagsFn != nil {
		return m.removeTagsFn(ctx, params, optFns...)
	}
	return &awselasticache.RemoveTagsFromResourceOutput{}, nil
}

// ── Constants ─────────────────────────────────────────────────────────────────

const (
	testGroupID  = "my-user-group"
	testGroupARN = "arn:aws:elasticache:us-east-1:609897127049:usergroup:my-user-group"
	testEngine   = "redis"
	testUserID1  = "user-alice"
	testUserID2  = "user-bob"
	testUserID3  = "user-charlie"
)

// ── Helpers ────────────────────────────────────────────────────────────────────

// newTestCR builds a minimal cluster-scoped UserGroupRAW for tests.
func newTestCR(name, extName string) *clusternative.UserGroupRAW {
	cr := &clusternative.UserGroupRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: clusternative.UserGroupRAWSpec{
			ForProvider: clusternative.UserGroupRAWParameters{
				Region: aws.String("us-east-1"),
				Engine: aws.String(testEngine),
			},
		},
	}
	if extName != "" {
		native.SetExternalName(cr, extName)
	}
	return cr
}

// userGroupResponse returns a DescribeUserGroupsOutput with a single group.
func userGroupResponse(id, arn, engine, status string, userIDs []string) *awselasticache.DescribeUserGroupsOutput {
	return &awselasticache.DescribeUserGroupsOutput{
		UserGroups: []ectypes.UserGroup{
			{
				UserGroupId: aws.String(id),
				ARN:         aws.String(arn),
				Engine:      aws.String(engine),
				Status:      aws.String(status),
				UserIds:     userIDs,
			},
		},
	}
}

// noopListTags returns an empty tag list without error.
func noopListTags(_ context.Context, _ *awselasticache.ListTagsForResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
	return &awselasticache.ListTagsForResourceOutput{TagList: []ectypes.Tag{}}, nil
}

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_EmptyExternalName(t *testing.T) {
	cr := newTestCR("my-ug", "")
	e := &ExternalClient{Client: &mockUGClient{}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty external name")
	}
}

func TestObserve_UserGroupNotFound(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)

	e := &ExternalClient{Client: &mockUGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUserGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUserGroupsOutput, error) {
			return nil, &ectypes.UserGroupNotFoundFault{Message: aws.String("not found")}
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for NotFound, got: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when user group not found")
	}
}

func TestObserve_EmptyList(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)

	e := &ExternalClient{Client: &mockUGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUserGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUserGroupsOutput, error) {
			return &awselasticache.DescribeUserGroupsOutput{UserGroups: []ectypes.UserGroup{}}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty UserGroups list")
	}
}

func TestObserve_Active_UpToDate(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)
	cr.Spec.ForProvider.UserIds = []*string{aws.String(testUserID1), aws.String(testUserID2)}

	e := &ExternalClient{Client: &mockUGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUserGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUserGroupsOutput, error) {
			return userGroupResponse(testGroupID, testGroupARN, testEngine, "active", []string{testUserID1, testUserID2}), nil
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
		t.Error("expected ResourceUpToDate=true when userIds match")
	}
	if cr.Status.AtProvider.ID == nil || *cr.Status.AtProvider.ID != testGroupID {
		t.Errorf("expected atProvider.ID=%s, got %v", testGroupID, cr.Status.AtProvider.ID)
	}
	if cr.Status.AtProvider.Arn == nil || *cr.Status.AtProvider.Arn != testGroupARN {
		t.Errorf("expected atProvider.Arn=%s, got %v", testGroupARN, cr.Status.AtProvider.Arn)
	}
}

func TestObserve_Active_Drifted_UserIds(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)
	cr.Spec.ForProvider.UserIds = []*string{aws.String(testUserID1), aws.String(testUserID2)}

	e := &ExternalClient{Client: &mockUGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUserGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUserGroupsOutput, error) {
			// AWS only has user1, spec wants user1+user2 → not up-to-date
			return userGroupResponse(testGroupID, testGroupARN, testEngine, "active", []string{testUserID1}), nil
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
		t.Error("expected ResourceUpToDate=false when userIds differ")
	}
}

// TestObserve_Creating_ReturnsUpToDate verifies that transitional state
// "creating" skips isUpToDate and returns UpToDate=true to prevent spurious
// Update calls that would cause InvalidUserGroupState errors.
func TestObserve_Creating_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)
	// Spec has different userIds than what's on AWS — but since we're "creating",
	// isUpToDate should NOT be called and UpToDate=true must be returned.
	cr.Spec.ForProvider.UserIds = []*string{aws.String(testUserID1), aws.String(testUserID2)}

	var isUpToDateCalled bool
	e := &ExternalClient{Client: &mockUGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUserGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUserGroupsOutput, error) {
			return userGroupResponse(testGroupID, testGroupARN, testEngine, "creating", nil), nil
		},
		// listTagsFn should NOT be called during a transitional state
		listTagsFn: func(_ context.Context, _ *awselasticache.ListTagsForResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
			isUpToDateCalled = true
			return &awselasticache.ListTagsForResourceOutput{}, nil
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
	if isUpToDateCalled {
		t.Error("isUpToDate should not be called for 'creating' state (ListTagsForResource was invoked)")
	}
}

// TestObserve_Modifying_ReturnsUpToDate verifies that "modifying" state returns
// UpToDate=true without calling isUpToDate.
func TestObserve_Modifying_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)

	var listTagsCalled bool
	e := &ExternalClient{Client: &mockUGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUserGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUserGroupsOutput, error) {
			return userGroupResponse(testGroupID, testGroupARN, testEngine, "modifying", nil), nil
		},
		listTagsFn: func(_ context.Context, _ *awselasticache.ListTagsForResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
			listTagsCalled = true
			return &awselasticache.ListTagsForResourceOutput{}, nil
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
	if listTagsCalled {
		t.Error("ListTagsForResource should not be called during 'modifying' state")
	}
}

// TestObserve_Deleting_ReturnsUpToDate verifies that "deleting" state returns
// UpToDate=true and sets Deleting condition without calling isUpToDate.
func TestObserve_Deleting_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)

	var listTagsCalled bool
	e := &ExternalClient{Client: &mockUGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUserGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUserGroupsOutput, error) {
			return userGroupResponse(testGroupID, testGroupARN, testEngine, "deleting", nil), nil
		},
		listTagsFn: func(_ context.Context, _ *awselasticache.ListTagsForResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
			listTagsCalled = true
			return &awselasticache.ListTagsForResourceOutput{}, nil
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
	if listTagsCalled {
		t.Error("ListTagsForResource should not be called during 'deleting' state")
	}
}

func TestObserve_DescribeError(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)

	e := &ExternalClient{Client: &mockUGClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUserGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUserGroupsOutput, error) {
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
	cr := newTestCR("my-ug", testGroupID)
	cr.Spec.ForProvider.UserIds = []*string{aws.String(testUserID1)}

	var gotID string
	var gotEngine string
	var gotUserIds []string
	e := &ExternalClient{Client: &mockUGClient{
		createFn: func(_ context.Context, params *awselasticache.CreateUserGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateUserGroupOutput, error) {
			gotID = aws.ToString(params.UserGroupId)
			gotEngine = aws.ToString(params.Engine)
			gotUserIds = params.UserIds
			return &awselasticache.CreateUserGroupOutput{
				UserGroupId: params.UserGroupId,
				ARN:         aws.String(testGroupARN),
			}, nil
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotID != testGroupID {
		t.Errorf("expected UserGroupId=%s, got %s", testGroupID, gotID)
	}
	if gotEngine != testEngine {
		t.Errorf("expected Engine=%s, got %s", testEngine, gotEngine)
	}
	if len(gotUserIds) != 1 || gotUserIds[0] != testUserID1 {
		t.Errorf("expected UserIds=[%s], got %v", testUserID1, gotUserIds)
	}
}

func TestCreate_Error(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)

	e := &ExternalClient{Client: &mockUGClient{
		createFn: func(_ context.Context, _ *awselasticache.CreateUserGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateUserGroupOutput, error) {
			return nil, errors.New("create failed")
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Create")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_AddUsers(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)
	// Desired: user1 + user2
	cr.Spec.ForProvider.UserIds = []*string{aws.String(testUserID1), aws.String(testUserID2)}
	// Observed (atProvider): only user1
	cr.Status.AtProvider = clusternative.UserGroupRAWObservation{
		Arn:     aws.String(testGroupARN),
		UserIds: []*string{aws.String(testUserID1)},
	}

	var gotToAdd []string
	var gotToRemove []string
	e := &ExternalClient{Client: &mockUGClient{
		modifyFn: func(_ context.Context, params *awselasticache.ModifyUserGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyUserGroupOutput, error) {
			gotToAdd = params.UserIdsToAdd
			gotToRemove = params.UserIdsToRemove
			return &awselasticache.ModifyUserGroupOutput{}, nil
		},
		listTagsFn: noopListTags,
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotToAdd) != 1 || gotToAdd[0] != testUserID2 {
		t.Errorf("expected UserIdsToAdd=[%s], got %v", testUserID2, gotToAdd)
	}
	if len(gotToRemove) != 0 {
		t.Errorf("expected UserIdsToRemove=[], got %v", gotToRemove)
	}
}

func TestUpdate_RemoveUsers(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)
	// Desired: only user1
	cr.Spec.ForProvider.UserIds = []*string{aws.String(testUserID1)}
	// Observed: user1 + user2
	cr.Status.AtProvider = clusternative.UserGroupRAWObservation{
		Arn:     aws.String(testGroupARN),
		UserIds: []*string{aws.String(testUserID1), aws.String(testUserID2)},
	}

	var gotToAdd []string
	var gotToRemove []string
	e := &ExternalClient{Client: &mockUGClient{
		modifyFn: func(_ context.Context, params *awselasticache.ModifyUserGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyUserGroupOutput, error) {
			gotToAdd = params.UserIdsToAdd
			gotToRemove = params.UserIdsToRemove
			return &awselasticache.ModifyUserGroupOutput{}, nil
		},
		listTagsFn: noopListTags,
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotToAdd) != 0 {
		t.Errorf("expected UserIdsToAdd=[], got %v", gotToAdd)
	}
	if len(gotToRemove) != 1 || gotToRemove[0] != testUserID2 {
		t.Errorf("expected UserIdsToRemove=[%s], got %v", testUserID2, gotToRemove)
	}
}

func TestUpdate_Mixed_AddAndRemove(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)
	// Desired: user1 + user3 (user2 removed, user3 added)
	cr.Spec.ForProvider.UserIds = []*string{aws.String(testUserID1), aws.String(testUserID3)}
	// Observed: user1 + user2
	cr.Status.AtProvider = clusternative.UserGroupRAWObservation{
		Arn:     aws.String(testGroupARN),
		UserIds: []*string{aws.String(testUserID1), aws.String(testUserID2)},
	}

	var gotToAdd []string
	var gotToRemove []string
	e := &ExternalClient{Client: &mockUGClient{
		modifyFn: func(_ context.Context, params *awselasticache.ModifyUserGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyUserGroupOutput, error) {
			gotToAdd = params.UserIdsToAdd
			gotToRemove = params.UserIdsToRemove
			return &awselasticache.ModifyUserGroupOutput{}, nil
		},
		listTagsFn: noopListTags,
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotToAdd) != 1 || gotToAdd[0] != testUserID3 {
		t.Errorf("expected UserIdsToAdd=[%s], got %v", testUserID3, gotToAdd)
	}
	if len(gotToRemove) != 1 || gotToRemove[0] != testUserID2 {
		t.Errorf("expected UserIdsToRemove=[%s], got %v", testUserID2, gotToRemove)
	}
}

func TestUpdate_NoUserIdChanges_NoModifyCalled(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)
	// Desired and observed are the same
	cr.Spec.ForProvider.UserIds = []*string{aws.String(testUserID1)}
	cr.Status.AtProvider = clusternative.UserGroupRAWObservation{
		Arn:     aws.String(testGroupARN),
		UserIds: []*string{aws.String(testUserID1)},
	}

	var modifyCalled bool
	e := &ExternalClient{Client: &mockUGClient{
		modifyFn: func(_ context.Context, _ *awselasticache.ModifyUserGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyUserGroupOutput, error) {
			modifyCalled = true
			return &awselasticache.ModifyUserGroupOutput{}, nil
		},
		listTagsFn: noopListTags,
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if modifyCalled {
		t.Error("ModifyUserGroup should not be called when userId list is unchanged")
	}
}

func TestUpdate_TagSync(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)
	cr.Spec.ForProvider.UserIds = []*string{aws.String(testUserID1)}
	cr.Spec.ForProvider.Tags = map[string]*string{"env": aws.String("prod")}
	cr.Status.AtProvider = clusternative.UserGroupRAWObservation{
		Arn:     aws.String(testGroupARN),
		UserIds: []*string{aws.String(testUserID1)},
	}

	var addTagsCalled bool
	e := &ExternalClient{Client: &mockUGClient{
		listTagsFn: noopListTags, // no existing tags
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
	cr := newTestCR("my-ug", testGroupID)

	e := &ExternalClient{Client: &mockUGClient{
		deleteFn: func(_ context.Context, params *awselasticache.DeleteUserGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteUserGroupOutput, error) {
			return &awselasticache.DeleteUserGroupOutput{}, nil
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDelete_NotFound_Idempotent verifies that UserGroupNotFoundFault is treated
// as success (idempotent delete).
func TestDelete_NotFound_Idempotent(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)

	e := &ExternalClient{Client: &mockUGClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteUserGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteUserGroupOutput, error) {
			return nil, &ectypes.UserGroupNotFoundFault{Message: aws.String("not found")}
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for UserGroupNotFoundFault (idempotent delete), got: %v", err)
	}
}

func TestDelete_OtherError(t *testing.T) {
	cr := newTestCR("my-ug", testGroupID)

	e := &ExternalClient{Client: &mockUGClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteUserGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteUserGroupOutput, error) {
			return nil, errors.New("some AWS error")
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Delete when AWS returns unexpected error")
	}
}

// ── diffUserIds tests ──────────────────────────────────────────────────────────

func TestDiffUserIds_AddOnly(t *testing.T) {
	desired := []*string{aws.String("a"), aws.String("b")}
	observed := []*string{aws.String("a")}

	toAdd, toRemove := diffUserIds(desired, observed)
	if len(toAdd) != 1 || toAdd[0] != "b" {
		t.Errorf("expected toAdd=[b], got %v", toAdd)
	}
	if len(toRemove) != 0 {
		t.Errorf("expected toRemove=[], got %v", toRemove)
	}
}

func TestDiffUserIds_RemoveOnly(t *testing.T) {
	desired := []*string{aws.String("a")}
	observed := []*string{aws.String("a"), aws.String("b")}

	toAdd, toRemove := diffUserIds(desired, observed)
	if len(toAdd) != 0 {
		t.Errorf("expected toAdd=[], got %v", toAdd)
	}
	if len(toRemove) != 1 || toRemove[0] != "b" {
		t.Errorf("expected toRemove=[b], got %v", toRemove)
	}
}

func TestDiffUserIds_Mixed(t *testing.T) {
	desired := []*string{aws.String("a"), aws.String("c")}
	observed := []*string{aws.String("a"), aws.String("b")}

	toAdd, toRemove := diffUserIds(desired, observed)
	if len(toAdd) != 1 || toAdd[0] != "c" {
		t.Errorf("expected toAdd=[c], got %v", toAdd)
	}
	if len(toRemove) != 1 || toRemove[0] != "b" {
		t.Errorf("expected toRemove=[b], got %v", toRemove)
	}
}

func TestDiffUserIds_NoChange(t *testing.T) {
	desired := []*string{aws.String("a"), aws.String("b")}
	observed := []*string{aws.String("b"), aws.String("a")} // different order

	toAdd, toRemove := diffUserIds(desired, observed)
	if len(toAdd) != 0 {
		t.Errorf("expected toAdd=[], got %v", toAdd)
	}
	if len(toRemove) != 0 {
		t.Errorf("expected toRemove=[], got %v", toRemove)
	}
}
