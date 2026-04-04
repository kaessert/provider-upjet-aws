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
	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
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

	// testSMTypeStandard is the default state machine type in AWS.
	testSMTypeStandard = "STANDARD"
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

	smType := testSMTypeStandard
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

// TestObserve_NilLoggingConfig_AcceptsAWSDefaults verifies that when spec omits
// loggingConfiguration, Observe returns ResourceUpToDate=true even when AWS
// returns default logging configuration (level=OFF). This prevents infinite
// update loops.
func TestObserve_NilLoggingConfig_AcceptsAWSDefaults(t *testing.T) {
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
				// AWS always returns default logging config
				LoggingConfiguration: &sfntypes.LoggingConfiguration{
					Level:                sfntypes.LogLevelOff,
					IncludeExecutionData: false,
				},
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	// Spec does NOT set loggingConfiguration
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		// LoggingConfiguration is nil
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() unexpected error: %v", err)
	}
	if !obs.ResourceUpToDate {
		t.Errorf("Observe() ResourceUpToDate = false, want true (nil logging spec should accept AWS defaults)")
	}
}

// TestObserve_NilEncryptionConfig_AcceptsAWSDefaults verifies that when spec
// omits encryptionConfiguration, Observe returns ResourceUpToDate=true even
// when AWS returns default encryption configuration (type=AWS_OWNED_KEY).
func TestObserve_NilEncryptionConfig_AcceptsAWSDefaults(t *testing.T) {
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
				// AWS always returns default encryption config
				EncryptionConfiguration: &sfntypes.EncryptionConfiguration{
					Type: sfntypes.EncryptionTypeAwsOwnedKey,
				},
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	// Spec does NOT set encryptionConfiguration
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		// EncryptionConfiguration is nil
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() unexpected error: %v", err)
	}
	if !obs.ResourceUpToDate {
		t.Errorf("Observe() ResourceUpToDate = false, want true (nil encryption spec should accept AWS defaults)")
	}
}

