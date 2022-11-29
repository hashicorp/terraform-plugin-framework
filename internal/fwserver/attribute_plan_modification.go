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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

	privateProviderData := privatestate.EmptyProviderData(ctx)

	if req.Private != nil {
		resp.Private = req.Private
		privateProviderData = req.Private
	}

	switch attributeWithPlanModifiers := a.(type) {
	// Legacy tfsdk.AttributePlanModifier handling
	case fwxschema.AttributeWithPlanModifiers:
		var requiresReplace bool

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

		if requiresReplace {
			resp.RequiresReplace = append(resp.RequiresReplace, req.AttributePath)
		}
	case fwxschema.AttributeWithBoolPlanModifiers:
		AttributePlanModifyBool(ctx, attributeWithPlanModifiers, req, resp)
	case fwxschema.AttributeWithFloat64PlanModifiers:
		AttributePlanModifyFloat64(ctx, attributeWithPlanModifiers, req, resp)
	case fwxschema.AttributeWithInt64PlanModifiers:
		AttributePlanModifyInt64(ctx, attributeWithPlanModifiers, req, resp)
	case fwxschema.AttributeWithListPlanModifiers:
		AttributePlanModifyList(ctx, attributeWithPlanModifiers, req, resp)
	case fwxschema.AttributeWithMapPlanModifiers:
		AttributePlanModifyMap(ctx, attributeWithPlanModifiers, req, resp)
	case fwxschema.AttributeWithNumberPlanModifiers:
		AttributePlanModifyNumber(ctx, attributeWithPlanModifiers, req, resp)
	case fwxschema.AttributeWithObjectPlanModifiers:
		AttributePlanModifyObject(ctx, attributeWithPlanModifiers, req, resp)
	case fwxschema.AttributeWithSetPlanModifiers:
		AttributePlanModifySet(ctx, attributeWithPlanModifiers, req, resp)
	case fwxschema.AttributeWithStringPlanModifiers:
		AttributePlanModifyString(ctx, attributeWithPlanModifiers, req, resp)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Null and unknown values should not have nested schema to modify.
	if resp.AttributePlan.IsNull() || resp.AttributePlan.IsUnknown() {
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

	nestedAttributeObject := nestedAttribute.GetNestedObject()

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

			objectReq := planmodifier.ObjectRequest{
				Config:         req.Config,
				ConfigValue:    configObject,
				Path:           attrPath,
				PathExpression: attrPath.Expression(),
				Plan:           req.Plan,
				PlanValue:      planObject,
				Private:        resp.Private,
				State:          req.State,
				StateValue:     stateObject,
			}
			objectResp := &ModifyAttributePlanResponse{
				AttributePlan: objectReq.PlanValue,
				Private:       objectReq.Private,
			}

			NestedAttributeObjectPlanModify(ctx, nestedAttributeObject, objectReq, objectResp)

			planElements[idx] = objectResp.AttributePlan
			resp.Diagnostics.Append(objectResp.Diagnostics...)
			resp.Private = objectResp.Private
			resp.RequiresReplace.Append(objectResp.RequiresReplace...)
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

			objectReq := planmodifier.ObjectRequest{
				Config:         req.Config,
				ConfigValue:    configObject,
				Path:           attrPath,
				PathExpression: attrPath.Expression(),
				Plan:           req.Plan,
				PlanValue:      planObject,
				Private:        resp.Private,
				State:          req.State,
				StateValue:     stateObject,
			}
			objectResp := &ModifyAttributePlanResponse{
				AttributePlan: objectReq.PlanValue,
				Private:       objectReq.Private,
			}

			NestedAttributeObjectPlanModify(ctx, nestedAttributeObject, objectReq, objectResp)

			planElements[idx] = objectResp.AttributePlan
			resp.Diagnostics.Append(objectResp.Diagnostics...)
			resp.Private = objectResp.Private
			resp.RequiresReplace.Append(objectResp.RequiresReplace...)
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

			objectReq := planmodifier.ObjectRequest{
				Config:         req.Config,
				ConfigValue:    configObject,
				Path:           attrPath,
				PathExpression: attrPath.Expression(),
				Plan:           req.Plan,
				PlanValue:      planObject,
				Private:        resp.Private,
				State:          req.State,
				StateValue:     stateObject,
			}
			objectResp := &ModifyAttributePlanResponse{
				AttributePlan: objectReq.PlanValue,
				Private:       objectReq.Private,
			}

			NestedAttributeObjectPlanModify(ctx, nestedAttributeObject, objectReq, objectResp)

			planElements[key] = objectResp.AttributePlan
			resp.Diagnostics.Append(objectResp.Diagnostics...)
			resp.Private = objectResp.Private
			resp.RequiresReplace.Append(objectResp.RequiresReplace...)
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

		objectReq := planmodifier.ObjectRequest{
			Config:         req.Config,
			ConfigValue:    configObject,
			Path:           req.AttributePath,
			PathExpression: req.AttributePathExpression,
			Plan:           req.Plan,
			PlanValue:      planObject,
			Private:        resp.Private,
			State:          req.State,
			StateValue:     stateObject,
		}
		objectResp := &ModifyAttributePlanResponse{
			AttributePlan: objectReq.PlanValue,
			Private:       objectReq.Private,
		}

		NestedAttributeObjectPlanModify(ctx, nestedAttributeObject, objectReq, objectResp)

		resp.AttributePlan = objectResp.AttributePlan
		resp.Diagnostics.Append(objectResp.Diagnostics...)
		resp.Private = objectResp.Private
		resp.RequiresReplace.Append(objectResp.RequiresReplace...)
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

// AttributePlanModifyBool performs all types.Bool plan modification.
func AttributePlanModifyBool(ctx context.Context, attribute fwxschema.AttributeWithBoolPlanModifiers, req tfsdk.ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	// Use types.BoolValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.BoolValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Bool Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Bool attribute plan modification. "+
				"The value type must implement the types.BoolValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToBoolValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planValuable, ok := req.AttributePlan.(types.BoolValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Bool Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Bool attribute plan modification. "+
				"The value type must implement the types.BoolValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributePlan),
		)

		return
	}

	planValue, diags := planValuable.ToBoolValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	stateValuable, ok := req.AttributeState.(types.BoolValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Bool Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Bool attribute plan modification. "+
				"The value type must implement the types.BoolValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeState),
		)

		return
	}

	stateValue, diags := stateValuable.ToBoolValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planModifyReq := planmodifier.BoolRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
		Plan:           req.Plan,
		PlanValue:      planValue,
		Private:        req.Private,
		State:          req.State,
		StateValue:     stateValue,
	}

	for _, planModifier := range attribute.BoolPlanModifiers() {
		// Instantiate a new response for each request to prevent plan modifiers
		// from modifying or removing diagnostics.
		planModifyResp := &planmodifier.BoolResponse{
			PlanValue: planModifyReq.PlanValue,
			Private:   resp.Private,
		}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined planmodifier.Bool",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifier.PlanModifyBool(ctx, planModifyReq, planModifyResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined planmodifier.Bool",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifyReq.PlanValue = planModifyResp.PlanValue
		resp.AttributePlan = planModifyResp.PlanValue
		resp.Diagnostics.Append(planModifyResp.Diagnostics...)
		resp.Private = planModifyResp.Private

		if planModifyResp.RequiresReplace {
			resp.RequiresReplace.Append(req.AttributePath)
		}

		// Only on new errors.
		if planModifyResp.Diagnostics.HasError() {
			return
		}
	}
}

