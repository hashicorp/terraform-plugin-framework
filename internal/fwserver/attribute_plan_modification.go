package fwserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/privatestate"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type ModifyAttributePlanResponse struct {
	AttributePlan   attr.Value
	Diagnostics     diag.Diagnostics
	RequiresReplace path.Paths
	Private         *privatestate.ProviderData
}

// AttributeModifyPlan runs all AttributePlanModifiers
//
// TODO: Clean up this abstraction back into an internal Attribute type method.
// The extra Attribute parameter is a carry-over of creating the proto6server
// package from the tfsdk package and not wanting to export the method.
// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/365
func AttributeModifyPlan(ctx context.Context, a fwschema.Attribute, req tfsdk.ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	ctx = logging.FrameworkWithAttributePath(ctx, req.AttributePath.String())

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
			resp.AttributePlan = modifyResp.AttributePlan
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

	if resp.Diagnostics.HasError() {
		return
	}

	// Null and unknown values should not have nested schema to modify.
	if req.AttributePlan.IsNull() || req.AttributePlan.IsUnknown() {
		return
	}

	nestedAttribute, ok := a.(fwschema.NestedAttribute)

	if !ok {
		return
	}

	// Temporarily handle tfsdk.Attribute, which always has a nesting mode, until its removed.
	if tfsdkAttribute, ok := a.(tfsdk.Attribute); ok && tfsdkAttribute.GetNestingMode() == fwschema.NestingModeUnknown {
		return
	}

	nm := nestedAttribute.GetNestingMode()
	switch nm {
	case fwschema.NestingModeList:
		configList, diags := coerceListValue(ctx, req.AttributePath, req.AttributeConfig)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		planList, diags := coerceListValue(ctx, req.AttributePath, req.AttributePlan)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		stateList, diags := coerceListValue(ctx, req.AttributePath, req.AttributeState)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		planElements := planList.Elements()

		for idx, planElem := range planElements {
			attrPath := req.AttributePath.AtListIndex(idx)

			configObject, diags := listElemObject(ctx, attrPath, configList, idx, fwschemadata.DataDescriptionConfiguration)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			planObject, diags := coerceObjectValue(ctx, attrPath, planElem)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			stateObject, diags := listElemObject(ctx, attrPath, stateList, idx, fwschemadata.DataDescriptionState)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			planAttributes := planObject.Attributes()

			for name, attr := range nestedAttribute.GetAttributes() {
				attrConfig, diags := objectAttributeValue(ctx, configObject, name, fwschemadata.DataDescriptionConfiguration)

				resp.Diagnostics.Append(diags...)

				if resp.Diagnostics.HasError() {
					return
				}

				attrPlan, diags := objectAttributeValue(ctx, planObject, name, fwschemadata.DataDescriptionPlan)

				resp.Diagnostics.Append(diags...)

				if resp.Diagnostics.HasError() {
					return
				}

				attrState, diags := objectAttributeValue(ctx, stateObject, name, fwschemadata.DataDescriptionState)

				resp.Diagnostics.Append(diags...)

				if resp.Diagnostics.HasError() {
					return
				}

				attrReq := tfsdk.ModifyAttributePlanRequest{
					AttributeConfig: attrConfig,
					AttributePath:   attrPath.AtName(name),
					AttributePlan:   attrPlan,
					AttributeState:  attrState,
					Config:          req.Config,
					Plan:            req.Plan,
					ProviderMeta:    req.ProviderMeta,
					State:           req.State,
					Private:         resp.Private,
				}
				attrResp := ModifyAttributePlanResponse{
					AttributePlan:   attrReq.AttributePlan,
					RequiresReplace: resp.RequiresReplace,
					Private:         attrReq.Private,
				}

				AttributeModifyPlan(ctx, attr, attrReq, &attrResp)

				planAttributes[name] = attrResp.AttributePlan
				resp.Diagnostics.Append(attrResp.Diagnostics...)
				resp.RequiresReplace = attrResp.RequiresReplace
				resp.Private = attrResp.Private
			}

			planElements[idx], diags = types.ObjectValue(planObject.AttributeTypes(ctx), planAttributes)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}
		}

		resp.AttributePlan, diags = types.ListValue(planList.ElementType(ctx), planElements)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}
	case fwschema.NestingModeSet:
		configSet, diags := coerceSetValue(ctx, req.AttributePath, req.AttributeConfig)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		planSet, diags := coerceSetValue(ctx, req.AttributePath, req.AttributePlan)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		stateSet, diags := coerceSetValue(ctx, req.AttributePath, req.AttributeState)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		planElements := planSet.Elements()

		for idx, planElem := range planElements {
			attrPath := req.AttributePath.AtSetValue(planElem)

			configObject, diags := setElemObject(ctx, attrPath, configSet, idx, fwschemadata.DataDescriptionConfiguration)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			planObject, diags := coerceObjectValue(ctx, attrPath, planElem)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			stateObject, diags := setElemObject(ctx, attrPath, stateSet, idx, fwschemadata.DataDescriptionState)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			planAttributes := planObject.Attributes()

			for name, attr := range nestedAttribute.GetAttributes() {
				attrConfig, diags := objectAttributeValue(ctx, configObject, name, fwschemadata.DataDescriptionConfiguration)

				resp.Diagnostics.Append(diags...)

				if resp.Diagnostics.HasError() {
					return
				}

				attrPlan, diags := objectAttributeValue(ctx, planObject, name, fwschemadata.DataDescriptionPlan)

				resp.Diagnostics.Append(diags...)

				if resp.Diagnostics.HasError() {
					return
				}

				attrState, diags := objectAttributeValue(ctx, stateObject, name, fwschemadata.DataDescriptionState)

				resp.Diagnostics.Append(diags...)

				if resp.Diagnostics.HasError() {
					return
				}

				attrReq := tfsdk.ModifyAttributePlanRequest{
					AttributeConfig: attrConfig,
					AttributePath:   attrPath.AtName(name),
					AttributePlan:   attrPlan,
					AttributeState:  attrState,
					Config:          req.Config,
					Plan:            req.Plan,
					ProviderMeta:    req.ProviderMeta,
					State:           req.State,
					Private:         resp.Private,
				}
				attrResp := ModifyAttributePlanResponse{
					AttributePlan:   attrReq.AttributePlan,
					RequiresReplace: resp.RequiresReplace,
					Private:         attrReq.Private,
				}

				AttributeModifyPlan(ctx, attr, attrReq, &attrResp)

				planAttributes[name] = attrResp.AttributePlan
				resp.Diagnostics.Append(attrResp.Diagnostics...)
				resp.RequiresReplace = attrResp.RequiresReplace
				resp.Private = attrResp.Private
			}

			planElements[idx], diags = types.ObjectValue(planObject.AttributeTypes(ctx), planAttributes)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}
		}

		resp.AttributePlan, diags = types.SetValue(planSet.ElementType(ctx), planElements)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}
	case fwschema.NestingModeMap:
		configMap, diags := coerceMapValue(ctx, req.AttributePath, req.AttributeConfig)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		planMap, diags := coerceMapValue(ctx, req.AttributePath, req.AttributePlan)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		stateMap, diags := coerceMapValue(ctx, req.AttributePath, req.AttributeState)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		planElements := planMap.Elements()

		for key, planElem := range planElements {
			attrPath := req.AttributePath.AtMapKey(key)

			configObject, diags := mapElemObject(ctx, attrPath, configMap, key, fwschemadata.DataDescriptionConfiguration)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			planObject, diags := coerceObjectValue(ctx, attrPath, planElem)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			stateObject, diags := mapElemObject(ctx, attrPath, stateMap, key, fwschemadata.DataDescriptionState)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			planAttributes := planObject.Attributes()

			for name, attr := range nestedAttribute.GetAttributes() {
				attrConfig, diags := objectAttributeValue(ctx, configObject, name, fwschemadata.DataDescriptionConfiguration)

				resp.Diagnostics.Append(diags...)

				if resp.Diagnostics.HasError() {
					return
				}

				attrPlan, diags := objectAttributeValue(ctx, planObject, name, fwschemadata.DataDescriptionPlan)

				resp.Diagnostics.Append(diags...)

				if resp.Diagnostics.HasError() {
					return
				}

				attrState, diags := objectAttributeValue(ctx, stateObject, name, fwschemadata.DataDescriptionState)

				resp.Diagnostics.Append(diags...)

				if resp.Diagnostics.HasError() {
					return
				}

				attrReq := tfsdk.ModifyAttributePlanRequest{
					AttributeConfig: attrConfig,
					AttributePath:   attrPath.AtName(name),
					AttributePlan:   attrPlan,
					AttributeState:  attrState,
					Config:          req.Config,
					Plan:            req.Plan,
					ProviderMeta:    req.ProviderMeta,
					State:           req.State,
					Private:         resp.Private,
				}
				attrResp := ModifyAttributePlanResponse{
					AttributePlan:   attrReq.AttributePlan,
					RequiresReplace: resp.RequiresReplace,
					Private:         attrReq.Private,
				}

				AttributeModifyPlan(ctx, attr, attrReq, &attrResp)

				planAttributes[name] = attrResp.AttributePlan
				resp.Diagnostics.Append(attrResp.Diagnostics...)
				resp.RequiresReplace = attrResp.RequiresReplace
				resp.Private = attrResp.Private
			}

			planElements[key], diags = types.ObjectValue(planObject.AttributeTypes(ctx), planAttributes)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}
		}

		resp.AttributePlan, diags = types.MapValue(planMap.ElementType(ctx), planElements)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}
	case fwschema.NestingModeSingle:
		configObject, diags := coerceObjectValue(ctx, req.AttributePath, req.AttributeConfig)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		planObject, diags := coerceObjectValue(ctx, req.AttributePath, req.AttributePlan)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		stateObject, diags := coerceObjectValue(ctx, req.AttributePath, req.AttributeState)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
		}

		if len(planObject.Attributes()) == 0 {
			return
		}

		planAttributes := planObject.Attributes()

		for name, attr := range nestedAttribute.GetAttributes() {
			attrConfig, diags := objectAttributeValue(ctx, configObject, name, fwschemadata.DataDescriptionConfiguration)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			attrPlan, diags := objectAttributeValue(ctx, planObject, name, fwschemadata.DataDescriptionPlan)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			attrState, diags := objectAttributeValue(ctx, stateObject, name, fwschemadata.DataDescriptionState)

			resp.Diagnostics.Append(diags...)

			if resp.Diagnostics.HasError() {
				return
			}

			attrReq := tfsdk.ModifyAttributePlanRequest{
				AttributeConfig: attrConfig,
				AttributePath:   req.AttributePath.AtName(name),
				AttributePlan:   attrPlan,
				AttributeState:  attrState,
				Config:          req.Config,
				Plan:            req.Plan,
				ProviderMeta:    req.ProviderMeta,
				State:           req.State,
				Private:         resp.Private,
			}
			attrResp := ModifyAttributePlanResponse{
				AttributePlan:   attrReq.AttributePlan,
				RequiresReplace: resp.RequiresReplace,
				Private:         attrReq.Private,
			}

			AttributeModifyPlan(ctx, attr, attrReq, &attrResp)

			planAttributes[name] = attrResp.AttributePlan
			resp.Diagnostics.Append(attrResp.Diagnostics...)
			resp.RequiresReplace = attrResp.RequiresReplace
			resp.Private = attrResp.Private
		}

		resp.AttributePlan, diags = types.ObjectValue(planObject.AttributeTypes(ctx), planAttributes)

		resp.Diagnostics.Append(diags...)

		if resp.Diagnostics.HasError() {
			return
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

func attributePlanModificationValueError(ctx context.Context, value attr.Value, description fwschemadata.DataDescription, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Attribute Plan Modification "+description.Title()+" Value Error",
		"An unexpected error occurred while fetching a "+value.Type(ctx).String()+" element value in the "+description.String()+". "+
			"This is an issue with the provider and should be reported to the provider developers.\n\n"+
			"Original Error: "+err.Error(),
	)
}

func attributePlanModificationWalkError(schemaPath path.Path, value attr.Value) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		schemaPath,
		"Attribute Plan Modification Walk Error",
		"An unexpected error occurred while walking the schema for attribute plan modification. "+
			"This is an issue with terraform-plugin-framework and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("unknown attribute value type (%T) at path: %s", value, schemaPath),
	)
}
