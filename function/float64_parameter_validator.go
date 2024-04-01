package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Float64Validator is a function validator for types.Float64 parameters.
type Float64Validator interface {

	// Validate should perform the validation.
	Validate(context.Context, Float64Request, *Float64Response)
}

// Float64Request is a request for types.Float64 schema validation.
type Float64Request struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int64

	// Value contains the value of the argument for validation.
	Value types.Float64
}

// Float64Response is a response to a Float64Request.
type Float64Response struct {
	// Error is a function error generated during validation of the Value.
	Error *FuncError
}
