// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

var _ defaults.Bool = Bool{}

// Declarative defaults.Bool for unit testing.
type Bool struct {
	// defaults.Describer interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string

	// defaults.Bool interface methods
	DefaultBoolMethod func(context.Context, defaults.BoolRequest, *defaults.BoolResponse)
}

// Description satisfies the defaults.Describer interface.
func (v Bool) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the defaults.Describer interface.
func (v Bool) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// DefaultBool satisfies the defaults.Bool interface.
func (v Bool) DefaultBool(ctx context.Context, req defaults.BoolRequest, resp *defaults.BoolResponse) {
	if v.DefaultBoolMethod == nil {
		return
	}

	v.DefaultBoolMethod(ctx, req, resp)
}
