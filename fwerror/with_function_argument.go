// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwerror

import (
	"errors"
)

var _ FunctionErrorWithFunctionArgument = withFunctionArgument{}

// withFunctionArgument wraps a function error with function argument information.
type withFunctionArgument struct {
	FunctionError

	functionArgument int
}

// Equal returns true if the other function error is wholly equivalent.
func (d withFunctionArgument) Equal(other FunctionError) bool {
	var o withFunctionArgument

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

// FunctionArgument returns the diagnostic function argument.
func (d withFunctionArgument) FunctionArgument() int {
	return d.functionArgument
}

// WithFunctionArgument wraps a function error with function argument information
// or overwrites the function argument.
func WithFunctionArgument(functionArgument int, f FunctionError) FunctionErrorWithFunctionArgument {
	var wp withFunctionArgument

	ok := errors.As(f, &wp)

	if !ok {
		return withFunctionArgument{
			FunctionError:    f,
			functionArgument: functionArgument,
		}
	}

	wp.functionArgument = functionArgument

	return wp
}