// AttributePlanModifyFloat64 performs all types.Float64 plan modification.
func AttributePlanModifyFloat64(ctx context.Context, attribute fwxschema.AttributeWithFloat64PlanModifiers, req tfsdk.ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	// Use types.Float64Valuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.Float64Valuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Float64 Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Float64 attribute plan modification. "+
				"The value type must implement the types.Float64Valuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToFloat64Value(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planValuable, ok := req.AttributePlan.(types.Float64Valuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Float64 Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Float64 attribute plan modification. "+
				"The value type must implement the types.Float64Valuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributePlan),
		)

		return
	}

	planValue, diags := planValuable.ToFloat64Value(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	stateValuable, ok := req.AttributeState.(types.Float64Valuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Float64 Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Float64 attribute plan modification. "+
				"The value type must implement the types.Float64Valuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeState),
		)

		return
	}

	stateValue, diags := stateValuable.ToFloat64Value(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planModifyReq := planmodifier.Float64Request{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
		Plan:           req.Plan,
		PlanValue:      planValue,
		Private:        req.Private,
		State:          req.State,
		StateValue:     stateValue,
	}

	for _, planModifier := range attribute.Float64PlanModifiers() {
		// Instantiate a new response for each request to prevent plan modifiers
		// from modifying or removing diagnostics.
		planModifyResp := &planmodifier.Float64Response{
			PlanValue: planModifyReq.PlanValue,
			Private:   resp.Private,
		}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined planmodifier.Float64",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifier.PlanModifyFloat64(ctx, planModifyReq, planModifyResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined planmodifier.Float64",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifyReq.PlanValue = planModifyResp.PlanValue
		resp.AttributePlan = planModifyResp.PlanValue
		resp.Diagnostics.Append(planModifyResp.Diagnostics...)
		resp.Private = planModifyResp.Private

		if planModifyResp.RequiresReplace {
			resp.RequiresReplace.Append(req.AttributePath)
		}

		// Only on new errors.
		if planModifyResp.Diagnostics.HasError() {
			return
		}
	}
}

