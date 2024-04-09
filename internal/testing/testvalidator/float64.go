// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.Float64                  = &Float64{}
	_ function.Float64ParameterValidator = &Float64{}
)

// Declarative validator.Float64 for unit testing.
type Float64 struct {
	// Float64 interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateFloat64Method     func(context.Context, validator.Float64Request, *validator.Float64Response)
	ValidateMethod            func(context.Context, function.Float64ParameterValidatorRequest, *function.Float64ParameterValidatorResponse)
}

// Description satisfies the validator.Float64 interface.
func (v Float64) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.Float64 interface.
func (v Float64) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// ValidateFloat64 satisfies the validator.Float64 interface.
func (v Float64) ValidateFloat64(ctx context.Context, req validator.Float64Request, resp *validator.Float64Response) {
	if v.ValidateFloat64Method == nil {
		return
	}

	v.ValidateFloat64Method(ctx, req, resp)
}

// Validate satisfies the function.Float64ParameterValidator interface.
func (v Float64) Validate(ctx context.Context, req function.Float64ParameterValidatorRequest, resp *function.Float64ParameterValidatorResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
