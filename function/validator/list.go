package validator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// List is a function validator for types.List parameters.
type List interface {

	// ValidateList should perform the validation.
	ValidateList(context.Context, ListRequest, *ListResponse)
}

// ListRequest is a request for types.List schema validation.
type ListRequest struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int

	// Value contains the value of the argument for validation.
	Value types.List
}

// ListResponse is a response to a ListRequest.
type ListResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *function.FuncError
}
