// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.Object                  = &Object{}
	_ function.ObjectParameterValidator = &Object{}
)

// Declarative validator.Object for unit testing.
type Object struct {
	// Object interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateObjectMethod      func(context.Context, validator.ObjectRequest, *validator.ObjectResponse)
	ValidateMethod            func(context.Context, function.ObjectParameterValidatorRequest, *function.ObjectParameterValidatorResponse)
}

// Description satisfies the validator.Object interface.
func (v Object) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.Object interface.
func (v Object) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// ValidateObject satisfies the validator.Object interface.
func (v Object) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if v.ValidateObjectMethod == nil {
		return
	}

	v.ValidateObjectMethod(ctx, req, resp)
}

// ValidateParameterObject satisfies the function.ObjectParameterValidator interface.
func (v Object) ValidateParameterObject(ctx context.Context, req function.ObjectParameterValidatorRequest, resp *function.ObjectParameterValidatorResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
