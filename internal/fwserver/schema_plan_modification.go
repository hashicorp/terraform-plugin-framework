package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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

	// Private is provider private state data.
	Private *privatestate.ProviderData
}

// ModifySchemaPlanResponse represents a response to a ModifySchemaPlanRequest.
type ModifySchemaPlanResponse struct {
	// Plan is the planned new state for the resource.
	Plan tfsdk.Plan

	// RequiresReplace is a list of attribute paths that require the
	// resource to be replaced. They should point to the specific field
	// that changed that requires the resource to be destroyed and
	// recreated.
	RequiresReplace path.Paths

	// Private is provider private state data following potential modifications.
	Private *privatestate.ProviderData

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
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/365
func SchemaModifyPlan(ctx context.Context, s fwschema.Schema, req ModifySchemaPlanRequest, resp *ModifySchemaPlanResponse) {
	for name, attribute := range s.GetAttributes() {
		attrReq := tfsdk.ModifyAttributePlanRequest{
			AttributePath: path.Root(name),
			Config:        req.Config,
			State:         req.State,
			Plan:          req.Plan,
			ProviderMeta:  req.ProviderMeta,
			Private:       req.Private,
		}

		AttributeModifyPlan(ctx, attribute, attrReq, resp)
	}

	for name, block := range s.GetBlocks() {
		blockReq := tfsdk.ModifyAttributePlanRequest{
			AttributePath: path.Root(name),
			Config:        req.Config,
			State:         req.State,
			Plan:          req.Plan,
			ProviderMeta:  req.ProviderMeta,
			Private:       req.Private,
		}

		BlockModifyPlan(ctx, block, blockReq, resp)
	}
}
