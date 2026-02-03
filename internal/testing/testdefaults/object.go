// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testdefaults

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

var _ defaults.Object = Object{}

// Declarative defaults.Object for unit testing.
type Object struct {
	// defaults.Describer interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string

	// defaults.Object interface methods
	DefaultObjectMethod func(context.Context, defaults.ObjectRequest, *defaults.ObjectResponse)
}

// Description satisfies the defaults.Describer interface.
func (v Object) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the defaults.Describer interface.
func (v Object) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// DefaultObject satisfies the defaults.Object interface.
func (v Object) DefaultObject(ctx context.Context, req defaults.ObjectRequest, resp *defaults.ObjectResponse) {
	if v.DefaultObjectMethod == nil {
		return
	}

	v.DefaultObjectMethod(ctx, req, resp)
}
