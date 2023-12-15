// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Definition is a function definition. Always set at least the Result field.
//
// NOTE: Provider-defined function support is in technical preview and offered
// without compatibility promises until Terraform 1.8 is generally available.
type Definition struct {
	// Parameters is the ordered list of function parameters and their
	// associated data types.
	Parameters []Parameter

	// VariadicParameter is an optional final parameter which can accept zero or
	// more arguments when the function is called. The argument data is sent as
	// an ordered list of the associated data type.
	VariadicParameter Parameter

	// Return is the function call response data type.
	Return Return

	// Summary is a short description of the function, preferably a single
	// sentence. Use the Description field for longer documentation about the
	// function and its implementation.
	Summary string

	// Description is the longer documentation for usage, such as editor
	// integrations, to give practitioners more information about the purpose of
	// the function and how its logic is implemented. It should be plaintext
	// formatted.
	Description string

	// MarkdownDescription is the longer documentation for usage, such as a
	// registry, to give practitioners more information about the purpose of the
	// function and how its logic is implemented.
	MarkdownDescription string

	// DeprecationMessage defines warning diagnostic details to display when
	// practitioner configurations use this function. The warning diagnostic
	// summary is automatically set to "Function Deprecated" along with
	// configuration source file and line information.
	DeprecationMessage string
}

// Parameter returns the Parameter for a given argument position. This may be
// from the Parameters field or, if defined, the VariadicParameter field. An
// error diagnostic is raised if the position is outside the expected arguments.
func (d Definition) Parameter(ctx context.Context, position int) (Parameter, diag.Diagnostics) {
	if d.VariadicParameter != nil && position >= len(d.Parameters) {
		return d.VariadicParameter, nil
	}

	if len(d.Parameters) == 0 {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid Parameter Position for Definition",
				"When determining the parameter for the given argument position, an invalid value was given. "+
					"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
					"Function does not implement parameters.\n"+
					fmt.Sprintf("Given position: %d", position),
			),
		}
	}

	if position >= len(d.Parameters) {
		return nil, diag.Diagnostics{
			diag.NewErrorDiagnostic(
				"Invalid Parameter Position for Definition",
				"When determining the parameter for the given argument position, an invalid value was given. "+
					"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
					fmt.Sprintf("Max argument position: %d\n", len(d.Parameters)-1)+
					fmt.Sprintf("Given position: %d", position),
			),
		}
	}

	return d.Parameters[position], nil
}

// ValidateImplementation contains logic for validating the provider-defined
// implementation of the definition to prevent unexpected errors or panics. This
// logic runs during the GetProviderSchema RPC, or via provider-defined unit
// testing, and should never include false positives.
func (d Definition) ValidateImplementation(ctx context.Context) diag.Diagnostics {
	var diags diag.Diagnostics

	if d.Return == nil {
		diags.AddError(
			"Invalid Function Definition",
			"When validating the function definition, an implementation issue was found. "+
				"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
				"Definition Return field is undefined",
		)
	} else if d.Return.GetType() == nil {
		diags.AddError(
			"Invalid Function Definition",
			"When validating the function definition, an implementation issue was found. "+
				"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
				"Definition return data type is undefined",
		)
	}

	return diags
}

// DefinitionRequest represents a request for the Function to return its
// definition, such as its ordered parameters and result. An instance of this
// request struct is supplied as an argument to the Function type Definition
// method.
type DefinitionRequest struct{}

// DefinitionResponse represents a response to a DefinitionRequest. An instance
// of this response struct is supplied as an argument to the Function type
// Definition method. Always set at least the Definition field.
type DefinitionResponse struct {
	// Definition is the function definition.
	Definition Definition

	// Diagnostics report errors or warnings related to defining the function.
	// An empty slice indicates success, with no warnings or errors generated.
	Diagnostics diag.Diagnostics
}
