// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpfake "github.com/crossplane/crossplane-runtime/v2/pkg/resource/fake"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// mockSvcClient is a trivial service client used to verify that the connector
// passes the right value to the clientFactory and newExternal functions.
type mockSvcClient struct {
	region string
}

// mockExternalClient is a minimal TypedExternalClient implementation for tests.
type mockExternalClient struct{}

func (m *mockExternalClient) Observe(_ context.Context, _ *xpfake.Managed) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, nil
}
func (m *mockExternalClient) Create(_ context.Context, _ *xpfake.Managed) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, nil
}
func (m *mockExternalClient) Update(_ context.Context, _ *xpfake.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}
func (m *mockExternalClient) Delete(_ context.Context, _ *xpfake.Managed) (managed.ExternalDelete, error) {
	return managed.ExternalDelete{}, nil
}
func (m *mockExternalClient) Disconnect(_ context.Context) error { return nil }

// ---------------------------------------------------------------------------
// TypedConnector tests
// ---------------------------------------------------------------------------

// TestTypedConnector_Connect_Success verifies the happy path: configFn succeeds,
// clientFactory is called with the resolved config, newExternal returns a client.
func TestTypedConnector_Connect_Success(t *testing.T) {
	wantRegion := "us-east-1"
	wantCfg := &aws.Config{Region: wantRegion}

	clientFactoryCalled := false
	newExternalCalled := false

	connector := &TypedConnector[*xpfake.Managed, *mockSvcClient]{
		kube: nil,
		clientFactory: func(cfg aws.Config) *mockSvcClient {
			clientFactoryCalled = true
			if cfg.Region != wantRegion {
				t.Errorf("clientFactory: region = %q, want %q", cfg.Region, wantRegion)
			}
			return &mockSvcClient{region: cfg.Region}
		},
		newExternal: func(svc *mockSvcClient, kube client.Client) managed.TypedExternalClient[*xpfake.Managed] {
			newExternalCalled = true
			if svc == nil {
				t.Error("newExternal called with nil svc client")
			}
			return &mockExternalClient{}
		},
		configFn: func(_ context.Context, _ client.Client, _ resource.Managed) (*aws.Config, error) {
			return wantCfg, nil
		},
	}

	cr := newFakeManaged()
	got, err := connector.Connect(context.Background(), cr)
	if err != nil {
		t.Fatalf("unexpected error from Connect: %v", err)
	}
	if got == nil {
		t.Error("Connect returned nil client")
	}
	if !clientFactoryCalled {
		t.Error("clientFactory was not called")
	}
	if !newExternalCalled {
		t.Error("newExternal was not called")
	}
}

// TestTypedConnector_Connect_PropagatesConfigError verifies that when configFn
// returns an error, Connect returns a wrapped error and nil client.
func TestTypedConnector_Connect_PropagatesConfigError(t *testing.T) {
	wantErr := errors.New("credentials error")

	connector := &TypedConnector[*xpfake.Managed, *mockSvcClient]{
		kube: nil,
		clientFactory: func(_ aws.Config) *mockSvcClient {
			t.Error("clientFactory should not be called when configFn fails")
			return nil
		},
		newExternal: func(_ *mockSvcClient, _ client.Client) managed.TypedExternalClient[*xpfake.Managed] {
			t.Error("newExternal should not be called when configFn fails")
			return nil
		},
		configFn: func(_ context.Context, _ client.Client, _ resource.Managed) (*aws.Config, error) {
			return nil, wantErr
		},
	}

	cr := newFakeManaged()
	got, err := connector.Connect(context.Background(), cr)
	if err == nil {
		t.Fatal("expected error from Connect when configFn fails, got nil")
	}
	if got != nil {
		t.Errorf("expected nil client when Connect fails, got %v", got)
	}
}

// TestTypedConnector_Connect_PassesKubeToNewExternal verifies that the kube
// client stored in the connector is forwarded to the newExternal factory.
func TestTypedConnector_Connect_PassesKubeToNewExternal(t *testing.T) {
	// Use a non-nil sentinel value to verify it is passed through.
	// We cannot use a real client.Client in a unit test, so we just check
	// that the same pointer value arrives in newExternal.
	//
	// client.Client is an interface; we need a concrete type to compare.
	// Use nil here — we verify the connector passes whatever kube is stored.
	kubePassedThrough := false

	connector := &TypedConnector[*xpfake.Managed, *mockSvcClient]{
		kube: nil, // sentinel
		clientFactory: func(_ aws.Config) *mockSvcClient {
			return &mockSvcClient{}
		},
		newExternal: func(_ *mockSvcClient, kube client.Client) managed.TypedExternalClient[*xpfake.Managed] {
			// kube should equal whatever was stored in connector.kube (nil in this test).
			if kube != nil {
				t.Errorf("expected kube=nil to be forwarded, got %v", kube)
			}
			kubePassedThrough = true
			return &mockExternalClient{}
		},
		configFn: func(_ context.Context, _ client.Client, _ resource.Managed) (*aws.Config, error) {
			return &aws.Config{}, nil
		},
	}

	cr := newFakeManaged()
	if _, err := connector.Connect(context.Background(), cr); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !kubePassedThrough {
		t.Error("kube client was not passed to newExternal")
	}
}

// TestNewTypedConnector_DefaultsToGetAWSConfigWithTracking verifies that
// NewTypedConnector sets a non-nil configFn (the default resolver).
func TestNewTypedConnector_DefaultsToGetAWSConfigWithTracking(t *testing.T) {
	connector := NewTypedConnector[*xpfake.Managed, *mockSvcClient](
		nil,
		func(_ aws.Config) *mockSvcClient { return &mockSvcClient{} },
		func(_ *mockSvcClient, _ client.Client) managed.TypedExternalClient[*xpfake.Managed] {
			return &mockExternalClient{}
		},
	)

	if connector.configFn == nil {
		t.Error("expected configFn to be set by NewTypedConnector, got nil")
	}
}