// AttributePlanModifyInt64 performs all types.Int64 plan modification.
func AttributePlanModifyInt64(ctx context.Context, attribute fwxschema.AttributeWithInt64PlanModifiers, req tfsdk.ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	// Use types.Int64Valuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.Int64Valuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Int64 Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Int64 attribute plan modification. "+
				"The value type must implement the types.Int64Valuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToInt64Value(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planValuable, ok := req.AttributePlan.(types.Int64Valuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Int64 Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Int64 attribute plan modification. "+
				"The value type must implement the types.Int64Valuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributePlan),
		)

		return
	}

	planValue, diags := planValuable.ToInt64Value(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	stateValuable, ok := req.AttributeState.(types.Int64Valuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Int64 Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Int64 attribute plan modification. "+
				"The value type must implement the types.Int64Valuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeState),
		)

		return
	}

	stateValue, diags := stateValuable.ToInt64Value(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planModifyReq := planmodifier.Int64Request{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
		Plan:           req.Plan,
		PlanValue:      planValue,
		Private:        req.Private,
		State:          req.State,
		StateValue:     stateValue,
	}

	for _, planModifier := range attribute.Int64PlanModifiers() {
		// Instantiate a new response for each request to prevent plan modifiers
		// from modifying or removing diagnostics.
		planModifyResp := &planmodifier.Int64Response{
			PlanValue: planModifyReq.PlanValue,
			Private:   resp.Private,
		}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined planmodifier.Int64",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifier.PlanModifyInt64(ctx, planModifyReq, planModifyResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined planmodifier.Int64",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifyReq.PlanValue = planModifyResp.PlanValue
		resp.AttributePlan = planModifyResp.PlanValue
		resp.Diagnostics.Append(planModifyResp.Diagnostics...)
		resp.Private = planModifyResp.Private

		if planModifyResp.RequiresReplace {
			resp.RequiresReplace.Append(req.AttributePath)
		}

		// Only on new errors.
		if planModifyResp.Diagnostics.HasError() {
			return
		}
	}
}

