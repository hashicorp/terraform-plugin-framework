// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.String = &String{}

// Declarative planmodifier.String for unit testing.
type String struct {
	// String interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	PlanModifyStringMethod    func(context.Context, planmodifier.StringRequest, *planmodifier.StringResponse)
}

// Description satisfies the planmodifier.String interface.
func (v String) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the planmodifier.String interface.
func (v String) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// PlanModify satisfies the planmodifier.String interface.
func (v String) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if v.PlanModifyStringMethod == nil {
		return
	}

	v.PlanModifyStringMethod(ctx, req, resp)
}
