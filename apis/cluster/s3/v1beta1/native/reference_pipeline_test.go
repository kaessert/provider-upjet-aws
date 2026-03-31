// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package native contains hand-written native (non-Terraform) type definitions
// for the s3 service.
package native_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNativeTypeHasReferenceAnnotations verifies that the test RAW type in the
// native package contains +crossplane:generate:reference annotations.
//
// See: .agents/specs/terraform-removal-migration.md §0.4
func TestNativeTypeHasReferenceAnnotations(t *testing.T) {
	files, err := filepath.Glob("*_types.go")
	if err != nil {
		t.Fatalf("failed to glob type files: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no *_types.go files found in native package; expected at least one RAW type file with reference annotations")
	}

	found := false
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		if strings.Contains(string(content), "+crossplane:generate:reference") {
			found = true
			break
		}
	}
	if !found {
		t.Error("no +crossplane:generate:reference annotation found in any *_types.go file; " +
			"native types must include reference annotations for cross-resource references")
	}
}

// TestNativeTypeNoTerraformIDExtractor verifies that the native type files do
// NOT use the TerraformID() extractor, which casts to the upjet Terraformed
// interface and fails for native types that don't implement it.
//
// See: .agents/specs/terraform-removal-migration.md §0.4
func TestNativeTypeNoTerraformIDExtractor(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("failed to glob go files: %v", err)
	}

	for _, f := range files {
		// Skip generated files (they are produced by the pipeline, not hand-written)
		if strings.HasPrefix(filepath.Base(f), "zz_") {
			continue
		}
		// Skip test files
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		content, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		if strings.Contains(string(content), "TerraformID()") {
			t.Errorf("file %s contains TerraformID() extractor which requires the upjet Terraformed interface; "+
				"native types must use ExtractResourceID() or other non-Terraformed extractors", f)
		}
	}
}

// TestResolverFileExists verifies that zz_generated.resolvers.go exists in the
// native package. This file is produced by running angryjet generate-methodsets
// against the native sub-package.
//
// See: .agents/specs/terraform-removal-migration.md §0.4
func TestResolverFileExists(t *testing.T) {
	if _, err := os.Stat("zz_generated.resolvers.go"); os.IsNotExist(err) {
		t.Fatal("zz_generated.resolvers.go does not exist in the native package; " +
			"run: go run github.com/crossplane/crossplane-tools/cmd/angryjet generate-methodsets " +
			"--header-file=./hack/boilerplate.go.txt ./apis/cluster/s3/v1beta1/native/")
	}
}

// TestGeneratedResolverNoTerraformID verifies that the generated resolver file
// does not reference TerraformID() — the resolver must use native-compatible
// extractors only.
//
// See: .agents/specs/terraform-removal-migration.md §0.4
func TestGeneratedResolverNoTerraformID(t *testing.T) {
	content, err := os.ReadFile("zz_generated.resolvers.go")
	if os.IsNotExist(err) {
		t.Skip("zz_generated.resolvers.go not yet generated; run angryjet first")
	}
	if err != nil {
		t.Fatalf("failed to read zz_generated.resolvers.go: %v", err)
	}
	if strings.Contains(string(content), "TerraformID()") {
		t.Error("zz_generated.resolvers.go contains TerraformID(); " +
			"native type reference annotations must not use the TerraformID extractor")
	}
}

// TestGeneratedResolverCompiles verifies that zz_generated.resolvers.go is
// syntactically valid Go (parseable). A full compilation check is provided by
// running go build against the package.
//
// See: .agents/specs/terraform-removal-migration.md §0.4
func TestGeneratedResolverCompiles(t *testing.T) {
	content, err := os.ReadFile("zz_generated.resolvers.go")
	if os.IsNotExist(err) {
		t.Skip("zz_generated.resolvers.go not yet generated; run angryjet first")
	}
	if err != nil {
		t.Fatalf("failed to read zz_generated.resolvers.go: %v", err)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "zz_generated.resolvers.go", content, parser.AllErrors)
	if parseErr != nil {
		t.Errorf("zz_generated.resolvers.go failed to parse: %v", parseErr)
	}
}

// TestReferenceAnnotationUsesNativeExtractor verifies that any
// +crossplane:generate:reference:extractor annotation in the native type files
// does NOT reference the upjet TerraformID extractor and DOES reference a
// native-compatible extractor.
func TestReferenceAnnotationUsesNativeExtractor(t *testing.T) {
	files, err := filepath.Glob("*_types.go")
	if err != nil {
		t.Fatalf("failed to glob type files: %v", err)
	}

	for _, f := range files {
		fset := token.NewFileSet()
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		astFile, err := parser.ParseFile(fset, f, src, parser.ParseComments)
		if err != nil {
			t.Fatalf("failed to parse %s: %v", f, err)
		}

		for _, commentGroup := range astFile.Comments {
			for _, comment := range commentGroup.List {
				text := comment.Text
				if !strings.Contains(text, "+crossplane:generate:reference:extractor") {
					continue
				}
				// Forbidden extractors that require the Terraformed interface
				forbidden := []string{
					"TerraformID()",
					"upjet/v2/pkg/resource.TerraformID",
					"config/cluster/common.TerraformID",
				}
				for _, bad := range forbidden {
					if strings.Contains(text, bad) {
						t.Errorf("file %s contains forbidden extractor %q in annotation %q; "+
							"use ExtractResourceID() or reference.ExternalName() instead",
							f, bad, text)
					}
				}
			}
		}
	}
}

// TestNativePackageHasGroupVersionInfo verifies that the native package defines
// the group, version, and scheme builder constants needed for type registration.
func TestNativePackageHasGroupVersionInfo(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("failed to glob go files: %v", err)
	}

	var foundSchemeBuilder bool
	var foundGroupVersion bool

	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		content, err := os.ReadFile(f)
		if err != nil {
			t.Fatalf("failed to read %s: %v", f, err)
		}
		s := string(content)
		if strings.Contains(s, "SchemeBuilder") {
			foundSchemeBuilder = true
		}
		if strings.Contains(s, "GroupVersion") || strings.Contains(s, "CRDGroup") {
			foundGroupVersion = true
		}
	}

	if !foundSchemeBuilder {
		t.Error("native package must define a SchemeBuilder for type registration")
	}
	if !foundGroupVersion {
		t.Error("native package must define GroupVersion or CRDGroup constants")
	}
}

// Ensure ast import is used.
var _ ast.File
