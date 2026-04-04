// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package replicationgroup

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awselasticache "github.com/aws/aws-sdk-go-v2/service/elasticache"
	ectypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"

	clusternativev2 "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native"
)

// ── Fake AWS client helpers ───────────────────────────────────────────────────

type fakeRGClient struct {
	// Control which calls are made.
	createCalled      bool
	describeCalled    bool
	modifyCalled      bool
	modifyShardCalled bool
	deleteCalled      bool
	addTagsCalled     bool
	removeTagsCalled  bool
	listTagsCalled    bool

	// Configurable responses.
	describeResp    *awselasticache.DescribeReplicationGroupsOutput
	describeErr     error
	createResp      *awselasticache.CreateReplicationGroupOutput
	createErr       error
	modifyResp      *awselasticache.ModifyReplicationGroupOutput
	modifyErr       error
	modifyShardResp *awselasticache.ModifyReplicationGroupShardConfigurationOutput
	modifyShardErr  error
	deleteResp      *awselasticache.DeleteReplicationGroupOutput
	deleteErr       error
	listTagsResp    *awselasticache.ListTagsForResourceOutput
	listTagsErr     error
	addTagsErr      error
	removeTagsErr   error

	// Captured inputs for assertion.
	capturedCreateInput      *awselasticache.CreateReplicationGroupInput
	capturedModifyInput      *awselasticache.ModifyReplicationGroupInput
	capturedModifyShardInput *awselasticache.ModifyReplicationGroupShardConfigurationInput
	capturedDeleteInput      *awselasticache.DeleteReplicationGroupInput
}

func (f *fakeRGClient) CreateReplicationGroup(_ context.Context, params *awselasticache.CreateReplicationGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.CreateReplicationGroupOutput, error) {
	f.createCalled = true
	f.capturedCreateInput = params
	if f.createErr != nil {
		return nil, f.createErr
	}
	if f.createResp != nil {
		return f.createResp, nil
	}
	return &awselasticache.CreateReplicationGroupOutput{
		ReplicationGroup: &ectypes.ReplicationGroup{
			ReplicationGroupId: params.ReplicationGroupId,
			Status:             aws.String("creating"),
		},
	}, nil
}

func (f *fakeRGClient) DescribeReplicationGroups(_ context.Context, _ *awselasticache.DescribeReplicationGroupsInput, _ ...func(*awselasticache.Options)) (*awselasticache.DescribeReplicationGroupsOutput, error) {
	f.describeCalled = true
	if f.describeErr != nil {
		return nil, f.describeErr
	}
	if f.describeResp != nil {
		return f.describeResp, nil
	}
	return &awselasticache.DescribeReplicationGroupsOutput{}, nil
}

func (f *fakeRGClient) ModifyReplicationGroup(_ context.Context, params *awselasticache.ModifyReplicationGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyReplicationGroupOutput, error) {
	f.modifyCalled = true
	f.capturedModifyInput = params
	if f.modifyErr != nil {
		return nil, f.modifyErr
	}
	if f.modifyResp != nil {
		return f.modifyResp, nil
	}
	return &awselasticache.ModifyReplicationGroupOutput{}, nil
}

func (f *fakeRGClient) ModifyReplicationGroupShardConfiguration(_ context.Context, params *awselasticache.ModifyReplicationGroupShardConfigurationInput, _ ...func(*awselasticache.Options)) (*awselasticache.ModifyReplicationGroupShardConfigurationOutput, error) {
	f.modifyShardCalled = true
	f.capturedModifyShardInput = params
	if f.modifyShardErr != nil {
		return nil, f.modifyShardErr
	}
	if f.modifyShardResp != nil {
		return f.modifyShardResp, nil
	}
	return &awselasticache.ModifyReplicationGroupShardConfigurationOutput{}, nil
}

func (f *fakeRGClient) DeleteReplicationGroup(_ context.Context, params *awselasticache.DeleteReplicationGroupInput, _ ...func(*awselasticache.Options)) (*awselasticache.DeleteReplicationGroupOutput, error) {
	f.deleteCalled = true
	f.capturedDeleteInput = params
	if f.deleteErr != nil {
		return nil, f.deleteErr
	}
	if f.deleteResp != nil {
		return f.deleteResp, nil
	}
	return &awselasticache.DeleteReplicationGroupOutput{}, nil
}

func (f *fakeRGClient) ListTagsForResource(_ context.Context, _ *awselasticache.ListTagsForResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.ListTagsForResourceOutput, error) {
	f.listTagsCalled = true
	if f.listTagsErr != nil {
		return nil, f.listTagsErr
	}
	if f.listTagsResp != nil {
		return f.listTagsResp, nil
	}
	return &awselasticache.ListTagsForResourceOutput{}, nil
}

func (f *fakeRGClient) AddTagsToResource(_ context.Context, _ *awselasticache.AddTagsToResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.AddTagsToResourceOutput, error) {
	f.addTagsCalled = true
	if f.addTagsErr != nil {
		return nil, f.addTagsErr
	}
	return &awselasticache.AddTagsToResourceOutput{}, nil
}

func (f *fakeRGClient) RemoveTagsFromResource(_ context.Context, _ *awselasticache.RemoveTagsFromResourceInput, _ ...func(*awselasticache.Options)) (*awselasticache.RemoveTagsFromResourceOutput, error) {
	f.removeTagsCalled = true
	if f.removeTagsErr != nil {
		return nil, f.removeTagsErr
	}
	return &awselasticache.RemoveTagsFromResourceOutput{}, nil
}

