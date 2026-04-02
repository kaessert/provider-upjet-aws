// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package main is the entry point for the kinesis provider binary.
// This file imports native (non-Terraform) controller packages so their
// init() functions run and register the NativeSetupHook_kinesis variable.
package main

import (
	// Import native cluster-scoped controllers so their init() function
	// registers them via NativeSetupHook_kinesis.
	_ "github.com/upbound/provider-aws/v2/internal/controller/cluster/kinesis"

	// Import native namespaced controllers so their init() function
	// registers them via NativeSetupHook_kinesis.
	_ "github.com/upbound/provider-aws/v2/internal/controller/namespaced/kinesis"
)
