// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwerror

// NewArgumentErrorFunctionError returns a new error severity function error with the
// given summary, detail, and function argument.
func NewArgumentErrorFunctionError(functionArgument int, summary string, detail string) FunctionErrorWithFunctionArgument {
	return withFunctionArgument{
		FunctionError:    NewErrorFunctionError(summary, detail),
		functionArgument: functionArgument,
	}
}
