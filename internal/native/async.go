// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
)

// Annotation keys used to persist async operation state on a managed resource's
// CR. All keys follow the native.aws.upbound.io/ prefix convention.
const (
	// AnnotationAsyncOperation holds the name of the in-flight operation, e.g.
	// "creating", "updating", or "deleting".
	AnnotationAsyncOperation = "native.aws.upbound.io/async-operation"

	// AnnotationAsyncStartedAt holds the RFC3339 timestamp at which the async
	// operation was initiated. Used for observability and timeout detection.
	AnnotationAsyncStartedAt = "native.aws.upbound.io/async-started-at"

	// AnnotationAsyncRequestID holds the AWS request ID for the in-flight call.
	// Stored to allow idempotent retries when a request is re-issued.
	AnnotationAsyncRequestID = "native.aws.upbound.io/async-request-id"
)

// AsyncState describes an in-flight AWS operation that has not yet completed.
// It is serialised to and from CR annotations so that the poll-based reconciler
// can track progress across multiple reconcile loops without a separate tracker
// store (unlike upjet's OperationTrackerStore approach).
//
// Observe() pattern:
//  1. Call GetAsyncState(cr) — returns nil when no operation is in-flight.
//  2. If non-nil, poll AWS for the resource's current status.
//  3. If still running: return ResourceExists=true, ResourceUpToDate=true so
//     the reconciler re-queues via WithPollInterval.
//  4. If complete: ClearAsyncState(cr), then return the observed state.
//  5. If failed: ClearAsyncState(cr), return an error.
//
// Create/Update/Delete() pattern:
//  1. Call the AWS API.
//  2. If the response indicates an async operation (status == "creating" etc.),
//     call SetAsyncState(cr, AsyncState{...}).
//  3. Return — the reconciler calls Observe on the next poll interval.
type AsyncState struct {
	// Operation is a human-readable name for the in-flight AWS operation.
	// Typical values: "creating", "updating", "deleting".
	Operation string

	// StartedAt is the UTC time at which the operation was initiated.
	StartedAt time.Time

	// RequestID is the AWS request ID returned by the initiating API call.
	// Stored so that idempotent retries can be issued if necessary.
	RequestID string
}

// GetAsyncState reads the three async annotation keys from the CR and returns a
// populated AsyncState. It returns nil if any of the three annotations are
// absent, ensuring callers always receive a complete state or none at all.
func GetAsyncState(cr resource.Managed) *AsyncState {
	annotations := cr.GetAnnotations()
	if annotations == nil {
		return nil
	}

	op, opOK := annotations[AnnotationAsyncOperation]
	ts, tsOK := annotations[AnnotationAsyncStartedAt]
	rid, ridOK := annotations[AnnotationAsyncRequestID]

	// All three annotations must be present to constitute a valid async state.
	if !opOK || !tsOK || !ridOK {
		return nil
	}

	startedAt, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		// Malformed timestamp — treat as no in-flight operation to avoid
		// the reconciler getting permanently stuck.
		return nil
	}

	return &AsyncState{
		Operation: op,
		StartedAt: startedAt.UTC(),
		RequestID: rid,
	}
}

// SetAsyncState writes all three async annotation keys to the CR. It
// initialises the annotation map if the CR has none, making it safe to call
// on a newly created resource.
func SetAsyncState(cr resource.Managed, state AsyncState) {
	annotations := cr.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}

	annotations[AnnotationAsyncOperation] = state.Operation
	annotations[AnnotationAsyncStartedAt] = state.StartedAt.UTC().Format(time.RFC3339)
	annotations[AnnotationAsyncRequestID] = state.RequestID

	cr.SetAnnotations(annotations)
}

// ClearAsyncState removes all three async annotation keys from the CR. It is a
// no-op when the annotation map is nil or the keys are not present.
func ClearAsyncState(cr resource.Managed) {
	annotations := cr.GetAnnotations()
	if annotations == nil {
		return
	}

	delete(annotations, AnnotationAsyncOperation)
	delete(annotations, AnnotationAsyncStartedAt)
	delete(annotations, AnnotationAsyncRequestID)

	cr.SetAnnotations(annotations)
}

// IsAsyncInProgress returns true when all three async annotation keys are
// present on the CR, indicating that an AWS operation has been initiated and
// has not yet been observed as complete.
func IsAsyncInProgress(cr resource.Managed) bool {
	return GetAsyncState(cr) != nil
}
