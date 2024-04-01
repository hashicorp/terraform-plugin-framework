package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// NumberValidator is a function validator for types.Number parameters.
type NumberValidator interface {

	// Validate should perform the validation.
	Validate(context.Context, NumberRequest, *NumberResponse)
}

// NumberRequest is a request for types.Number schema validation.
type NumberRequest struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int64

	// Value contains the value of the argument for validation.
	Value types.Number
}

// NumberResponse is a response to a NumberRequest.
type NumberResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *FuncError
}
