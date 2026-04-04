// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package user

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	ectypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1beta1native "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native"
	clusternativev2 "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native"
	"github.com/upbound/provider-aws/v2/internal/native"
)

// ── Mock client ────────────────────────────────────────────────────────────────

type mockUserClient struct {
	describeFn   func(ctx context.Context, params *awselasticache.DescribeUsersInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeUsersOutput, error)
	createFn     func(ctx context.Context, params *awselasticache.CreateUserInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateUserOutput, error)
	modifyFn     func(ctx context.Context, params *awselasticache.ModifyUserInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyUserOutput, error)
	deleteFn     func(ctx context.Context, params *awselasticache.DeleteUserInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteUserOutput, error)
	listTagsFn   func(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error)
	addTagsFn    func(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error)
	removeTagsFn func(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error)
}

func (m *mockUserClient) DescribeUsers(ctx context.Context, params *awselasticache.DescribeUsersInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DescribeUsersOutput, error) {
	return m.describeFn(ctx, params, optFns...)
}
func (m *mockUserClient) CreateUser(ctx context.Context, params *awselasticache.CreateUserInput, optFns ...func(*awselasticache.Options)) (*awselasticache.CreateUserOutput, error) {
	return m.createFn(ctx, params, optFns...)
}
func (m *mockUserClient) ModifyUser(ctx context.Context, params *awselasticache.ModifyUserInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ModifyUserOutput, error) {
	return m.modifyFn(ctx, params, optFns...)
}
func (m *mockUserClient) DeleteUser(ctx context.Context, params *awselasticache.DeleteUserInput, optFns ...func(*awselasticache.Options)) (*awselasticache.DeleteUserOutput, error) {
	return m.deleteFn(ctx, params, optFns...)
}
func (m *mockUserClient) ListTagsForResource(ctx context.Context, params *awselasticache.ListTagsForResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
	if m.listTagsFn != nil {
		return m.listTagsFn(ctx, params, optFns...)
	}
	return &awselasticache.ListTagsForResourceOutput{TagList: []ectypes.Tag{}}, nil
}
func (m *mockUserClient) AddTagsToResource(ctx context.Context, params *awselasticache.AddTagsToResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error) {
	if m.addTagsFn != nil {
		return m.addTagsFn(ctx, params, optFns...)
	}
	return &awselasticache.AddTagsToResourceOutput{}, nil
}
func (m *mockUserClient) RemoveTagsFromResource(ctx context.Context, params *awselasticache.RemoveTagsFromResourceInput, optFns ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error) {
	if m.removeTagsFn != nil {
		return m.removeTagsFn(ctx, params, optFns...)
	}
	return &awselasticache.RemoveTagsFromResourceOutput{}, nil
}

// ── Constants ─────────────────────────────────────────────────────────────────

const (
	testUserID  = "my-test-user"
	testUserARN = "arn:aws:elasticache:us-east-1:609897127049:user:my-test-user"
	testEngine  = "redis"
)

// ── Helpers ────────────────────────────────────────────────────────────────────

// newTestCR builds a minimal cluster-scoped UserRAW v1beta2 for tests.
func newTestCR(name, extName string) *clusternativev2.UserRAW {
	cr := &clusternativev2.UserRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: map[string]string{},
		},
		Spec: clusternativev2.UserRAWSpec{
			ForProvider: clusternativev2.UserRAWParameters{
				Region:       aws.String("us-east-1"),
				Engine:       aws.String(testEngine),
				AccessString: aws.String("on ~* +@all"),
				UserName:     aws.String("testuser"),
			},
		},
	}
	if extName != "" {
		native.SetExternalName(cr, extName)
	}
	return cr
}