// AttributePlanModifyList performs all types.List plan modification.
func AttributePlanModifyList(ctx context.Context, attribute fwxschema.AttributeWithListPlanModifiers, req tfsdk.ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	// Use types.ListValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.ListValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid List Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform List attribute plan modification. "+
				"The value type must implement the types.ListValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToListValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planValuable, ok := req.AttributePlan.(types.ListValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid List Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform List attribute plan modification. "+
				"The value type must implement the types.ListValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributePlan),
		)

		return
	}

	planValue, diags := planValuable.ToListValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	stateValuable, ok := req.AttributeState.(types.ListValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid List Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform List attribute plan modification. "+
				"The value type must implement the types.ListValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeState),
		)

		return
	}

	stateValue, diags := stateValuable.ToListValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planModifyReq := planmodifier.ListRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
		Plan:           req.Plan,
		PlanValue:      planValue,
		Private:        req.Private,
		State:          req.State,
		StateValue:     stateValue,
	}

	for _, planModifier := range attribute.ListPlanModifiers() {
		// Instantiate a new response for each request to prevent plan modifiers
		// from modifying or removing diagnostics.
		planModifyResp := &planmodifier.ListResponse{
			PlanValue: planModifyReq.PlanValue,
			Private:   resp.Private,
		}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined planmodifier.List",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifier.PlanModifyList(ctx, planModifyReq, planModifyResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined planmodifier.List",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifyReq.PlanValue = planModifyResp.PlanValue
		resp.AttributePlan = planModifyResp.PlanValue
		resp.Diagnostics.Append(planModifyResp.Diagnostics...)
		resp.Private = planModifyResp.Private

		if planModifyResp.RequiresReplace {
			resp.RequiresReplace.Append(req.AttributePath)
		}

		// Only on new errors.
		if planModifyResp.Diagnostics.HasError() {
			return
		}
	}
}

// AttributePlanModifyMap performs all types.Map plan modification.
func AttributePlanModifyMap(ctx context.Context, attribute fwxschema.AttributeWithMapPlanModifiers, req tfsdk.ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	// Use types.MapValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.MapValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Map Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Map attribute plan modification. "+
				"The value type must implement the types.MapValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToMapValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planValuable, ok := req.AttributePlan.(types.MapValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Map Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Map attribute plan modification. "+
				"The value type must implement the types.MapValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributePlan),
		)

		return
	}

	planValue, diags := planValuable.ToMapValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	stateValuable, ok := req.AttributeState.(types.MapValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Map Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Map attribute plan modification. "+
				"The value type must implement the types.MapValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeState),
		)

		return
	}

	stateValue, diags := stateValuable.ToMapValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planModifyReq := planmodifier.MapRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
		Plan:           req.Plan,
		PlanValue:      planValue,
		Private:        req.Private,
		State:          req.State,
		StateValue:     stateValue,
	}

	for _, planModifier := range attribute.MapPlanModifiers() {
		// Instantiate a new response for each request to prevent plan modifiers
		// from modifying or removing diagnostics.
		planModifyResp := &planmodifier.MapResponse{
			PlanValue: planModifyReq.PlanValue,
			Private:   resp.Private,
		}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined planmodifier.Map",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifier.PlanModifyMap(ctx, planModifyReq, planModifyResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined planmodifier.Map",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifyReq.PlanValue = planModifyResp.PlanValue
		resp.AttributePlan = planModifyResp.PlanValue
		resp.Diagnostics.Append(planModifyResp.Diagnostics...)
		resp.Private = planModifyResp.Private

		if planModifyResp.RequiresReplace {
			resp.RequiresReplace.Append(req.AttributePath)
		}

		// Only on new errors.
		if planModifyResp.Diagnostics.HasError() {
			return
		}
	}
}

