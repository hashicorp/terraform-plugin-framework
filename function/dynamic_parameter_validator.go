package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// DynamicValidator is a function validator for types.Dynamic parameters.
type DynamicValidator interface {

	// Validate should perform the validation.
	Validate(context.Context, DynamicRequest, *DynamicResponse)
}

// DynamicRequest is a request for types.Dynamic schema validation.
type DynamicRequest struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int

	// Value contains the value of the argument for validation.
	Value types.Dynamic
}

// DynamicResponse is a response to a DynamicRequest.
type DynamicResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *FuncError
}
