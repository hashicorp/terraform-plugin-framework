package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Int64Validator is a function validator for types.Int64 parameters.
type Int64Validator interface {

	// Validate should perform the validation.
	Validate(context.Context, Int64Request, *Int64Response)
}

// Int64Request is a request for types.Int64 schema validation.
type Int64Request struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int64

	// Value contains the value of the argument for validation.
	Value types.Int64
}

// Int64Response is a response to a Int64Request.
type Int64Response struct {
	// Error is a function error generated during validation of the Value.
	Error *FuncError
}
