// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

var _ defaults.Number = Number{}

// Declarative defaults.Number for unit testing.
type Number struct {
	// defaults.Describer interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string

	// defaults.Number interface methods
	DefaultNumberMethod func(context.Context, defaults.NumberRequest, *defaults.NumberResponse)
}

// Description satisfies the defaults.Describer interface.
func (v Number) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the defaults.Describer interface.
func (v Number) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// DefaultNumber satisfies the defaults.Number interface.
func (v Number) DefaultNumber(ctx context.Context, req defaults.NumberRequest, resp *defaults.NumberResponse) {
	if v.DefaultNumberMethod == nil {
		return
	}

	v.DefaultNumberMethod(ctx, req, resp)
}
