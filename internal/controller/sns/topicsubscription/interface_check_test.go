// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package topicsubscription_test

import (
	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native"
	namespacednative "github.com/upbound/provider-aws/v2/apis/namespaced/sns/v1beta1/native"
	snssub "github.com/upbound/provider-aws/v2/internal/controller/sns/topicsubscription"
)

// Compile-time interface satisfaction checks.
// If either of these lines fails to compile, the RAW types don't implement TopicSubscriptionCR.
var _ snssub.TopicSubscriptionCR = &clusternative.TopicSubscriptionRAW{}
var _ snssub.TopicSubscriptionCR = &namespacednative.TopicSubscriptionRAW{}