// AttributePlanModifyNumber performs all types.Number plan modification.
func AttributePlanModifyNumber(ctx context.Context, attribute fwxschema.AttributeWithNumberPlanModifiers, req tfsdk.ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	// Use types.NumberValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.NumberValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Number Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Number attribute plan modification. "+
				"The value type must implement the types.NumberValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToNumberValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planValuable, ok := req.AttributePlan.(types.NumberValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Number Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Number attribute plan modification. "+
				"The value type must implement the types.NumberValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributePlan),
		)

		return
	}

	planValue, diags := planValuable.ToNumberValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	stateValuable, ok := req.AttributeState.(types.NumberValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Number Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Number attribute plan modification. "+
				"The value type must implement the types.NumberValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeState),
		)

		return
	}

	stateValue, diags := stateValuable.ToNumberValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planModifyReq := planmodifier.NumberRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
		Plan:           req.Plan,
		PlanValue:      planValue,
		Private:        req.Private,
		State:          req.State,
		StateValue:     stateValue,
	}

	for _, planModifier := range attribute.NumberPlanModifiers() {
		// Instantiate a new response for each request to prevent plan modifiers
		// from modifying or removing diagnostics.
		planModifyResp := &planmodifier.NumberResponse{
			PlanValue: planModifyReq.PlanValue,
			Private:   resp.Private,
		}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined planmodifier.Number",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifier.PlanModifyNumber(ctx, planModifyReq, planModifyResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined planmodifier.Number",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifyReq.PlanValue = planModifyResp.PlanValue
		resp.AttributePlan = planModifyResp.PlanValue
		resp.Diagnostics.Append(planModifyResp.Diagnostics...)
		resp.Private = planModifyResp.Private

		if planModifyResp.RequiresReplace {
			resp.RequiresReplace.Append(req.AttributePath)
		}

		// Only on new errors.
		if planModifyResp.Diagnostics.HasError() {
			return
		}
	}
}

// AttributePlanModifyObject performs all types.Object plan modification.
func AttributePlanModifyObject(ctx context.Context, attribute fwxschema.AttributeWithObjectPlanModifiers, req tfsdk.ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	// Use types.ObjectValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.ObjectValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Object Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Object attribute plan modification. "+
				"The value type must implement the types.ObjectValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToObjectValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planValuable, ok := req.AttributePlan.(types.ObjectValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Object Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Object attribute plan modification. "+
				"The value type must implement the types.ObjectValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributePlan),
		)

		return
	}

	planValue, diags := planValuable.ToObjectValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	stateValuable, ok := req.AttributeState.(types.ObjectValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Object Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Object attribute plan modification. "+
				"The value type must implement the types.ObjectValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeState),
		)

		return
	}

	stateValue, diags := stateValuable.ToObjectValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planModifyReq := planmodifier.ObjectRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
		Plan:           req.Plan,
		PlanValue:      planValue,
		Private:        req.Private,
		State:          req.State,
		StateValue:     stateValue,
	}

	for _, planModifier := range attribute.ObjectPlanModifiers() {
		// Instantiate a new response for each request to prevent plan modifiers
		// from modifying or removing diagnostics.
		planModifyResp := &planmodifier.ObjectResponse{
			PlanValue: planModifyReq.PlanValue,
			Private:   resp.Private,
		}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined planmodifier.Object",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifier.PlanModifyObject(ctx, planModifyReq, planModifyResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined planmodifier.Object",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifyReq.PlanValue = planModifyResp.PlanValue
		resp.AttributePlan = planModifyResp.PlanValue
		resp.Diagnostics.Append(planModifyResp.Diagnostics...)
		resp.Private = planModifyResp.Private

		if planModifyResp.RequiresReplace {
			resp.RequiresReplace.Append(req.AttributePath)
		}

		// Only on new errors.
		if planModifyResp.Diagnostics.HasError() {
			return
		}
	}
}

