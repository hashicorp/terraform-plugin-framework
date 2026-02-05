// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

var _ defaults.Int32 = Int32{}

// Declarative defaults.Int32 for unit testing.
type Int32 struct {
	// defaults.Describer interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string

	// defaults.Int32 interface methods
	DefaultInt32Method func(context.Context, defaults.Int32Request, *defaults.Int32Response)
}

// Description satisfies the defaults.Describer interface.
func (v Int32) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the defaults.Describer interface.
func (v Int32) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// DefaultInt32 satisfies the defaults.Int32 interface.
func (v Int32) DefaultInt32(ctx context.Context, req defaults.Int32Request, resp *defaults.Int32Response) {
	if v.DefaultInt32Method == nil {
		return
	}

	v.DefaultInt32Method(ctx, req, resp)
}
