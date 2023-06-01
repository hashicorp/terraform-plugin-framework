// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Object = &Object{}

// Declarative validator.Object for unit testing.
type Object struct {
	// Object interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateObjectMethod      func(context.Context, validator.ObjectRequest, *validator.ObjectResponse)
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

// Validate satisfies the validator.Object interface.
func (v Object) ValidateObject(ctx context.Context, req validator.ObjectRequest, resp *validator.ObjectResponse) {
	if v.ValidateObjectMethod == nil {
		return
	}

	v.ValidateObjectMethod(ctx, req, resp)
}
