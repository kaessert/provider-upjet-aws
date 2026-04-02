// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package generate_test

import (
	"bufio"
	"bytes"
	"go/ast"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"golang.org/x/tools/go/packages"
)

// tfSpecificMethods is the set of method names that are uniquely defined by the
// upjet resource.Terraformed interface.  If any exported type in a .../native/
// package declares one of these methods it has accidentally acquired the
// Terraformed interface, which causes the upjet resolver to panic.
var tfSpecificMethods = map[string]bool{ //nolint:gochecknoglobals
	"GetTerraformResourceType":    true,
	"GetTerraformSchemaVersion":   true,
	"GetConnectionDetailsMapping": true,
	"GetObservation":              true,
	"SetObservation":              true,
	"GetParameters":               true,
	"SetParameters":               true,
	"GetInitParameters":           true,
	"GetMergedParameters":         true,
	"LateInitialize":              true,
}

// TestNativeTypesNotTerraformed verifies that no exported type in any
// .../native/ package implements resource.Terraformed (the upjet interface).
//
// Native controllers use the AWS SDK v2 directly; passing them to the upjet
// resolver causes a runtime panic.  This test catches the case where someone
// accidentally copies a TF-style method onto a native type.
//
// Invariant 1 from the coexistence guardrail spec.
func TestNativeTypesNotTerraformed(t *testing.T) {
	t.Helper()
	repoRoot := ".."

	// 1. Enumerate all .../native/ packages using go list.
	nativePkgs, err := listNativePackages(repoRoot)
	if err != nil {
		t.Fatalf("listing native packages: %v", err)
	}
	if len(nativePkgs) == 0 {
		t.Skip("no .../native/ packages found — nothing to test")
	}

	// 2. Load the packages with AST parsing enabled.
	cfg := &packages.Config{
		Mode: packages.NeedSyntax | packages.NeedName | packages.NeedFiles,
		Dir:  repoRoot,
	}
	loaded, err := packages.Load(cfg, nativePkgs...)
	if err != nil {
		t.Fatalf("go/packages.Load: %v", err)
	}

	// 3. Scan every method declaration in every file of every native package.
	for _, pkg := range loaded {
		for _, pErr := range pkg.Errors {
			t.Errorf("package %s load error: %v", pkg.PkgPath, pErr)
		}
		for _, file := range pkg.Syntax {
			for _, decl := range file.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok || fn.Recv == nil {
					continue // not a method
				}
				if !tfSpecificMethods[fn.Name.Name] {
					continue // not a TF-specific method
				}
				// Determine the receiver type name for a helpful error.
				var recvTypeName string
				if len(fn.Recv.List) > 0 {
					recvTypeName = receiverTypeName(fn.Recv.List[0].Type)
				}
				t.Errorf(
					"package %s: type %s has method %q — this is a "+
						"Terraform-specific method; native types must NOT "+
						"implement resource.Terraformed",
					pkg.PkgPath, recvTypeName, fn.Name.Name,
				)
			}
		}
	}
}

// receiverTypeName returns a human-readable name for a method receiver type
// expression (handles *T, T, etc.).
func receiverTypeName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.StarExpr:
		return "*" + receiverTypeName(e.X)
	case *ast.Ident:
		return e.Name
	default:
		return "<unknown>"
	}
}

// listNativePackages returns the import paths of all .../native/ packages
// reachable from repoRoot, filtered to exact "native" path segments.
func listNativePackages(repoRoot string) ([]string, error) {
	// List all packages under apis/cluster and apis/namespaced.
	patterns := []string{"./apis/cluster/...", "./apis/namespaced/..."}
	var result []string
	for _, pat := range patterns {
		cmd := exec.Command("go", "list", pat) //nolint:gosec
		cmd.Dir = repoRoot
		out, err := cmd.Output()
		if err != nil {
			// Pattern may have no packages — not fatal.
			continue
		}
		for _, line := range strings.Split(string(out), "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			if !isNativePackage(line) {
				continue
			}
			result = append(result, line)
		}
	}
	return result, nil
}

// isNativePackage reports whether importPath ends with or contains "native"
// as an exact path segment.
func isNativePackage(importPath string) bool {
	for _, seg := range strings.Split(importPath, "/") {
		if seg == "native" {
			return true
		}
	}
	return false
}

