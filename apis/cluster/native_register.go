// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package cluster contains Kubernetes API for the provider.
// This file registers native (non-Terraform) types with the cluster-scoped
// API scheme. It must NOT have a zz_ prefix since it is hand-written and
// should survive make generate.
package cluster

import (
	nativesfn "github.com/upbound/provider-aws/v2/apis/cluster/sfn/v1beta2/native"
	natives3 "github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1/native"
	nativesns "github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native"
	nativesqs "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
)

func init() {
	// Register native S3 types (BucketPolicyRAW and BucketPolicyRAWList)
	// so that cross-resource reference resolvers can look them up by GVK.
	AddToSchemes = append(AddToSchemes, natives3.SchemeBuilder.AddToScheme)

	// Register native sfn types (StateMachineRAW and StateMachineRAWList)
	// so that the controller manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativesfn.SchemeBuilder.AddToScheme)

	// Register native sns types (TopicRAW, TopicSubscriptionRAW and their list
	// types) so that the controller manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativesns.SchemeBuilder.AddToScheme)

	// Register native sqs types (QueueRAW, QueuePolicyRAW, QueueRedrivePolicyRAW,
	// QueueRedriveAllowPolicyRAW and their list types) so that the controller
	// manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativesqs.SchemeBuilder.AddToScheme)
}
