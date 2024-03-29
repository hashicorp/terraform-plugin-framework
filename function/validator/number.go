package validator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Number is a function validator for types.Number parameters.
type Number interface {

	// ValidateNumber should perform the validation.
	ValidateNumber(context.Context, NumberRequest, *NumberResponse)
}

// NumberRequest is a request for types.Number schema validation.
type NumberRequest struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int

	// Value contains the value of the argument for validation.
	Value types.Number
}

// NumberResponse is a response to a NumberRequest.
type NumberResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *function.FuncError
}
