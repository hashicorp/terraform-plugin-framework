// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

var _ defaults.Float64 = Float64{}

// Declarative defaults.Float64 for unit testing.
type Float64 struct {
	// defaults.Describer interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string

	// defaults.Float64 interface methods
	DefaultFloat64Method func(context.Context, defaults.Float64Request, *defaults.Float64Response)
}

// Description satisfies the defaults.Describer interface.
func (v Float64) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the defaults.Describer interface.
func (v Float64) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// DefaultFloat64 satisfies the defaults.Float64 interface.
func (v Float64) DefaultFloat64(ctx context.Context, req defaults.Float64Request, resp *defaults.Float64Response) {
	if v.DefaultFloat64Method == nil {
		return
	}

	v.DefaultFloat64Method(ctx, req, resp)
}
