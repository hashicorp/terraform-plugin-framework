package validator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Set is a function validator for types.Set parameters.
type Set interface {

	// ValidateSet should perform the validation.
	ValidateSet(context.Context, SetRequest, *SetResponse)
}

// SetRequest is a request for types.Set schema validation.
type SetRequest struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int

	// Value contains the value of the argument for validation.
	Value types.Set
}

// SetResponse is a response to a SetRequest.
type SetResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *function.FuncError
}