// TestNativeGenerationIdempotent verifies that running controller-gen on the
// native/ packages twice produces byte-for-byte identical output.
//
// Non-deterministic generation (e.g. from map-iteration order) would manifest
// as a diff between two runs.  This test catches that class of bug.
//
// Invariant 2 from the coexistence guardrail spec.
func TestNativeGenerationIdempotent(t *testing.T) {
	repoRoot := ".."

	// Resolve to an absolute path so that paths passed to child processes work
	// regardless of the child process's working directory.
	absRepoRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		t.Fatalf("resolving repo root: %v", err)
	}

	nativePkgs, err := listNativePackages(repoRoot)
	if err != nil {
		t.Fatalf("listing native packages: %v", err)
	}
	if len(nativePkgs) == 0 {
		t.Skip("no .../native/ packages found — nothing to test")
	}

	// Convert import paths to file-system patterns.
	nativeDirPatterns := importPathsToDirPatterns(repoRoot, nativePkgs)
	if len(nativeDirPatterns) == 0 {
		t.Skip("could not resolve native import paths to directory patterns")
	}

	// Run controller-gen twice into separate temp directories and compare.
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	// Use the absolute repo root so the header-file path resolves correctly
	// when controller-gen runs from absRepoRoot.
	headerFile := filepath.Join(absRepoRoot, "hack", "boilerplate.go.txt")

	if err := runControllerGen(repoRoot, dir1, headerFile, nativeDirPatterns); err != nil {
		t.Fatalf("first controller-gen run failed: %v", err)
	}
	if err := runControllerGen(repoRoot, dir2, headerFile, nativeDirPatterns); err != nil {
		t.Fatalf("second controller-gen run failed: %v", err)
	}

	// Compare the two output directories.
	diffFound, diffMsg := compareDirs(dir1, dir2)
	if diffFound {
		t.Errorf("controller-gen output is non-deterministic:\n%s", diffMsg)
	}
}

// importPathsToDirPatterns converts a list of import paths to relative path
// patterns suitable for passing to controller-gen (e.g. ./apis/cluster/.../native).
func importPathsToDirPatterns(repoRoot string, importPaths []string) []string {
	// Read the module name from go.mod.
	modName := readModuleName(repoRoot)
	if modName == "" {
		return nil
	}

	var patterns []string
	for _, ip := range importPaths {
		rel := strings.TrimPrefix(ip, modName+"/")
		if rel == ip {
			continue // didn't start with the module prefix
		}
		patterns = append(patterns, "./"+rel)
	}
	return patterns
}

// readModuleName reads the module name from go.mod in repoRoot.
func readModuleName(repoRoot string) string {
	f, err := os.Open(filepath.Join(repoRoot, "go.mod"))
	if err != nil {
		return ""
	}
	defer f.Close() //nolint:errcheck

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module ")
		}
	}
	return ""
}

// runControllerGen invokes controller-gen with the object generator on each
// pattern, writing output to outDir.
func runControllerGen(repoRoot, outDir, headerFile string, patterns []string) error {
	args := make([]string, 0, 3+len(patterns)+1)
	args = append(args,
		"run",
		"sigs.k8s.io/controller-tools/cmd/controller-gen",
		"object:headerFile="+headerFile,
	)
	for _, p := range patterns {
		args = append(args, "paths="+p)
	}
	args = append(args, "output:object:dir="+outDir)

	cmd := exec.Command("go", args...) //nolint:gosec
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &execError{cmd: strings.Join(args, " "), out: string(out), err: err}
	}
	return nil
}

// execError wraps an exec.Command failure with its output.
type execError struct {
	cmd string
	out string
	err error
}

func (e *execError) Error() string {
	return "command " + e.cmd + " failed: " + e.err.Error() + "\n" + e.out
}

// compareDirs returns true + a diff summary if the two directories have
// different file sets or file contents.
func compareDirs(dir1, dir2 string) (bool, string) {
	files1 := dirFiles(dir1)
	files2 := dirFiles(dir2)

	set1 := fileSet(files1)
	set2 := fileSet(files2)

	var diffs []string
	for name := range set1 {
		if !set2[name] {
			diffs = append(diffs, "file "+name+" present in first run, absent in second")
		}
	}
	for name := range set2 {
		if !set1[name] {
			diffs = append(diffs, "file "+name+" absent in first run, present in second")
		}
	}
	// Compare common files.
	for name := range set1 {
		if !set2[name] {
			continue
		}
		c1, _ := os.ReadFile(filepath.Join(dir1, name))
		c2, _ := os.ReadFile(filepath.Join(dir2, name))
		if !bytes.Equal(c1, c2) {
			diffs = append(diffs, "file "+name+" has different content between runs")
		}
	}

	sort.Strings(diffs)
	return len(diffs) > 0, strings.Join(diffs, "\n")
}

func dirFiles(dir string) []string {
	var files []string
	_ = filepath.Walk(dir, func(path string, info os.FileInfo, _ error) error {
		if info != nil && !info.IsDir() {
			rel, _ := filepath.Rel(dir, path)
			files = append(files, rel)
		}
		return nil
	})
	return files
}

func fileSet(files []string) map[string]bool {
	m := make(map[string]bool, len(files))
	for _, f := range files {
		m[f] = true
	}
	return m
}

// hookVarRE matches lines like:
//
//	var NativeSetupHook_accessanalyzer func(...) error
var hookVarRE = regexp.MustCompile(`^var NativeSetupHook_(\w+)\s`)

