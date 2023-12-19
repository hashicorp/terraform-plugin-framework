// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package diag

var _ DiagnosticWithFunctionArgument = withFunctionArgument{}

// withFunctionArgument wraps a diagnostic with function argument information.
type withFunctionArgument struct {
	Diagnostic

	functionArgument int
}

// Equal returns true if the other diagnostic is wholly equivalent.
func (d withFunctionArgument) Equal(other Diagnostic) bool {
	o, ok := other.(withFunctionArgument)

	if !ok {
		return false
	}

	if d.functionArgument != o.functionArgument {
		return false
	}

	if d.Diagnostic == nil {
		return d.Diagnostic == o.Diagnostic
	}

	return d.Diagnostic.Equal(o.Diagnostic)
}

// FunctionArgument returns the diagnostic function argument.
func (d withFunctionArgument) FunctionArgument() int {
	return d.functionArgument
}

// WithFunctionArgument wraps a diagnostic with function argument information
// or overwrites the function argument.
func WithFunctionArgument(functionArgument int, d Diagnostic) DiagnosticWithFunctionArgument {
	wp, ok := d.(withFunctionArgument)

	if !ok {
		return withFunctionArgument{
			Diagnostic:       d,
			functionArgument: functionArgument,
		}
	}

	wp.functionArgument = functionArgument

	return wp
}
