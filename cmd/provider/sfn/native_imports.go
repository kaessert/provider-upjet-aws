// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package main is the entry point for the sfn provider binary.
// This file imports native (non-Terraform) controller packages so their
// init() functions run and register the NativeSetupHook_sfn variable.
package main

import (
	// Import native cluster-scoped controllers so their init() functions
	// register them via NativeSetupHook_sfn.
	_ "github.com/upbound/provider-aws/v2/internal/controller/cluster/sfn/statemachineraw"

	// Import native namespaced controllers so their init() functions
	// register them via NativeSetupHook_sfn.
	_ "github.com/upbound/provider-aws/v2/internal/controller/namespaced/sfn/statemachineraw"
)
