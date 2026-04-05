// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package cluster contains Kubernetes API for the provider.
// This file registers native (non-Terraform) types with the cluster-scoped
// API scheme. It must NOT have a zz_ prefix since it is hand-written and
// should survive make generate.
package cluster

import (
	nativeelasticache1 "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta1/native"
	nativeelasticache2 "github.com/upbound/provider-aws/v2/apis/cluster/elasticache/v1beta2/native"
	nativekinesis1 "github.com/upbound/provider-aws/v2/apis/cluster/kinesis/v1beta1/native"
	nativekinesis2 "github.com/upbound/provider-aws/v2/apis/cluster/kinesis/v1beta2/native"
	nativememorydb "github.com/upbound/provider-aws/v2/apis/cluster/memorydb/v1beta1/native"
	natives3 "github.com/upbound/provider-aws/v2/apis/cluster/s3/v1beta1/native"
	nativesecretsmanager "github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native"
	nativesecretsmanager2 "github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta2/native"
	nativesfn "github.com/upbound/provider-aws/v2/apis/cluster/sfn/v1beta2/native"
	nativesns "github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native"
	nativesqs "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
)

func init() {
	// Register native elasticache types (SubnetGroupRAW, ParameterGroupRAW, UserGroupRAW,
	// ClusterRAW, GlobalReplicationGroupRAW, ServerlessCacheRAW + UserRAW spoke, ReplicationGroupRAW spoke)
	AddToSchemes = append(AddToSchemes, nativeelasticache1.SchemeBuilder.AddToScheme)
	// Register native elasticache v1beta2 types (UserRAW hub, ReplicationGroupRAW hub)
	AddToSchemes = append(AddToSchemes, nativeelasticache2.SchemeBuilder.AddToScheme)

	// Register native kinesis types (StreamRAW, StreamConsumerRAW and their list
	// types) so that the controller manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativekinesis1.SchemeBuilder.AddToScheme)
	AddToSchemes = append(AddToSchemes, nativekinesis2.SchemeBuilder.AddToScheme)

	// Register native S3 types (BucketPolicyRAW and BucketPolicyRAWList)
	// so that cross-resource reference resolvers can look them up by GVK.
	AddToSchemes = append(AddToSchemes, natives3.SchemeBuilder.AddToScheme)

	// Register native sfn types (StateMachineRAW and StateMachineRAWList)
	// so that the controller manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativesfn.SchemeBuilder.AddToScheme)

	// Register native sns types (TopicRAW, TopicSubscriptionRAW and their list
	// types) so that the controller manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativesns.SchemeBuilder.AddToScheme)

	// Register native secretsmanager types (SecretRAW, SecretPolicyRAW,
	// SecretVersionRAW, SecretRotationRAW-v1beta1 and their list types).
	AddToSchemes = append(AddToSchemes, nativesecretsmanager.SchemeBuilder.AddToScheme)
	// Register native secretsmanager v1beta2 types (SecretRotationRAW — storage
	// version / conversion hub).
	AddToSchemes = append(AddToSchemes, nativesecretsmanager2.SchemeBuilder.AddToScheme)

	// Register native sqs types (QueueRAW, QueuePolicyRAW, QueueRedrivePolicyRAW,
	// QueueRedriveAllowPolicyRAW and their list types) so that the controller
	// manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativesqs.SchemeBuilder.AddToScheme)

	// Register native memorydb types (ACLRAW and its list type)
	// so that the controller manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativememorydb.SchemeBuilder.AddToScheme)
}
