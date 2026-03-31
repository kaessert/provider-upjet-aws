//go:build generate
// +build generate

// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// NOTE: See the below link for details on what is happening here.
// https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

// Remove existing CRDs
//go:generate rm -rf ../package/crds

// Remove generated files
//go:generate bash -c "find ../apis \\( -iname 'zz_generated.conversion_hubs.go' -o -iname 'zz_generated.conversion_spokes.go' -o -iname 'zz_generated.resolvers.go' \\) -delete"
//go:generate bash -c "find ../apis -type d -empty -delete"
//go:generate bash -c "find ../internal/controller -iname 'zz_*' -delete"
//go:generate bash -c "find ../internal/controller -type d -empty -delete"
//go:generate bash -c "find ../cmd/provider -name 'zz_*' -type f -delete"
//go:generate bash -c "find ../cmd/provider -type d -maxdepth 1 -mindepth 1 -empty -delete"

// Scrape metadata from Terraform registry
//go:generate go run github.com/crossplane/upjet/v2/cmd/scraper -n hashicorp/terraform-provider-aws -r ../.work/terraform-provider-aws/website/docs/r -o ../config/provider-metadata.yaml

// NOTE(muvaf): Some of Terraform AWS provider files have "!generate" build tag
// that prevent us from using it for generator program.

// Run Terrajet generator
//go:generate go run ../cmd/generator/main.go ..

// Generate deepcopy methodsets and CRD manifests
//go:generate go run -tags generate sigs.k8s.io/controller-tools/cmd/controller-gen object:headerFile=../hack/boilerplate.go.txt paths=../apis/... crd:allowDangerousTypes=true,crdVersions=v1 output:artifacts:config=../package/crds

// Generate crossplane-runtime methodsets (resource.Claim, etc)
//go:generate go run -tags generate github.com/crossplane/crossplane-tools/cmd/angryjet generate-methodsets --header-file=../hack/boilerplate.go.txt ../apis/...

// Run upjet's transformer for the generated resolvers to get rid of the cross
// API-group imports and to prevent import cycles.
//
// Native types (in native/ sub-packages) are excluded from the resolver
// invocation because they don't implement resource.Terraformed — passing them
// to the resolver would cause a panic or generate broken code.
//
// Package enumeration is delegated to generate/cmd/pkgfilter which:
//   - Uses exact path-segment matching (not a substring grep) to exclude /native
//   - Propagates go list errors loudly (no stderr redirection to /dev/null)
//   - Outputs "-p <pkg>" flags (one per line) consumed by xargs
//
// The resolver is invoked with bash -e -o pipefail so that a pkgfilter failure
// (non-zero exit) aborts the entire go generate run via pipefail.
//
// Reference resolution for native types is handled as a separate step using
// crossplane-tools (crossplane-gen) run directly against the native
// sub-package, for example:
//
//	go run ./vendor/github.com/crossplane/crossplane-tools/cmd/crossplane-gen/... \
//	    -p ./apis/cluster/<service>/<version>/native/
//
// The resulting zz_resolve_references.go file is committed to source control.
// This step is NOT part of make generate — run it explicitly when
// +crossplane:generate:reference annotations change on native types.
//go:generate bash -e -o pipefail -c "go run ./cmd/pkgfilter .. ./apis/cluster/... | xargs go run github.com/crossplane/upjet/v2/cmd/resolver -g aws.upbound.io -a github.com/upbound/provider-aws/v2/internal/apis -s"
//go:generate bash -e -o pipefail -c "go run ./cmd/pkgfilter .. ./apis/namespaced/... | xargs go run github.com/crossplane/upjet/v2/cmd/resolver -g aws.m.upbound.io -a github.com/upbound/provider-aws/v2/internal/apis -s"

package generate

import (
	_ "sigs.k8s.io/controller-tools/cmd/controller-gen" //nolint:typecheck

	_ "github.com/crossplane/crossplane-tools/cmd/angryjet" //nolint:typecheck

	_ "github.com/crossplane/upjet/v2/cmd/scraper"

	_ "github.com/crossplane/upjet/v2/cmd/resolver"
)