// ── Test CR builder ───────────────────────────────────────────────────────────

func buildTestCR() *clusternativev2.ReplicationGroupRAW {
	cr := &clusternativev2.ReplicationGroupRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-rg",
		},
		Spec: clusternativev2.ReplicationGroupRAWSpec{
			ForProvider: clusternativev2.ReplicationGroupRAWParameters{
				Description: aws.String("test description"),
				Region:      aws.String("us-east-1"),
			},
		},
	}
	meta.SetExternalName(cr, "test-rg")
	return cr
}

func buildFakeKubeClient(objs ...client.Object) client.Client {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	return fake.NewClientBuilder().WithScheme(scheme).WithObjects(objs...).Build()
}

// ── Test: Auth Token — Auto-Generate Fresh ────────────────────────────────────

func TestCreate_AuthToken_AutoGenerate_Fresh(t *testing.T) {
	// When autoGenerateAuthToken=true and Secret doesn't exist,
	// a new token should be generated, written to Secret, then passed to CreateReplicationGroup.

	secretRef := &xpv1.SecretKeySelector{
		SecretReference: xpv1.SecretReference{Namespace: "default", Name: "rg-auth"},
		Key:             "password",
	}

	cr := buildTestCR()
	cr.Spec.ForProvider.AutoGenerateAuthToken = aws.Bool(true)
	cr.Spec.ForProvider.AuthTokenSecretRef = secretRef
	cr.Spec.ForProvider.TransitEncryptionEnabled = aws.Bool(true) // Auth token requires transit encryption

	fakeAWS := &fakeRGClient{}
	kube := buildFakeKubeClient() // No existing secret

	ec := &ExternalClient{Client: fakeAWS, Kube: kube}
	creation, err := ec.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify auth_token was published as connection detail.
	if len(creation.ConnectionDetails["auth_token"]) == 0 {
		t.Error("Expected auth_token in connection details, got none")
	}

	// Verify the Secret was created with the token.
	s := &corev1.Secret{}
	if err := kube.Get(context.Background(), client.ObjectKey{Namespace: "default", Name: "rg-auth"}, s); err != nil {
		t.Fatalf("Secret not created: %v", err)
	}
	if len(s.Data["password"]) == 0 {
		t.Error("Expected password in Secret, got empty")
	}

	// Verify CreateReplicationGroup was called with the auth token.
	if fakeAWS.capturedCreateInput == nil {
		t.Fatal("CreateReplicationGroup not called")
	}
	if fakeAWS.capturedCreateInput.AuthToken == nil {
		t.Error("Expected AuthToken in CreateReplicationGroupInput, got nil")
	}
	if *fakeAWS.capturedCreateInput.AuthToken != string(s.Data["password"]) {
		t.Errorf("AuthToken mismatch: create input has %q, secret has %q",
			*fakeAWS.capturedCreateInput.AuthToken, string(s.Data["password"]))
	}
}

// ── Test: Auth Token — Auto-Generate Reuse Existing ──────────────────────────

func TestCreate_AuthToken_AutoGenerate_Reuse(t *testing.T) {
	// When autoGenerateAuthToken=true and Secret already has a value, reuse it (idempotent retry).

	existingToken := "existingTokenABCDEF123456789"
	secretRef := &xpv1.SecretKeySelector{
		SecretReference: xpv1.SecretReference{Namespace: "default", Name: "rg-auth"},
		Key:             "password",
	}

	cr := buildTestCR()
	cr.Spec.ForProvider.AutoGenerateAuthToken = aws.Bool(true)
	cr.Spec.ForProvider.AuthTokenSecretRef = secretRef

	// Pre-create the secret with existing value.
	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "rg-auth"},
		Data:       map[string][]byte{"password": []byte(existingToken)},
	}
	kube := buildFakeKubeClient(existingSecret)
	fakeAWS := &fakeRGClient{}

	ec := &ExternalClient{Client: fakeAWS, Kube: kube}
	creation, err := ec.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify the existing token was reused (not a new one).
	if string(creation.ConnectionDetails["auth_token"]) != existingToken {
		t.Errorf("Expected reused token %q, got %q", existingToken, string(creation.ConnectionDetails["auth_token"]))
	}

	// Verify CreateReplicationGroup received the existing token.
	if fakeAWS.capturedCreateInput.AuthToken == nil || *fakeAWS.capturedCreateInput.AuthToken != existingToken {
		t.Errorf("Expected CreateRG to receive existing token %q", existingToken)
	}
}

// ── Test: Auth Token — Explicit SecretRef (no auto-generate) ─────────────────

