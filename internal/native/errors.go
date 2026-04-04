// SPDX-FileCopyrightText: 2024 The Crossplane Authors <https://crossplane.io>
//
// SPDX-License-Identifier: Apache-2.0

package native

import (
	"errors"
	"strings"

	smithy "github.com/aws/smithy-go"
	xperrors "github.com/crossplane/crossplane-runtime/v2/pkg/errors"
)

// Wrap annotates err with a stack trace at the point Wrap was called and a
// message. Wrap returns nil when err is nil.
func Wrap(err error, msg string) error {
	return xperrors.Wrap(err, msg)
}

// IsNotFound returns true when err represents an AWS "not found" error.
// It inspects the error chain for a smithy.APIError and checks the error code
// against common not-found patterns used across AWS services.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var ae smithy.APIError
	if !errors.As(err, &ae) {
		return false
	}
	code := ae.ErrorCode()
	return strings.Contains(code, "NotFound") ||
		strings.Contains(code, "NoSuch") ||
		code == "404" ||
		code == "ResourceNotFound"
}

// IsAccessDenied returns true when err represents an AWS "access denied" or
// "unauthorized" error. It inspects the error chain for a smithy.APIError and
// checks the error code against common access-denied patterns.
func IsAccessDenied(err error) bool {
	if err == nil {
		return false
	}
	var ae smithy.APIError
	if !errors.As(err, &ae) {
		return false
	}
	code := ae.ErrorCode()
	return strings.Contains(code, "AccessDenied") ||
		strings.Contains(code, "Unauthorized") ||
		strings.Contains(code, "AuthorizationError") ||
		code == "403" ||
		code == "Forbidden"
}

// IsErrorCode returns true when err is an AWS API error with the given code.
// This is useful for service-specific error codes that are not covered by
// IsNotFound or IsAccessDenied.
func IsErrorCode(err error, code string) bool {
	if err == nil {
		return false
	}
	var ae smithy.APIError
	if !errors.As(err, &ae) {
		return false
	}
	return ae.ErrorCode() == code
}
