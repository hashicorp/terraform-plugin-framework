// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Number = &Number{}

// Declarative validator.Number for unit testing.
type Number struct {
	// Number interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateNumberMethod      func(context.Context, validator.NumberRequest, *validator.NumberResponse)
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

// Validate satisfies the validator.Number interface.
func (v Number) ValidateNumber(ctx context.Context, req validator.NumberRequest, resp *validator.NumberResponse) {
	if v.ValidateNumberMethod == nil {
		return
	}

	v.ValidateNumberMethod(ctx, req, resp)
}
