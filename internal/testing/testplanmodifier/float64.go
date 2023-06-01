// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Float64 = &Float64{}

// Declarative planmodifier.Float64 for unit testing.
type Float64 struct {
	// Float64 interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	PlanModifyFloat64Method   func(context.Context, planmodifier.Float64Request, *planmodifier.Float64Response)
}

// Description satisfies the planmodifier.Float64 interface.
func (v Float64) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the planmodifier.Float64 interface.
func (v Float64) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// PlanModify satisfies the planmodifier.Float64 interface.
func (v Float64) PlanModifyFloat64(ctx context.Context, req planmodifier.Float64Request, resp *planmodifier.Float64Response) {
	if v.PlanModifyFloat64Method == nil {
		return
	}

	v.PlanModifyFloat64Method(ctx, req, resp)
}
