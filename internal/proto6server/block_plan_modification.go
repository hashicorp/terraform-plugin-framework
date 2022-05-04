package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// BlockModifyPlan performs all Block plan modification.
//
// TODO: Clean up this abstraction back into an internal Block type method.
// The extra Block parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/215
func BlockModifyPlan(ctx context.Context, b tfsdk.Block, req tfsdk.ModifyAttributePlanRequest, resp *ModifySchemaPlanResponse) {
	attributeConfig, diags := ConfigGetAttributeValue(ctx, req.Config, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	req.AttributeConfig = attributeConfig

	attributePlan, diags := PlanGetAttributeValue(ctx, req.Plan, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	req.AttributePlan = attributePlan

	attributeState, diags := StateGetAttributeValue(ctx, req.State, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	req.AttributeState = attributeState

	var requiresReplace bool
	for _, planModifier := range b.PlanModifiers {
		modifyResp := &tfsdk.ModifyAttributePlanResponse{
			AttributePlan:   req.AttributePlan,
			RequiresReplace: requiresReplace,
		}

		planModifier.Modify(ctx, req, modifyResp)

		req.AttributePlan = modifyResp.AttributePlan
		resp.Diagnostics.Append(modifyResp.Diagnostics...)
		requiresReplace = modifyResp.RequiresReplace

		// Only on new errors.
		if modifyResp.Diagnostics.HasError() {
			return
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

	nm := b.NestingMode
	switch nm {
	case tfsdk.BlockNestingModeList:
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
			for name, attr := range b.Attributes {
				attrReq := tfsdk.ModifyAttributePlanRequest{
					AttributePath: req.AttributePath.WithElementKeyInt(idx).WithAttributeName(name),
					Config:        req.Config,
					Plan:          resp.Plan,
					ProviderMeta:  req.ProviderMeta,
					State:         req.State,
				}

				AttributeModifyPlan(ctx, attr, attrReq, resp)
			}

			for name, block := range b.Blocks {
				blockReq := tfsdk.ModifyAttributePlanRequest{
					AttributePath: req.AttributePath.WithElementKeyInt(idx).WithAttributeName(name),
					Config:        req.Config,
					Plan:          resp.Plan,
					ProviderMeta:  req.ProviderMeta,
					State:         req.State,
				}

				BlockModifyPlan(ctx, block, blockReq, resp)
			}
		}
	case tfsdk.BlockNestingModeSet:
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
			tfValue, err := value.ToTerraformValue(ctx)
			if err != nil {
				err := fmt.Errorf("error running ToTerraformValue on element value: %v", value)
				resp.Diagnostics.AddAttributeError(
					req.AttributePath,
					"Block Plan Modification Error",
					"Block plan modification cannot convert element into a Terraform value. Report this to the provider developer:\n\n"+err.Error(),
				)

				return
			}

			for name, attr := range b.Attributes {
				attrReq := tfsdk.ModifyAttributePlanRequest{
					AttributePath: req.AttributePath.WithElementKeyValue(tfValue).WithAttributeName(name),
					Config:        req.Config,
					Plan:          resp.Plan,
					ProviderMeta:  req.ProviderMeta,
					State:         req.State,
				}

				AttributeModifyPlan(ctx, attr, attrReq, resp)
			}

			for name, block := range b.Blocks {
				blockReq := tfsdk.ModifyAttributePlanRequest{
					AttributePath: req.AttributePath.WithElementKeyValue(tfValue).WithAttributeName(name),
					Config:        req.Config,
					Plan:          resp.Plan,
					ProviderMeta:  req.ProviderMeta,
					State:         req.State,
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