// TestObserve_NilTracingConfig_AcceptsAWSDefaults verifies that when spec omits
// tracingConfiguration, Observe returns ResourceUpToDate=true when AWS returns
// default tracing (enabled=false).
func TestObserve_NilTracingConfig_AcceptsAWSDefaults(t *testing.T) {
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
				// AWS returns default tracing (disabled)
				TracingConfiguration: &sfntypes.TracingConfiguration{Enabled: false},
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	// Spec does NOT set tracingConfiguration
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		// TracingConfiguration is nil
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() unexpected error: %v", err)
	}
	if !obs.ResourceUpToDate {
		t.Errorf("Observe() ResourceUpToDate = false, want true (nil tracing spec should accept AWS defaults)")
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

// ── Late initialization tests ─────────────────────────────────────────────────

// TestObserve_LateInit_TypeDefaultedByAWS verifies that when spec.type is nil
// and AWS returns STANDARD (the default), Observe performs late initialization:
// it sets spec.forProvider.type = "STANDARD" and returns ResourceLateInitialized=true.
func TestObserve_LateInit_TypeDefaultedByAWS(t *testing.T) {
	now := time.Now()
	mock := &mockSFNClient{
		describeStateMachineFn: func(ctx context.Context, params *awssfn.DescribeStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
			return &awssfn.DescribeStateMachineOutput{
				StateMachineArn: aws.String(testARN),
				Name:            aws.String(testName),
				Definition:      aws.String(testDefinition),
				RoleArn:         aws.String(testRoleARN),
				Type:            sfntypes.StateMachineTypeStandard, // AWS defaults to STANDARD
				Status:          sfntypes.StateMachineStatusActive,
				CreationDate:    &now,
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	// spec.Type is nil — not yet set by user, AWS will default it
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		// Type is intentionally nil
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() late init: unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Errorf("Observe() ResourceExists = false, want true")
	}
	if !obs.ResourceLateInitialized {
		t.Errorf("Observe() ResourceLateInitialized = false, want true when type was nil and AWS returned STANDARD")
	}
	if cr.Spec.ForProvider.Type == nil {
		t.Errorf("Observe() late init: spec.forProvider.type is still nil, want %q", testSMTypeStandard)
	} else if *cr.Spec.ForProvider.Type != testSMTypeStandard {
		t.Errorf("Observe() late init: spec.forProvider.type = %q, want %q", *cr.Spec.ForProvider.Type, testSMTypeStandard)
	}
}

// TestObserve_LateInit_TypeAlreadySet verifies that when spec.type is already set,
// Observe does NOT late-initialize it (ResourceLateInitialized=false).
func TestObserve_LateInit_TypeAlreadySet(t *testing.T) {
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

	specType := testSMTypeStandard
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		Type:       &specType, // already set
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() late init already set: unexpected error: %v", err)
	}
	if obs.ResourceLateInitialized {
		t.Errorf("Observe() ResourceLateInitialized = true, want false when type is already set")
	}
}

// ── Condition tests ────────────────────────────────────────────────────────────

// TestObserve_DeletingState_SetsUnavailable verifies that when a state machine
// is in the DELETING state, Observe sets the Unavailable condition (not Available).
func TestObserve_DeletingState_SetsUnavailable(t *testing.T) {
	now := time.Now()
	mock := &mockSFNClient{
		describeStateMachineFn: func(ctx context.Context, params *awssfn.DescribeStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
			return &awssfn.DescribeStateMachineOutput{
				StateMachineArn: aws.String(testARN),
				Name:            aws.String(testName),
				Definition:      aws.String(testDefinition),
				RoleArn:         aws.String(testRoleARN),
				Type:            sfntypes.StateMachineTypeStandard,
				Status:          sfntypes.StateMachineStatusDeleting, // DELETING state
				CreationDate:    &now,
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	smType := testSMTypeStandard
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		Type:       &smType,
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	_, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() DELETING state: unexpected error: %v", err)
	}

	readyCond := cr.GetCondition(xpv1.TypeReady)
	if readyCond.Status != corev1.ConditionFalse {
		t.Errorf("Observe() DELETING state: Ready condition status = %q, want %q (Unavailable)",
			readyCond.Status, corev1.ConditionFalse)
	}
}

// TestObserve_ActiveState_SetsAvailable verifies that when a state machine
// is ACTIVE, Observe sets the Available condition.
func TestObserve_ActiveState_SetsAvailable(t *testing.T) {
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

	smType := testSMTypeStandard
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		Type:       &smType,
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	_, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() ACTIVE state: unexpected error: %v", err)
	}

	readyCond := cr.GetCondition(xpv1.TypeReady)
	if readyCond.Status != corev1.ConditionTrue {
		t.Errorf("Observe() ACTIVE state: Ready condition status = %q, want %q (Available)",
			readyCond.Status, corev1.ConditionTrue)
	}
}

// ── KMSDataKeyReusePeriodSeconds drift detection tests ─────────────────────────

// TestObserve_EncryptionKMSDataKeyReusePeriodSecondsDrift_ReturnsNotUpToDate verifies
// that when spec.kmsDataKeyReusePeriodSeconds differs from AWS state, isUpToDate returns false.
func TestObserve_EncryptionKMSDataKeyReusePeriodSecondsDrift_ReturnsNotUpToDate(t *testing.T) {
	now := time.Now()
	specPeriod := float64(900)
	mock := &mockSFNClient{
		describeStateMachineFn: func(ctx context.Context, params *awssfn.DescribeStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
			awsPeriod := int32(300) // AWS has 300 seconds, spec wants 900
			return &awssfn.DescribeStateMachineOutput{
				StateMachineArn: aws.String(testARN),
				Name:            aws.String(testName),
				Definition:      aws.String(testDefinition),
				RoleArn:         aws.String(testRoleARN),
				Type:            sfntypes.StateMachineTypeStandard,
				Status:          sfntypes.StateMachineStatusActive,
				CreationDate:    &now,
				EncryptionConfiguration: &sfntypes.EncryptionConfiguration{
					Type:                         sfntypes.EncryptionTypeCustomerManagedKmsKey,
					KmsKeyId:                     aws.String("arn:aws:kms:us-east-1:123456789012:key/test-key"),
					KmsDataKeyReusePeriodSeconds: &awsPeriod,
				},
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	kmsKeyID := "arn:aws:kms:us-east-1:123456789012:key/test-key"
	encType := "CUSTOMER_MANAGED_KMS_KEY"
	smType := testSMTypeStandard
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		Type:       &smType,
		EncryptionConfiguration: &clusternative.EncryptionConfigurationRAWParameters{
			Type:                         &encType,
			KMSKeyID:                     &kmsKeyID,
			KMSDataKeyReusePeriodSeconds: &specPeriod, // 900s in spec
		},
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() KMSDataKeyReuse drift: unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Errorf("Observe() ResourceExists = false, want true")
	}
	if obs.ResourceUpToDate {
		t.Errorf("Observe() ResourceUpToDate = true, want false (KMSDataKeyReusePeriodSeconds drift: spec=900, AWS=300)")
	}
}

// TestObserve_EncryptionKMSDataKeyReusePeriodSecondsMatch_ReturnsUpToDate verifies
// that when spec.kmsDataKeyReusePeriodSeconds matches AWS state, isUpToDate returns true.
func TestObserve_EncryptionKMSDataKeyReusePeriodSecondsMatch_ReturnsUpToDate(t *testing.T) {
	now := time.Now()
	specPeriod := float64(300)
	mock := &mockSFNClient{
		describeStateMachineFn: func(ctx context.Context, params *awssfn.DescribeStateMachineInput, optFns ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
			awsPeriod := int32(300) // Both spec and AWS have 300
			return &awssfn.DescribeStateMachineOutput{
				StateMachineArn: aws.String(testARN),
				Name:            aws.String(testName),
				Definition:      aws.String(testDefinition),
				RoleArn:         aws.String(testRoleARN),
				Type:            sfntypes.StateMachineTypeStandard,
				Status:          sfntypes.StateMachineStatusActive,
				CreationDate:    &now,
				EncryptionConfiguration: &sfntypes.EncryptionConfiguration{
					Type:                         sfntypes.EncryptionTypeCustomerManagedKmsKey,
					KmsKeyId:                     aws.String("arn:aws:kms:us-east-1:123456789012:key/test-key"),
					KmsDataKeyReusePeriodSeconds: &awsPeriod,
				},
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	kmsKeyID := "arn:aws:kms:us-east-1:123456789012:key/test-key"
	encType := "CUSTOMER_MANAGED_KMS_KEY"
	smType := testSMTypeStandard
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		Type:       &smType,
		EncryptionConfiguration: &clusternative.EncryptionConfigurationRAWParameters{
			Type:                         &encType,
			KMSKeyID:                     &kmsKeyID,
			KMSDataKeyReusePeriodSeconds: &specPeriod, // same as AWS
		},
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	obs, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() KMSDataKeyReuse match: unexpected error: %v", err)
	}
	if !obs.ResourceExists {
		t.Errorf("Observe() ResourceExists = false, want true")
	}
	if !obs.ResourceUpToDate {
		t.Errorf("Observe() ResourceUpToDate = false, want true (KMSDataKeyReusePeriodSeconds match)")
	}
}

// ── Observation Tags/Region reflection tests ───────────────────────────────────

// TestObserve_ReflectsTagsFromSpec verifies that after a successful Observe call,
// status.atProvider.tags is populated with the tags from spec.forProvider.tags,
// mirroring TF StateMachine behavior (atProvider.tags reflects user-specified tags).
func TestObserve_ReflectsTagsFromSpec(t *testing.T) {
	now := time.Now()
	mock := &mockSFNClient{
		describeStateMachineFn: func(_ context.Context, _ *awssfn.DescribeStateMachineInput, _ ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
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
		listTagsForResourceFn: func(_ context.Context, _ *awssfn.ListTagsForResourceInput, _ ...func(*awssfn.Options)) (*awssfn.ListTagsForResourceOutput, error) {
			return &awssfn.ListTagsForResourceOutput{
				Tags: []sfntypes.Tag{
					{Key: aws.String("env"), Value: aws.String("prod")},
				},
			}, nil
		},
	}
	ec := &statemachine.ExternalClient{Client: mock}

	tagVal := "prod"
	smType := testSMTypeStandard
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		Type:       &smType,
		Tags:       map[string]*string{"env": &tagVal},
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	_, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() Tags reflection: unexpected error: %v", err)
	}

	atProvider := cr.GetAtProvider()
	if atProvider.Tags == nil {
		t.Fatal("Observe() atProvider.tags is nil, want map with 'env' key")
	}
	v, ok := atProvider.Tags["env"]
	if !ok {
		t.Errorf("Observe() atProvider.tags missing 'env' key, got %v", atProvider.Tags)
	} else if v == nil || *v != "prod" {
		t.Errorf("Observe() atProvider.tags[env] = %v, want 'prod'", v)
	}
}

// TestObserve_ReflectsRegionFromSpec verifies that after a successful Observe call,
// status.atProvider.region is populated with the region from spec.forProvider.region,
// mirroring TF StateMachine behavior (atProvider.region reflects the spec region).
func TestObserve_ReflectsRegionFromSpec(t *testing.T) {
	now := time.Now()
	mock := &mockSFNClient{
		describeStateMachineFn: func(_ context.Context, _ *awssfn.DescribeStateMachineInput, _ ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
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

	region := "us-east-1"
	smType := testSMTypeStandard
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		Type:       &smType,
		Region:     &region,
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	_, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() Region reflection: unexpected error: %v", err)
	}

	atProvider := cr.GetAtProvider()
	if atProvider.Region == nil {
		t.Fatal("Observe() atProvider.region is nil, want 'us-east-1'")
	}
	if *atProvider.Region != "us-east-1" {
		t.Errorf("Observe() atProvider.region = %q, want 'us-east-1'", *atProvider.Region)
	}
}

// TestObserve_ReflectsNilTagsAndRegion verifies that nil Tags and nil Region in
// spec do not cause panics and result in nil atProvider.tags/region.
func TestObserve_ReflectsNilTagsAndRegion(t *testing.T) {
	now := time.Now()
	mock := &mockSFNClient{
		describeStateMachineFn: func(_ context.Context, _ *awssfn.DescribeStateMachineInput, _ ...func(*awssfn.Options)) (*awssfn.DescribeStateMachineOutput, error) {
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

	smType := testSMTypeStandard
	cr := testCR(testName, clusternative.StateMachineRAWParameters{
		Definition: aws.String(testDefinition),
		RoleArn:    aws.String(testRoleARN),
		Type:       &smType,
		// No Tags, no Region
	}, clusternative.StateMachineRAWObservation{
		Arn: aws.String(testARN),
	})

	_, err := ec.Observe(context.Background(), cr)
	if err != nil {
		t.Fatalf("Observe() nil Tags/Region: unexpected error: %v", err)
	}

	atProvider := cr.GetAtProvider()
	if atProvider.Tags != nil {
		t.Errorf("Observe() atProvider.tags = %v, want nil when no spec tags", atProvider.Tags)
	}
	if atProvider.Region != nil {
		t.Errorf("Observe() atProvider.region = %v, want nil when no spec region", atProvider.Region)
	}
}
