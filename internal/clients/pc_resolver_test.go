// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package clients

import (
	"context"
	"testing"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	xpfake "github.com/crossplane/crossplane-runtime/v2/pkg/resource/fake"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	clusterv1beta1 "github.com/upbound/provider-aws/v2/apis/cluster/v1beta1"
)

// TestTrackClusterScopedModernPCU verifies that trackClusterScopedModernPCU creates a
// cluster-scoped ProviderConfigUsage for a cluster-scoped ModernManaged resource.
//
// Background: the standard ProviderConfigUsageTracker sets namespace from
// mg.GetNamespace(), which is "" for cluster-scoped resources.  Using the namespaced
// PCU type with an empty namespace causes "an empty namespace may not be set during
// creation".  This test guards against that regression.
func TestTrackClusterScopedModernPCU(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clusterv1beta1.SchemeBuilder.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}

	uid := types.UID("abc-123-uid")

	mg := &xpfake.ModernManaged{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-state-machine",
			// No namespace → cluster-scoped resource
			UID: uid,
		},
		TypedProviderConfigReferencer: xpfake.TypedProviderConfigReferencer{
			Ref: &xpv1.ProviderConfigReference{
				Kind: "ClusterProviderConfig",
				Name: "default",
			},
		},
	}

	fc := fake.NewClientBuilder().WithScheme(scheme).Build()

	if err := trackClusterScopedModernPCU(context.Background(), fc, mg); err != nil {
		t.Fatalf("trackClusterScopedModernPCU() unexpected error: %v", err)
	}

	// The PCU should be created with name == string(mg.GetUID()) and no namespace.
	pcu := &clusterv1beta1.ProviderConfigUsage{}
	if err := fc.Get(context.Background(), types.NamespacedName{Name: string(uid)}, pcu); err != nil {
		t.Fatalf("PCU was not created: %v", err)
	}

	if got, want := pcu.GetProviderConfigReference().Name, "default"; got != want {
		t.Errorf("PCU providerConfigRef.Name = %q, want %q", got, want)
	}
	if got := pcu.GetNamespace(); got != "" {
		t.Errorf("PCU namespace = %q, want empty (cluster-scoped)", got)
	}

	// Label for provider name should be set.
	labels := pcu.GetLabels()
	if labels[xpv1.LabelKeyProviderName] != "default" {
		t.Errorf("PCU label %q = %q, want %q", xpv1.LabelKeyProviderName, labels[xpv1.LabelKeyProviderName], "default")
	}
}

// TestTrackClusterScopedModernPCU_NilRef verifies that a nil ProviderConfigReference
// returns an error without panicking.
func TestTrackClusterScopedModernPCU_NilRef(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clusterv1beta1.SchemeBuilder.AddToScheme(scheme); err != nil {
		t.Fatalf("AddToScheme: %v", err)
	}

	mg := &xpfake.ModernManaged{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nil-ref-resource",
			UID:  types.UID("uid-nil"),
		},
		TypedProviderConfigReferencer: xpfake.TypedProviderConfigReferencer{
			Ref: nil, // no ref
		},
	}

	fc := fake.NewClientBuilder().WithScheme(scheme).Build()

	if err := trackClusterScopedModernPCU(context.Background(), fc, mg); err == nil {
		t.Error("trackClusterScopedModernPCU() expected error for nil ref, got nil")
	}
}

// TestResolveProviderConfigModern_ClusterScoped_NoPanicOnEmptyNamespace is a
// regression guard for the "empty namespace" bug.  It verifies that calling
// resolveProviderConfigModern with a cluster-scoped (namespace == "") resource does
// NOT return an "empty namespace may not be set" error.
func TestResolveProviderConfigModern_ClusterScoped_NoPanicOnEmptyNamespace(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := clusterv1beta1.SchemeBuilder.AddToScheme(scheme); err != nil {
		t.Fatalf("clusterv1beta1 AddToScheme: %v", err)
	}

	// We don't register namespacedv1beta1 because we only test the PCU path here;
	// the test will fail at the Get(ClusterProviderConfig) step, not at the PCU step.
	// That's acceptable: we just need to confirm no "empty namespace" error.
	uid := types.UID("cluster-uid-999")
	mg := &xpfake.ModernManaged{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cluster-resource",
			UID:  uid,
			// No namespace
		},
		TypedProviderConfigReferencer: xpfake.TypedProviderConfigReferencer{
			Ref: &xpv1.ProviderConfigReference{
				Kind: "ClusterProviderConfig",
				Name: "default",
			},
		},
	}

	fc := fake.NewClientBuilder().WithScheme(scheme).Build()

	// resolveProviderConfigModern is expected to fail (no PC in the fake store),
	// but it must NOT return an "empty namespace may not be set" error.
	_, err := resolveProviderConfigModern(context.Background(), fc, mg)
	if err == nil {
		t.Fatal("resolveProviderConfigModern() expected an error (PC not found), got nil")
	}
	if isEmptyNamespaceErr(err) {
		t.Errorf("resolveProviderConfigModern() returned 'empty namespace' error for cluster-scoped resource: %v", err)
	}
}

// isEmptyNamespaceErr returns true if the error message contains the "empty namespace"
// string that the API server returns when a namespaced object is created without one.
func isEmptyNamespaceErr(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "empty namespace may not be set") ||
		contains(msg, "empty namespace")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
