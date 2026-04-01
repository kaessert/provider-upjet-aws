// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package statemachine_test

import (
	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/sfn/v1beta2/native"
	namespacednative "github.com/upbound/provider-aws/v2/apis/namespaced/sfn/v1beta2/native"
	statemachine "github.com/upbound/provider-aws/v2/internal/controller/sfn/statemachine"
)

// Compile-time verification that both scope types implement the StateMachineCR interface.
var _ statemachine.StateMachineCR = (*clusternative.StateMachineRAW)(nil)
var _ statemachine.StateMachineCR = (*namespacednative.StateMachineRAW)(nil)