// AttributePlanModifySet performs all types.Set plan modification.
func AttributePlanModifySet(ctx context.Context, attribute fwxschema.AttributeWithSetPlanModifiers, req tfsdk.ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	// Use types.SetValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.SetValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Set Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Set attribute plan modification. "+
				"The value type must implement the types.SetValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToSetValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planValuable, ok := req.AttributePlan.(types.SetValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Set Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Set attribute plan modification. "+
				"The value type must implement the types.SetValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributePlan),
		)

		return
	}

	planValue, diags := planValuable.ToSetValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	stateValuable, ok := req.AttributeState.(types.SetValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid Set Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform Set attribute plan modification. "+
				"The value type must implement the types.SetValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeState),
		)

		return
	}

	stateValue, diags := stateValuable.ToSetValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planModifyReq := planmodifier.SetRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
		Plan:           req.Plan,
		PlanValue:      planValue,
		Private:        req.Private,
		State:          req.State,
		StateValue:     stateValue,
	}

	for _, planModifier := range attribute.SetPlanModifiers() {
		// Instantiate a new response for each request to prevent plan modifiers
		// from modifying or removing diagnostics.
		planModifyResp := &planmodifier.SetResponse{
			PlanValue: planModifyReq.PlanValue,
			Private:   resp.Private,
		}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined planmodifier.Set",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifier.PlanModifySet(ctx, planModifyReq, planModifyResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined planmodifier.Set",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifyReq.PlanValue = planModifyResp.PlanValue
		resp.AttributePlan = planModifyResp.PlanValue
		resp.Diagnostics.Append(planModifyResp.Diagnostics...)
		resp.Private = planModifyResp.Private

		if planModifyResp.RequiresReplace {
			resp.RequiresReplace.Append(req.AttributePath)
		}

		// Only on new errors.
		if planModifyResp.Diagnostics.HasError() {
			return
		}
	}
}

// AttributePlanModifyString performs all types.String plan modification.
func AttributePlanModifyString(ctx context.Context, attribute fwxschema.AttributeWithStringPlanModifiers, req tfsdk.ModifyAttributePlanRequest, resp *ModifyAttributePlanResponse) {
	// Use types.StringValuable until custom types cannot re-implement
	// ValueFromTerraform. Until then, custom types are not technically
	// required to implement this interface. This opts to enforce the
	// requirement before compatibility promises would interfere.
	configValuable, ok := req.AttributeConfig.(types.StringValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid String Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform String attribute plan modification. "+
				"The value type must implement the types.StringValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeConfig),
		)

		return
	}

	configValue, diags := configValuable.ToStringValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planValuable, ok := req.AttributePlan.(types.StringValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid String Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform String attribute plan modification. "+
				"The value type must implement the types.StringValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributePlan),
		)

		return
	}

	planValue, diags := planValuable.ToStringValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	stateValuable, ok := req.AttributeState.(types.StringValuable)

	if !ok {
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Invalid String Attribute Plan Modifier Value Type",
			"An unexpected value type was encountered while attempting to perform String attribute plan modification. "+
				"The value type must implement the types.StringValuable interface. "+
				"Please report this to the provider developers.\n\n"+
				fmt.Sprintf("Incoming Value Type: %T", req.AttributeState),
		)

		return
	}

	stateValue, diags := stateValuable.ToStringValue(ctx)

	resp.Diagnostics.Append(diags...)

	// Only return early on new errors as the resp.Diagnostics may have errors
	// from other attributes.
	if diags.HasError() {
		return
	}

	planModifyReq := planmodifier.StringRequest{
		Config:         req.Config,
		ConfigValue:    configValue,
		Path:           req.AttributePath,
		PathExpression: req.AttributePathExpression,
		Plan:           req.Plan,
		PlanValue:      planValue,
		Private:        req.Private,
		State:          req.State,
		StateValue:     stateValue,
	}

	for _, planModifier := range attribute.StringPlanModifiers() {
		// Instantiate a new response for each request to prevent plan modifiers
		// from modifying or removing diagnostics.
		planModifyResp := &planmodifier.StringResponse{
			PlanValue: planModifyReq.PlanValue,
			Private:   resp.Private,
		}

		logging.FrameworkDebug(
			ctx,
			"Calling provider defined planmodifier.String",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifier.PlanModifyString(ctx, planModifyReq, planModifyResp)

		logging.FrameworkDebug(
			ctx,
			"Called provider defined planmodifier.String",
			map[string]interface{}{
				logging.KeyDescription: planModifier.Description(ctx),
			},
		)

		planModifyReq.PlanValue = planModifyResp.PlanValue
		resp.AttributePlan = planModifyResp.PlanValue
		resp.Diagnostics.Append(planModifyResp.Diagnostics...)
		resp.Private = planModifyResp.Private

		if planModifyResp.RequiresReplace {
			resp.RequiresReplace.Append(req.AttributePath)
		}

		// Only on new errors.
		if planModifyResp.Diagnostics.HasError() {
			return
		}
	}
}

