// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpfake "github.com/crossplane/crossplane-runtime/v2/pkg/resource/fake"
)

// newFakeManaged returns a fresh fake managed resource with an empty annotation map.
func newFakeManaged() *xpfake.Managed {
	return &xpfake.Managed{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{},
		},
	}
}

// ---------------------------------------------------------------------------
// Annotation key tests
// ---------------------------------------------------------------------------

// TestAnnotationKeyPrefixes verifies that all annotation constants follow the
// native.aws.upbound.io/ prefix convention.
func TestAnnotationKeyPrefixes(t *testing.T) {
	for _, key := range []string{
		AnnotationAsyncOperation,
		AnnotationAsyncStartedAt,
		AnnotationAsyncRequestID,
	} {
		const wantPrefix = "native.aws.upbound.io/"
		if len(key) < len(wantPrefix) || key[:len(wantPrefix)] != wantPrefix {
			t.Errorf("annotation key %q does not have the required prefix %q", key, wantPrefix)
		}
	}
}

// ---------------------------------------------------------------------------
// GetAsyncState tests
// ---------------------------------------------------------------------------

// TestGetAsyncState_NoAnnotations verifies that GetAsyncState returns nil when
// no async annotations are present on the CR.
func TestGetAsyncState_NoAnnotations(t *testing.T) {
	cr := newFakeManaged()
	state := GetAsyncState(cr)
	if state != nil {
		t.Errorf("expected nil state when no annotations present, got %+v", state)
	}
}

// TestGetAsyncState_PartialAnnotations verifies that GetAsyncState returns nil
// when only some async annotations are present (incomplete state).
func TestGetAsyncState_PartialAnnotations(t *testing.T) {
	cr := newFakeManaged()
	// Only set the operation annotation, not the others.
	cr.Annotations[AnnotationAsyncOperation] = "creating"

	state := GetAsyncState(cr)
	if state != nil {
		t.Errorf("expected nil state when annotations are incomplete, got %+v", state)
	}
}

// TestGetAsyncState_AllAnnotations verifies that GetAsyncState correctly parses
// all three annotations into an AsyncState.
func TestGetAsyncState_AllAnnotations(t *testing.T) {
	cr := newFakeManaged()
	now := time.Now().UTC().Truncate(time.Second)

	cr.Annotations[AnnotationAsyncOperation] = "creating"
	cr.Annotations[AnnotationAsyncStartedAt] = now.Format(time.RFC3339)
	cr.Annotations[AnnotationAsyncRequestID] = "aws-req-id-12345"

	state := GetAsyncState(cr)
	if state == nil {
		t.Fatal("expected non-nil AsyncState when all annotations are present")
	}
	if state.Operation != "creating" {
		t.Errorf("expected Operation=%q, got %q", "creating", state.Operation)
	}
	if !state.StartedAt.Equal(now) {
		t.Errorf("expected StartedAt=%v, got %v", now, state.StartedAt)
	}
	if state.RequestID != "aws-req-id-12345" {
		t.Errorf("expected RequestID=%q, got %q", "aws-req-id-12345", state.RequestID)
	}
}

// ---------------------------------------------------------------------------
// SetAsyncState tests
// ---------------------------------------------------------------------------

// TestSetAsyncState_WritesAllAnnotations verifies that SetAsyncState writes all
// three annotations to the CR.
func TestSetAsyncState_WritesAllAnnotations(t *testing.T) {
	cr := newFakeManaged()
	now := time.Now().UTC().Truncate(time.Second)

	SetAsyncState(cr, AsyncState{
		Operation: "updating",
		StartedAt: now,
		RequestID: "req-abc",
	})

	annotations := cr.GetAnnotations()

	if op, ok := annotations[AnnotationAsyncOperation]; !ok || op != "updating" {
		t.Errorf("expected operation annotation %q=%q, got %q", AnnotationAsyncOperation, "updating", op)
	}
	if ts, ok := annotations[AnnotationAsyncStartedAt]; !ok || ts != now.Format(time.RFC3339) {
		t.Errorf("expected started-at annotation %q=%q, got %q", AnnotationAsyncStartedAt, now.Format(time.RFC3339), ts)
	}
	if rid, ok := annotations[AnnotationAsyncRequestID]; !ok || rid != "req-abc" {
		t.Errorf("expected request-id annotation %q=%q, got %q", AnnotationAsyncRequestID, "req-abc", rid)
	}
}

// TestSetAsyncState_NilAnnotationMap verifies that SetAsyncState initialises
// the annotation map if the CR had none.
func TestSetAsyncState_NilAnnotationMap(t *testing.T) {
	cr := &xpfake.Managed{} // no ObjectMeta.Annotations initialised
	now := time.Now().UTC()

	SetAsyncState(cr, AsyncState{
		Operation: "deleting",
		StartedAt: now,
		RequestID: "",
	})

	if cr.GetAnnotations() == nil {
		t.Error("expected annotations map to be initialised, got nil")
	}
	if op := cr.GetAnnotations()[AnnotationAsyncOperation]; op != "deleting" {
		t.Errorf("expected operation=%q, got %q", "deleting", op)
	}
}

// ---------------------------------------------------------------------------
// ClearAsyncState tests
// ---------------------------------------------------------------------------

// TestClearAsyncState_RemovesAllAnnotations verifies that ClearAsyncState
// removes all three async annotation keys from the CR.
func TestClearAsyncState_RemovesAllAnnotations(t *testing.T) {
	cr := newFakeManaged()
	now := time.Now().UTC()
	SetAsyncState(cr, AsyncState{
		Operation: "creating",
		StartedAt: now,
		RequestID: "rid",
	})

	ClearAsyncState(cr)

	annotations := cr.GetAnnotations()
	for _, key := range []string{
		AnnotationAsyncOperation,
		AnnotationAsyncStartedAt,
		AnnotationAsyncRequestID,
	} {
		if _, exists := annotations[key]; exists {
			t.Errorf("expected annotation %q to be removed after ClearAsyncState, but it still exists", key)
		}
	}
}