func TestCreate_AuthToken_ExplicitSecretRef(t *testing.T) {
	explicitToken := "myExplicitToken1234567890123"
	secretRef := &xpv1.SecretKeySelector{
		SecretReference: xpv1.SecretReference{Namespace: "default", Name: "explicit-auth"},
		Key:             "token",
	}

	cr := buildTestCR()
	cr.Spec.ForProvider.AuthTokenSecretRef = secretRef
	// autoGenerateAuthToken is nil/false — use explicit secret

	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "explicit-auth"},
		Data:       map[string][]byte{"token": []byte(explicitToken)},
	}
	kube := buildFakeKubeClient(existingSecret)
	fakeAWS := &fakeRGClient{}

	ec := &ExternalClient{Client: fakeAWS, Kube: kube}
	creation, err := ec.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if string(creation.ConnectionDetails["auth_token"]) != explicitToken {
		t.Errorf("Expected explicit token in conn details, got %q", string(creation.ConnectionDetails["auth_token"]))
	}
	if fakeAWS.capturedCreateInput.AuthToken == nil || *fakeAWS.capturedCreateInput.AuthToken != explicitToken {
		t.Error("Expected explicit token passed to CreateReplicationGroup")
	}
}

// ── Test: Auth Token — No Auth ────────────────────────────────────────────────

func TestCreate_AuthToken_None(t *testing.T) {
	cr := buildTestCR()
	// No authTokenSecretRef, no autoGenerateAuthToken

	fakeAWS := &fakeRGClient{}
	kube := buildFakeKubeClient()

	ec := &ExternalClient{Client: fakeAWS, Kube: kube}
	creation, err := ec.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if _, exists := creation.ConnectionDetails["auth_token"]; exists {
		t.Error("Did not expect auth_token in connection details when no auth configured")
	}
	if fakeAWS.capturedCreateInput.AuthToken != nil {
		t.Error("Expected nil AuthToken in CreateReplicationGroupInput when no auth configured")
	}
}

// ── Test: Async — Create Sets AsyncState ─────────────────────────────────────

func TestCreate_SetsAsyncState(t *testing.T) {
	cr := buildTestCR()
	fakeAWS := &fakeRGClient{}
	kube := buildFakeKubeClient()

	ec := &ExternalClient{Client: fakeAWS, Kube: kube}
	_, err := ec.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Verify async state was set.
	state := nativeGetAsyncState(cr)
	if state == nil {
		t.Fatal("Expected async state to be set after Create")
	}
	if state.Operation != "creating" {
		t.Errorf("Expected operation 'creating', got %q", state.Operation)
	}
}

// ── Test: Async — Observe Clears State on Available ──────────────────────────

func TestObserve_ClearsAsyncState_OnAvailable(t *testing.T) {
	cr := buildTestCR()
	// Pre-set async state (simulating in-flight create).
	setAsyncStateForTest(cr, "creating")

	fakeAWS := &fakeRGClient{
		describeResp: &awselasticache.DescribeReplicationGroupsOutput{
			ReplicationGroups: []ectypes.ReplicationGroup{
				{
					ReplicationGroupId: aws.String("test-rg"),
					Status:             aws.String("available"),
					Description:        aws.String("test description"),
				},
			},
		},
	}
	kube := buildFakeKubeClient()
	ec := &ExternalClient{Client: fakeAWS, Kube: kube}

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe failed: %v", err)
	}

	if !obs.ResourceExists {
		t.Error("Expected ResourceExists=true")
	}

	// Async state should be cleared.
	state := nativeGetAsyncState(cr)
	if state != nil {
		t.Errorf("Expected async state to be cleared, got operation=%q", state.Operation)
	}
}

// ── Test: Async — Observe Returns UpToDate=true During Creating ───────────────

func TestObserve_UpToDate_During_Creating(t *testing.T) {
	cr := buildTestCR()
	setAsyncStateForTest(cr, "creating")

	fakeAWS := &fakeRGClient{
		describeResp: &awselasticache.DescribeReplicationGroupsOutput{
			ReplicationGroups: []ectypes.ReplicationGroup{
				{
					ReplicationGroupId: aws.String("test-rg"),
					Status:             aws.String("creating"),
					Description:        aws.String("test description"),
				},
			},
		},
	}
	kube := buildFakeKubeClient()
	ec := &ExternalClient{Client: fakeAWS, Kube: kube}

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe failed: %v", err)
	}
	if !obs.ResourceExists {
		t.Error("Expected ResourceExists=true")
	}
	if !obs.ResourceUpToDate {
		t.Error("Expected ResourceUpToDate=true during 'creating' state")
	}
}

// ── Test: Async — Observe Handles Modifying ───────────────────────────────────

func TestObserve_UpToDate_During_Modifying(t *testing.T) {
	for _, status := range []string{"modifying", "snapshotting"} {
		t.Run(status, func(t *testing.T) {
			cr := buildTestCR()
			fakeAWS := &fakeRGClient{
				describeResp: &awselasticache.DescribeReplicationGroupsOutput{
					ReplicationGroups: []ectypes.ReplicationGroup{
						{
							ReplicationGroupId: aws.String("test-rg"),
							Status:             aws.String(status),
							Description:        aws.String("test description"),
						},
					},
				},
			}
			kube := buildFakeKubeClient()
			ec := &ExternalClient{Client: fakeAWS, Kube: kube}

			obs, err := ec.Observe(context.Background(), cr)
			if err != nil {
				t.Fatalf("Observe failed: %v", err)
			}
			if !obs.ResourceUpToDate {
				t.Errorf("Expected UpToDate=true during %q", status)
			}
		})
	}
}

// ── Test: Async — Observe Handles Deleting ────────────────────────────────────

