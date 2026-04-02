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
	nativesfn "github.com/upbound/provider-aws/v2/apis/namespaced/sfn/v1beta2/native"
	nativesns "github.com/upbound/provider-aws/v2/apis/namespaced/sns/v1beta1/native"
	nativesqs "github.com/upbound/provider-aws/v2/apis/namespaced/sqs/v1beta1/native"
)

func init() {
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