// TestClearAsyncState_NoopOnEmpty verifies that ClearAsyncState is safe to call
// when no async annotations are present.
func TestClearAsyncState_NoopOnEmpty(t *testing.T) {
	cr := newFakeManaged()
	// Should not panic.
	ClearAsyncState(cr)
}

// TestClearAsyncState_NilAnnotationMap verifies that ClearAsyncState does not
// panic when the CR's annotation map is nil.
func TestClearAsyncState_NilAnnotationMap(t *testing.T) {
	cr := &xpfake.Managed{} // nil annotations
	// Should not panic.
	ClearAsyncState(cr)
}

// ---------------------------------------------------------------------------
// IsAsyncInProgress tests
// ---------------------------------------------------------------------------

// TestIsAsyncInProgress_False_NoAnnotations verifies false when no async
// annotations are on the CR.
func TestIsAsyncInProgress_False_NoAnnotations(t *testing.T) {
	cr := newFakeManaged()
	if IsAsyncInProgress(cr) {
		t.Error("expected IsAsyncInProgress=false when no annotations present")
	}
}

// TestIsAsyncInProgress_True_WhenAnnotationsPresent verifies true when all
// async annotations are set.
func TestIsAsyncInProgress_True_WhenAnnotationsPresent(t *testing.T) {
	cr := newFakeManaged()
	SetAsyncState(cr, AsyncState{
		Operation: "creating",
		StartedAt: time.Now().UTC(),
		RequestID: "rid",
	})

	if !IsAsyncInProgress(cr) {
		t.Error("expected IsAsyncInProgress=true after SetAsyncState")
	}
}

// TestIsAsyncInProgress_False_AfterClear verifies false after ClearAsyncState.
func TestIsAsyncInProgress_False_AfterClear(t *testing.T) {
	cr := newFakeManaged()
	SetAsyncState(cr, AsyncState{
		Operation: "deleting",
		StartedAt: time.Now().UTC(),
		RequestID: "rid",
	})
	ClearAsyncState(cr)

	if IsAsyncInProgress(cr) {
		t.Error("expected IsAsyncInProgress=false after ClearAsyncState")
	}
}

// ---------------------------------------------------------------------------
// Full lifecycle test
// ---------------------------------------------------------------------------

// TestAsyncLifecycle_SetReadClear demonstrates the full async annotation
// lifecycle: set state, read state, clear state.
func TestAsyncLifecycle_SetReadClear(t *testing.T) {
	cr := newFakeManaged()

	// Initially no async operation in progress.
	if IsAsyncInProgress(cr) {
		t.Fatal("expected no in-progress operation on fresh CR")
	}
	if state := GetAsyncState(cr); state != nil {
		t.Fatalf("expected nil state on fresh CR, got %+v", state)
	}

	// Start a creation operation.
	started := time.Now().UTC().Truncate(time.Second)
	SetAsyncState(cr, AsyncState{
		Operation: "creating",
		StartedAt: started,
		RequestID: "aws-req-001",
	})

	if !IsAsyncInProgress(cr) {
		t.Fatal("expected in-progress after SetAsyncState")
	}

	state := GetAsyncState(cr)
	if state == nil {
		t.Fatal("expected non-nil state after SetAsyncState")
	}
	if state.Operation != "creating" {
		t.Errorf("operation: want %q, got %q", "creating", state.Operation)
	}
	if !state.StartedAt.Equal(started) {
		t.Errorf("started-at: want %v, got %v", started, state.StartedAt)
	}
	if state.RequestID != "aws-req-001" {
		t.Errorf("request-id: want %q, got %q", "aws-req-001", state.RequestID)
	}

	// Simulate polling — still in progress — state should remain.
	if !IsAsyncInProgress(cr) {
		t.Fatal("expected still in-progress while polling")
	}

	// Operation completes — clear the state.
	ClearAsyncState(cr)

	if IsAsyncInProgress(cr) {
		t.Fatal("expected no in-progress after ClearAsyncState")
	}
	if state := GetAsyncState(cr); state != nil {
		t.Fatalf("expected nil state after ClearAsyncState, got %+v", state)
	}
}

// TestAsyncLifecycle_UpdateOperation verifies that an in-progress state can
// be overwritten with a new state (e.g., update replaces create).
func TestAsyncLifecycle_UpdateOperation(t *testing.T) {
	cr := newFakeManaged()
	t1 := time.Now().UTC().Add(-1 * time.Minute).Truncate(time.Second)
	t2 := time.Now().UTC().Truncate(time.Second)

	SetAsyncState(cr, AsyncState{
		Operation: "creating",
		StartedAt: t1,
		RequestID: "req-1",
	})

	// Overwrite with an updating state.
	SetAsyncState(cr, AsyncState{
		Operation: "updating",
		StartedAt: t2,
		RequestID: "req-2",
	})

	state := GetAsyncState(cr)
	if state == nil {
		t.Fatal("expected non-nil state")
	}
	if state.Operation != "updating" {
		t.Errorf("expected overwritten operation=%q, got %q", "updating", state.Operation)
	}
	if state.RequestID != "req-2" {
		t.Errorf("expected overwritten request-id=%q, got %q", "req-2", state.RequestID)
	}
	if !state.StartedAt.Equal(t2) {
		t.Errorf("expected overwritten started-at=%v, got %v", t2, state.StartedAt)
	}
}
