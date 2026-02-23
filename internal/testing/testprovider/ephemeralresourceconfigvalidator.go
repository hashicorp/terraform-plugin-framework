// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
)

var _ ephemeral.ConfigValidator = &EphemeralResourceConfigValidator{}

// Declarative ephemeral.ConfigValidator for unit testing.
type EphemeralResourceConfigValidator struct {
	// EphemeralResourceConfigValidator interface methods
	DescriptionMethod               func(context.Context) string
	MarkdownDescriptionMethod       func(context.Context) string
	ValidateEphemeralResourceMethod func(context.Context, ephemeral.ValidateConfigRequest, *ephemeral.ValidateConfigResponse)
}

// Description satisfies the ephemeral.ConfigValidator interface.
func (v *EphemeralResourceConfigValidator) Description(ctx context.Context) string {
	if v.DescriptionMethod == nil {
		return ""
	}

	return v.DescriptionMethod(ctx)
}

// MarkdownDescription satisfies the ephemeral.ConfigValidator interface.
func (v *EphemeralResourceConfigValidator) MarkdownDescription(ctx context.Context) string {
	if v.MarkdownDescriptionMethod == nil {
		return ""
	}

	return v.MarkdownDescriptionMethod(ctx)
}

// Validate satisfies the ephemeral.ConfigValidator interface.
func (v *EphemeralResourceConfigValidator) ValidateEphemeralResource(ctx context.Context, req ephemeral.ValidateConfigRequest, resp *ephemeral.ValidateConfigResponse) {
	if v.ValidateEphemeralResourceMethod == nil {
		return
	}

	v.ValidateEphemeralResourceMethod(ctx, req, resp)
}
