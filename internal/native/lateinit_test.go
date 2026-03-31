// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpfake "github.com/crossplane/crossplane-runtime/v2/pkg/resource/fake"
)

// ---------------------------------------------------------------------------
// Test struct types used across multiple test cases
// ---------------------------------------------------------------------------

type lateInitSpec struct {
	Name        *string       `json:"name,omitempty"`
	Description *string       `json:"description,omitempty"`
	Count       *int64        `json:"count,omitempty"`
	Enabled     *bool         `json:"enabled,omitempty"`
	Ratio       *float64      `json:"ratio,omitempty"`
	Nested      *lateInitNest `json:"nested,omitempty"`
	Tags        *string       `json:"tags,omitempty"`
	// unexported field — must never be touched
	secret *string //nolint:unused
}

type lateInitNest struct {
	Value    *string `json:"value,omitempty"`
	Priority *int64  `json:"priority,omitempty"`
}

// strPtr / int64Ptr / boolPtr / float64Ptr are local helpers (avoid collision
// with the ptr() helper defined in tags_test.go, which is also in package native).
func strP(s string) *string   { return &s }
func i64P(n int64) *int64     { return &n }
func boolP(b bool) *bool      { return &b }
func f64P(f float64) *float64 { return &f }

// newCR returns a fake managed resource for conditional-ignore tests.
func newCR() *xpfake.Managed {
	return &xpfake.Managed{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
	}
}

// ---------------------------------------------------------------------------
// IsIgnored
// ---------------------------------------------------------------------------

func TestIsIgnored_FieldInList(t *testing.T) {
	cfg := LateInitConfig{IgnoredFields: []string{"name", "description"}}
	if !IsIgnored("name", cfg) {
		t.Error("expected 'name' to be ignored")
	}
}

func TestIsIgnored_FieldNotInList(t *testing.T) {
	cfg := LateInitConfig{IgnoredFields: []string{"name"}}
	if IsIgnored("other", cfg) {
		t.Error("expected 'other' NOT to be ignored")
	}
}

func TestIsIgnored_EmptyList(t *testing.T) {
	cfg := LateInitConfig{}
	if IsIgnored("name", cfg) {
		t.Error("expected false when IgnoredFields is empty")
	}
}

// ---------------------------------------------------------------------------
// Individual typed helpers
// ---------------------------------------------------------------------------

func TestLateInitializeStringPtr_NilDst(t *testing.T) {
	src := strP("hello")
	var dst *string
	changed := LateInitializeStringPtr(&dst, src)
	if !changed {
		t.Fatal("expected changed=true")
	}
	if dst != src {
		t.Errorf("expected dst==%p, got %p", src, dst)
	}
}

func TestLateInitializeStringPtr_AlreadySet(t *testing.T) {
	existing := strP("existing")
	src := strP("new")
	dst := existing
	changed := LateInitializeStringPtr(&dst, src)
	if changed {
		t.Fatal("expected changed=false when dst already set")
	}
	if dst != existing {
		t.Error("expected dst to remain unchanged")
	}
}

func TestLateInitializeStringPtr_NilSrc(t *testing.T) {
	var dst *string
	changed := LateInitializeStringPtr(&dst, nil)
	if changed {
		t.Fatal("expected changed=false when src is nil")
	}
	if dst != nil {
		t.Error("expected dst to remain nil")
	}
}

func TestLateInitializeBoolPtr_NilDst(t *testing.T) {
	src := boolP(true)
	var dst *bool
	changed := LateInitializeBoolPtr(&dst, src)
	if !changed {
		t.Fatal("expected changed=true")
	}
	if dst != src {
		t.Errorf("expected dst==%p", src)
	}
}

func TestLateInitializeBoolPtr_AlreadySet(t *testing.T) {
	existing := boolP(false)
	dst := existing
	changed := LateInitializeBoolPtr(&dst, boolP(true))
	if changed {
		t.Fatal("expected changed=false when dst already set")
	}
	if dst != existing {
		t.Error("expected dst to remain unchanged")
	}
}