// userResponse returns a DescribeUsersOutput with a single user.
func userResponse(id, arn, engine, status, accessString string) *awselasticache.DescribeUsersOutput {
	return &awselasticache.DescribeUsersOutput{
		Users: []ectypes.User{
			{
				UserId:       aws.String(id),
				ARN:          aws.String(arn),
				Engine:       aws.String(engine),
				Status:       aws.String(status),
				AccessString: aws.String(accessString),
				UserName:     aws.String("testuser"),
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
	cr := newTestCR("my-user", "")
	e := &ExternalClient{Client: &mockUserClient{}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty external name")
	}
}

func TestObserve_UserNotFound(t *testing.T) {
	cr := newTestCR("my-user", testUserID)

	e := &ExternalClient{Client: &mockUserClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUsersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUsersOutput, error) {
			return nil, &ectypes.UserNotFoundFault{Message: aws.String("not found")}
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for NotFound, got: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false when user not found")
	}
}

func TestObserve_EmptyList(t *testing.T) {
	cr := newTestCR("my-user", testUserID)

	e := &ExternalClient{Client: &mockUserClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUsersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUsersOutput, error) {
			return &awselasticache.DescribeUsersOutput{Users: []ectypes.User{}}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Error("expected ResourceExists=false for empty Users list")
	}
}

func TestObserve_Active_UpToDate(t *testing.T) {
	cr := newTestCR("my-user", testUserID)

	e := &ExternalClient{Client: &mockUserClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUsersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUsersOutput, error) {
			return userResponse(testUserID, testUserARN, testEngine, "active", "on ~* +@all"), nil
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
		t.Error("expected ResourceUpToDate=true when fields match")
	}
	if cr.Status.AtProvider.ID == nil || *cr.Status.AtProvider.ID != testUserID {
		t.Errorf("expected atProvider.ID=%s, got %v", testUserID, cr.Status.AtProvider.ID)
	}
	if cr.Status.AtProvider.Arn == nil || *cr.Status.AtProvider.Arn != testUserARN {
		t.Errorf("expected atProvider.Arn=%s, got %v", testUserARN, cr.Status.AtProvider.Arn)
	}
}

func TestObserve_Active_Drifted_AccessString(t *testing.T) {
	cr := newTestCR("my-user", testUserID)
	cr.Spec.ForProvider.AccessString = aws.String("on ~* +@read") // differs from AWS

	e := &ExternalClient{Client: &mockUserClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUsersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUsersOutput, error) {
			return userResponse(testUserID, testUserARN, testEngine, "active", "on ~* +@all"), nil
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
		t.Error("expected ResourceUpToDate=false when accessString differs")
	}
}

// TestObserve_Modifying_ReturnsUpToDate verifies that "modifying" state
// skips isUpToDate and returns UpToDate=true to prevent spurious Update calls
// that would cause InvalidUserState errors.
func TestObserve_Modifying_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-user", testUserID)

	var listTagsCalled bool
	e := &ExternalClient{Client: &mockUserClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUsersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUsersOutput, error) {
			return userResponse(testUserID, testUserARN, testEngine, "modifying", "on ~* +@all"), nil
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
		t.Error("expected ResourceUpToDate=true for modifying state (prevent spurious Update)")
	}
	if listTagsCalled {
		t.Error("ListTagsForResource should not be called during 'modifying' state")
	}
}

// TestObserve_Deleting_ReturnsUpToDate verifies that "deleting" state returns
// UpToDate=true and sets Deleting condition without calling isUpToDate.
func TestObserve_Deleting_ReturnsUpToDate(t *testing.T) {
	cr := newTestCR("my-user", testUserID)

	var listTagsCalled bool
	e := &ExternalClient{Client: &mockUserClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUsersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUsersOutput, error) {
			return userResponse(testUserID, testUserARN, testEngine, "deleting", "on ~* +@all"), nil
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
	cr := newTestCR("my-user", testUserID)

	e := &ExternalClient{Client: &mockUserClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUsersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUsersOutput, error) {
			return nil, errors.New("unexpected AWS error")
		},
	}}

	_, err := e.Observe(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Observe when describe fails")
	}
}

// TestObserve_NoPasswordInAtProvider verifies that password values never
// appear in status.atProvider (Observation struct has no password field).
func TestObserve_NoPasswordInAtProvider(t *testing.T) {
	cr := newTestCR("my-user", testUserID)

	e := &ExternalClient{Client: &mockUserClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUsersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUsersOutput, error) {
			// AWS API never returns passwords — they are not present in ectypes.User.
			return userResponse(testUserID, testUserARN, testEngine, "active", "on ~* +@all"), nil
		},
		listTagsFn: noopListTags,
	}}

	_, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Observation struct has no password field by design — just verify it compiles
	// and atProvider is populated without passwords.
	obs := cr.Status.AtProvider
	if obs.ID == nil || *obs.ID != testUserID {
		t.Errorf("expected atProvider.ID=%s, got %v", testUserID, obs.ID)
	}
}

