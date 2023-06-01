// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.Set = &Set{}

// Declarative planmodifier.Set for unit testing.
type Set struct {
	// Set interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	PlanModifySetMethod       func(context.Context, planmodifier.SetRequest, *planmodifier.SetResponse)
}

// Description satisfies the planmodifier.Set interface.
func (v Set) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the planmodifier.Set interface.
func (v Set) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// PlanModify satisfies the planmodifier.Set interface.
func (v Set) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	if v.PlanModifySetMethod == nil {
		return
	}

	v.PlanModifySetMethod(ctx, req, resp)
}
