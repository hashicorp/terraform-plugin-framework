// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BoolValidator is a function validator for types.Bool parameters.
type BoolValidator interface {

	// Validate should perform the validation.
	Validate(context.Context, BoolRequest, *BoolResponse)
}

// BoolRequest is a request for types.Bool schema validation.
type BoolRequest struct {
	// ArgumentPosition contains the position of the argument for validation.
	// Use this position for any response diagnostics.
	ArgumentPosition int64

	// Value contains the value of the argument for validation.
	Value types.Bool
}

// BoolResponse is a response to a BoolRequest.
type BoolResponse struct {
	// Error is a function error generated during validation of the Value.
	Error *FuncError
}
