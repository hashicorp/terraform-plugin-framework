// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"errors"
)

var _ FunctionErrorWithFunctionArgument = argumentFunctionError{}

// NewArgumentFunctionError returns a new function error with the
// given message, and function argument.
func NewArgumentFunctionError(functionArgument int, msg string) FunctionErrorWithFunctionArgument {
	return argumentFunctionError{
		FunctionError:    NewFunctionError(msg),
		functionArgument: functionArgument,
	}
}

// argumentFunctionError wraps a function error with function argument information.
type argumentFunctionError struct {
	FunctionError

	functionArgument int
}

// Equal returns true if the other function error is wholly equivalent.
func (d argumentFunctionError) Equal(other FunctionError) bool {
	var o argumentFunctionError

	ok := errors.As(other, &o)

	if !ok {
		return false
	}

	if d.functionArgument != o.functionArgument {
		return false
	}

	if d.FunctionError == nil {
		return o.FunctionError == nil
	}

	return d.FunctionError.Equal(o.FunctionError)
}

// FunctionArgument returns the function argument.
func (d argumentFunctionError) FunctionArgument() int {
	return d.functionArgument
}

// WithFunctionArgument wraps a function error with function argument information
// or overwrites the function argument.
func WithFunctionArgument(functionArgument int, f FunctionError) FunctionErrorWithFunctionArgument {
	var afe argumentFunctionError

	ok := errors.As(f, &afe)

	if !ok {
		return argumentFunctionError{
			FunctionError:    f,
			functionArgument: functionArgument,
		}
	}

	afe.functionArgument = functionArgument

	return afe
}
