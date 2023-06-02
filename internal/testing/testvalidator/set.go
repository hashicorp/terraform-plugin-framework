// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Set = &Set{}

// Declarative validator.Set for unit testing.
type Set struct {
	// Set interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateSetMethod         func(context.Context, validator.SetRequest, *validator.SetResponse)
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

// Validate satisfies the validator.Set interface.
func (v Set) ValidateSet(ctx context.Context, req validator.SetRequest, resp *validator.SetResponse) {
	if v.ValidateSetMethod == nil {
		return
	}

	v.ValidateSetMethod(ctx, req, resp)
}
