// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// Return is the interface for defining function return data.
type Return interface {
	// GetType should return the data type for the return, which determines
	// what data type Terraform requires for configurations receiving the
	// response of a function call and the return data type required from the
	// Function type Run method.
	GetType() attr.Type

	// NewResultData should return a new ResultData with an unknown value (or
	// best approximation of an invalid value) of the corresponding data type.
	// The Function type Run method is expected to overwrite the value before
	// returning.
	NewResultData(context.Context) (ResultData, *FuncError)
}

// ReturnWithValidateImplementation is an optional interface on
// Return which enables validation of the provider-defined implementation
// for the Return. This logic runs during the GetProviderSchema RPC, or via
// provider-defined unit testing, to ensure the provider's definition is valid
// before further usage could cause other unexpected errors or panics.
type ReturnWithValidateImplementation interface {
	Return

	// ValidateImplementation should contain the logic which validates
	// the Return implementation. Since this logic can prevent the provider
	// from being usable, it should be very targeted and defensive against
	// false positives.
	ValidateImplementation(context.Context, ValidateReturnImplementationRequest, *ValidateReturnImplementationResponse)
}

// ValidateReturnImplementationRequest contains the information available
// during a ValidateImplementation call to validate the Return
// definition. ValidateReturnImplementationResponse is the type used for
// responses.
type ValidateReturnImplementationRequest struct{}

// ValidateReturnImplementationResponse contains the returned data from a
// ValidateImplementation method call to validate the Return
// implementation. ValidateReturnImplementationRequest is the type used for
// requests.
type ValidateReturnImplementationResponse struct {
	// Diagnostics report errors or warnings related to validating the
	// definition of the Return. An empty slice indicates success, with no
	// warnings or errors generated.
	Diagnostics diag.Diagnostics
}
