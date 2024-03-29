package validator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Map is a function validator for types.Map parameters.
type Map interface {

	// ValidateMap should perform the validation.
	ValidateMap(context.Context, MapRequest, *MapResponse)
}

// MapRequest is a request for types.Map schema validation.
type MapRequest struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int

	// Value contains the value of the argument for validation.
	Value types.Map
}

// MapResponse is a response to a MapRequest.
type MapResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *function.FuncError
}
