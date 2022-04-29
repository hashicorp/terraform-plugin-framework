package proto6server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AttributeModifyPlan runs all AttributePlanModifiers
//
// TODO: Clean up this abstraction back into an internal Attribute type method.
// The extra Attribute parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/215
func AttributeModifyPlan(ctx context.Context, a tfsdk.Attribute, req tfsdk.ModifyAttributePlanRequest, resp *ModifySchemaPlanResponse) {
	ctx = logging.FrameworkWithAttributePath(ctx, req.AttributePath.String())

	attrConfig, diags := ConfigGetAttributeValue(ctx, req.Config, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	// Only on new errors.
	if diags.HasError() {
		return
	}
	req.AttributeConfig = attrConfig

	attrState, diags := StateGetAttributeValue(ctx, req.State, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	// Only on new errors.
	if diags.HasError() {
		return
	}
	req.AttributeState = attrState

	attrPlan, diags := PlanGetAttributeValue(ctx, req.Plan, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	// Only on new errors.
	if diags.HasError() {
		return
	}
	req.AttributePlan = attrPlan

	var requiresReplace bool
	for _, planModifier := range a.PlanModifiers {
		modifyResp := &tfsdk.ModifyAttributePlanResponse{
			AttributePlan:   req.AttributePlan,
			RequiresReplace: requiresReplace,
		}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined AttributePlanModifier",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)
		planModifier.Modify(ctx, req, modifyResp)
		logging.FrameworkDebug(
			ctx,
			"Called provider defined AttributePlanModifier",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

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

	if a.Attributes == nil || len(a.Attributes.GetAttributes()) == 0 {
		return
	}

	nm := a.Attributes.GetNestingMode()
	switch nm {
	case tfsdk.NestingModeList:
		l, ok := req.AttributePlan.(types.List)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributePlan, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Plan Modification Error",
				"Attribute plan modifier cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for idx := range l.Elems {
			for name, attr := range a.Attributes.GetAttributes() {
				attrReq := tfsdk.ModifyAttributePlanRequest{
					AttributePath: req.AttributePath.WithElementKeyInt(idx).WithAttributeName(name),
					Config:        req.Config,
					Plan:          resp.Plan,
					ProviderMeta:  req.ProviderMeta,
					State:         req.State,
				}

				AttributeModifyPlan(ctx, attr, attrReq, resp)
			}
		}
	case tfsdk.NestingModeSet:
		s, ok := req.AttributePlan.(types.Set)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributePlan, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Plan Modification Error",
				"Attribute plan modifier cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for _, value := range s.Elems {
			tfValue, err := value.ToTerraformValue(ctx)
			if err != nil {
				err := fmt.Errorf("error running ToTerraformValue on element value: %v", value)
				resp.Diagnostics.AddAttributeError(
					req.AttributePath,
					"Attribute Plan Modification Error",
					"Attribute plan modification cannot convert element into a Terraform value. Report this to the provider developer:\n\n"+err.Error(),
				)

				return
			}

			for name, attr := range a.Attributes.GetAttributes() {
				attrReq := tfsdk.ModifyAttributePlanRequest{
					AttributePath: req.AttributePath.WithElementKeyValue(tfValue).WithAttributeName(name),
					Config:        req.Config,
					Plan:          resp.Plan,
					ProviderMeta:  req.ProviderMeta,
					State:         req.State,
				}

				AttributeModifyPlan(ctx, attr, attrReq, resp)
			}
		}
	case tfsdk.NestingModeMap:
		m, ok := req.AttributePlan.(types.Map)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributePlan, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Plan Modification Error",
				"Attribute plan modifier cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for key := range m.Elems {
			for name, attr := range a.Attributes.GetAttributes() {
				attrReq := tfsdk.ModifyAttributePlanRequest{
					AttributePath: req.AttributePath.WithElementKeyString(key).WithAttributeName(name),
					Config:        req.Config,
					Plan:          resp.Plan,
					ProviderMeta:  req.ProviderMeta,
					State:         req.State,
				}

				AttributeModifyPlan(ctx, attr, attrReq, resp)
			}
		}
	case tfsdk.NestingModeSingle:
		o, ok := req.AttributePlan.(types.Object)

		if !ok {
			err := fmt.Errorf("unknown attribute value type (%T) for nesting mode (%T) at path: %s", req.AttributePlan, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Attribute Plan Modification Error",
				"Attribute plan modifier cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		if len(o.Attrs) == 0 {
			return
		}

		for name, attr := range a.Attributes.GetAttributes() {
			attrReq := tfsdk.ModifyAttributePlanRequest{
				AttributePath: req.AttributePath.WithAttributeName(name),
				Config:        req.Config,
				Plan:          resp.Plan,
				ProviderMeta:  req.ProviderMeta,
				State:         req.State,
			}

			AttributeModifyPlan(ctx, attr, attrReq, resp)
		}
	default:
		err := fmt.Errorf("unknown attribute nesting mode (%T: %v) at path: %s", nm, nm, req.AttributePath)
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Attribute Plan Modification Error",
			"Attribute plan modifier cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
		)

		return
	}
}
