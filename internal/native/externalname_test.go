// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpmeta "github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	xpfake "github.com/crossplane/crossplane-runtime/v2/pkg/resource/fake"
)

// ---------------------------------------------------------------------------
// GetExternalName tests
// ---------------------------------------------------------------------------

// TestGetExternalName_Empty verifies that GetExternalName returns an empty
// string when the external-name annotation is not set.
func TestGetExternalName_Empty(t *testing.T) {
	cr := &xpfake.Managed{
		ObjectMeta: metav1.ObjectMeta{},
	}
	if name := GetExternalName(cr); name != "" {
		t.Errorf("expected empty external name, got %q", name)
	}
}

// TestGetExternalName_ReturnsAnnotationValue verifies that GetExternalName
// returns the value of the crossplane.io/external-name annotation.
func TestGetExternalName_ReturnsAnnotationValue(t *testing.T) {
	const want = "my-resource-name"
	cr := &xpfake.Managed{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				xpmeta.AnnotationKeyExternalName: want,
			},
		},
	}
	if got := GetExternalName(cr); got != want {
		t.Errorf("GetExternalName() = %q, want %q", got, want)
	}
}

// TestGetExternalName_NilAnnotations verifies that GetExternalName handles nil
// annotations safely.
func TestGetExternalName_NilAnnotations(t *testing.T) {
	cr := &xpfake.Managed{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: nil,
		},
	}
	// Must not panic; returns "".
	got := GetExternalName(cr)
	if got != "" {
		t.Errorf("expected empty string for nil annotations, got %q", got)
	}
}

// ---------------------------------------------------------------------------
// SetExternalName tests
// ---------------------------------------------------------------------------

// TestSetExternalName_SetsAnnotation verifies that SetExternalName writes the
// crossplane.io/external-name annotation to the CR.
func TestSetExternalName_SetsAnnotation(t *testing.T) {
	cr := &xpfake.Managed{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
	}
	const want = "provider-assigned-id"
	SetExternalName(cr, want)

	got := cr.GetAnnotations()[xpmeta.AnnotationKeyExternalName]
	if got != want {
		t.Errorf("SetExternalName: annotation = %q, want %q", got, want)
	}
}

// TestSetExternalName_InitialisesAnnotationMap verifies that SetExternalName
// initialises the annotation map when it is nil.
func TestSetExternalName_InitialisesAnnotationMap(t *testing.T) {
	cr := &xpfake.Managed{} // nil annotations
	const want = "my-external-id"
	SetExternalName(cr, want)

	annotations := cr.GetAnnotations()
	if annotations == nil {
		t.Fatal("expected annotations map to be initialised after SetExternalName")
	}
	if got := annotations[xpmeta.AnnotationKeyExternalName]; got != want {
		t.Errorf("SetExternalName: annotation = %q, want %q", got, want)
	}
}

// TestSetExternalName_OverwritesExistingValue verifies that SetExternalName
// replaces any pre-existing external-name annotation.
func TestSetExternalName_OverwritesExistingValue(t *testing.T) {
	const initial = "old-name"
	const updated = "new-name"
	cr := &xpfake.Managed{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				xpmeta.AnnotationKeyExternalName: initial,
			},
		},
	}
	SetExternalName(cr, updated)

	got := cr.GetAnnotations()[xpmeta.AnnotationKeyExternalName]
	if got != updated {
		t.Errorf("SetExternalName overwrite: annotation = %q, want %q", got, updated)
	}
}

// ---------------------------------------------------------------------------
// Round-trip test
// ---------------------------------------------------------------------------

// TestExternalNameRoundTrip verifies that setting and then getting the external
// name returns the original value.
func TestExternalNameRoundTrip(t *testing.T) {
	cr := newFakeManaged()
	const name = "round-trip-id"

	SetExternalName(cr, name)
	got := GetExternalName(cr)
	if got != name {
		t.Errorf("round-trip: GetExternalName after SetExternalName = %q, want %q", got, name)
	}
}
