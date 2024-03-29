// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var (
	_ validator.Map         = &Map{}
	_ function.MapValidator = &Map{}
)

// Declarative validator.Map for unit testing.
type Map struct {
	// Map interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateMapMethod         func(context.Context, validator.MapRequest, *validator.MapResponse)
	ValidateMethod            func(context.Context, function.MapRequest, *function.MapResponse)
}

// Description satisfies the validator.Map interface.
func (v Map) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.Map interface.
func (v Map) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// ValidateMap satisfies the validator.Map interface.
func (v Map) ValidateMap(ctx context.Context, req validator.MapRequest, resp *validator.MapResponse) {
	if v.ValidateMapMethod == nil {
		return
	}

	v.ValidateMapMethod(ctx, req, resp)
}

// Validate satisfies the function.MapValidator interface.
func (v Map) Validate(ctx context.Context, req function.MapRequest, resp *function.MapResponse) {
	if v.ValidateMethod == nil {
		return
	}

	v.ValidateMethod(ctx, req, resp)
}
