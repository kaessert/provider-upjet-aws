// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package generate_test

import (
	"os"
	"strings"
	"testing"

	"github.com/upbound/provider-aws/v2/generate"
)

// TestResolverDirectivesExcludeNativePackages verifies that the upjet resolver
// generator directives in generate.go exclude native/ sub-packages via the
// pkgfilter helper program instead of the old inline bash/grep pipeline.
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
		"-p ./apis/cluster/...",
		"-p ./apis/namespaced/...",
	}
	for _, f := range forbidden {
		if strings.Contains(s, f) {
			t.Errorf("generate.go contains %q which includes native/ sub-packages; "+
				"must filter /native out before passing to the upjet resolver", f)
		}
	}

	// The old inline bash pipeline must be gone.
	oldPatterns := []string{
		"grep -v /native",
		"2>/dev/null",
		"xargs printf",
	}
	for _, p := range oldPatterns {
		if strings.Contains(s, p) {
			t.Errorf("generate.go still contains old fragile pattern %q; "+
				"should be replaced by pkgfilter", p)
		}
	}

	// Must use the pkgfilter helper for package enumeration.
	if !strings.Contains(s, "pkgfilter") {
		t.Error("generate.go must use the pkgfilter helper to enumerate packages for the upjet resolver")
	}

	// Must document the /native exclusion rationale.
	if !strings.Contains(s, "/native") {
		t.Error("generate.go must contain logic to exclude /native packages from resolver invocation")
	}

	// Must use pipefail to propagate pkgfilter errors loudly.
	if !strings.Contains(s, "pipefail") {
		t.Error("generate.go resolver directives must use 'pipefail' to propagate pkgfilter failures")
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

// TestHasNativeSegment verifies that HasNativeSegment correctly identifies
// the "native" path segment using exact matching (not substring matching).
func TestHasNativeSegment(t *testing.T) {
	cases := []struct {
		importPath string
		want       bool
	}{
		// True: "native" is a path segment
		{"github.com/foo/apis/cluster/s3/v1beta1/native", true},
		{"github.com/foo/apis/cluster/s3/v1beta1/native/sub", true},
		{"native", true},
		{"a/native/b", true},

		// False: "native" only appears as part of a larger segment
		{"github.com/foo/apis/cluster/nativefeatures/v1", false},
		{"github.com/foo/apis/cluster/mynative/v1", false},
		{"github.com/foo/apis/cluster/s3/v1beta1", false},
		{"", false},
	}
	for _, tc := range cases {
		got := generate.HasNativeSegment(tc.importPath)
		if got != tc.want {
			t.Errorf("HasNativeSegment(%q) = %v, want %v", tc.importPath, got, tc.want)
		}
	}
}

// TestFilterNativePackagesExcludesNative verifies that FilterNativePackages
// runs against the real package tree and excludes native/ sub-packages while
// retaining all other packages.
//
// This test runs "go list ./apis/cluster/..." from the repo root and checks
// that the known native package is excluded.
func TestFilterNativePackagesExcludesNative(t *testing.T) {
	// generate_test.go lives in generate/, so ".." is the repo root.
	repoRoot := ".."

	pkgs, err := generate.FilterNativePackages(repoRoot, "./apis/cluster/...")
	if err != nil {
		t.Fatalf("FilterNativePackages returned error: %v", err)
	}

	if len(pkgs) == 0 {
		t.Fatal("FilterNativePackages returned no packages — expected at least one cluster package")
	}

	// The known native package must be absent.
	const nativePkg = "github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1/native"
	for _, p := range pkgs {
		if p == nativePkg {
			t.Errorf("native package %q was NOT excluded by FilterNativePackages", nativePkg)
		}
		// General rule: no package with "native" as an exact segment must appear.
		if generate.HasNativeSegment(p) {
			t.Errorf("package %q contains 'native' segment and should have been excluded", p)
		}
	}

	// At least one non-native cluster package must be present.
	const s3Pkg = "github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1"
	found := false
	for _, p := range pkgs {
		if p == s3Pkg {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected non-native package %q to be present in results, but it was not found", s3Pkg)
	}
}

// TestFilterNativePackagesErrorPropagation verifies that FilterNativePackages
// returns a non-nil error (rather than silently returning empty) when "go list"
// is given an invalid pattern.
func TestFilterNativePackagesErrorPropagation(t *testing.T) {
	_, err := generate.FilterNativePackages("..", "./this/does/not/exist/...")
	if err == nil {
		t.Error("FilterNativePackages should return an error for a non-existent pattern, got nil")
	}
}