func TestLateInitializeInt64Ptr_NilDst(t *testing.T) {
	src := i64P(42)
	var dst *int64
	changed := LateInitializeInt64Ptr(&dst, src)
	if !changed {
		t.Fatal("expected changed=true")
	}
	if dst != src {
		t.Errorf("expected dst==%p", src)
	}
}

func TestLateInitializeInt64Ptr_NilSrc(t *testing.T) {
	var dst *int64
	changed := LateInitializeInt64Ptr(&dst, nil)
	if changed {
		t.Fatal("expected changed=false when src is nil")
	}
}

func TestLateInitializeInt32Ptr_NilDst(t *testing.T) {
	v := int32(7)
	var dst *int32
	changed := LateInitializeInt32Ptr(&dst, &v)
	if !changed {
		t.Fatal("expected changed=true")
	}
	if dst != &v {
		t.Errorf("expected dst to point to %d", v)
	}
}

func TestLateInitializeFloat64Ptr_NilDst(t *testing.T) {
	src := f64P(3.14)
	var dst *float64
	changed := LateInitializeFloat64Ptr(&dst, src)
	if !changed {
		t.Fatal("expected changed=true")
	}
	if dst != src {
		t.Errorf("expected dst==%p", src)
	}
}

func TestLateInitializeMapStringPtr_NilDst(t *testing.T) {
	v := "bar"
	src := map[string]*string{"foo": &v}
	var dst map[string]*string
	changed := LateInitializeMapStringPtr(&dst, src)
	if !changed {
		t.Fatal("expected changed=true")
	}
	if len(dst) != 1 {
		t.Errorf("expected dst to have 1 entry, got %d", len(dst))
	}
}

func TestLateInitializeMapStringPtr_AlreadySet(t *testing.T) {
	v := "bar"
	existing := map[string]*string{"existing": &v}
	dst := existing
	changed := LateInitializeMapStringPtr(&dst, map[string]*string{})
	if changed {
		t.Fatal("expected changed=false when dst already set")
	}
}

func TestLateInitializeMapStringPtr_NilSrc(t *testing.T) {
	var dst map[string]*string
	changed := LateInitializeMapStringPtr(&dst, nil)
	if changed {
		t.Fatal("expected changed=false when src is nil")
	}
}

// ---------------------------------------------------------------------------
// LateInitialize — reflection-based
// ---------------------------------------------------------------------------

func TestLateInitialize_SimpleField_Nil(t *testing.T) {
	spec := &lateInitSpec{}
	observed := &lateInitSpec{Name: strP("alice")}

	changed := LateInitialize(nil, spec, observed, LateInitConfig{})
	if !changed {
		t.Fatal("expected changed=true")
	}
	if spec.Name == nil || *spec.Name != "alice" {
		t.Errorf("expected spec.Name=alice, got %v", spec.Name)
	}
}

func TestLateInitialize_AlreadySet_NotOverwritten(t *testing.T) {
	spec := &lateInitSpec{Name: strP("bob")}
	observed := &lateInitSpec{Name: strP("alice")}

	changed := LateInitialize(nil, spec, observed, LateInitConfig{})
	if changed {
		t.Fatal("expected changed=false — field already set")
	}
	if *spec.Name != "bob" {
		t.Errorf("expected spec.Name to stay bob, got %s", *spec.Name)
	}
}

func TestLateInitialize_MultipleFields(t *testing.T) {
	spec := &lateInitSpec{Name: strP("bob")}
	observed := &lateInitSpec{
		Name:        strP("alice"),
		Description: strP("my desc"),
		Count:       i64P(10),
	}

	changed := LateInitialize(nil, spec, observed, LateInitConfig{})
	if !changed {
		t.Fatal("expected changed=true")
	}
	// Name should not change (already set)
	if *spec.Name != "bob" {
		t.Errorf("spec.Name: want bob, got %s", *spec.Name)
	}
	// Description should be initialized
	if spec.Description == nil || *spec.Description != "my desc" {
		t.Errorf("spec.Description: want 'my desc', got %v", spec.Description)
	}
	// Count should be initialized
	if spec.Count == nil || *spec.Count != 10 {
		t.Errorf("spec.Count: want 10, got %v", spec.Count)
	}
}

