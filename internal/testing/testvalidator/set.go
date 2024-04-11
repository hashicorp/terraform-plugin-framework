// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.Set                  = &Set{}
	_ function.SetParameterValidator = &Set{}
)

// Declarative validator.Set for unit testing.
type Set struct {
	// Set interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateSetMethod         func(context.Context, validator.SetRequest, *validator.SetResponse)
	ValidateMethod            func(context.Context, function.SetParameterValidatorRequest, *function.SetParameterValidatorResponse)
}

// Description satisfies the validator.Set interface.
func (v Set) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.Set interface.
func (v Set) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// ValidateSet satisfies the validator.Set interface.
func (v Set) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if v.ValidateSetMethod == nil {
		return
	}

	v.ValidateSetMethod(ctx, req, resp)
}

// ValidateParameterSet satisfies the function.SetParameterValidator interface.
func (v Set) ValidateParameterSet(ctx context.Context, req function.SetParameterValidatorRequest, resp *function.SetParameterValidatorResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