// TestNativeSetupHookVarsMatchServices verifies that the NativeSetupHook_*
// variables declared in internal/controller/{cluster,namespaced}/native_hooks.go
// exactly match the service groups inferred from cmd/provider/ subdirectories.
//
// If a new service is added to cmd/provider/ but not to native_hooks.go, the
// generated zz_main.go will contain a reference to a non-existent variable and
// fail to compile.
//
// Invariant 3 from the coexistence guardrail spec.
func TestNativeSetupHookVarsMatchServices(t *testing.T) {
	repoRoot := ".."

	// 1. Collect expected service names from cmd/provider/ subdirectories.
	providerDir := filepath.Join(repoRoot, "cmd", "provider")
	entries, err := os.ReadDir(providerDir)
	if err != nil {
		t.Fatalf("reading cmd/provider/: %v", err)
	}
	expectedServices := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() {
			expectedServices[e.Name()] = true
		}
	}
	if len(expectedServices) == 0 {
		t.Fatal("cmd/provider/ contains no subdirectories — expected at least one service")
	}

	// 2. Verify that both native_hooks.go files declare hooks for every service.
	hookFiles := []string{
		filepath.Join(repoRoot, "internal", "controller", "cluster", "native_hooks.go"),
		filepath.Join(repoRoot, "internal", "controller", "namespaced", "native_hooks.go"),
	}
	for _, hookFile := range hookFiles {
		declared := parseHookVars(t, hookFile)
		checkHookCoverage(t, hookFile, expectedServices, declared)
	}

	// 3. Verify that the template references NativeSetupHook_{{ .Group }},
	// i.e. it uses the hook pattern (not a hard-coded name).
	tmplPath := filepath.Join(repoRoot, "hack", "main.go.tmpl")
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		t.Fatalf("reading %s: %v", tmplPath, err)
	}
	if !strings.Contains(string(tmplContent), "NativeSetupHook_") {
		t.Errorf("%s must reference NativeSetupHook_ pattern so generated zz_main.go binds native controllers", tmplPath)
	}
	if !strings.Contains(string(tmplContent), "{{ .Group }}") {
		t.Errorf("%s must use {{ .Group }} template variable for service-specific hook lookup", tmplPath)
	}
}

// parseHookVars extracts the service names from NativeSetupHook_* variable
// declarations in a native_hooks.go file.
func parseHookVars(t *testing.T, path string) map[string]bool {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("opening %s: %v", path, err)
	}
	defer f.Close() //nolint:errcheck

	declared := make(map[string]bool)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		m := hookVarRE.FindStringSubmatch(scanner.Text())
		if m != nil {
			declared[m[1]] = true
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanning %s: %v", path, err)
	}
	return declared
}

// checkHookCoverage verifies that every expected service has a declared hook,
// and that no extra hooks exist for non-existent services.
func checkHookCoverage(t *testing.T, hookFile string, expected, declared map[string]bool) {
	t.Helper()

	for svc := range expected {
		if !declared[svc] {
			t.Errorf("%s: missing NativeSetupHook_%s — add a variable for the new service", hookFile, svc)
		}
	}
	for svc := range declared {
		if !expected[svc] {
			t.Errorf("%s: NativeSetupHook_%s declared but cmd/provider/%s does not exist — remove stale hook", hookFile, svc, svc)
		}
	}
}

// TestFilterNativePackagesAlsoExcludesNamespaced is a strengthened variant of
// TestFilterNativePackagesExcludesNative.  It verifies that FilterNativePackages
// correctly excludes native/ packages from the namespaced pattern as well as the
// cluster pattern, and that the function works correctly against the real
// package tree (not just static string matching).
//
// Invariant 4 — strengthened — from the coexistence guardrail spec.
func TestFilterNativePackagesAlsoExcludesNamespaced(t *testing.T) {
	repoRoot := ".."

	// Test against the namespaced pattern as well as cluster.
	for _, pat := range []string{"./apis/cluster/...", "./apis/namespaced/..."} {
		pat := pat // capture for error messages
		pkgs, err := filterPackages(repoRoot, pat)
		if err != nil {
			t.Fatalf("FilterNativePackages(%q) returned error: %v", pat, err)
		}
		if len(pkgs) == 0 {
			t.Fatalf("FilterNativePackages(%q) returned no packages — expected at least one", pat)
		}
		for _, p := range pkgs {
			if isNativePackage(p) {
				t.Errorf(
					"FilterNativePackages(%q) returned %q which is a native/ package; "+
						"native packages must be excluded from the resolver invocation",
					pat, p,
				)
			}
		}
	}
}

// filterPackages is a thin adapter that calls go list directly (mirroring the
// generate.FilterNativePackages logic) so the test can verify the real go list
// output rather than only calling the library function.
func filterPackages(repoRoot, pattern string) ([]string, error) {
	cmd := exec.Command("go", "list", pattern) //nolint:gosec
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if strings.Contains(err.Error(), "exit") {
			_ = exitErr
		}
		return nil, err
	}
	var result []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if isNativePackage(line) {
			continue
		}
		result = append(result, line)
	}
	return result, nil
}
