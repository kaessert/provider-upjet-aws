// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package generate_test

import (
	"os"
	"strings"
	"testing"
)

// TestResolverDirectivesExcludeNativePackages verifies that the upjet resolver
// generator directives in generate.go exclude native/ sub-packages.
//
// Native types don't implement resource.Terraformed — passing them to the
// resolver causes a panic. The resolver must receive an explicit package list
// that filters out any .../native paths.
//
// See: .agents/specs/terraform-removal-migration.md §0.15
func TestResolverDirectivesExcludeNativePackages(t *testing.T) {
	content, err := os.ReadFile("generate.go")
	if err != nil {
		t.Fatalf("failed to read generate.go: %v", err)
	}
	s := string(content)

	// The raw "..." patterns must NOT be used for the resolver — they would
	// recurse into native/ sub-packages when those packages exist.
	forbidden := []string{
		"-p ../apis/cluster/...",
		"-p ../apis/namespaced/...",
	}
	for _, f := range forbidden {
		if strings.Contains(s, f) {
			t.Errorf("generate.go contains %q which includes native/ sub-packages; "+
				"must filter /native out before passing to the upjet resolver", f)
		}
	}

	// Must contain logic that excludes /native packages from the resolver.
	if !strings.Contains(s, "/native") {
		t.Error("generate.go must contain logic to exclude /native packages from resolver invocation")
	}
}

// TestNativeCrosstoolsStepDocumented verifies that generate.go contains a
// comment documenting the separate crossplane-tools step for native types.
//
// Reference resolution for native types is NOT part of make generate — it
// must be run explicitly via crossplane-tools against native/ sub-packages.
func TestNativeCrosstoolsStepDocumented(t *testing.T) {
	content, err := os.ReadFile("generate.go")
	if err != nil {
		t.Fatalf("failed to read generate.go: %v", err)
	}
	s := string(content)

	// Either "crossplane-tools" or "crossplane-gen" as the tool reference.
	if !strings.Contains(s, "crossplane-tools") && !strings.Contains(s, "crossplane-gen") {
		t.Error("generate.go must document that crossplane-tools is run separately for native types")
	}
}
