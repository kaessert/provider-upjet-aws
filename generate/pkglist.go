// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package generate holds the code-generation directives and supporting
// utilities for the provider-aws build pipeline.
package generate

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// FilterNativePackages runs "go list" for each pattern (rooted at rootDir) and
// returns only packages whose import paths do NOT contain "native" as an exact
// path segment.
//
// Unlike a simple strings.Contains("/native") check this uses path-segment
// splitting, so a future package like "apis/cluster/nativefeatures/v1beta1"
// would NOT be excluded — only paths with a segment that is exactly "native".
//
// Returns an error if "go list" fails for any pattern so that callers fail
// loudly rather than silently producing an incomplete package list.
func FilterNativePackages(rootDir string, patterns ...string) ([]string, error) {
	var result []string
	for _, pat := range patterns {
		pkgs, err := listAndFilter(rootDir, pat)
		if err != nil {
			return nil, fmt.Errorf("listing packages for pattern %q: %w", pat, err)
		}
		result = append(result, pkgs...)
	}
	return result, nil
}

// listAndFilter runs "go list <pattern>" from rootDir and returns packages
// that do not contain "native" as an exact path segment.
func listAndFilter(rootDir, pattern string) ([]string, error) {
	cmd := exec.Command("go", "list", pattern)
	cmd.Dir = rootDir
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return nil, fmt.Errorf("go list %q failed (exit %d):\n%s", pattern, exitErr.ExitCode(), exitErr.Stderr)
		}
		return nil, fmt.Errorf("go list %q: %w", pattern, err)
	}

	var pkgs []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if HasNativeSegment(line) {
			continue
		}
		pkgs = append(pkgs, line)
	}
	return pkgs, nil
}

// HasNativeSegment reports whether importPath contains "native" as an exact
// forward-slash-delimited path segment.
//
// Examples:
//
//	HasNativeSegment("github.com/foo/apis/cluster/s3/v1beta1/native") == true
//	HasNativeSegment("github.com/foo/apis/cluster/s3/v1beta1/native/sub") == true
//	HasNativeSegment("github.com/foo/apis/cluster/nativefeatures/v1") == false
//	HasNativeSegment("github.com/foo/apis/cluster/s3/v1beta1") == false
func HasNativeSegment(importPath string) bool {
	for _, seg := range strings.Split(importPath, "/") {
		if seg == "native" {
			return true
		}
	}
	return false
}
