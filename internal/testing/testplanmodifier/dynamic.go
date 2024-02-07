// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Dynamic = &Dynamic{}

// Declarative planmodifier.Dynamic for unit testing.
type Dynamic struct {
	// Dynamic interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	PlanModifyDynamicMethod   func(context.Context, planmodifier.DynamicRequest, *planmodifier.DynamicResponse)
}

// Description satisfies the planmodifier.Dynamic interface.
func (v Dynamic) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the planmodifier.Dynamic interface.
func (v Dynamic) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// PlanModify satisfies the planmodifier.Dynamic interface.
func (v Dynamic) PlanModifyDynamic(ctx context.Context, req planmodifier.DynamicRequest, resp *planmodifier.DynamicResponse) {
	if v.PlanModifyDynamicMethod == nil {
		return
	}

	v.PlanModifyDynamicMethod(ctx, req, resp)
}
