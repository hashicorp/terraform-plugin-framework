// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

var _ defaults.Map = Map{}

// Declarative defaults.Map for unit testing.
type Map struct {
	// defaults.Describer interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string

	// defaults.Map interface methods
	DefaultMapMethod func(context.Context, defaults.MapRequest, *defaults.MapResponse)
}

// Description satisfies the defaults.Describer interface.
func (v Map) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the defaults.Describer interface.
func (v Map) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// DefaultMap satisfies the defaults.Map interface.
func (v Map) DefaultMap(ctx context.Context, req defaults.MapRequest, resp *defaults.MapResponse) {
	if v.DefaultMapMethod == nil {
		return
	}

	v.DefaultMapMethod(ctx, req, resp)
}