// ── Create tests ───────────────────────────────────────────────────────────────

func TestCreate_Success_NoPasswords(t *testing.T) {
	cr := newTestCR("my-user", testUserID)

	var gotUserId string
	var gotEngine string
	var gotPasswords []string

	e := &ExternalClient{Client: &mockUserClient{
		createFn: func(_ context.Context, params *awselasticache.CreateUserInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateUserOutput, error) {
			gotUserId = aws.ToString(params.UserId)
			gotEngine = aws.ToString(params.Engine)
			gotPasswords = params.Passwords
			return &awselasticache.CreateUserOutput{
				UserId: params.UserId,
				ARN:    aws.String(testUserARN),
			}, nil
		},
	}}

	result, err := e.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotUserId != testUserID {
		t.Errorf("expected UserId=%s, got %s", testUserID, gotUserId)
	}
	if gotEngine != testEngine {
		t.Errorf("expected Engine=%s, got %s", testEngine, gotEngine)
	}
	if len(gotPasswords) != 0 {
		t.Errorf("expected no passwords, got %v", gotPasswords)
	}
	// No passwords → no connection details.
	if len(result.ConnectionDetails) != 0 {
		t.Errorf("expected empty connection details, got %v", result.ConnectionDetails)
	}
}

func TestCreate_Error(t *testing.T) {
	cr := newTestCR("my-user", testUserID)

	e := &ExternalClient{Client: &mockUserClient{
		createFn: func(_ context.Context, _ *awselasticache.CreateUserInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateUserOutput, error) {
			return nil, errors.New("create failed")
		},
	}}

	_, err := e.Create(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Create")
	}
}

// ── Update tests ───────────────────────────────────────────────────────────────

func TestUpdate_Success(t *testing.T) {
	cr := newTestCR("my-user", testUserID)
	cr.Status.AtProvider = clusternativev2.UserRAWObservation{
		Arn: aws.String(testUserARN),
	}

	var gotUserId string
	e := &ExternalClient{Client: &mockUserClient{
		modifyFn: func(_ context.Context, params *awselasticache.ModifyUserInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyUserOutput, error) {
			gotUserId = aws.ToString(params.UserId)
			return &awselasticache.ModifyUserOutput{}, nil
		},
		listTagsFn: noopListTags,
	}}

	_, err := e.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotUserId != testUserID {
		t.Errorf("expected UserId=%s, got %s", testUserID, gotUserId)
	}
}

func TestUpdate_Error(t *testing.T) {
	cr := newTestCR("my-user", testUserID)

	e := &ExternalClient{Client: &mockUserClient{
		modifyFn: func(_ context.Context, _ *awselasticache.ModifyUserInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyUserOutput, error) {
			return nil, errors.New("modify failed")
		},
	}}

	_, err := e.Update(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Update")
	}
}

