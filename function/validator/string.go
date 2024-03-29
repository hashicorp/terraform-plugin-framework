package validator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// String is a function validator for types.String parameters.
type String interface {

	// ValidateString should perform the validation.
	ValidateString(context.Context, StringRequest, *StringResponse)
}

// StringRequest is a request for types.String schema validation.
type StringRequest struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int

	// Value contains the value of the argument for validation.
	Value types.String
}

// StringResponse is a response to a StringRequest.
type StringResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *function.FuncError
}
