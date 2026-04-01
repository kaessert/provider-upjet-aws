// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package statemachine_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssfn "github.com/aws/aws-sdk-go-v2/service/sfn"
	sfntypes "github.com/aws/aws-sdk-go-v2/service/sfn/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sfn/v1beta2/native"
	"github.com/upbound/provider-aws/v2/internal/controller/sfn/statemachine"
)

// mockSFNClient is a test double for the SFNClient interface.
type mockSFNClient struct {
	describeStateMachineFn func(ctx context.Context, params *awssfn.DescribeStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error)
	createStateMachineFn   func(ctx context.Context, params *awssfn.CreateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.CreateStateMachineOutput, error)
	updateStateMachineFn   func(ctx context.Context, params *awssfn.UpdateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.UpdateStateMachineOutput, error)
	deleteStateMachineFn   func(ctx context.Context, params *awssfn.DeleteStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DeleteStateMachineOutput, error)
	listTagsForResourceFn  func(ctx context.Context, params *awssfn.ListTagsForResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.ListTagsForResourceOutput, error)
	tagResourceFn          func(ctx context.Context, params *awssfn.TagResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.TagResourceOutput, error)
	untagResourceFn        func(ctx context.Context, params *awssfn.UntagResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.UntagResourceOutput, error)
}

