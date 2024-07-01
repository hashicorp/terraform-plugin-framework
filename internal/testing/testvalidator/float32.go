// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.Float32                  = &Float32{}
	_ function.Float32ParameterValidator = &Float32{}
)

// Declarative validator.Float32 for unit testing.
type Float32 struct {
	// Float32 interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateFloat32Method     func(context.Context, validator.Float32Request, *validator.Float32Response)
	ValidateMethod            func(context.Context, function.Float32ParameterValidatorRequest, *function.Float32ParameterValidatorResponse)
}

// Description satisfies the validator.Float32 interface.
func (v Float32) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.Float32 interface.
func (v Float32) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// ValidateFloat32 satisfies the validator.Float32 interface.
func (v Float32) ValidateFloat32(ctx context.Context, req validator.Float32Request, resp *validator.Float32Response) {
	if v.ValidateFloat32Method == nil {
		return
	}

	v.ValidateFloat32Method(ctx, req, resp)
}

// ValidateParameterFloat32 satisfies the function.Float32ParameterValidator interface.
func (v Float32) ValidateParameterFloat32(ctx context.Context, req function.Float32ParameterValidatorRequest, resp *function.Float32ParameterValidatorResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
