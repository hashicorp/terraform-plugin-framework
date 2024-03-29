package validator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Bool is a function validator for types.Bool parameters.
type Bool interface {

	// ValidateBool should perform the validation.
	ValidateBool(context.Context, BoolRequest, *BoolResponse)
}

// BoolRequest is a request for types.Bool schema validation.
type BoolRequest struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int

	// Value contains the value of the argument for validation.
	Value types.Bool
}

// BoolResponse is a response to a BoolRequest.
type BoolResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *function.FuncError
}
