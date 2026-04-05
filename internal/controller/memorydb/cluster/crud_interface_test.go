// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package cluster_test

import (
	"testing"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/memorydb/v1beta1/native"
	namespacednative "github.com/upbound/provider-aws/v2/apis/namespaced/memorydb/v1beta1/native"
	clustershared "github.com/upbound/provider-aws/v2/internal/controller/memorydb/cluster"
)

// Compile-time interface checks: both cluster-scoped and namespaced ClusterRAW
// must implement the ClusterCR interface used by the shared ExternalClient.
var _ clustershared.ClusterCR = &clusternative.ClusterRAW{}
var _ clustershared.ClusterCR = &namespacednative.ClusterRAW{}

func TestClusterCRInterface(t *testing.T) {
	// This test exists solely to verify the interface constraints compile.
	// The actual logic is verified through the compile-time assertions above.
	t.Log("ClusterCR interface satisfied by both cluster-scoped and namespaced ClusterRAW")
}
