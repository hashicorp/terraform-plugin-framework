// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

var _ defaults.String = String{}

// Declarative defaults.String for unit testing.
type String struct {
	// defaults.Describer interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string

	// defaults.String interface methods
	DefaultStringMethod func(context.Context, defaults.StringRequest, *defaults.StringResponse)
}

// Description satisfies the defaults.Describer interface.
func (v String) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the defaults.Describer interface.
func (v String) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// DefaultString satisfies the defaults.String interface.
func (v String) DefaultString(ctx context.Context, req defaults.StringRequest, resp *defaults.StringResponse) {
	if v.DefaultStringMethod == nil {
		return
	}

	v.DefaultStringMethod(ctx, req, resp)
}
