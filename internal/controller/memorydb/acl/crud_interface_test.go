// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package acl_test

import (
	"testing"

	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/memorydb/v1beta1/native"
	namespacednative "github.com/upbound/provider-aws/v2/apis/namespaced/memorydb/v1beta1/native"
	aclshared "github.com/upbound/provider-aws/v2/internal/controller/memorydb/acl"
)

// Compile-time interface checks: both cluster-scoped and namespaced ACLRAW
// must implement the ACLCR interface used by the shared ExternalClient.
var _ aclshared.ACLCR = &clusternative.ACLRAW{}
var _ aclshared.ACLCR = &namespacednative.ACLRAW{}

func TestACLCRInterface(t *testing.T) {
	// This test exists solely to verify the interface constraints compile.
	// The actual logic is verified through the compile-time assertions above.
	t.Log("ACLCR interface satisfied by both cluster-scoped and namespaced ACLRAW")
}