func NestedAttributeObjectPlanModify(ctx context.Context, o fwschema.NestedAttributeObject, req planmodifier.ObjectRequest, resp *ModifyAttributePlanResponse) {
	if objectWithPlanModifiers, ok := o.(fwxschema.NestedAttributeObjectWithPlanModifiers); ok {
		for _, objectValidator := range objectWithPlanModifiers.ObjectPlanModifiers() {
			// Instantiate a new response for each request to prevent plan modifiers
			// from modifying or removing diagnostics.
			planModifyResp := &planmodifier.ObjectResponse{
				PlanValue: req.PlanValue,
				Private:   resp.Private,
			}

			logging.FrameworkDebug(
				ctx,
				"Calling provider defined planmodifier.Object",
				map[string]interface{}{
					logging.KeyDescription: objectValidator.Description(ctx),
				},
			)

			objectValidator.PlanModifyObject(ctx, req, planModifyResp)

			logging.FrameworkDebug(
				ctx,
				"Called provider defined planmodifier.Object",
				map[string]interface{}{
					logging.KeyDescription: objectValidator.Description(ctx),
				},
			)

			req.PlanValue = planModifyResp.PlanValue
			resp.AttributePlan = planModifyResp.PlanValue
			resp.Diagnostics.Append(planModifyResp.Diagnostics...)
			resp.Private = planModifyResp.Private

			if planModifyResp.RequiresReplace {
				resp.RequiresReplace.Append(req.Path)
			}

			// only on new errors
			if planModifyResp.Diagnostics.HasError() {
				return
			}
		}
	}

	newPlanValueAttributes := req.PlanValue.Attributes()

	for nestedName, nestedAttr := range o.GetAttributes() {
		nestedAttrConfig, diags := objectAttributeValue(ctx, req.ConfigValue, nestedName, fwschemadata.DataDescriptionConfiguration)

		resp.Diagnostics.Append(diags...)

		if diags.HasError() {
			return
		}

		nestedAttrPlan, diags := objectAttributeValue(ctx, req.PlanValue, nestedName, fwschemadata.DataDescriptionPlan)

		resp.Diagnostics.Append(diags...)

		if diags.HasError() {
			return
		}

		nestedAttrState, diags := objectAttributeValue(ctx, req.StateValue, nestedName, fwschemadata.DataDescriptionState)

		resp.Diagnostics.Append(diags...)

		if diags.HasError() {
			return
		}

		nestedAttrReq := tfsdk.ModifyAttributePlanRequest{
			AttributeConfig:         nestedAttrConfig,
			AttributePath:           req.Path.AtName(nestedName),
			AttributePathExpression: req.PathExpression.AtName(nestedName),
			AttributePlan:           nestedAttrPlan,
			AttributeState:          nestedAttrState,
			Config:                  req.Config,
			Plan:                    req.Plan,
			Private:                 resp.Private,
			State:                   req.State,
		}
		nestedAttrResp := &ModifyAttributePlanResponse{
			AttributePlan:   nestedAttrReq.AttributePlan,
			RequiresReplace: resp.RequiresReplace,
			Private:         nestedAttrReq.Private,
		}

		AttributeModifyPlan(ctx, nestedAttr, nestedAttrReq, nestedAttrResp)

		newPlanValueAttributes[nestedName] = nestedAttrResp.AttributePlan
		resp.Diagnostics.Append(nestedAttrResp.Diagnostics...)
		resp.Private = nestedAttrResp.Private
		resp.RequiresReplace.Append(nestedAttrResp.RequiresReplace...)
	}

	newPlanValue, diags := types.ObjectValue(req.PlanValue.AttributeTypes(ctx), newPlanValueAttributes)

	resp.Diagnostics.Append(diags...)

	resp.AttributePlan = newPlanValue
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
