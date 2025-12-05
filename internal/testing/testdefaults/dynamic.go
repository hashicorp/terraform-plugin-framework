// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

var _ defaults.Dynamic = Dynamic{}

// Declarative defaults.Dynamic for unit testing.
type Dynamic struct {
	// defaults.Describer interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string

	// defaults.Dynamic interface methods
	DefaultDynamicMethod func(context.Context, defaults.DynamicRequest, *defaults.DynamicResponse)
}

// Description satisfies the defaults.Describer interface.
func (v Dynamic) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the defaults.Describer interface.
func (v Dynamic) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// DefaultDynamic satisfies the defaults.Dynamic interface.
func (v Dynamic) DefaultDynamic(ctx context.Context, req defaults.DynamicRequest, resp *defaults.DynamicResponse) {
	if v.DefaultDynamicMethod == nil {
		return
	}

	v.DefaultDynamicMethod(ctx, req, resp)
}
