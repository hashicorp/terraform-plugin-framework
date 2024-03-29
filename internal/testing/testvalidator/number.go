// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.Number         = &Number{}
	_ function.NumberValidator = &Number{}
)

// Declarative validator.Number for unit testing.
type Number struct {
	// Number interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateNumberMethod      func(context.Context, validator.NumberRequest, *validator.NumberResponse)
	ValidateMethod            func(context.Context, function.NumberRequest, *function.NumberResponse)
}

// Description satisfies the validator.Number interface.
func (v Number) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.Number interface.
func (v Number) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// ValidateNumber satisfies the validator.Number interface.
func (v Number) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	if v.ValidateNumberMethod == nil {
		return
	}

	v.ValidateNumberMethod(ctx, req, resp)
}

// Validate satisfies the function.NumberValidator interface.
func (v Number) Validate(ctx context.Context, req function.NumberRequest, resp *function.NumberResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
