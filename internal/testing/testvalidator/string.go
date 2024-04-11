// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.String                  = &String{}
	_ function.StringParameterValidator = &String{}
)

// Declarative validator.String for unit testing.
type String struct {
	// String interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateStringMethod      func(context.Context, validator.StringRequest, *validator.StringResponse)
	ValidateMethod            func(context.Context, function.StringParameterValidatorRequest, *function.StringParameterValidatorResponse)
}

// Description satisfies the validator.String interface.
func (v String) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.String interface.
func (v String) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// ValidateString satisfies the validator.String interface.
func (v String) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if v.ValidateStringMethod == nil {
		return
	}

	v.ValidateStringMethod(ctx, req, resp)
}

// ValidateParameterString satisfies the function.StringParameterValidator interface.
func (v String) ValidateParameterString(ctx context.Context, req function.StringParameterValidatorRequest, resp *function.StringParameterValidatorResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