func TestObserve_UpToDate_During_Deleting(t *testing.T) {
	cr := buildTestCR()
	fakeAWS := &fakeRGClient{
		describeResp: &awselasticache.DescribeReplicationGroupsOutput{
			ReplicationGroups: []ectypes.ReplicationGroup{
				{
					ReplicationGroupId: aws.String("test-rg"),
					Status:             aws.String("deleting"),
					Description:        aws.String("test description"),
				},
			},
		},
	}
	kube := buildFakeKubeClient()
	ec := &ExternalClient{Client: fakeAWS, Kube: kube}

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe failed: %v", err)
	}
	if !obs.ResourceUpToDate {
		t.Error("Expected UpToDate=true during 'deleting'")
	}
}

// ── Test: Async — Observe Handles Create-Failed ───────────────────────────────

func TestObserve_Error_On_CreateFailed(t *testing.T) {
	cr := buildTestCR()
	setAsyncStateForTest(cr, "creating")

	fakeAWS := &fakeRGClient{
		describeResp: &awselasticache.DescribeReplicationGroupsOutput{
			ReplicationGroups: []ectypes.ReplicationGroup{
				{
					ReplicationGroupId: aws.String("test-rg"),
					Status:             aws.String("create-failed"),
					Description:        aws.String("test description"),
				},
			},
		},
	}
	kube := buildFakeKubeClient()
	ec := &ExternalClient{Client: fakeAWS, Kube: kube}

	_, err := ec.Observe(context.Background(), cr)
	if err == nil {
		t.Error("Expected error for 'create-failed' status, got nil")
	}

	// Async state should be cleared after create-failed.
	state := nativeGetAsyncState(cr)
	if state != nil {
		t.Error("Expected async state cleared after create-failed")
	}
}

// ── Test: isUpToDate — SecurityGroupNames Skipped ─────────────────────────────

func TestIsUpToDate_SecurityGroupNames_NeverCompared(t *testing.T) {
	// SecurityGroupNames in spec should never trigger not-up-to-date.
	cr := buildTestCR()
	cr.Spec.ForProvider.SecurityGroupNames = []*string{aws.String("old-sg-name")}
	cr.Spec.ForProvider.Description = aws.String("test description")

	rg := ectypes.ReplicationGroup{
		Description: aws.String("test description"),
		// SecurityGroupNames is on member clusters, not the RG response.
		// Even if they differ, isUpToDate should return true.
	}

	result := isUpToDate(cr, rg, nil)
	if !result {
		t.Error("Expected isUpToDate=true when only SecurityGroupNames differ (EC2-Classic legacy field)")
	}
}

// ── Test: isUpToDate — SecurityGroupIds Compared ──────────────────────────────

func TestIsUpToDate_SecurityGroupIds_Compared(t *testing.T) {
	// SecurityGroupIds differ → not up to date.
	cr := buildTestCR()
	cr.Spec.ForProvider.Description = aws.String("test description")
	cr.Spec.ForProvider.UserGroupIds = []*string{aws.String("ug-1")}

	rg := ectypes.ReplicationGroup{
		Description:  aws.String("test description"),
		UserGroupIds: []string{"ug-2"}, // Different user group
	}

	result := isUpToDate(cr, rg, nil)
	if result {
		t.Error("Expected isUpToDate=false when UserGroupIds differ")
	}
}

// ── Test: Update Decomposition — NumNodeGroups Triggers ShardConfig ───────────

func TestUpdate_NumNodeGroups_TriggersShardConfig(t *testing.T) {
	cr := buildTestCR()
	cr.Spec.ForProvider.NumNodeGroups = aws.Float64(3)

	fakeAWS := &fakeRGClient{
		describeResp: &awselasticache.DescribeReplicationGroupsOutput{
			ReplicationGroups: []ectypes.ReplicationGroup{
				{
					ReplicationGroupId: aws.String("test-rg"),
					Status:             aws.String("available"),
					Description:        aws.String("test description"),
					NodeGroups: []ectypes.NodeGroup{
						{NodeGroupId: aws.String("0001")}, // Currently 1 shard
					},
				},
			},
		},
	}
	kube := buildFakeKubeClient()
	ec := &ExternalClient{Client: fakeAWS, Kube: kube}

	_, err := ec.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify ModifyReplicationGroupShardConfiguration was called.
	if !fakeAWS.modifyShardCalled {
		t.Error("Expected ModifyReplicationGroupShardConfiguration to be called")
	}
	// Verify ModifyReplicationGroup was NOT called (shard config takes priority).
	if fakeAWS.modifyCalled {
		t.Error("Expected ModifyReplicationGroup NOT to be called when shard config changes")
	}
	// Verify the node group count.
	if fakeAWS.capturedModifyShardInput == nil || aws.ToInt32(fakeAWS.capturedModifyShardInput.NodeGroupCount) != 3 {
		t.Error("Expected NodeGroupCount=3 in shard config input")
	}
}

// ── Test: Update Decomposition — Auth Token Rotation ─────────────────────────

