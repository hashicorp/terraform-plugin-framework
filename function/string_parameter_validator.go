package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringValidator is a function validator for types.String parameters.
type StringValidator interface {

	// ValidateString should perform the validation.
	Validate(context.Context, StringRequest, *StringResponse)
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
	Error *FuncError
}
