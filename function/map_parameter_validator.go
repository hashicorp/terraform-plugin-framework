package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// MapValidator is a function validator for types.Map parameters.
type MapValidator interface {

	// Validate should perform the validation.
	Validate(context.Context, MapRequest, *MapResponse)
}

// MapRequest is a request for types.Map schema validation.
type MapRequest struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int64

	// Value contains the value of the argument for validation.
	Value types.Map
}

// MapResponse is a response to a MapRequest.
type MapResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *FuncError
}
