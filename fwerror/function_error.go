// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwerror

// FunctionError is an interface for errors generated during the execution
// of provider-defined functions.
//
// See the NewFunctionError constructor for a generic implementation.
//
// To add argument position information to an existing function error,
// see the WithFunctionArgument function.
type FunctionError interface {
	// Equal returns true if the other function error is wholly equivalent.
	Equal(FunctionError) bool

	error
}

// FunctionErrorWithFunctionArgument is a function error associated with a
// function argument.
type FunctionErrorWithFunctionArgument interface {
	FunctionError

	// FunctionArgument points to a specific function argument position.
	FunctionArgument() int
}