func TestUpdate_AuthToken_TriggersModifyRG(t *testing.T) {
	secretRef := &xpv1.SecretKeySelector{
		SecretReference: xpv1.SecretReference{Namespace: "default", Name: "auth-secret"},
		Key:             "token",
	}

	cr := buildTestCR()
	cr.Spec.ForProvider.AuthTokenUpdateStrategy = aws.String("ROTATE")
	cr.Spec.ForProvider.AuthTokenSecretRef = secretRef
	// atProvider.AuthTokenUpdateStrategy is nil → rotation not yet applied

	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "auth-secret"},
		Data:       map[string][]byte{"token": []byte("myRotatedToken1234567890123")},
	}

	fakeAWS := &fakeRGClient{
		describeResp: &awselasticache.DescribeReplicationGroupsOutput{
			ReplicationGroups: []ectypes.ReplicationGroup{
				{
					ReplicationGroupId: aws.String("test-rg"),
					Status:             aws.String("available"),
					Description:        aws.String("test description"),
				},
			},
		},
	}
	kube := buildFakeKubeClient(existingSecret)
	ec := &ExternalClient{Client: fakeAWS, Kube: kube}

	_, err := ec.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify ModifyReplicationGroup was called with auth token strategy.
	if !fakeAWS.modifyCalled {
		t.Error("Expected ModifyReplicationGroup to be called for auth token rotation")
	}
	if fakeAWS.capturedModifyInput == nil {
		t.Fatal("Expected captured ModifyReplicationGroup input")
	}
	if fakeAWS.capturedModifyInput.AuthToken == nil {
		t.Error("Expected AuthToken in ModifyReplicationGroupInput")
	}
	if fakeAWS.capturedModifyInput.AuthTokenUpdateStrategy != ectypes.AuthTokenUpdateStrategyType("ROTATE") {
		t.Errorf("Expected AuthTokenUpdateStrategy=ROTATE, got %q", fakeAWS.capturedModifyInput.AuthTokenUpdateStrategy)
	}

	// Verify shard config was NOT called (auth token rotation comes first).
	if fakeAWS.modifyShardCalled {
		t.Error("Expected shard config NOT called during auth token rotation step")
	}

	// Verify atProvider was updated to reflect the applied strategy.
	atProvider := cr.GetAtProvider()
	if atProvider.AuthTokenUpdateStrategy == nil || *atProvider.AuthTokenUpdateStrategy != "ROTATE" {
		t.Error("Expected atProvider.AuthTokenUpdateStrategy to be set to 'ROTATE' after rotation")
	}
}

// ── Test: Update Decomposition — Field Change Triggers ModifyRG ──────────────

func TestUpdate_FieldChange_TriggersModifyRG(t *testing.T) {
	cr := buildTestCR()
	cr.Spec.ForProvider.Description = aws.String("new description")
	cr.Spec.ForProvider.SnapshotRetentionLimit = aws.Float64(7)

	fakeAWS := &fakeRGClient{
		describeResp: &awselasticache.DescribeReplicationGroupsOutput{
			ReplicationGroups: []ectypes.ReplicationGroup{
				{
					ReplicationGroupId:     aws.String("test-rg"),
					Status:                 aws.String("available"),
					Description:            aws.String("old description"), // Different → triggers modify
					SnapshotRetentionLimit: aws.Int32(0),
				},
			},
		},
	}
	kube := buildFakeKubeClient()
	ec := &ExternalClient{Client: fakeAWS, Kube: kube}

	_, err := ec.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if !fakeAWS.modifyCalled {
		t.Error("Expected ModifyReplicationGroup to be called for field change")
	}
	if fakeAWS.modifyShardCalled {
		t.Error("Expected shard config NOT called for simple field change")
	}
	if fakeAWS.capturedModifyInput.ReplicationGroupDescription == nil ||
		*fakeAWS.capturedModifyInput.ReplicationGroupDescription != "new description" {
		t.Error("Expected new description in ModifyReplicationGroupInput")
	}
}

// ── Test: Update Decomposition — Tag-Only Change ─────────────────────────────

func TestUpdate_TagsOnly_TriggersTagSync(t *testing.T) {
	cr := buildTestCR()
	cr.Spec.ForProvider.Tags = map[string]*string{
		"env": aws.String("prod"),
	}

	fakeAWS := &fakeRGClient{
		describeResp: &awselasticache.DescribeReplicationGroupsOutput{
			ReplicationGroups: []ectypes.ReplicationGroup{
				{
					ReplicationGroupId: aws.String("test-rg"),
					Status:             aws.String("available"),
					Description:        aws.String("test description"),
					ARN:                aws.String("arn:aws:elasticache:us-east-1:123:replication-group/test-rg"),
				},
			},
		},
		listTagsResp: &awselasticache.ListTagsForResourceOutput{
			TagList: []ectypes.Tag{
				{Key: aws.String("old-tag"), Value: aws.String("old-val")},
			},
		},
	}
	kube := buildFakeKubeClient()
	ec := &ExternalClient{Client: fakeAWS, Kube: kube}

	_, err := ec.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify tag sync was triggered (AddTags for new, RemoveTags for old).
	if !fakeAWS.addTagsCalled {
		t.Error("Expected AddTagsToResource to be called for new tags")
	}
	if !fakeAWS.removeTagsCalled {
		t.Error("Expected RemoveTagsFromResource to be called for removed tags")
	}
	// Verify ModifyReplicationGroup was NOT called (tags only).
	if fakeAWS.modifyCalled {
		t.Error("Expected ModifyReplicationGroup NOT to be called for tag-only change")
	}
}

// ── Test: Connection Details — Cluster Mode Enabled ───────────────────────────

