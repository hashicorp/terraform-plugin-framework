package proto6server

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ModifySchemaPlanRequest represents a request for a schema to run all
// attribute plan modification functions.
type ModifySchemaPlanRequest struct {
	// Config is the configuration the user supplied for the resource.
	Config tfsdk.Config

	// State is the current state of the resource.
	State tfsdk.State

	// Plan is the planned new state for the resource.
	Plan tfsdk.Plan

	// ProviderMeta is metadata from the provider_meta block of the module.
	ProviderMeta tfsdk.Config
}

// ModifySchemaPlanResponse represents a response to a ModifySchemaPlanRequest.
type ModifySchemaPlanResponse struct {
	// Plan is the planned new state for the resource.
	Plan tfsdk.Plan

	// RequiresReplace is a list of tftypes.AttributePaths that require the
	// resource to be replaced. They should point to the specific field
	// that changed that requires the resource to be destroyed and
	// recreated.
	RequiresReplace []*tftypes.AttributePath

	// Diagnostics report errors or warnings related to running all attribute
	// plan modifiers. Returning an empty slice indicates a successful
	// plan modification with no warnings or errors generated.
	Diagnostics diag.Diagnostics
}

// SchemaModifyPlan runs all AttributePlanModifiers in all schema attributes
// and blocks.
//
// TODO: Clean up this abstraction back into an internal Schema type method.
// The extra Schema parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/215
func SchemaModifyPlan(ctx context.Context, s tfsdk.Schema, req ModifySchemaPlanRequest, resp *ModifySchemaPlanResponse) {
	for name, attr := range s.Attributes {
		attrReq := tfsdk.ModifyAttributePlanRequest{
			AttributePath: tftypes.NewAttributePath().WithAttributeName(name),
			Config:        req.Config,
			State:         req.State,
			Plan:          req.Plan,
			ProviderMeta:  req.ProviderMeta,
		}

		AttributeModifyPlan(ctx, attr, attrReq, resp)
	}

	//nolint:staticcheck // Block support is required within the framework.
	for name, block := range s.Blocks {
		blockReq := tfsdk.ModifyAttributePlanRequest{
			AttributePath: tftypes.NewAttributePath().WithAttributeName(name),
			Config:        req.Config,
			State:         req.State,
			Plan:          req.Plan,
			ProviderMeta:  req.ProviderMeta,
		}

		BlockModifyPlan(ctx, block, blockReq, resp)
	}
}
