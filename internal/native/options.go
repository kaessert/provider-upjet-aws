// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

// Package native provides framework utilities for building native AWS SDK v2
// controllers without any upjet/Terraform dependencies.
package native

import (
	"time"

	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
)

// Options extends crossplane-runtime's controller.Options with AWS-specific
// fields for native controllers. Use this type in all native controller
// Setup() functions instead of tjcontroller.Options (upjet).
//
// PollInterval shadows the embedded controller.Options.PollInterval and is
// used specifically for async resource polling (e.g. EKS cluster creation).
type Options struct {
	xpcontroller.Options

	// PollInterval is the interval at which async AWS operations are polled
	// for completion (e.g., during EKS cluster or RDS instance creation).
	// Distinct from the reconciler poll interval in the embedded Options.
	PollInterval time.Duration
}
