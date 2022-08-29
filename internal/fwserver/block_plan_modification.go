package fwserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BlockModifyPlan performs all Block plan modification.
//
// TODO: Clean up this abstraction back into an internal Block type method.
// The extra Block parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/365
func BlockModifyPlan(ctx context.Context, b fwschema.Block, req tfsdk.ModifyAttributePlanRequest, resp *ModifySchemaPlanResponse) {
	configData := &fwschemadata.Data{
		Description:    fwschemadata.DataDescriptionConfiguration,
		Schema:         req.Config.Schema,
		TerraformValue: req.Config.Raw,
	}

	attributeConfig, diags := configData.ValueAtPath(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	req.AttributeConfig = attributeConfig

	planData := &fwschemadata.Data{
		Description:    fwschemadata.DataDescriptionPlan,
		Schema:         req.Plan.Schema,
		TerraformValue: req.Plan.Raw,
	}

	attributePlan, diags := planData.ValueAtPath(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	req.AttributePlan = attributePlan

	stateData := &fwschemadata.Data{
		Description:    fwschemadata.DataDescriptionState,
		Schema:         req.State.Schema,
		TerraformValue: req.State.Raw,
	}

	attributeState, diags := stateData.ValueAtPath(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	req.AttributeState = attributeState

	var requiresReplace bool

	privateProviderData := privatestate.EmptyProviderData(ctx)

	if req.Private != nil {
		resp.Private = req.Private
		privateProviderData = req.Private
	}

	if blockWithPlanModifiers, ok := b.(fwxschema.BlockWithPlanModifiers); ok {
		for _, planModifier := range blockWithPlanModifiers.GetPlanModifiers() {
			modifyResp := &tfsdk.ModifyAttributePlanResponse{
				AttributePlan:   req.AttributePlan,
				RequiresReplace: requiresReplace,
				Private:         privateProviderData,
			}

			planModifier.Modify(ctx, req, modifyResp)

			req.AttributePlan = modifyResp.AttributePlan
			resp.Diagnostics.Append(modifyResp.Diagnostics...)
			requiresReplace = modifyResp.RequiresReplace
			resp.Private = modifyResp.Private

			// Only on new errors.
			if modifyResp.Diagnostics.HasError() {
				return
			}
		}
	}

	if requiresReplace {
		resp.RequiresReplace = append(resp.RequiresReplace, req.AttributePath)
	}

	setAttrDiags := resp.Plan.SetAttribute(ctx, req.AttributePath, req.AttributePlan)
	resp.Diagnostics.Append(setAttrDiags...)

	if setAttrDiags.HasError() {
		return
	}

	nm := b.GetNestingMode()
	switch nm {
	case fwschema.BlockNestingModeList:
		l, ok := req.AttributePlan.(types.List)

		if !ok {
			err := fmt.Errorf("unknown block value type (%s) for nesting mode (%T) at path: %s", req.AttributeConfig.Type(ctx), nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Block Plan Modification Error",
				"Block validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for idx := range l.Elems {
			for name, attr := range b.GetAttributes() {
				attrReq := tfsdk.ModifyAttributePlanRequest{
					AttributePath: req.AttributePath.AtListIndex(idx).AtName(name),
					Config:        req.Config,
					Plan:          resp.Plan,
					ProviderMeta:  req.ProviderMeta,
					State:         req.State,
					Private:       resp.Private,
				}

				AttributeModifyPlan(ctx, attr, attrReq, resp)
			}

			for name, block := range b.GetBlocks() {
				blockReq := tfsdk.ModifyAttributePlanRequest{
					AttributePath: req.AttributePath.AtListIndex(idx).AtName(name),
					Config:        req.Config,
					Plan:          resp.Plan,
					ProviderMeta:  req.ProviderMeta,
					State:         req.State,
					Private:       resp.Private,
				}

				BlockModifyPlan(ctx, block, blockReq, resp)
			}
		}
	case fwschema.BlockNestingModeSet:
		s, ok := req.AttributePlan.(types.Set)

		if !ok {
			err := fmt.Errorf("unknown block value type (%s) for nesting mode (%T) at path: %s", req.AttributeConfig.Type(ctx), nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Block Plan Modification Error",
				"Block plan modification cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for _, value := range s.Elems {
			for name, attr := range b.GetAttributes() {
				attrReq := tfsdk.ModifyAttributePlanRequest{
					AttributePath: req.AttributePath.AtSetValue(value).AtName(name),
					Config:        req.Config,
					Plan:          resp.Plan,
					ProviderMeta:  req.ProviderMeta,
					State:         req.State,
					Private:       resp.Private,
				}

				AttributeModifyPlan(ctx, attr, attrReq, resp)
			}

			for name, block := range b.GetBlocks() {
				blockReq := tfsdk.ModifyAttributePlanRequest{
					AttributePath: req.AttributePath.AtSetValue(value).AtName(name),
					Config:        req.Config,
					Plan:          resp.Plan,
					ProviderMeta:  req.ProviderMeta,
					State:         req.State,
					Private:       resp.Private,
				}

				BlockModifyPlan(ctx, block, blockReq, resp)
			}
		}
	default:
		err := fmt.Errorf("unknown block plan modification nesting mode (%T: %v) at path: %s", nm, nm, req.AttributePath)
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Block Plan Modification Error",
			"Block plan modification cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
		)

		return
	}
}