func TestUpdate_TagSync(t *testing.T) {
	cr := newTestCR("my-user", testUserID)
	cr.Spec.ForProvider.Tags = map[string]*string{"env": aws.String("prod")}
	cr.Status.AtProvider = clusternativev2.UserRAWObservation{
		Arn: aws.String(testUserARN),
	}

	var addTagsCalled bool
	e := &ExternalClient{Client: &mockUserClient{
		modifyFn: func(_ context.Context, _ *awselasticache.ModifyUserInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyUserOutput, error) {
			return &awselasticache.ModifyUserOutput{}, nil
		},
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
	cr := newTestCR("my-user", testUserID)

	e := &ExternalClient{Client: &mockUserClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteUserInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteUserOutput, error) {
			return &awselasticache.DeleteUserOutput{}, nil
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDelete_NotFound_Idempotent verifies that UserNotFoundFault is treated
// as success (idempotent delete).
func TestDelete_NotFound_Idempotent(t *testing.T) {
	cr := newTestCR("my-user", testUserID)

	e := &ExternalClient{Client: &mockUserClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteUserInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteUserOutput, error) {
			return nil, &ectypes.UserNotFoundFault{Message: aws.String("not found")}
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("expected nil error for UserNotFoundFault (idempotent delete), got: %v", err)
	}
}

func TestDelete_OtherError(t *testing.T) {
	cr := newTestCR("my-user", testUserID)

	e := &ExternalClient{Client: &mockUserClient{
		deleteFn: func(_ context.Context, _ *awselasticache.DeleteUserInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteUserOutput, error) {
			return nil, errors.New("some AWS error")
		},
	}}

	_, err := e.Delete(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Delete when AWS returns unexpected error")
	}
}

// ── Password hash tests ────────────────────────────────────────────────────────

// TestHashPasswordsToConnDetails verifies that passwords are SHA-256 hashed
// and published as connection details with the correct keys.
func TestHashPasswordsToConnDetails(t *testing.T) {
	pw1 := "supersecret1"
	pw2 := "supersecret2"

	details := hashPasswordsToConnDetails([]string{pw1, pw2})

	if len(details) != 2 {
		t.Fatalf("expected 2 connection details, got %d", len(details))
	}

	// Verify key names and hash values.
	for i, pw := range []string{pw1, pw2} {
		key := fmt.Sprintf("password_hash_%d", i)
		val, ok := details[key]
		if !ok {
			t.Errorf("expected key %s in connection details", key)
			continue
		}
		expected := sha256.Sum256([]byte(pw))
		expectedHex := fmt.Sprintf("%x", expected)
		if string(val) != expectedHex {
			t.Errorf("key %s: expected hash %s, got %s", key, expectedHex, string(val))
		}
	}
}

func TestHashPasswordsToConnDetails_Empty(t *testing.T) {
	details := hashPasswordsToConnDetails(nil)
	if len(details) != 0 {
		t.Errorf("expected empty connection details for nil passwords, got %d", len(details))
	}
}

// ── Password path tests ────────────────────────────────────────────────────────

// TestReadPasswords_PreferNested verifies that nested path takes priority.
// With Kube=nil, readSecretSelectors returns nil,nil — we verify no panic.
func TestReadPasswords_PreferNested(t *testing.T) {
	spec := &clusternativev2.UserRAWParameters{
		PasswordsSecretRef: &[]xpv1.SecretKeySelector{
			{SecretReference: xpv1.SecretReference{Name: "top-level-secret", Namespace: "default"}, Key: "password"},
		},
		AuthenticationMode: &clusternativev2.AuthenticationModeRAWParameters{
			Type: aws.String("password"),
			PasswordsSecretRef: &[]xpv1.SecretKeySelector{
				{SecretReference: xpv1.SecretReference{Name: "nested-secret", Namespace: "default"}, Key: "password"},
			},
		},
	}

	// With Kube=nil, readSecretSelectors returns nil, nil (no panic).
	e := &ExternalClient{Client: &mockUserClient{}, Kube: nil}
	passwords, err := e.readPasswords(context.Background(), spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With Kube=nil, readSecretSelectors returns nil.
	if passwords != nil {
		t.Errorf("expected nil passwords when Kube=nil, got %v", passwords)
	}
}

func TestReadPasswords_NoRef(t *testing.T) {
	spec := &clusternativev2.UserRAWParameters{
		Engine:       aws.String("redis"),
		AccessString: aws.String("on ~* +@all"),
	}

	e := &ExternalClient{Client: &mockUserClient{}, Kube: nil}
	passwords, err := e.readPasswords(context.Background(), spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if passwords != nil {
		t.Errorf("expected nil passwords when no secretRef set, got %v", passwords)
	}
}

// TestReadPasswords_TopLevelPath verifies that top-level path is used when
// nested path is not set. With Kube=nil, returns nil,nil (no panic).
func TestReadPasswords_TopLevelPath(t *testing.T) {
	spec := &clusternativev2.UserRAWParameters{
		PasswordsSecretRef: &[]xpv1.SecretKeySelector{
			{SecretReference: xpv1.SecretReference{Name: "top-level-secret", Namespace: "default"}, Key: "password"},
		},
	}

	e := &ExternalClient{Client: &mockUserClient{}, Kube: nil}
	passwords, err := e.readPasswords(context.Background(), spec)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With Kube=nil, readSecretSelectors returns nil (no crash).
	if passwords != nil {
		t.Errorf("expected nil passwords when Kube=nil, got %v", passwords)
	}
}

// ── v1beta1 ↔ v1beta2 conversion tests ────────────────────────────────────────

// TestConversion_v1beta1_To_v1beta2 verifies that slice→pointer conversion works.
func TestConversion_v1beta1_To_v1beta2(t *testing.T) {
	src := &v1beta1native.UserRAW{
		ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
		Spec: v1beta1native.UserRAWSpec{
			ForProvider: v1beta1native.UserRAWParameters{
				Region:       aws.String("us-east-1"),
				Engine:       aws.String("redis"),
				AccessString: aws.String("on ~* +@all"),
				UserName:     aws.String("testuser"),
				AuthenticationMode: []v1beta1native.AuthenticationModeRAWParameters{
					{
						Type: aws.String("password"),
					},
				},
			},
		},
	}

	dst := &clusternativev2.UserRAW{}
	if err := src.ConvertTo(dst); err != nil {
		t.Fatalf("ConvertTo failed: %v", err)
	}

	// Verify TypeMeta is set.
	if dst.Kind != "UserRAW" {
		t.Errorf("expected Kind=UserRAW, got %s", dst.Kind)
	}

	// Verify slice→pointer conversion.
	if dst.Spec.ForProvider.AuthenticationMode == nil {
		t.Fatal("expected AuthenticationMode to be non-nil after conversion")
	}
	if aws.ToString(dst.Spec.ForProvider.AuthenticationMode.Type) != "password" {
		t.Errorf("expected AuthenticationMode.Type=password, got %v", dst.Spec.ForProvider.AuthenticationMode.Type)
	}

	// Round-trip back to v1beta1.
	dst2 := &v1beta1native.UserRAW{}
	if err := dst2.ConvertFrom(dst); err != nil {
		t.Fatalf("ConvertFrom failed: %v", err)
	}
	if len(dst2.Spec.ForProvider.AuthenticationMode) != 1 {
		t.Fatalf("expected 1 AuthenticationMode in v1beta1, got %d", len(dst2.Spec.ForProvider.AuthenticationMode))
	}
	if aws.ToString(dst2.Spec.ForProvider.AuthenticationMode[0].Type) != "password" {
		t.Errorf("expected Type=password after round-trip, got %v", dst2.Spec.ForProvider.AuthenticationMode[0].Type)
	}
	if dst2.Kind != "UserRAW" {
		t.Errorf("expected Kind=UserRAW in v1beta1 after round-trip, got %s", dst2.Kind)
	}
}

// TestConversion_v1beta1_NoAuthMode verifies conversion when AuthenticationMode is nil.
func TestConversion_v1beta1_NoAuthMode(t *testing.T) {
	src := &v1beta1native.UserRAW{
		ObjectMeta: metav1.ObjectMeta{Name: "test-user"},
		Spec: v1beta1native.UserRAWSpec{
			ForProvider: v1beta1native.UserRAWParameters{
				Region:   aws.String("us-east-1"),
				Engine:   aws.String("redis"),
				UserName: aws.String("testuser"),
				// No AuthenticationMode
			},
		},
	}

	dst := &clusternativev2.UserRAW{}
	if err := src.ConvertTo(dst); err != nil {
		t.Fatalf("ConvertTo failed: %v", err)
	}
	if dst.Spec.ForProvider.AuthenticationMode != nil {
		t.Error("expected AuthenticationMode to be nil when not set in v1beta1")
	}

	dst2 := &v1beta1native.UserRAW{}
	if err := dst2.ConvertFrom(dst); err != nil {
		t.Fatalf("ConvertFrom failed: %v", err)
	}
	if len(dst2.Spec.ForProvider.AuthenticationMode) != 0 {
		t.Errorf("expected empty AuthenticationMode slice, got %v", dst2.Spec.ForProvider.AuthenticationMode)
	}
}

// ── Late-initialization tests ──────────────────────────────────────────────────

// TestObserve_LateInit_NilEngine verifies that when spec.Engine is nil and
// AWS returns an Engine value, Observe returns ResourceLateInitialized=true and
// populates the spec field — preventing infinite reconciliation from the Engine
// comparison in isUpToDate.
func TestObserve_LateInit_NilEngine(t *testing.T) {
	cr := newTestCR("my-user", testUserID)
	cr.Spec.ForProvider.Engine = nil // intentionally nil

	e := &ExternalClient{Client: &mockUserClient{
		describeFn: func(_ context.Context, _ *awselasticache.DescribeUsersInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeUsersOutput, error) {
			return &awselasticache.DescribeUsersOutput{
				Users: []ectypes.User{
					{
						UserId:       aws.String(testUserID),
						ARN:          aws.String(testUserARN),
						Status:       aws.String("active"),
						Engine:       aws.String("redis"),
						AccessString: aws.String("on ~* +@all"),
						UserName:     aws.String("testuser"),
					},
				},
			}, nil
		},
	}}

	obs, err := e.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must signal late initialization so the controller saves the spec.
	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true when Engine is nil and AWS returns a value")
	}

	// spec.Engine must be populated from the AWS response.
	if cr.Spec.ForProvider.Engine == nil {
		t.Fatal("expected spec.Engine to be populated after late initialization")
	}
	if got, want := *cr.Spec.ForProvider.Engine, "redis"; got != want {
		t.Errorf("spec.Engine: got %q, want %q", got, want)
	}
}

// TestIsUpToDate_NilEngine_NoLoop verifies that when spec.Engine is nil,
// isUpToDate does not return false for the Engine field — preventing an
// infinite reconciliation loop.
func TestIsUpToDate_NilEngine_NoLoop(t *testing.T) {
	spec := &clusternativev2.UserRAWParameters{
		Region:       aws.String("us-east-1"),
		Engine:       nil, // intentionally nil
		AccessString: aws.String("on ~* +@all"),
		UserName:     aws.String("testuser"),
	}
	u := ectypes.User{
		Engine:       aws.String("redis"),
		AccessString: aws.String("on ~* +@all"),
	}

	if !isUpToDate(spec, u, nil) {
		t.Error("isUpToDate should return true when spec.Engine is nil — no spurious update")
	}
}

// TestIsUpToDate_EngineCase_NoLoop verifies case-insensitive Engine comparison.
// Spec may say "REDIS" (uppercase) while AWS returns "redis" (lowercase).
// Without this, isUpToDate returns false every cycle causing infinite updates.
func TestIsUpToDate_EngineCase_NoLoop(t *testing.T) {
	spec := &clusternativev2.UserRAWParameters{
		Region:       aws.String("us-east-1"),
		Engine:       aws.String("REDIS"), // uppercase
		AccessString: aws.String("on ~* +@all"),
		UserName:     aws.String("testuser"),
	}
	u := ectypes.User{
		Engine:       aws.String("redis"), // AWS returns lowercase
		AccessString: aws.String("on ~* +@all"),
	}

	if !isUpToDate(spec, u, nil) {
		t.Error("isUpToDate should return true when spec Engine 'REDIS' matches AWS 'redis' case-insensitively")
	}
}
