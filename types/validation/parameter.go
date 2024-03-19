// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validation

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
)

// ValidateableParameter defines an interface for validating a parameter value.
type ValidateableParameter interface {
	// ValidateParameter returns any error generated during validation
	// of  the parameter. It is generally used to check the data format and ensure
	// that it complies with the requirements of the Value.
	ValidateParameter(context.Context, ValidateParameterRequest, *ValidateParameterResponse)
}

// ValidateParameterRequest represents a request for the Value to call its
// validation logic. An instance of this request struct is supplied as an
// argument to the Value type ValidateParameter method.
type ValidateParameterRequest struct {
	// Position is the zero-ordered position of the parameter being validated.
	Position int64
}

// ValidateParameterResponse represents a response to a ValidateParameterRequest.
// An instance of this response struct is supplied as an argument to the
// ValidateParameter method.
type ValidateParameterResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *function.FuncError
}
