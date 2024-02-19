// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwerror

import (
	"errors"
)

// NewFunctionError returns a new function error with the
// given message.
func NewFunctionError(msg string) FunctionError {
	return &functionError{
		msg: msg,
	}
}

type functionError struct {
	msg string
}

// Equal returns true if the other function error is wholly equivalent.
func (fe *functionError) Equal(other FunctionError) bool {
	var funcErr *functionError

	ok := errors.As(other, &funcErr)

	if !ok {
		return false
	}

	if fe == nil && funcErr == nil {
		return true
	}

	if fe == nil || funcErr == nil {
		return false
	}

	return fe.Error() == funcErr.Error()
}

// Error returns the function error message.
func (fe *functionError) Error() string {
	return fe.msg
}
