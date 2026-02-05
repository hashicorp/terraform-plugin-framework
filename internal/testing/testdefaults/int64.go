// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

var _ defaults.Int64 = Int64{}

// Declarative defaults.Int64 for unit testing.
type Int64 struct {
	// defaults.Describer interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string

	// defaults.Int64 interface methods
	DefaultInt64Method func(context.Context, defaults.Int64Request, *defaults.Int64Response)
}

// Description satisfies the defaults.Describer interface.
func (v Int64) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the defaults.Describer interface.
func (v Int64) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// DefaultInt64 satisfies the defaults.Int64 interface.
func (v Int64) DefaultInt64(ctx context.Context, req defaults.Int64Request, resp *defaults.Int64Response) {
	if v.DefaultInt64Method == nil {
		return
	}

	v.DefaultInt64Method(ctx, req, resp)
}
