// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testvalidator

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Int64 = &Int64{}

// Declarative validator.Int64 for unit testing.
type Int64 struct {
	// Int64 interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateInt64Method       func(context.Context, validator.Int64Request, *validator.Int64Response)
}

// Description satisfies the validator.Int64 interface.
func (v Int64) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the validator.Int64 interface.
func (v Int64) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the validator.Int64 interface.
func (v Int64) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if v.ValidateInt64Method == nil {
		return
	}

	v.ValidateInt64Method(ctx, req, resp)
}
