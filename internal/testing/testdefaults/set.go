// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

var _ defaults.Set = Set{}

// Declarative defaults.Set for unit testing.
type Set struct {
	// defaults.Describer interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string

	// defaults.Set interface methods
	DefaultSetMethod func(context.Context, defaults.SetRequest, *defaults.SetResponse)
}

// Description satisfies the defaults.Describer interface.
func (v Set) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the defaults.Describer interface.
func (v Set) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// DefaultSet satisfies the defaults.Set interface.
func (v Set) DefaultSet(ctx context.Context, req defaults.SetRequest, resp *defaults.SetResponse) {
	if v.DefaultSetMethod == nil {
		return
	}

	v.DefaultSetMethod(ctx, req, resp)
}