func TestLateInitialize_ReturnsFalse_WhenNothingInitialized(t *testing.T) {
	spec := &lateInitSpec{
		Name:        strP("bob"),
		Description: strP("desc"),
	}
	observed := &lateInitSpec{
		Name:        strP("alice"),
		Description: strP("other"),
	}

	changed := LateInitialize(nil, spec, observed, LateInitConfig{})
	if changed {
		t.Fatal("expected changed=false — all already-set fields")
	}
}

func TestLateInitialize_IgnoredFields_NotInitialized(t *testing.T) {
	spec := &lateInitSpec{}
	observed := &lateInitSpec{Name: strP("alice"), Description: strP("desc")}

	cfg := LateInitConfig{IgnoredFields: []string{"name"}}
	changed := LateInitialize(nil, spec, observed, cfg)
	if !changed {
		t.Fatal("expected changed=true (description should be initialized)")
	}
	// Name should remain nil (ignored)
	if spec.Name != nil {
		t.Errorf("expected spec.Name=nil (ignored), got %v", spec.Name)
	}
	// Description should be set
	if spec.Description == nil || *spec.Description != "desc" {
		t.Errorf("expected spec.Description=desc, got %v", spec.Description)
	}
}

func TestLateInitialize_IgnoredFields_AllIgnored_ReturnsFalse(t *testing.T) {
	spec := &lateInitSpec{}
	observed := &lateInitSpec{Name: strP("alice")}

	cfg := LateInitConfig{IgnoredFields: []string{"name"}}
	changed := LateInitialize(nil, spec, observed, cfg)
	if changed {
		t.Fatal("expected changed=false — only field is ignored")
	}
}

func TestLateInitialize_ConditionalIgnored_WhenTrue_FieldSkipped(t *testing.T) {
	cr := newCR()
	spec := &lateInitSpec{}
	observed := &lateInitSpec{Name: strP("alice"), Description: strP("desc")}

	cfg := LateInitConfig{
		ConditionalIgnored: map[string]func(resource.Managed) bool{
			"name": func(_ resource.Managed) bool { return true }, // always ignore name
		},
	}
	changed := LateInitialize(cr, spec, observed, cfg)
	if !changed {
		t.Fatal("expected changed=true (description still initialized)")
	}
	if spec.Name != nil {
		t.Errorf("expected spec.Name=nil (conditional ignore), got %v", spec.Name)
	}
	if spec.Description == nil || *spec.Description != "desc" {
		t.Errorf("expected spec.Description=desc, got %v", spec.Description)
	}
}

func TestLateInitialize_ConditionalIgnored_WhenFalse_FieldInitialized(t *testing.T) {
	cr := newCR()
	spec := &lateInitSpec{}
	observed := &lateInitSpec{Name: strP("alice")}

	cfg := LateInitConfig{
		ConditionalIgnored: map[string]func(resource.Managed) bool{
			"name": func(_ resource.Managed) bool { return false }, // don't ignore
		},
	}
	changed := LateInitialize(cr, spec, observed, cfg)
	if !changed {
		t.Fatal("expected changed=true")
	}
	if spec.Name == nil || *spec.Name != "alice" {
		t.Errorf("expected spec.Name=alice, got %v", spec.Name)
	}
}

func TestLateInitialize_ConditionalIgnored_ReceivesCR(t *testing.T) {
	cr := newCR()
	cr.Annotations["test-key"] = "test-val"

	spec := &lateInitSpec{}
	observed := &lateInitSpec{Name: strP("alice")}

	var receivedCR resource.Managed
	cfg := LateInitConfig{
		ConditionalIgnored: map[string]func(resource.Managed) bool{
			"name": func(c resource.Managed) bool {
				receivedCR = c
				return false
			},
		},
	}
	_ = LateInitialize(cr, spec, observed, cfg)

	if receivedCR == nil {
		t.Fatal("expected conditional function to receive the CR")
	}
	if receivedCR.GetAnnotations()["test-key"] != "test-val" {
		t.Error("expected CR annotation to be accessible inside conditional function")
	}
}

