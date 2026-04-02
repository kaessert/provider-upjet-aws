// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package native contains hand-written native (non-Terraform) type definitions
// for the sfn service (namespaced scope).
package native_test

import (
	"os"
	"strings"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	native "github.com/upbound/provider-aws/v2/apis/namespaced/sfn/v1beta2/native"
)

// TestNamespacedStateMachineRAWIsModernManaged verifies that the namespaced
// StateMachineRAW satisfies resource.ModernManaged (xpv2 interface with typed
// ProviderConfigReference) and does NOT satisfy resource.LegacyManaged.
func TestNamespacedStateMachineRAWIsModernManaged(t *testing.T) {
	var raw resource.Managed = &native.StateMachineRAW{}
	_, isModern := raw.(resource.ModernManaged)
	if !isModern {
		t.Error("namespaced StateMachineRAW does not satisfy resource.ModernManaged; " +
			"StateMachineRAWSpec must embed xpv2.ManagedResourceSpec (not xpv1.ResourceSpec)")
	}
}

// TestNamespacedStateMachineRAWIsNotLegacyManaged verifies that the namespaced
// StateMachineRAW does NOT satisfy resource.LegacyManaged.
func TestNamespacedStateMachineRAWIsNotLegacyManaged(t *testing.T) {
	var raw resource.Managed = &native.StateMachineRAW{}
	_, isLegacy := raw.(resource.LegacyManaged)
	if isLegacy {
		t.Error("namespaced StateMachineRAW must NOT satisfy resource.LegacyManaged; " +
			"StateMachineRAWSpec must embed xpv2.ManagedResourceSpec (not xpv1.ResourceSpec) " +
			"which removes DeletionPolicy/Orphanable from the interface set")
	}
}

// TestNamespacedResolverFileExists verifies that zz_generated.resolvers.go exists
// in the namespaced native sfn package. This file is produced by running
// angryjet generate-methodsets against the namespaced native sub-package.
// Without it, cross-resource references such as roleArnRef and kmsKeyIdRef
// are never resolved in namespaced scope.
func TestNamespacedResolverFileExists(t *testing.T) {
	if _, err := os.Stat("zz_generated.resolvers.go"); os.IsNotExist(err) {
		t.Fatal("zz_generated.resolvers.go does not exist in the namespaced sfn package; " +
			"run: make generate.native")
	}
}

// TestNamespacedResolverUsesNamespacedIAMRole verifies that the generated resolver
// references the NAMESPACED IAM Role type (not the cluster-scoped Role).
// This is required so that roleArnRef resolves correctly in namespaced mode.
func TestNamespacedResolverUsesNamespacedIAMRole(t *testing.T) {
	content, err := os.ReadFile("zz_generated.resolvers.go")
	if os.IsNotExist(err) {
		t.Skip("zz_generated.resolvers.go not yet generated")
	}
	if err != nil {
		t.Fatalf("failed to read zz_generated.resolvers.go: %v", err)
	}
	s := string(content)

	// The resolver must reference namespaced IAM, not cluster IAM.
	if !strings.Contains(s, "namespaced/iam") {
		t.Error("zz_generated.resolvers.go does not reference namespaced IAM; " +
			"roleArnRef will resolve against cluster-scoped Roles instead of namespaced Roles")
	}
	if strings.Contains(s, "cluster/iam") {
		t.Error("zz_generated.resolvers.go references cluster IAM; " +
			"namespaced resolver must use namespaced/iam, not cluster/iam")
	}
}

// TestNamespacedResolverUsesNamespacedKMSKey verifies that the generated resolver
// references the NAMESPACED KMS Key type (not the cluster-scoped Key).
func TestNamespacedResolverUsesNamespacedKMSKey(t *testing.T) {
	content, err := os.ReadFile("zz_generated.resolvers.go")
	if os.IsNotExist(err) {
		t.Skip("zz_generated.resolvers.go not yet generated")
	}
	if err != nil {
		t.Fatalf("failed to read zz_generated.resolvers.go: %v", err)
	}
	s := string(content)

	// The resolver must reference namespaced KMS, not cluster KMS.
	if !strings.Contains(s, "namespaced/kms") {
		t.Error("zz_generated.resolvers.go does not reference namespaced KMS; " +
			"kmsKeyIdRef will resolve against cluster-scoped Keys instead of namespaced Keys")
	}
	if strings.Contains(s, "cluster/kms") {
		t.Error("zz_generated.resolvers.go references cluster KMS; " +
			"namespaced resolver must use namespaced/kms, not cluster/kms")
	}
}

// TestNamespacedParamsHasNoClusterParamsImport verifies at a source-code level
// that the namespaced statemachineraw_types.go defines its own Parameters and
// InitParameters (the spec.forProvider/spec.initProvider fields are NOT typed
// as cluster parameter types). The methods GetForProvider/GetInitProvider may
// still return *clusternative.* types (required by the shared CRUD interface),
// but the struct fields themselves must be local types so that angryjet
// generates correct namespaced resolvers.
func TestNamespacedParamsHasNoClusterParamsImport(t *testing.T) {
	content, err := os.ReadFile("statemachineraw_types.go")
	if err != nil {
		t.Fatalf("failed to read statemachineraw_types.go: %v", err)
	}
	s := string(content)

	// The ForProvider and InitProvider spec fields must be local types, NOT
	// imported cluster parameter types.
	for _, forbidden := range []string{
		"ForProvider clusternative.StateMachineRAWParameters",
		"InitProvider clusternative.StateMachineRAWInitParameters",
	} {
		if strings.Contains(s, forbidden) {
			t.Errorf("statemachineraw_types.go uses cluster parameter type %q as a spec field type; "+
				"namespaced types must define their own Parameters and InitParameters structs "+
				"so angryjet generates resolvers with correct namespaced type references", forbidden)
		}
	}

	// Verify that the local parameter types ARE defined in this file.
	for _, required := range []string{
		"type StateMachineRAWParameters struct",
		"type StateMachineRAWInitParameters struct",
	} {
		if !strings.Contains(s, required) {
			t.Errorf("statemachineraw_types.go is missing %q; "+
				"each scope must define its own Parameters and InitParameters structs", required)
		}
	}
}
