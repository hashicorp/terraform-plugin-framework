// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = &String{}

// Declarative validator.String for unit testing.
type String struct {
	// String interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateStringMethod      func(context.Context, validator.StringRequest, *validator.StringResponse)
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

// Validate satisfies the validator.String interface.
func (v String) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if v.ValidateStringMethod == nil {
		return
	}

	v.ValidateStringMethod(ctx, req, resp)
}
