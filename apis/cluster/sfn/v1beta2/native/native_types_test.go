// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package native contains hand-written native (non-Terraform) type definitions
// for the sfn service.
package native_test

import (
	"os"
	"strings"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	native "github.com/upbound/provider-aws/v2/apis/cluster/sfn/v1beta2/native"
)

// TestStateMachineRAWIsModernManaged verifies that StateMachineRAW satisfies
// resource.ModernManaged (xpv2 interface with typed ProviderConfigReference and
// LocalSecretReference) and does NOT satisfy resource.LegacyManaged (which
// would imply xpv1.ResourceSpec with DeletionPolicy, untyped ProviderConfigReference,
// and SecretReference with namespace).
//
// ModernManaged = Managed + LocalConnectionSecretWriterTo + TypedProviderConfigReferencer
// LegacyManaged = Managed + ConnectionSecretWriterTo + ProviderConfigReferencer + Orphanable
func TestStateMachineRAWIsModernManaged(t *testing.T) {
	var raw resource.Managed = &native.StateMachineRAW{}
	_, isModern := raw.(resource.ModernManaged)
	if !isModern {
		t.Error("StateMachineRAW does not satisfy resource.ModernManaged; " +
			"StateMachineRAWSpec must embed xpv2.ManagedResourceSpec (not xpv1.ResourceSpec)")
	}
}

// TestStateMachineRAWIsNotLegacyManaged verifies that StateMachineRAW does NOT
// satisfy resource.LegacyManaged. Embedding xpv2.ManagedResourceSpec removes
// DeletionPolicy, so the type cannot satisfy the Orphanable interface required
// by LegacyManaged.
func TestStateMachineRAWIsNotLegacyManaged(t *testing.T) {
	var raw resource.Managed = &native.StateMachineRAW{}
	_, isLegacy := raw.(resource.LegacyManaged)
	if isLegacy {
		t.Error("StateMachineRAW must NOT satisfy resource.LegacyManaged; " +
			"StateMachineRAWSpec must embed xpv2.ManagedResourceSpec (not xpv1.ResourceSpec) " +
			"which removes DeletionPolicy/Orphanable from the interface set")
	}
}

// TestResolverFileExists verifies that zz_generated.resolvers.go exists in the
// native package. This file is produced by running angryjet generate-methodsets
// against the native sub-package. Without it, cross-resource references such as
// roleArnRef and kmsKeyIdRef are never resolved.
func TestResolverFileExists(t *testing.T) {
	if _, err := os.Stat("zz_generated.resolvers.go"); os.IsNotExist(err) {
		t.Fatal("zz_generated.resolvers.go does not exist in the native sfn package; " +
			"run: go run github.com/crossplane/crossplane-tools/cmd/angryjet generate-methodsets " +
			"--header-file=./hack/boilerplate.go.txt ./apis/cluster/sfn/v1beta2/native/")
	}
}

// TestGeneratedResolverHasRoleArnResolution verifies that the generated resolver
// includes logic to resolve the roleArnRef reference, which populates roleArn.
// This is critical for state machine creation — without it, roleArn would be nil
// and the AWS API would reject the CreateStateMachine call.
func TestGeneratedResolverHasRoleArnResolution(t *testing.T) {
	content, err := os.ReadFile("zz_generated.resolvers.go")
	if os.IsNotExist(err) {
		t.Skip("zz_generated.resolvers.go not yet generated")
	}
	if err != nil {
		t.Fatalf("failed to read zz_generated.resolvers.go: %v", err)
	}
	s := string(content)
	if !strings.Contains(s, "RoleArnRef") {
		t.Error("zz_generated.resolvers.go does not contain RoleArnRef resolution; " +
			"roleArn will never be populated from a roleArnRef cross-resource reference")
	}
	if !strings.Contains(s, "NewAPINamespacedResolver") {
		t.Error("zz_generated.resolvers.go does not use NewAPINamespacedResolver; " +
			"cluster ModernManaged types must use the namespaced resolver variant")
	}
}
