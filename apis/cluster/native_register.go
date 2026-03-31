// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package cluster contains Kubernetes API for the provider.
// This file registers native (non-Terraform) types with the cluster-scoped
// API scheme. It must NOT have a zz_ prefix since it is hand-written and
// should survive make generate.
package cluster

import (
	natives3 "github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1/native"
)

func init() {
	// Register native S3 types (BucketPolicyRAW and BucketPolicyRAWList)
	// so that cross-resource reference resolvers can look them up by GVK.
	AddToSchemes = append(AddToSchemes, natives3.SchemeBuilder.AddToScheme)
}