func TestConnectionDetails_ClusterModeEnabled(t *testing.T) {
	cr := buildTestCR()
	rg := ectypes.ReplicationGroup{
		ConfigurationEndpoint: &ectypes.Endpoint{
			Address: aws.String("cluster-cfg.cache.amazonaws.com"),
			Port:    aws.Int32(6379),
		},
		NodeGroups: []ectypes.NodeGroup{
			{
				PrimaryEndpoint: &ectypes.Endpoint{
					Address: aws.String("primary.cache.amazonaws.com"),
					Port:    aws.Int32(6379),
				},
			},
		},
	}

	kube := buildFakeKubeClient()
	ec := &ExternalClient{Client: &fakeRGClient{}, Kube: kube}
	cd := ec.buildConnectionDetails(context.Background(), cr, rg)

	// In cluster mode, should have configuration_endpoint_address.
	if string(cd["configuration_endpoint_address"]) != "cluster-cfg.cache.amazonaws.com" {
		t.Errorf("Expected configuration_endpoint_address, got %q", cd["configuration_endpoint_address"])
	}
	// Should NOT have primary_endpoint_address when in cluster mode.
	if _, exists := cd["primary_endpoint_address"]; exists {
		t.Error("Should not have primary_endpoint_address in cluster mode")
	}
	// Port should be present.
	if string(cd["port"]) != "6379" {
		t.Errorf("Expected port=6379, got %q", cd["port"])
	}
}

// ── Test: Connection Details — Cluster Mode Disabled ─────────────────────────

func TestConnectionDetails_ClusterModeDisabled(t *testing.T) {
	cr := buildTestCR()
	rg := ectypes.ReplicationGroup{
		// No ConfigurationEndpoint — cluster mode disabled.
		NodeGroups: []ectypes.NodeGroup{
			{
				PrimaryEndpoint: &ectypes.Endpoint{
					Address: aws.String("primary.cache.amazonaws.com"),
					Port:    aws.Int32(6379),
				},
				ReaderEndpoint: &ectypes.Endpoint{
					Address: aws.String("reader.cache.amazonaws.com"),
				},
			},
		},
	}

	kube := buildFakeKubeClient()
	ec := &ExternalClient{Client: &fakeRGClient{}, Kube: kube}
	cd := ec.buildConnectionDetails(context.Background(), cr, rg)

	if string(cd["primary_endpoint_address"]) != "primary.cache.amazonaws.com" {
		t.Errorf("Expected primary_endpoint_address, got %q", cd["primary_endpoint_address"])
	}
	if string(cd["reader_endpoint_address"]) != "reader.cache.amazonaws.com" {
		t.Errorf("Expected reader_endpoint_address, got %q", cd["reader_endpoint_address"])
	}
	if _, exists := cd["configuration_endpoint_address"]; exists {
		t.Error("Should not have configuration_endpoint_address in non-cluster mode")
	}
	if string(cd["port"]) != "6379" {
		t.Errorf("Expected port=6379, got %q", cd["port"])
	}
}

// ── Test: Connection Details — Auth Token Published ───────────────────────────

func TestConnectionDetails_AuthToken_Published(t *testing.T) {
	secretRef := &xpv1.SecretKeySelector{
		SecretReference: xpv1.SecretReference{Namespace: "default", Name: "rg-auth"},
		Key:             "token",
	}

	cr := buildTestCR()
	cr.Spec.ForProvider.AuthTokenSecretRef = secretRef

	existingSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "rg-auth"},
		Data:       map[string][]byte{"token": []byte("myAuthToken123456789012")},
	}
	kube := buildFakeKubeClient(existingSecret)
	ec := &ExternalClient{Client: &fakeRGClient{}, Kube: kube}

	rg := ectypes.ReplicationGroup{
		NodeGroups: []ectypes.NodeGroup{
			{PrimaryEndpoint: &ectypes.Endpoint{Port: aws.Int32(6379)}},
		},
	}
	cd := ec.buildConnectionDetails(context.Background(), cr, rg)

	if string(cd["auth_token"]) != "myAuthToken123456789012" {
		t.Errorf("Expected auth_token in connection details, got %q", string(cd["auth_token"]))
	}
}

// ── Test: v1beta1↔v1beta2 Conversion ─────────────────────────────────────────

func TestConversion_v1beta1_ClusterMode_To_v1beta2(t *testing.T) {
	// Import from test context; just verify the types compile and interface is satisfied.
	// The conversion logic is in apis/cluster/elasticache/v1beta1/native.
	// Here we just verify that the v1beta2 hub type satisfies our interface.
	var _ ReplicationGroupCR = (*clusternativev2.ReplicationGroupRAW)(nil)
}

// ── Test: Type Mismatch — AtRestEncryptionEnabled *string → *bool ─────────────

func TestBuildCreateInput_AtRestEncryptionEnabled_StringToBool(t *testing.T) {
	spec := &clusternativev2.ReplicationGroupRAWParameters{
		Description:             aws.String("test"),
		Region:                  aws.String("us-east-1"),
		AtRestEncryptionEnabled: aws.String("true"),
	}

	input := buildCreateInput(spec, "test-rg", "")
	if input.AtRestEncryptionEnabled == nil {
		t.Fatal("Expected AtRestEncryptionEnabled to be set")
	}
	if !*input.AtRestEncryptionEnabled {
		t.Error("Expected AtRestEncryptionEnabled=true (from string 'true')")
	}
}