func (m *mockSFNClient) DescribeStateMachine(ctx context.Context, params *awssfn.DescribeStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
	if m.describeStateMachineFn != nil {
		return m.describeStateMachineFn(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockSFNClient) CreateStateMachine(ctx context.Context, params *awssfn.CreateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.CreateStateMachineOutput, error) {
	if m.createStateMachineFn != nil {
		return m.createStateMachineFn(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockSFNClient) UpdateStateMachine(ctx context.Context, params *awssfn.UpdateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.UpdateStateMachineOutput, error) {
	if m.updateStateMachineFn != nil {
		return m.updateStateMachineFn(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockSFNClient) DeleteStateMachine(ctx context.Context, params *awssfn.DeleteStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DeleteStateMachineOutput, error) {
	if m.deleteStateMachineFn != nil {
		return m.deleteStateMachineFn(ctx, params, optFns...)
	}
	return nil, nil
}

func (m *mockSFNClient) ListTagsForResource(ctx context.Context, params *awssfn.ListTagsForResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.ListTagsForResourceOutput, error) {
	if m.listTagsForResourceFn != nil {
		return m.listTagsForResourceFn(ctx, params, optFns...)
	}
	return &awssfn.ListTagsForResourceOutput{}, nil
}

func (m *mockSFNClient) TagResource(ctx context.Context, params *awssfn.TagResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.TagResourceOutput, error) {
	if m.tagResourceFn != nil {
		return m.tagResourceFn(ctx, params, optFns...)
	}
	return &awssfn.TagResourceOutput{}, nil
}

func (m *mockSFNClient) UntagResource(ctx context.Context, params *awssfn.UntagResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.UntagResourceOutput, error) {
	if m.untagResourceFn != nil {
		return m.untagResourceFn(ctx, params, optFns...)
	}
	return &awssfn.UntagResourceOutput{}, nil
}

// testCR creates a StateMachineRAW for testing.
func testCR(extName string, spec clusternative.StateMachineRAWParameters, obs clusternative.StateMachineRAWObservation) *clusternative.StateMachineRAW {
	cr := &clusternative.StateMachineRAW{
		ObjectMeta: metav1.ObjectMeta{
			Name: extName,
			Annotations: map[string]string{
				"crossplane.io/external-name": extName,
			},
		},
		Spec: clusternative.StateMachineRAWSpec{
			ForProvider: spec,
		},
		Status: clusternative.StateMachineRAWStatus{
			AtProvider: obs,
		},
	}
	return cr
}

const (
	testARN        = "arn:aws:states:us-east-1:123456789012:stateMachine:my-state-machine"
	testName       = "my-state-machine"
	testDefinition = `{"Comment":"Test","StartAt":"Hello","States":{"Hello":{"Type":"Pass","End":true}}}`
	testRoleARN    = "arn:aws:iam::123456789012:role/my-role"
)

// ── Observe tests ──────────────────────────────────────────────────────────────

func TestObserve_EmptyExternalName_ReturnsNotExists(t *testing.T) {
	ec := &statemachine.ExternalClient{
		Client: &mockSFNClient{},
	}
	// CR with no external name annotation
	cr := &clusternative.StateMachineRAW{
		ObjectMeta: metav1.ObjectMeta{Name: "test"},
	}
	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Errorf("Observe() ResourceExists = true, want false when external name is empty")
	}
}

func TestObserve_NoARN_ReturnsNotExists(t *testing.T) {
	ec := &statemachine.ExternalClient{
		Client: &mockSFNClient{},
	}
	// CR with external name set but no ARN in status (and not an ARN itself)
	cr := testCR(testName, clusternative.StateMachineRAWParameters{}, clusternative.StateMachineRAWObservation{})
	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Errorf("Observe() ResourceExists = true, want false when no ARN available")
	}
}

func TestObserve_StateMachineDoesNotExist_ReturnsNotExists(t *testing.T) {
	notFoundErr := &sfntypes.StateMachineDoesNotExist{
		Message: aws.String("state machine does not exist"),
	}
	mock := &mockSFNClient{
		describeStateMachineFn: func(ctx context.Context, params *awssfn.DescribeStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
			return nil, notFoundErr
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}
	cr := testCR(testName, clusternative.StateMachineRAWParameters{}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})
	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() unexpected error: %v", err)
	}
	if obs.ResourceExists {
		t.Errorf("Observe() ResourceExists = true, want false for StateMachineDoesNotExist")
	}
}

func TestObserve_DescribeError_ReturnsError(t *testing.T) {
	describeErr := errors.New("internal server error")
	mock := &mockSFNClient{
		describeStateMachineFn: func(ctx context.Context, params *awssfn.DescribeStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
			return nil, describeErr
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}
	cr := testCR(testName, clusternative.StateMachineRAWParameters{}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})
	_, err := ec.Observe(context.Background(), cr)
	if err == nil {
		t.Errorf("Observe() expected error, got nil")
	}
}

func TestObserve_UpToDate_ReturnsUpToDate(t *testing.T) {
	now := time.Now()
	mock := &mockSFNClient{
		describeStateMachineFn: func(ctx context.Context, params *awssfn.DescribeStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
			return &awssfn.DescribeStateMachineOutput{
				StateMachineArn: aws.String(testARN),
				Name:            aws.String(testName),
				Definition:      aws.String(testDefinition),
				RoleArn:         aws.String(testRoleARN),
				Type:            sfntypes.StateMachineTypeStandard,
				Status:          sfntypes.StateMachineStatusActive,
				CreationDate:    &now,
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	smType := "STANDARD"
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		Type:       &smType,
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Errorf("Observe() ResourceExists = false, want true")
	}
	if !obs.ResourceUpToDate {
		t.Errorf("Observe() ResourceUpToDate = false, want true")
	}
	// Status ARN should be set
	if cr.Status.AtProvider.Arn == nil || *cr.Status.AtProvider.Arn != testARN {
		t.Errorf("Observe() status.atProvider.arn = %v, want %q", cr.Status.AtProvider.Arn, testARN)
	}
}

func TestObserve_DefinitionDrift_ReturnsNotUpToDate(t *testing.T) {
	now := time.Now()
	awsDefinition := `{"Comment":"AWS version","StartAt":"Hello","States":{"Hello":{"Type":"Pass","End":true}}}`
	specDefinition := `{"Comment":"Spec version","StartAt":"Hello","States":{"Hello":{"Type":"Pass","End":true}}}`

	mock := &mockSFNClient{
		describeStateMachineFn: func(ctx context.Context, params *awssfn.DescribeStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
			return &awssfn.DescribeStateMachineOutput{
				StateMachineArn: aws.String(testARN),
				Name:            aws.String(testName),
				Definition:      aws.String(awsDefinition),
				RoleArn:         aws.String(testRoleARN),
				Type:            sfntypes.StateMachineTypeStandard,
				Status:          sfntypes.StateMachineStatusActive,
				CreationDate:    &now,
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(specDefinition),
		RoleArn:    aws.String(testRoleARN),
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Errorf("Observe() ResourceExists = false, want true")
	}
	if obs.ResourceUpToDate {
		t.Errorf("Observe() ResourceUpToDate = true, want false (definition drift)")
	}
}

func TestObserve_RoleArnDrift_ReturnsNotUpToDate(t *testing.T) {
	now := time.Now()
	awsRoleARN := "arn:aws:iam::123456789012:role/old-role"
	specRoleARN := "arn:aws:iam::123456789012:role/new-role"

	mock := &mockSFNClient{
		describeStateMachineFn: func(ctx context.Context, params *awssfn.DescribeStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
			return &awssfn.DescribeStateMachineOutput{
				StateMachineArn: aws.String(testARN),
				Name:            aws.String(testName),
				Definition:      aws.String(testDefinition),
				RoleArn:         aws.String(awsRoleARN),
				Type:            sfntypes.StateMachineTypeStandard,
				Status:          sfntypes.StateMachineStatusActive,
				CreationDate:    &now,
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(specRoleARN),
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Errorf("Observe() ResourceExists = false, want true")
	}
	if obs.ResourceUpToDate {
		t.Errorf("Observe() ResourceUpToDate = true, want false (roleArn drift)")
	}
}

// ── Create tests ──────────────────────────────────────────────────────────────

func TestCreate_Success_SetsExternalNameAndARN(t *testing.T) {
	creationDate := time.Now()
	mock := &mockSFNClient{
		createStateMachineFn: func(ctx context.Context, params *awssfn.CreateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.CreateStateMachineOutput, error) {
			if aws.ToString(params.Name) != testName {
				t.Errorf("Create: Name = %q, want %q", aws.ToString(params.Name), testName)
			}
			if aws.ToString(params.Definition) != testDefinition {
				t.Errorf("Create: Definition mismatch")
			}
			return &awssfn.CreateStateMachineOutput{
				StateMachineArn: aws.String(testARN),
				CreationDate:    &creationDate,
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
	}, clusternative.StateMachineRAWObservation{})

	result, err := ec.Create(context.Background(), cr)
	if err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}
	_ = result

	// External name should now be the FULL ARN (not just the short name).
	// This is required so that getARN() can reconstruct the ARN from the external
	// name annotation on subsequent Observe calls, even if status.atProvider.arn
	// was reset by the reconciler's annotation update step.
	externalName := cr.Annotations["crossplane.io/external-name"]
	if externalName != testARN {
		t.Errorf("Create() external name = %q, want full ARN %q", externalName, testARN)
	}

	// ARN should also be stored in status (best-effort; may be reset by reconciler).
	if cr.Status.AtProvider.Arn == nil || *cr.Status.AtProvider.Arn != testARN {
		t.Errorf("Create() status.atProvider.arn = %v, want %q", cr.Status.AtProvider.Arn, testARN)
	}
}

// TestCreate_ARNExternalName_BuildsCorrectName verifies that when a previous
// Create already set the external name to the full ARN, subsequent Create
// calls (due to the idempotent-create loop) still send just the short name
// to the AWS API (not the full ARN as the state machine name).
func TestCreate_ARNExternalName_BuildsCorrectName(t *testing.T) {
	var capturedName string
	mock := &mockSFNClient{
		createStateMachineFn: func(ctx context.Context, params *awssfn.CreateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.CreateStateMachineOutput, error) {
			capturedName = aws.ToString(params.Name)
			return &awssfn.CreateStateMachineOutput{
				StateMachineArn: aws.String(testARN),
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	// Simulate a second Create call: external-name is already set to the full ARN.
	cr := testCR(testARN, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
	}, clusternative.StateMachineRAWObservation{})

	if _, err := ec.Create(context.Background(), cr); err != nil {
		t.Fatalf("Create() unexpected error: %v", err)
	}

	// The AWS API should receive the short name, not the full ARN.
	if capturedName != testName {
		t.Errorf("Create() API Name = %q, want short name %q", capturedName, testName)
	}
}

func TestCreate_Error_ReturnsError(t *testing.T) {
	createErr := errors.New("create failed")
	mock := &mockSFNClient{
		createStateMachineFn: func(ctx context.Context, params *awssfn.CreateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.CreateStateMachineOutput, error) {
			return nil, createErr
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
	}, clusternative.StateMachineRAWObservation{})

	_, err := ec.Create(context.Background(), cr)
	if err == nil {
		t.Errorf("Create() expected error, got nil")
	}
}

// ── Update tests ──────────────────────────────────────────────────────────────

func TestUpdate_Success_CallsUpdateStateMachine(t *testing.T) {
	var updateCalled bool
	newRoleARN := "arn:aws:iam::123456789012:role/new-role"

	mock := &mockSFNClient{
		updateStateMachineFn: func(ctx context.Context, params *awssfn.UpdateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.UpdateStateMachineOutput, error) {
			updateCalled = true
			if aws.ToString(params.StateMachineArn) != testARN {
				t.Errorf("Update: StateMachineArn = %q, want %q", aws.ToString(params.StateMachineArn), testARN)
			}
			if aws.ToString(params.RoleArn) != newRoleARN {
				t.Errorf("Update: RoleArn = %q, want %q", aws.ToString(params.RoleArn), newRoleARN)
			}
			return &awssfn.UpdateStateMachineOutput{
				UpdateDate: aws.Time(time.Now()),
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(newRoleARN),
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	_, err := ec.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("Update() unexpected error: %v", err)
	}
	if !updateCalled {
		t.Errorf("Update() did not call UpdateStateMachine")
	}
}

func TestUpdate_Error_ReturnsError(t *testing.T) {
	updateErr := errors.New("update failed")
	mock := &mockSFNClient{
		updateStateMachineFn: func(ctx context.Context, params *awssfn.UpdateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.UpdateStateMachineOutput, error) {
			return nil, updateErr
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	_, err := ec.Update(context.Background(), cr)
	if err == nil {
		t.Errorf("Update() expected error, got nil")
	}
}

// ── Delete tests ──────────────────────────────────────────────────────────────

func TestDelete_Success_CallsDeleteStateMachine(t *testing.T) {
	var deleteCalled bool
	mock := &mockSFNClient{
		deleteStateMachineFn: func(ctx context.Context, params *awssfn.DeleteStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DeleteStateMachineOutput, error) {
			deleteCalled = true
			if aws.ToString(params.StateMachineArn) != testARN {
				t.Errorf("Delete: StateMachineArn = %q, want %q", aws.ToString(params.StateMachineArn), testARN)
			}
			return &awssfn.DeleteStateMachineOutput{}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}
	cr := testCR(testName, clusternative.StateMachineRAWParameters{}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	_, err := ec.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("Delete() unexpected error: %v", err)
	}
	if !deleteCalled {
		t.Errorf("Delete() did not call DeleteStateMachine")
	}
}

func TestDelete_AlreadyGone_ReturnsNil(t *testing.T) {
	notFoundErr := &sfntypes.StateMachineDoesNotExist{
		Message: aws.String("state machine does not exist"),
	}
	mock := &mockSFNClient{
		deleteStateMachineFn: func(ctx context.Context, params *awssfn.DeleteStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DeleteStateMachineOutput, error) {
			return nil, notFoundErr
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}
	cr := testCR(testName, clusternative.StateMachineRAWParameters{}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	_, err := ec.Delete(context.Background(), cr)
	if err != nil {
		t.Errorf("Delete() expected nil error for already-gone resource, got: %v", err)
	}
}

func TestDelete_NoARN_ReturnsNil(t *testing.T) {
	var deleteCalled bool
	mock := &mockSFNClient{
		deleteStateMachineFn: func(ctx context.Context, params *awssfn.DeleteStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DeleteStateMachineOutput, error) {
			deleteCalled = true
			return &awssfn.DeleteStateMachineOutput{}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}
	// CR with no ARN in status
	cr := testCR(testName, clusternative.StateMachineRAWParameters{}, clusternative.StateMachineRAWObservation{})

	_, err := ec.Delete(context.Background(), cr)
	if err != nil {
		t.Fatalf("Delete() with no ARN: unexpected error: %v", err)
	}
	if deleteCalled {
		t.Errorf("Delete() with no ARN: should not call DeleteStateMachine")
	}
}

func TestDelete_Error_ReturnsError(t *testing.T) {
	deleteErr := errors.New("delete failed")
	mock := &mockSFNClient{
		deleteStateMachineFn: func(ctx context.Context, params *awssfn.DeleteStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DeleteStateMachineOutput, error) {
			return nil, deleteErr
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}
	cr := testCR(testName, clusternative.StateMachineRAWParameters{}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	_, err := ec.Delete(context.Background(), cr)
	if err == nil {
		t.Errorf("Delete() expected error, got nil")
	}
}

// ── Tag reconciliation tests ──────────────────────────────────────────────────

func TestUpdate_TagsAdded_CallsTagResource(t *testing.T) {
	var tagCalled bool
	mock := &mockSFNClient{
		updateStateMachineFn: func(ctx context.Context, params *awssfn.UpdateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.UpdateStateMachineOutput, error) {
			return &awssfn.UpdateStateMachineOutput{UpdateDate: aws.Time(time.Now())}, nil
		},
		listTagsForResourceFn: func(ctx context.Context, params *awssfn.ListTagsForResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.ListTagsForResourceOutput, error) {
			// Return no existing tags
			return &awssfn.ListTagsForResourceOutput{Tags: []sfntypes.Tag{}}, nil
		},
		tagResourceFn: func(ctx context.Context, params *awssfn.TagResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.TagResourceOutput, error) {
			tagCalled = true
			return &awssfn.TagResourceOutput{}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	env := "prod"
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		Tags:       map[string]*string{"env": &env},
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	_, err := ec.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("Update() with tags: unexpected error: %v", err)
	}
	if !tagCalled {
		t.Errorf("Update() did not call TagResource when new tags need adding")
	}
}

func TestUpdate_TagsRemoved_CallsUntagResource(t *testing.T) {
	var untagCalled bool
	mock := &mockSFNClient{
		updateStateMachineFn: func(ctx context.Context, params *awssfn.UpdateStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.UpdateStateMachineOutput, error) {
			return &awssfn.UpdateStateMachineOutput{UpdateDate: aws.Time(time.Now())}, nil
		},
		listTagsForResourceFn: func(ctx context.Context, params *awssfn.ListTagsForResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.ListTagsForResourceOutput, error) {
			// Return an existing tag that's not in spec
			return &awssfn.ListTagsForResourceOutput{
				Tags: []sfntypes.Tag{{Key: aws.String("old-tag"), Value: aws.String("old-value")}},
			}, nil
		},
		untagResourceFn: func(ctx context.Context, params *awssfn.UntagResourceInput, optFns ...func(*awssfn.Options)) (*awssfn.UntagResourceOutput, error) {
			untagCalled = true
			return &awssfn.UntagResourceOutput{}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	// Spec has no tags → old tags should be removed
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		Tags:       map[string]*string{},
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	_, err := ec.Update(context.Background(), cr)
	if err != nil {
		t.Fatalf("Update() remove tags: unexpected error: %v", err)
	}
	if !untagCalled {
		t.Errorf("Update() did not call UntagResource when tags need removing")
	}
}
