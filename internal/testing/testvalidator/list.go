// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.List         = &List{}
	_ function.ListValidator = &List{}
)

// Declarative validator.List for unit testing.
type List struct {
	// List interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateListMethod        func(context.Context, validator.ListRequest, *validator.ListResponse)
	ValidateMethod            func(context.Context, function.ListRequest, *function.ListResponse)
}

// Description satisfies the validator.List interface.
func (v List) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.List interface.
func (v List) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// ValidateList satisfies the validator.List interface.
func (v List) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if v.ValidateListMethod == nil {
		return
	}

	v.ValidateListMethod(ctx, req, resp)
}

// Validate satisfies the function.ListValidator interface.
func (v List) Validate(ctx context.Context, req function.ListRequest, resp *function.ListResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
