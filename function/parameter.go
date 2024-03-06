// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

const (
	// DefaultParameterNamePrefix is the prefix used to default the name of parameters which do not declare
	// a name. Use this to prevent Terraform errors for missing names. This prefix is used with the parameter
	// position in a function definition to create a unique name (param1, param2, etc.)
	DefaultParameterNamePrefix = "param"

	// DefaultVariadicParameterName is the default name given to a variadic parameter that does not declare
	// a name. Use this to prevent Terraform errors for missing names.
	DefaultVariadicParameterName = "varparam"
)

// Parameter is the interface for defining function parameters.
type Parameter interface {
	// GetAllowNullValue should return if the parameter accepts a null value.
	GetAllowNullValue() bool

	// GetAllowUnknownValues should return if the parameter accepts an unknown
	// value.
	GetAllowUnknownValues() bool

	// GetDescription should return the plaintext documentation for the
	// parameter.
	GetDescription() string

	// GetMarkdownDescription should return the Markdown documentation for the
	// parameter.
	GetMarkdownDescription() string

	// GetName should return a usage name for the parameter. Parameters are
	// positional, so this name has no meaning except documentation.
	//
	// If the name is returned as an empty string, a default name will be used to prevent Terraform errors for missing names.
	// The default name will be the prefix "param" with a suffix of the position the parameter is in the function definition. (`param1`, `param2`, etc.)
	// If the parameter is variadic, the default name will be `varparam`.
	GetName() string

	// GetType should return the data type for the parameter, which determines
	// what data type Terraform requires for configurations setting the argument
	// during a function call and the argument data type received by the
	// Function type Run method.
	GetType() attr.Type
}

// ParameterWithValidateImplementation is an optional interface on
// Parameter which enables validation of the provider-defined implementation
// for the Parameter. This logic runs during the GetProviderSchema RPC, or via
// provider-defined unit testing, to ensure the provider's definition is valid
// before further usage could cause other unexpected errors or panics.
type ParameterWithValidateImplementation interface {
	Parameter

	// ValidateImplementation should contain the logic which validates
	// the Parameter implementation. Since this logic can prevent the provider
	// from being usable, it should be very targeted and defensive against
	// false positives.
	ValidateImplementation(context.Context, ValidateParameterImplementationRequest, *ValidateParameterImplementationResponse)
}

// ValidateParameterImplementationRequest contains the information available
// during a ValidateImplementation call to validate the Parameter
// definition. ValidateParameterImplementationResponse is the type used for
// responses.
type ValidateParameterImplementationRequest struct {
	// FunctionArgument is the positional function argument for reporting diagnostics.
	FunctionArgument int64
}

// ValidateParameterImplementationResponse contains the returned data from a
// ValidateImplementation method call to validate the Parameter
// implementation. ValidateParameterImplementationRequest is the type used for
// requests.
type ValidateParameterImplementationResponse struct {
	// Diagnostics report errors or warnings related to validating the
	// definition of the Parameter. An empty slice indicates success, with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}
