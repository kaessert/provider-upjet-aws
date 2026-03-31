// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"testing"
	"time"

	xpcontroller "github.com/crossplane/crossplane-runtime/v2/pkg/controller"
)

// TestOptions_EmbedControllerOptions verifies that native.Options embeds
// crossplane-runtime's controller.Options and exposes all its fields.
func TestOptions_EmbedControllerOptions(t *testing.T) {
	// Verify that the embedded Options field is accessible and of the right type.
	var base xpcontroller.Options = xpcontroller.DefaultOptions()
	opts := Options{
		Options:      base,
		PollInterval: 5 * time.Minute,
	}

	// The PollInterval on the outer struct should be the async poll interval.
	if opts.PollInterval != 5*time.Minute {
		t.Errorf("expected PollInterval=%v, got %v", 5*time.Minute, opts.PollInterval)
	}

	// Logger from embedded controller.Options must still be accessible.
	if opts.Logger == nil {
		t.Error("expected Logger to be non-nil from embedded controller.Options")
	}

	// MaxConcurrentReconciles from embedded controller.Options must be accessible.
	if opts.MaxConcurrentReconciles != base.MaxConcurrentReconciles {
		t.Errorf("expected MaxConcurrentReconciles=%d, got %d",
			base.MaxConcurrentReconciles, opts.MaxConcurrentReconciles)
	}
}

// TestOptions_ZeroValue verifies the zero value compiles without panics.
func TestOptions_ZeroValue(t *testing.T) {
	var o Options
	// Just verifying it compiles and has the right type for PollInterval.
	_ = o.PollInterval
	_ = o.Options
}

// TestOptions_SetupFnCompatibility verifies that Options can be used in a
// native Setup function without importing any upjet types.
func TestOptions_SetupFnCompatibility(t *testing.T) {
	// This function signature matches the expected Setup pattern for native
	// controllers — using native.Options, NOT tjcontroller.Options.
	type setupFn func(opts Options) error

	fn := setupFn(func(o Options) error {
		// Fields from embedded controller.Options are accessible.
		_ = o.Logger
		_ = o.MaxConcurrentReconciles
		// Native PollInterval is accessible.
		_ = o.PollInterval
		return nil
	})

	if err := fn(Options{
		Options:      xpcontroller.DefaultOptions(),
		PollInterval: 30 * time.Second,
	}); err != nil {
		t.Errorf("unexpected error from setup fn: %v", err)
	}
}
