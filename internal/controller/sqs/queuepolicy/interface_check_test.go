// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package queuepolicy_test

import (
	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
	namespacednative "github.com/upbound/provider-aws/v2/apis/namespaced/sqs/v1beta1/native"
	"github.com/upbound/provider-aws/v2/internal/controller/sqs/queuepolicy"
)

// Compile-time verification that both scope types implement the QueuePolicyCR interface.
var _ queuepolicy.QueuePolicyCR = (*clusternative.QueuePolicyRAW)(nil)
var _ queuepolicy.QueuePolicyCR = (*namespacednative.QueuePolicyRAW)(nil)
