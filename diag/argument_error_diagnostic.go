// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diag

// NewArgumentErrorDiagnostic returns a new error severity diagnostic with the
// given summary, detail, and function argument.
func NewArgumentErrorDiagnostic(functionArgument int, summary string, detail string) DiagnosticWithFunctionArgument {
	return withFunctionArgument{
		Diagnostic:       NewErrorDiagnostic(summary, detail),
		functionArgument: functionArgument,
	}
}
