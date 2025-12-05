// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/action"
)

var _ action.ConfigValidator = &ActionConfigValidator{}

// Declarative action.ConfigValidator for unit testing.
type ActionConfigValidator struct {
	// ActionConfigValidator interface methods
	DescriptionMethod         func(context.Context) string
	MarkdownDescriptionMethod func(context.Context) string
	ValidateActionMethod      func(context.Context, action.ValidateConfigRequest, *action.ValidateConfigResponse)
}

// Description satisfies the action.ConfigValidator interface.
func (v *ActionConfigValidator) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the action.ConfigValidator interface.
func (v *ActionConfigValidator) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the action.ConfigValidator interface.
func (v *ActionConfigValidator) ValidateAction(ctx context.Context, req action.ValidateConfigRequest, resp *action.ValidateConfigResponse) {
	if v.ValidateActionMethod == nil {
		return
	}

	v.ValidateActionMethod(ctx, req, resp)
}
