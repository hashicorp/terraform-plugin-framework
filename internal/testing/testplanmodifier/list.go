// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testplanmodifier

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
)

var _ planmodifier.List = &List{}

// Declarative planmodifier.List for unit testing.
type List struct {
	// List interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	PlanModifyListMethod      func(context.Context, planmodifier.ListRequest, *planmodifier.ListResponse)
}

// Description satisfies the planmodifier.List interface.
func (v List) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the planmodifier.List interface.
func (v List) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// PlanModify satisfies the planmodifier.List interface.
func (v List) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	if v.PlanModifyListMethod == nil {
		return
	}

	v.PlanModifyListMethod(ctx, req, resp)
}
