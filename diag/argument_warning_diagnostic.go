// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diag

// NewArgumentWarningDiagnostic returns a new warning severity diagnostic with
// the given summary, detail, and function argument.
func NewArgumentWarningDiagnostic(functionArgument int, summary string, detail string) DiagnosticWithFunctionArgument {
	return withFunctionArgument{
		Diagnostic:       NewWarningDiagnostic(summary, detail),
		functionArgument: functionArgument,
	}
}
