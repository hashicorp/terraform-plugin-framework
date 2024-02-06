// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Dynamic = &Dynamic{}

// Declarative validator.Dynamic for unit testing.
type Dynamic struct {
	// Dynamic interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateDynamicMethod     func(context.Context, validator.DynamicRequest, *validator.DynamicResponse)
}

// Description satisfies the validator.Dynamic interface.
func (v Dynamic) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.Dynamic interface.
func (v Dynamic) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the validator.Dynamic interface.
func (v Dynamic) ValidateDynamic(ctx context.Context, req validator.DynamicRequest, resp *validator.DynamicResponse) {
	if v.ValidateDynamicMethod == nil {
		return
	}

	v.ValidateDynamicMethod(ctx, req, resp)
}
