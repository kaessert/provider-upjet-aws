// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package queue_test

import (
	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sqs/v1beta1/native"
	namespacednative "github.com/upbound/provider-aws/v2/apis/namespaced/sqs/v1beta1/native"
	"github.com/upbound/provider-aws/v2/internal/controller/sqs/queue"
)

// Compile-time verification that both scope types implement the QueueCR interface.
var _ queue.QueueCR = (*clusternative.QueueRAW)(nil)
var _ queue.QueueCR = (*namespacednative.QueueRAW)(nil)
