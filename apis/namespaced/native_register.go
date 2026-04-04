// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package namespaced contains Kubernetes API for the provider.
// This file registers native (non-Terraform) types with the namespaced-scoped
// API scheme. It must NOT have a zz_ prefix since it is hand-written and
// should survive make generate.
//
// As native (RAW) types are scaffolded for each service, append their
// SchemeBuilders here following the pattern below:
//
//	import nativeFOO "github.com/upbound/provider-aws/v2/apis/namespaced/foo/v1beta1/native"
//
//	func init() {
//	    AddToSchemes = append(AddToSchemes, nativeFOO.SchemeBuilder.AddToScheme)
//	}
//
// Go init() functions within a package run in filename alphabetical order.
// native_register.go (n) executes before zz_register.go (z), so native entries
// are appended to AddToSchemes first; the generated TF entries are appended on
// top during the same init() pass. Both end up in the same SchemeBuilder slice.
package namespaced

import (
	nativeelasticache1 "github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta1/native"
	nativeelasticache2 "github.com/upbound/provider-aws/v2/apis/namespaced/elasticache/v1beta2/native"
	nativekinesis "github.com/upbound/provider-aws/v2/apis/namespaced/kinesis/v1beta1/native"
	nativekinesis2 "github.com/upbound/provider-aws/v2/apis/namespaced/kinesis/v1beta2/native"
	nativesecretsmanager "github.com/upbound/provider-aws/v2/apis/namespaced/secretsmanager/v1beta1/native"
	nativesfn "github.com/upbound/provider-aws/v2/apis/namespaced/sfn/v1beta2/native"
	nativesns "github.com/upbound/provider-aws/v2/apis/namespaced/sns/v1beta1/native"
	nativesqs "github.com/upbound/provider-aws/v2/apis/namespaced/sqs/v1beta1/native"
)

func init() {
	// Register native elasticache types for namespaced scope
	AddToSchemes = append(AddToSchemes, nativeelasticache1.SchemeBuilder.AddToScheme)
	// Register native elasticache v1beta2 types (UserRAW hub, ReplicationGroupRAW hub)
	AddToSchemes = append(AddToSchemes, nativeelasticache2.SchemeBuilder.AddToScheme)

	// Register native kinesis types (StreamRAW, StreamConsumerRAW and their list
	// types) so that the controller manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativekinesis.SchemeBuilder.AddToScheme)

	// Register native kinesis v1beta2 types (StreamRAW served-but-not-stored alias)
	// so that the API server accepts manifests using kinesis.aws.m.upbound.io/v1beta2.
	AddToSchemes = append(AddToSchemes, nativekinesis2.SchemeBuilder.AddToScheme)

	// Register native sfn types (StateMachineRAW and StateMachineRAWList)
	// so that the controller manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativesfn.SchemeBuilder.AddToScheme)

	// Register native sns types (TopicRAW, TopicSubscriptionRAW and their list
	// types) so that the controller manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativesns.SchemeBuilder.AddToScheme)

	// Register native secretsmanager types (SecretRAW and SecretRAWList)
	// so that the controller manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativesecretsmanager.SchemeBuilder.AddToScheme)

	// Register native sqs types (QueueRAW, QueuePolicyRAW, QueueRedrivePolicyRAW,
	// QueueRedriveAllowPolicyRAW and their list types) so that the controller
	// manager can discover and watch them.
	AddToSchemes = append(AddToSchemes, nativesqs.SchemeBuilder.AddToScheme)
}
