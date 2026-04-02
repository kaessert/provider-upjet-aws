// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package topic_test

import (
	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sns/v1beta1/native"
	namespacednative "github.com/upbound/provider-aws/v2/apis/namespaced/sns/v1beta1/native"
	snstopic "github.com/upbound/provider-aws/v2/internal/controller/sns/topic"
)

// Compile-time interface satisfaction checks.
// If either of these lines fails to compile, the RAW types don't implement TopicCR.
var _ snstopic.TopicCR = &clusternative.TopicRAW{}
var _ snstopic.TopicCR = &namespacednative.TopicRAW{}
