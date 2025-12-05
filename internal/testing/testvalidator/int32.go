// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.Int32                  = &Int32{}
	_ function.Int32ParameterValidator = &Int32{}
)

// Declarative validator.Int32 for unit testing.
type Int32 struct {
	// Int32 interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateInt32Method       func(context.Context, validator.Int32Request, *validator.Int32Response)
	ValidateMethod            func(context.Context, function.Int32ParameterValidatorRequest, *function.Int32ParameterValidatorResponse)
}

// Description satisfies the validator.Int32 interface.
func (v Int32) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.Int32 interface.
func (v Int32) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// ValidateInt32 satisfies the validator.Int32 interface.
func (v Int32) ValidateInt32(ctx context.Context, req validator.Int32Request, resp *validator.Int32Response) {
	if v.ValidateInt32Method == nil {
		return
	}

	v.ValidateInt32Method(ctx, req, resp)
}

// ValidateParameterInt32 satisfies the function.Int32ParameterValidator interface.
func (v Int32) ValidateParameterInt32(ctx context.Context, req function.Int32ParameterValidatorRequest, resp *function.Int32ParameterValidatorResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
