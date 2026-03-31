// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package namespaced_test verifies the hand-written registration infrastructure
// for native (non-Terraform) types in the namespaced API scope.
package namespaced_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
)

// TestNativeRegisterFileExists verifies that apis/namespaced/native_register.go
// exists as a hand-written file (no zz_ prefix) so that make generate does NOT
// overwrite it.
//
// See: .agents/specs/terraform-removal-migration.md §0.16
func TestNativeRegisterFileExists(t *testing.T) {
	const filename = "native_register.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("%s does not exist in apis/namespaced; "+
			"create this file to register native (RAW) types with the controller-manager scheme. "+
			"It must NOT have a zz_ prefix so that make generate does not overwrite it.", filename)
	}
}

// TestNativeRegisterNoZZPrefix verifies that there is no zz_native_register.go
// which would mean the registration file is generated and would be overwritten.
func TestNativeRegisterNoZZPrefix(t *testing.T) {
	if _, err := os.Stat("zz_native_register.go"); !os.IsNotExist(err) {
		t.Error("zz_native_register.go must not exist; the native register file must be " +
			"hand-written (native_register.go) so it survives make generate")
	}
}

// TestNativeRegisterPackageDeclaration verifies that native_register.go
// declares the correct package (namespaced) and is parseable Go.
//
// See: .agents/specs/terraform-removal-migration.md §0.16
func TestNativeRegisterPackageDeclaration(t *testing.T) {
	const filename = "native_register.go"
	content, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		t.Skip("native_register.go does not yet exist; skipping content checks")
	}
	if err != nil {
		t.Fatalf("failed to read %s: %v", filename, err)
	}

	if !strings.Contains(string(content), "package namespaced") {
		t.Errorf("%s must declare 'package namespaced'", filename)
	}
}

// TestNativeRegisterIsParseable verifies that native_register.go is valid Go
// syntax.
func TestNativeRegisterIsParseable(t *testing.T) {
	const filename = "native_register.go"
	src, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		t.Skip("native_register.go does not yet exist; skipping parse check")
	}
	if err != nil {
		t.Fatalf("failed to read %s: %v", filename, err)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, filename, src, parser.AllErrors)
	if parseErr != nil {
		t.Errorf("%s failed to parse as valid Go: %v", filename, parseErr)
	}
}

// TestClusterNativeRegisterFileExists verifies that the cluster-scoped
// apis/cluster/native_register.go also exists.
//
// See: .agents/specs/terraform-removal-migration.md §0.16
func TestClusterNativeRegisterFileExists(t *testing.T) {
	const filename = "../cluster/native_register.go"
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("%s does not exist; "+
			"both cluster and namespaced scopes need a native_register.go", filename)
	}
}

// TestNativeRegisterHasLicenseHeader verifies that native_register.go has the
// standard SPDX license header (project convention).
func TestNativeRegisterHasLicenseHeader(t *testing.T) {
	const filename = "native_register.go"
	content, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		t.Skip("native_register.go does not yet exist; skipping license check")
	}
	if err != nil {
		t.Fatalf("failed to read %s: %v", filename, err)
	}

	if !strings.Contains(string(content), "SPDX-License-Identifier") {
		t.Errorf("%s must include an SPDX license header", filename)
	}
}

// TestNativeRegisterPatternComment verifies that native_register.go contains a
// comment explaining how to add new native types (template for future
// implementors).
func TestNativeRegisterPatternComment(t *testing.T) {
	const filename = "native_register.go"
	content, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		t.Skip("native_register.go does not yet exist; skipping comment check")
	}
	if err != nil {
		t.Fatalf("failed to read %s: %v", filename, err)
	}

	// The file should document the pattern for adding new native types.
	wantOneOf := []string{"AddToSchemes", "SchemeBuilder", "native"}
	found := false
	for _, want := range wantOneOf {
		if strings.Contains(string(content), want) {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("%s must reference AddToSchemes, SchemeBuilder, or 'native' to document the registration pattern", filename)
	}
}

// Ensure ast package is used (required by the import).
var _ *ast.File