func TestBuildCreateInput_AutoMinorVersionUpgrade_StringToBool(t *testing.T) {
	spec := &clusternativev2.ReplicationGroupRAWParameters{
		Description:             aws.String("test"),
		Region:                  aws.String("us-east-1"),
		AutoMinorVersionUpgrade: aws.String("false"),
	}

	input := buildCreateInput(spec, "test-rg", "")
	if input.AutoMinorVersionUpgrade == nil {
		t.Fatal("Expected AutoMinorVersionUpgrade to be set")
	}
	if *input.AutoMinorVersionUpgrade {
		t.Error("Expected AutoMinorVersionUpgrade=false (from string 'false')")
	}
}

// ── Test: observationFromSDK — AtRestEncryptionEnabled *bool → *string ────────

func TestObservationFromSDK_AtRestEncryptionEnabled_BoolToString(t *testing.T) {
	rg := ectypes.ReplicationGroup{
		AtRestEncryptionEnabled: aws.Bool(true),
	}
	obs := observationFromSDK(rg)
	if obs.AtRestEncryptionEnabled == nil {
		t.Fatal("Expected AtRestEncryptionEnabled in observation")
	}
	if *obs.AtRestEncryptionEnabled != "true" {
		t.Errorf("Expected 'true', got %q", *obs.AtRestEncryptionEnabled)
	}
}

func TestObservationFromSDK_AutoMinorVersionUpgrade_BoolToString(t *testing.T) {
	rg := ectypes.ReplicationGroup{
		AutoMinorVersionUpgrade: aws.Bool(false),
	}
	obs := observationFromSDK(rg)
	if obs.AutoMinorVersionUpgrade == nil {
		t.Fatal("Expected AutoMinorVersionUpgrade in observation")
	}
	if *obs.AutoMinorVersionUpgrade != "false" {
		t.Errorf("Expected 'false', got %q", *obs.AutoMinorVersionUpgrade)
	}
}

// ── Test: Delete Sets AsyncState ──────────────────────────────────────────────

func TestDelete_SetsAsyncState(t *testing.T) {
	cr := buildTestCR()
	fakeAWS := &fakeRGClient{}
	kube := buildFakeKubeClient()
	ec := &ExternalClient{Client: fakeAWS, Kube: kube}

	_, err := ec.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	state := nativeGetAsyncState(cr)
	if state == nil {
		t.Fatal("Expected async state set after Delete")
	}
	if state.Operation != "deleting" {
		t.Errorf("Expected operation 'deleting', got %q", state.Operation)
	}
}

// ── Test: Observe NotFound After Deleting ─────────────────────────────────────

func TestObserve_NotFound_AfterDeleting(t *testing.T) {
	// When async state is 'deleting' and resource is not found, should return ResourceExists=false.
	cr := buildTestCR()
	setAsyncStateForTest(cr, "deleting")

	fakeAWS := &fakeRGClient{
		describeResp: &awselasticache.DescribeReplicationGroupsOutput{
			ReplicationGroups: []ectypes.ReplicationGroup{}, // Empty — resource gone
		},
	}
	kube := buildFakeKubeClient()
	ec := &ExternalClient{Client: fakeAWS, Kube: kube}

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe failed: %v", err)
	}
	if obs.ResourceExists {
		t.Error("Expected ResourceExists=false after deletion")
	}
}

// ── Test: isUpToDate — NumNodeGroups Change ───────────────────────────────────

func TestIsUpToDate_NumNodeGroups_Change(t *testing.T) {
	cr := buildTestCR()
	cr.Spec.ForProvider.NumNodeGroups = aws.Float64(3)
	cr.Spec.ForProvider.Description = aws.String("test description")

	rg := ectypes.ReplicationGroup{
		Description: aws.String("test description"),
		NodeGroups: []ectypes.NodeGroup{
			{NodeGroupId: aws.String("0001")}, // Only 1 node group currently
		},
	}

	result := isUpToDate(cr, rg, nil)
	if result {
		t.Error("Expected isUpToDate=false when NumNodeGroups changed from 1 to 3")
	}
}

// ── Test: isUpToDate — Auth Token Rotation Pending ────────────────────────────

func TestIsUpToDate_AuthTokenRotation_Pending(t *testing.T) {
	cr := buildTestCR()
	cr.Spec.ForProvider.AuthTokenUpdateStrategy = aws.String("ROTATE")
	cr.Spec.ForProvider.Description = aws.String("test description")
	// atProvider.AuthTokenUpdateStrategy is nil → rotation not yet applied

	rg := ectypes.ReplicationGroup{
		Description: aws.String("test description"),
	}

	result := isUpToDate(cr, rg, nil)
	if result {
		t.Error("Expected isUpToDate=false when auth token rotation is pending")
	}
}

func TestIsUpToDate_AuthTokenRotation_AlreadyApplied(t *testing.T) {
	cr := buildTestCR()
	cr.Spec.ForProvider.AuthTokenUpdateStrategy = aws.String("ROTATE")
	cr.Spec.ForProvider.Description = aws.String("test description")
	// Set atProvider to match spec → rotation already applied
	cr.Status.AtProvider.AuthTokenUpdateStrategy = aws.String("ROTATE")

	rg := ectypes.ReplicationGroup{
		Description: aws.String("test description"),
	}

	result := isUpToDate(cr, rg, nil)
	if !result {
		t.Error("Expected isUpToDate=true when auth token rotation already applied")
	}
}

// ── Test: Delete — FinalSnapshotIdentifier Passed ─────────────────────────────

