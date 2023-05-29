// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Bool = &Bool{}

// Declarative validator.Bool for unit testing.
type Bool struct {
	// Bool interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateBoolMethod        func(context.Context, validator.BoolRequest, *validator.BoolResponse)
}

// Description satisfies the validator.Bool interface.
func (v Bool) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.Bool interface.
func (v Bool) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the validator.Bool interface.
func (v Bool) ValidateBool(ctx context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if v.ValidateBoolMethod == nil {
		return
	}

	v.ValidateBoolMethod(ctx, req, resp)
}
