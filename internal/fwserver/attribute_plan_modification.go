package fwserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// AttributeModifyPlan runs all AttributePlanModifiers
//
// TODO: Clean up this abstraction back into an internal Attribute type method.
// The extra Attribute parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/365
func AttributeModifyPlan(ctx context.Context, a fwschema.Attribute, req tfsdk.ModifyAttributePlanRequest, resp *ModifySchemaPlanResponse) {
	ctx = logging.FrameworkWithAttributePath(ctx, req.AttributePath.String())

	configData := &fwschemadata.Data{
		Description:    fwschemadata.DataDescriptionConfiguration,
		Schema:         req.Config.Schema,
		TerraformValue: req.Config.Raw,
	}

	attrConfig, diags := configData.ValueAtPath(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	// Only on new errors.
	if diags.HasError() {
		return
	}
	req.AttributeConfig = attrConfig

	stateData := &fwschemadata.Data{
		Description:    fwschemadata.DataDescriptionState,
		Schema:         req.State.Schema,
		TerraformValue: req.State.Raw,
	}

	attrState, diags := stateData.ValueAtPath(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	// Only on new errors.
	if diags.HasError() {
		return
	}
	req.AttributeState = attrState

	planData := &fwschemadata.Data{
		Description:    fwschemadata.DataDescriptionPlan,
		Schema:         req.Plan.Schema,
		TerraformValue: req.Plan.Raw,
	}

	attrPlan, diags := planData.ValueAtPath(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	// Only on new errors.
	if diags.HasError() {
		return
	}
	req.AttributePlan = attrPlan

	var requiresReplace bool

	privateProviderData := privatestate.EmptyProviderData(ctx)

	if req.Private != nil {
		resp.Private = req.Private
		privateProviderData = req.Private
	}

	if attributeWithPlanModifiers, ok := a.(fwxschema.AttributeWithPlanModifiers); ok {
		for _, planModifier := range attributeWithPlanModifiers.GetPlanModifiers() {
			modifyResp := &tfsdk.ModifyAttributePlanResponse{
				AttributePlan:   req.AttributePlan,
				RequiresReplace: requiresReplace,
				Private:         privateProviderData,
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

	if a.GetAttributes() == nil || len(a.GetAttributes().GetAttributes()) == 0 {
		return
	}

	nm := a.GetAttributes().GetNestingMode()
	switch nm {
	case fwschema.NestingModeList:
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
			for name, attr := range a.GetAttributes().GetAttributes() {
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
		}
	case fwschema.NestingModeSet:
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
			for name, attr := range a.GetAttributes().GetAttributes() {
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
		}
	case fwschema.NestingModeMap:
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
			for name, attr := range a.GetAttributes().GetAttributes() {
				attrReq := tfsdk.ModifyAttributePlanRequest{
					AttributePath: req.AttributePath.AtMapKey(key).AtName(name),
					Config:        req.Config,
					Plan:          resp.Plan,
					ProviderMeta:  req.ProviderMeta,
					State:         req.State,
					Private:       resp.Private,
				}

				AttributeModifyPlan(ctx, attr, attrReq, resp)
			}
		}
	case fwschema.NestingModeSingle:
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

		for name, attr := range a.GetAttributes().GetAttributes() {
			attrReq := tfsdk.ModifyAttributePlanRequest{
				AttributePath: req.AttributePath.AtName(name),
				Config:        req.Config,
				Plan:          resp.Plan,
				ProviderMeta:  req.ProviderMeta,
				State:         req.State,
				Private:       resp.Private,
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
