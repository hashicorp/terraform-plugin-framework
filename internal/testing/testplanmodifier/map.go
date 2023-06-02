// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Map = &Map{}

// Declarative planmodifier.Map for unit testing.
type Map struct {
	// Map interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	PlanModifyMapMethod       func(context.Context, planmodifier.MapRequest, *planmodifier.MapResponse)
}

// Description satisfies the planmodifier.Map interface.
func (v Map) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the planmodifier.Map interface.
func (v Map) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// PlanModify satisfies the planmodifier.Map interface.
func (v Map) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if v.PlanModifyMapMethod == nil {
		return
	}

	v.PlanModifyMapMethod(ctx, req, resp)
}