func TestLateInitialize_NestedStruct_NilPointer(t *testing.T) {
	spec := &lateInitSpec{}
	observed := &lateInitSpec{
		Nested: &lateInitNest{Value: strP("nested-val")},
	}

	changed := LateInitialize(nil, spec, observed, LateInitConfig{})
	if !changed {
		t.Fatal("expected changed=true (nested pointer set)")
	}
	if spec.Nested == nil {
		t.Fatal("expected spec.Nested to be non-nil")
	}
	if spec.Nested.Value == nil || *spec.Nested.Value != "nested-val" {
		t.Errorf("expected spec.Nested.Value=nested-val, got %v", spec.Nested.Value)
	}
}

func TestLateInitialize_NestedStruct_BothNonNil_RecursesInto(t *testing.T) {
	spec := &lateInitSpec{
		Nested: &lateInitNest{}, // non-nil but empty
	}
	observed := &lateInitSpec{
		Nested: &lateInitNest{Value: strP("nested-val"), Priority: i64P(5)},
	}

	changed := LateInitialize(nil, spec, observed, LateInitConfig{})
	if !changed {
		t.Fatal("expected changed=true (nested fields initialized)")
	}
	if spec.Nested.Value == nil || *spec.Nested.Value != "nested-val" {
		t.Errorf("expected spec.Nested.Value=nested-val, got %v", spec.Nested.Value)
	}
	if spec.Nested.Priority == nil || *spec.Nested.Priority != 5 {
		t.Errorf("expected spec.Nested.Priority=5, got %v", spec.Nested.Priority)
	}
}

func TestLateInitialize_NilSpec_ReturnsFalse(t *testing.T) {
	// Passing nil spec should not panic and return false.
	changed := LateInitialize(nil, nil, &lateInitSpec{Name: strP("x")}, LateInitConfig{})
	if changed {
		t.Fatal("expected changed=false for nil spec")
	}
}

func TestLateInitialize_NilObserved_ReturnsFalse(t *testing.T) {
	spec := &lateInitSpec{}
	changed := LateInitialize(nil, spec, nil, LateInitConfig{})
	if changed {
		t.Fatal("expected changed=false for nil observed")
	}
}

func TestLateInitialize_BoolField(t *testing.T) {
	spec := &lateInitSpec{}
	observed := &lateInitSpec{Enabled: boolP(true)}

	changed := LateInitialize(nil, spec, observed, LateInitConfig{})
	if !changed {
		t.Fatal("expected changed=true")
	}
	if spec.Enabled == nil || !*spec.Enabled {
		t.Errorf("expected spec.Enabled=true, got %v", spec.Enabled)
	}
}

func TestLateInitialize_Float64Field(t *testing.T) {
	spec := &lateInitSpec{}
	observed := &lateInitSpec{Ratio: f64P(1.5)}

	changed := LateInitialize(nil, spec, observed, LateInitConfig{})
	if !changed {
		t.Fatal("expected changed=true")
	}
	if spec.Ratio == nil || *spec.Ratio != 1.5 {
		t.Errorf("expected spec.Ratio=1.5, got %v", spec.Ratio)
	}
}

// TestLateInitialize_ObservedNilNested ensures we don't crash when observed has a nil pointer.
func TestLateInitialize_ObservedNilNested_NoChange(t *testing.T) {
	spec := &lateInitSpec{}
	observed := &lateInitSpec{Nested: nil}

	changed := LateInitialize(nil, spec, observed, LateInitConfig{})
	if changed {
		t.Fatal("expected changed=false — observed nested is nil")
	}
	if spec.Nested != nil {
		t.Error("expected spec.Nested to remain nil")
	}
}