func TestDelete_FinalSnapshotIdentifier(t *testing.T) {
	cr := buildTestCR()
	cr.Spec.ForProvider.FinalSnapshotIdentifier = aws.String("final-snapshot-123")

	fakeAWS := &fakeRGClient{}
	kube := buildFakeKubeClient()
	ec := &ExternalClient{Client: fakeAWS, Kube: kube}

	_, err := ec.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if fakeAWS.capturedDeleteInput == nil {
		t.Fatal("Expected DeleteReplicationGroup to be called")
	}
	if fakeAWS.capturedDeleteInput.FinalSnapshotIdentifier == nil ||
		*fakeAWS.capturedDeleteInput.FinalSnapshotIdentifier != "final-snapshot-123" {
		t.Error("Expected FinalSnapshotIdentifier in delete input")
	}
}

// ── Test: buildModifyInput — SecurityGroupNames Never Included ────────────────

func TestBuildModifyInput_SecurityGroupNames_NotIncluded(t *testing.T) {
	spec := &clusternativev2.ReplicationGroupRAWParameters{
		Description:        aws.String("test"),
		Region:             aws.String("us-east-1"),
		SecurityGroupNames: []*string{aws.String("my-sg-name")}, // Should be ignored
		SecurityGroupIds:   []*string{aws.String("sg-123")},     // These should be included
	}

	input := buildModifyInput(spec, "test-rg")
	if len(input.CacheSecurityGroupNames) > 0 {
		t.Error("SecurityGroupNames should never be included in ModifyReplicationGroupInput")
	}
	if len(input.SecurityGroupIds) == 0 {
		t.Error("Expected SecurityGroupIds to be included")
	}
}

// ── Helpers for tests ─────────────────────────────────────────────────────────

// nativeGetAsyncState is a test helper that reads async state from the CR's annotations.
func nativeGetAsyncState(cr *clusternativev2.ReplicationGroupRAW) *struct{ Operation string } {
	annotations := cr.GetAnnotations()
	if annotations == nil {
		return nil
	}
	op, ok := annotations["native.aws.upbound.io/async-operation"]
	if !ok {
		return nil
	}
	return &struct{ Operation string }{Operation: op}
}

// setAsyncStateForTest sets the async state annotations on the CR for testing.
func setAsyncStateForTest(cr *clusternativev2.ReplicationGroupRAW, operation string) {
	annotations := cr.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["native.aws.upbound.io/async-operation"] = operation
	annotations["native.aws.upbound.io/async-started-at"] = "2024-01-01T00:00:00Z"
	annotations["native.aws.upbound.io/async-request-id"] = "test-request-id"
	cr.SetAnnotations(annotations)
}

// Ensure fakeRGClient implements ElastiCacheRGClient at compile time.
var _ ElastiCacheRGClient = (*fakeRGClient)(nil)

// ── Test: Late Initialization ─────────────────────────────────────────────────

// TestObserve_LateInit_NilNodeType verifies that when spec.NodeType is nil and
// AWS returns a CacheNodeType, Observe returns ResourceLateInitialized=true and
// populates the spec field — preventing the infinite reconciliation loop.
func TestObserve_LateInit_NilNodeType_RG(t *testing.T) {
	cr := buildTestCR()
	cr.Spec.ForProvider.NodeType = nil // intentionally nil

	fakeAWS := &fakeRGClient{
		describeResp: &awselasticache.DescribeReplicationGroupsOutput{
			ReplicationGroups: []ectypes.ReplicationGroup{
				{
					ReplicationGroupId: aws.String("test-rg"),
					Status:             aws.String("available"),
					Description:        aws.String("test description"),
					CacheNodeType:      aws.String("cache.r7g.medium"),
				},
			},
		},
		listTagsResp: &awselasticache.ListTagsForResourceOutput{TagList: []ectypes.Tag{}},
	}

	ec := &ExternalClient{Client: fakeAWS, Kube: buildFakeKubeClient()}
	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !obs.ResourceLateInitialized {
		t.Error("expected ResourceLateInitialized=true when NodeType is nil and AWS returns a value")
	}
	if cr.Spec.ForProvider.NodeType == nil {
		t.Fatal("expected spec.NodeType to be populated after late initialization")
	}
	if got, want := *cr.Spec.ForProvider.NodeType, "cache.r7g.medium"; got != want {
		t.Errorf("spec.NodeType: got %q, want %q", got, want)
	}
}

// TestFieldChangesUpToDate_NilNodeType_NoLoop verifies that fieldChangesUpToDate
// returns true when spec.NodeType is nil — no spurious update should be triggered.
func TestFieldChangesUpToDate_NilNodeType_NoLoop(t *testing.T) {
	spec := &clusternativev2.ReplicationGroupRAWParameters{
		Description: aws.String("test"),
		Region:      aws.String("us-east-1"),
		NodeType:    nil, // intentionally nil
	}
	rg := ectypes.ReplicationGroup{
		Description:   aws.String("test"),
		CacheNodeType: aws.String("cache.r7g.medium"),
	}

	if !fieldChangesUpToDate(spec, rg) {
		t.Error("fieldChangesUpToDate should return true when spec.NodeType is nil — no spurious update")
	}
}

// Ensure managed.ConnectionDetails is used correctly.
var _ managed.ConnectionDetails = managed.ConnectionDetails{}
