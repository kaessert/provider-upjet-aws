// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"errors"
	"testing"

	smithy "github.com/aws/smithy-go"
)

// mockAPIError is a smithy.APIError implementation for testing.
type mockAPIError struct {
	code    string
	message string
}

func (e *mockAPIError) Error() string                 { return e.message }
func (e *mockAPIError) ErrorCode() string             { return e.code }
func (e *mockAPIError) ErrorMessage() string          { return e.message }
func (e *mockAPIError) ErrorFault() smithy.ErrorFault { return smithy.FaultUnknown }

// ---------------------------------------------------------------------------
// Wrap tests
// ---------------------------------------------------------------------------

// TestWrap_NilError verifies that Wrap returns nil when given a nil error.
func TestWrap_NilError(t *testing.T) {
	if err := Wrap(nil, "context"); err != nil {
		t.Errorf("expected Wrap(nil) to return nil, got %v", err)
	}
}

// TestWrap_NonNilError verifies that Wrap returns a non-nil error wrapping the
// original error with the provided message.
func TestWrap_NonNilError(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, "context message")
	if wrapped == nil {
		t.Fatal("expected non-nil wrapped error")
	}
	msg := wrapped.Error()
	if msg == "" {
		t.Error("expected non-empty error message from Wrap")
	}
}

// ---------------------------------------------------------------------------
// IsNotFound tests
// ---------------------------------------------------------------------------

// TestIsNotFound_NilError verifies that IsNotFound returns false for nil.
func TestIsNotFound_NilError(t *testing.T) {
	if IsNotFound(nil) {
		t.Error("expected IsNotFound(nil) = false")
	}
}

// TestIsNotFound_TrueForNotFoundCode verifies that IsNotFound returns true when
// the smithy APIError code contains "NotFound".
func TestIsNotFound_TrueForNotFoundCode(t *testing.T) {
	for _, code := range []string{
		"NotFoundException",
		"ResourceNotFoundException",
		"NotFound",
		"UserNotFoundException",
	} {
		err := &mockAPIError{code: code, message: "resource not found"}
		if !IsNotFound(err) {
			t.Errorf("IsNotFound(%q) = false, want true", code)
		}
	}
}

// TestIsNotFound_TrueForNoSuchCode verifies that IsNotFound returns true when
// the smithy APIError code contains "NoSuch".
func TestIsNotFound_TrueForNoSuchCode(t *testing.T) {
	for _, code := range []string{
		"NoSuchBucket",
		"NoSuchKey",
		"NoSuchEntity",
	} {
		err := &mockAPIError{code: code, message: "no such resource"}
		if !IsNotFound(err) {
			t.Errorf("IsNotFound(%q) = false, want true", code)
		}
	}
}

// TestIsNotFound_FalseForOtherCode verifies that IsNotFound returns false for
// non-not-found error codes.
func TestIsNotFound_FalseForOtherCode(t *testing.T) {
	err := &mockAPIError{code: "AccessDenied", message: "access denied"}
	if IsNotFound(err) {
		t.Error("IsNotFound(AccessDenied) = true, want false")
	}
}

// TestIsNotFound_FalseForNonSmithyError verifies that IsNotFound returns false
// for plain Go errors that do not implement smithy.APIError.
func TestIsNotFound_FalseForNonSmithyError(t *testing.T) {
	err := errors.New("some generic error")
	if IsNotFound(err) {
		t.Error("IsNotFound(plain error) = true, want false")
	}
}

// TestIsNotFound_WrappedSmithyError verifies that IsNotFound works even when
// the smithy error is wrapped by errors.Is/As traversal.
func TestIsNotFound_WrappedSmithyError(t *testing.T) {
	inner := &mockAPIError{code: "NotFoundException", message: "not found"}
	wrapped := errors.Join(inner)
	if !IsNotFound(wrapped) {
		t.Error("IsNotFound(wrapped NotFoundException) = false, want true")
	}
}

// ---------------------------------------------------------------------------
// IsAccessDenied tests
// ---------------------------------------------------------------------------

// TestIsAccessDenied_NilError verifies that IsAccessDenied returns false for nil.
func TestIsAccessDenied_NilError(t *testing.T) {
	if IsAccessDenied(nil) {
		t.Error("expected IsAccessDenied(nil) = false")
	}
}

// TestIsAccessDenied_TrueForAccessDeniedCode verifies that IsAccessDenied
// returns true when the smithy APIError code contains "AccessDenied".
func TestIsAccessDenied_TrueForAccessDeniedCode(t *testing.T) {
	for _, code := range []string{
		"AccessDenied",
		"AccessDeniedException",
	} {
		err := &mockAPIError{code: code, message: "access denied"}
		if !IsAccessDenied(err) {
			t.Errorf("IsAccessDenied(%q) = false, want true", code)
		}
	}
}

// TestIsAccessDenied_TrueForUnauthorized verifies that IsAccessDenied returns
// true for Unauthorized and similar codes.
func TestIsAccessDenied_TrueForUnauthorized(t *testing.T) {
	for _, code := range []string{
		"UnauthorizedOperation",
		"Unauthorized",
		"AuthorizationError",
	} {
		err := &mockAPIError{code: code, message: "unauthorized"}
		if !IsAccessDenied(err) {
			t.Errorf("IsAccessDenied(%q) = false, want true", code)
		}
	}
}

// TestIsAccessDenied_FalseForNotFoundCode verifies that IsAccessDenied returns
// false for not-found error codes.
func TestIsAccessDenied_FalseForNotFoundCode(t *testing.T) {
	err := &mockAPIError{code: "NotFoundException", message: "not found"}
	if IsAccessDenied(err) {
		t.Error("IsAccessDenied(NotFoundException) = true, want false")
	}
}

// TestIsAccessDenied_FalseForNonSmithyError verifies that IsAccessDenied
// returns false for plain Go errors.
func TestIsAccessDenied_FalseForNonSmithyError(t *testing.T) {
	err := errors.New("some generic error")
	if IsAccessDenied(err) {
		t.Error("IsAccessDenied(plain error) = true, want false")
	}
}
