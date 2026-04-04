// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package secret_test

import (
	clusternative "github.com/upbound/provider-aws/v2/apis/cluster/secretsmanager/v1beta1/native"
	namespacednative "github.com/upbound/provider-aws/v2/apis/namespaced/secretsmanager/v1beta1/native"
	secretcontroller "github.com/upbound/provider-aws/v2/internal/controller/secretsmanager/secret"
)

// Compile-time interface satisfaction checks.
// If either of these lines fails to compile, the RAW types don't implement SecretCR.
var _ secretcontroller.SecretCR = &clusternative.SecretRAW{}
var _ secretcontroller.SecretCR = &namespacednative.SecretRAW{}
