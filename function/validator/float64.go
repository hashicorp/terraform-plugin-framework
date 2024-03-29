package validator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Float64 is a function validator for types.Float64 parameters.
type Float64 interface {

	// ValidateFloat64 should perform the validation.
	ValidateFloat64(context.Context, Float64Request, *Float64Response)
}

// Float64Request is a request for types.Float64 schema validation.
type Float64Request struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int

	// Value contains the value of the argument for validation.
	Value types.Float64
}

// Float64Response is a response to a Float64Request.
type Float64Response struct {
	// Error is a function error generated during validation of the Value.
	Error *function.FuncError
}
