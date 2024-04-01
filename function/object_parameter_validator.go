package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ObjectValidator is a function validator for types.Object parameters.
type ObjectValidator interface {

	// Validate should perform the validation.
	Validate(context.Context, ObjectRequest, *ObjectResponse)
}

// ObjectRequest is a request for types.Object schema validation.
type ObjectRequest struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int64

	// Value contains the value of the argument for validation.
	Value types.Object
}

// ObjectResponse is a response to a ObjectRequest.
type ObjectResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *FuncError
}
