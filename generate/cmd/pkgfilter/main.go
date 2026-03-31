// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// pkgfilter lists Go packages matching the given patterns, excludes any
// package whose import path contains "native" as an exact path segment, and
// writes the remaining packages as "-p <pkg>" flags (one per line) to stdout.
//
// This is used by the //go:generate directives in generate/generate.go to
// pass a filtered package list to the upjet resolver — native types must be
// excluded because they do not implement resource.Terraformed.
//
// Usage:
//
//	go run ./cmd/pkgfilter <root-dir> <pattern> [pattern...]
//
// Example (run from generate/):
//
//	go run ./cmd/pkgfilter .. ./apis/cluster/... ./apis/namespaced/...
//
// Exits non-zero if "go list" fails for any pattern so that the generate step
// fails loudly rather than silently producing an incomplete package list.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/upbound/provider-aws/v2/generate"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: pkgfilter <root-dir> <pattern> [pattern...]")
		os.Exit(1)
	}

	rootDir := os.Args[1]
	patterns := os.Args[2:]

	pkgs, err := generate.FilterNativePackages(rootDir, patterns...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pkgfilter: %v\n", err)
		os.Exit(1)
	}

	// Output one "-p <pkg>" per line so the caller can feed it to xargs or
	// capture it with $() and word-split into flags.
	var sb strings.Builder
	for _, pkg := range pkgs {
		sb.WriteString("-p ")
		sb.WriteString(pkg)
		sb.WriteByte('\n')
	}
	fmt.Print(sb.String())
}
