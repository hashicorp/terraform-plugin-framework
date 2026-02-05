// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.Int64                  = &Int64{}
	_ function.Int64ParameterValidator = &Int64{}
)

// Declarative validator.Int64 for unit testing.
type Int64 struct {
	// Int64 interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateInt64Method       func(context.Context, validator.Int64Request, *validator.Int64Response)
	ValidateMethod            func(context.Context, function.Int64ParameterValidatorRequest, *function.Int64ParameterValidatorResponse)
}

// Description satisfies the validator.Int64 interface.
func (v Int64) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.Int64 interface.
func (v Int64) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// ValidateInt64 satisfies the validator.Int64 interface.
func (v Int64) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if v.ValidateInt64Method == nil {
		return
	}

	v.ValidateInt64Method(ctx, req, resp)
}

// ValidateParameterInt64 satisfies the function.Int64ParameterValidator interface.
func (v Int64) ValidateParameterInt64(ctx context.Context, req function.Int64ParameterValidatorRequest, resp *function.Int64ParameterValidatorResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
