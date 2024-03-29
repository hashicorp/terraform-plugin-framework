package validator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Object is a function validator for types.Object parameters.
type Object interface {

	// ValidateObject should perform the validation.
	ValidateObject(context.Context, ObjectRequest, *ObjectResponse)
}

// ObjectRequest is a request for types.Object schema validation.
type ObjectRequest struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int

	// Value contains the value of the argument for validation.
	Value types.Object
}

// ObjectResponse is a response to a ObjectRequest.
type ObjectResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *function.FuncError
}
