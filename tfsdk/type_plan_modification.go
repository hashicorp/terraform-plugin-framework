package tfsdk

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func runTypePlanModifiers(ctx context.Context, state, plan tftypes.Value, schema Schema, resp *planResourceChangeResponse) (tftypes.Value, bool) {
	if plan.IsNull() {
		return plan, true
	}
	rawPlan := map[string]tftypes.Value{}
	err := plan.As(&rawPlan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error parsing plan",
			"An unexpected error was encountered trying to parse the plan. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return plan, false
	}
	rawState := map[string]tftypes.Value{}
	if !state.IsNull() {
		err = state.As(&rawState)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error parsing state",
				"An unexpected error was encountered trying to parse the prior state. This is always an error in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return plan, false
		}
	}
	for attrName, a := range schema.Attributes {
		path := tftypes.NewAttributePath().WithAttributeName(attrName)
		planAttr, ok := rawPlan[attrName]
		if !ok {
			resp.Diagnostics.AddError(
				"Error modifying plan",
				// TODO: this isn't ideal
				// but should it be a pathed diagnostic?
				// Terraform is horribly broken at this point
				// there may be nothing in the config to point to
				fmt.Sprintf("An attribute %s in the schema was not present in the plan. This is possibly a bug with Terraform. Please report it to the provider developer.", path),
			)
			return plan, false
		}
		stateAttr, ok := rawState[attrName]
		if !ok {
			stateAttr = tftypes.NewValue(a.terraformType(ctx), nil)
		}
		newPlan, diags := attributeTypeModifyPlan(ctx, a.attributeType(), stateAttr, planAttr, path)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return plan, false
		}
		rawNewPlan, err := attr.ValueToTerraform(ctx, newPlan)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error converting value",
				"An unexpected error was encountered converting a value to its protocol type during plan modification. This is always a bug in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
			)
			return plan, false
		}
		rawPlan[attrName] = rawNewPlan
	}
	err = tftypes.ValidateValue(plan.Type(), rawPlan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error modifying plan",
			"An unexpected error was encountered validating the modified plan. This is always a bug in the provider. Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return plan, false
	}
	plan = tftypes.NewValue(plan.Type(), rawPlan)
	return plan, true
}

func attributeTypeModifyPlanObject(ctx context.Context, typ attr.Type, state, plan tftypes.Value, path *tftypes.AttributePath) (tftypes.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	wat, ok := typ.(attr.TypeWithAttributeTypes)
	if !ok {
		diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nObject at "+path.String()+" isn't of an attr.Type that implements TypeWithAttributeType. This isn't a valid type for this value.")
		return plan, diags
	}
	stateVals := map[string]tftypes.Value{}

	// we can handle null states, though; if the state is null and the plan
	// is not, the state gets set to the null value for the purposes of
	// plan modification for that attribute
	if state.IsNull() {
		for attr, typ := range wat.AttributeTypes() {
			stateVals[attr] = tftypes.NewValue(typ.TerraformType(ctx), nil)
		}
	} else {
		err := state.As(&stateVals)
		if err != nil {
			diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nError converting state: "+err.Error())
			return plan, diags
		}
	}

	planVals := map[string]tftypes.Value{}
	err := plan.As(&planVals)
	if err != nil {
		diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nError converting plan: "+err.Error())
		return plan, diags
	}
	for attrName, attrPlan := range planVals {
		attrType, ok := wat.AttributeTypes()[attrName]
		if !ok {
			diags.AddAttributeError(path, "Error generating plan",
				fmt.Sprintf("An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nThe type of %s has no %q attribute. This should never happen.", path, attrName),
			)
			return plan, diags
		}
		attrState, ok := stateVals[attrName]
		if !ok {
			attrState = tftypes.NewValue(attrType.TerraformType(ctx), nil)
		}
		newPlan, ds := attributeTypeModifyPlan(ctx, attrType, attrState, attrPlan, path.WithAttributeName(attrName))
		diags.Append(ds...)
		if diags.HasError() {
			return plan, diags
		}
		rawNewPlan, err := attr.ValueToTerraform(ctx, newPlan)
		if err != nil {
			diags.AddAttributeError(path, "Error generating plan",
				fmt.Sprintf("An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nCouldn't convert the attr.Value at %s back to a tftypes.Value: %s", path.WithAttributeName(attrName), err),
			)
			return plan, diags
		}
		planVals[attrName] = rawNewPlan
	}
	err = tftypes.ValidateValue(typ.TerraformType(ctx), planVals)
	if err != nil {
		diags.AddAttributeError(path, "Error generating plan",
			fmt.Sprintf("An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nThe modified plan was no longer compatible with the Terraform type: %s", err),
		)
		return plan, diags
	}
	return tftypes.NewValue(typ.TerraformType(ctx), planVals), diags
}

func attributeTypeModifyPlanMap(ctx context.Context, typ attr.Type, state, plan tftypes.Value, path *tftypes.AttributePath) (tftypes.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	wet, ok := typ.(attr.TypeWithElementType)
	if !ok {
		diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nMap at "+path.String()+" isn't of an attr.Type that implements TypeWithElementType. This isn't a valid type for this value.")
		return plan, diags
	}
	stateVals := map[string]tftypes.Value{}
	if !state.IsNull() {
		err := state.As(&stateVals)
		if err != nil {
			diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nError converting state: "+err.Error())
			return plan, diags
		}
	}
	planVals := map[string]tftypes.Value{}
	err := plan.As(&planVals)
	if err != nil {
		diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nError converting plan: "+err.Error())
		return plan, diags
	}
	for key, elemPlan := range planVals {
		elemState, ok := stateVals[key]
		if !ok {
			elemState = tftypes.NewValue(wet.ElementType().TerraformType(ctx), nil)
		}
		newPlan, ds := attributeTypeModifyPlan(ctx, wet.ElementType(), elemState, elemPlan, path.WithElementKeyString(key))
		diags.Append(ds...)
		if diags.HasError() {
			return plan, diags
		}
		rawNewPlan, err := attr.ValueToTerraform(ctx, newPlan)
		if err != nil {
			diags.AddAttributeError(path, "Error generating plan",
				fmt.Sprintf("An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nCouldn't convert the attr.Value at %s back to a tftypes.Value: %s", path.WithElementKeyString(key), err),
			)
			return plan, diags
		}
		planVals[key] = rawNewPlan
	}
	err = tftypes.ValidateValue(typ.TerraformType(ctx), planVals)
	if err != nil {
		diags.AddAttributeError(path, "Error generating plan",
			fmt.Sprintf("An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nThe modified plan was no longer compatible with the Terraform type: %s", err),
		)
		return plan, diags
	}
	return tftypes.NewValue(typ.TerraformType(ctx), planVals), diags
}

func attributeTypeModifyPlanList(ctx context.Context, typ attr.Type, state, plan tftypes.Value, path *tftypes.AttributePath) (tftypes.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	wet, ok := typ.(attr.TypeWithElementType)
	if !ok {
		diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nList at "+path.String()+" isn't of an attr.Type that implements TypeWithElementType. This isn't a valid type for this value.")
		return plan, diags
	}
	stateVals := []tftypes.Value{}
	if !state.IsNull() {
		err := state.As(&stateVals)
		if err != nil {
			diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nError converting state: "+err.Error())
			return plan, diags
		}
	}
	planVals := []tftypes.Value{}
	err := plan.As(&planVals)
	if err != nil {
		diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nError converting plan: "+err.Error())
		return plan, diags
	}
	for index, elemPlan := range planVals {
		elemState := tftypes.NewValue(wet.ElementType().TerraformType(ctx), nil)
		if index < len(stateVals) {
			elemState = stateVals[index]
		}
		newPlan, ds := attributeTypeModifyPlan(ctx, wet.ElementType(), elemState, elemPlan, path.WithElementKeyInt(index))
		diags.Append(ds...)
		if diags.HasError() {
			return plan, diags
		}
		rawNewPlan, err := attr.ValueToTerraform(ctx, newPlan)
		if err != nil {
			diags.AddAttributeError(path, "Error generating plan",
				fmt.Sprintf("An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nCouldn't convert the attr.Value at %s back to a tftypes.Value: %s", path.WithElementKeyInt(index), err),
			)
			return plan, diags
		}
		planVals[index] = rawNewPlan
	}
	err = tftypes.ValidateValue(typ.TerraformType(ctx), planVals)
	if err != nil {
		diags.AddAttributeError(path, "Error generating plan",
			fmt.Sprintf("An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nThe modified plan was no longer compatible with the Terraform type: %s", err),
		)
		return plan, diags
	}
	return tftypes.NewValue(typ.TerraformType(ctx), planVals), diags
}

func attributeTypeModifyPlanTuple(ctx context.Context, typ attr.Type, state, plan tftypes.Value, path *tftypes.AttributePath) (tftypes.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	wets, ok := typ.(attr.TypeWithElementTypes)
	if !ok {
		diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nTuple at "+path.String()+" isn't of an attr.Type that implements TypeWithElementTypes. This isn't a valid type for this value.")
		return plan, diags
	}
	stateVals := []tftypes.Value{}
	if !state.IsNull() {
		err := state.As(&stateVals)
		if err != nil {
			diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nError converting state: "+err.Error())
			return plan, diags
		}
	}
	planVals := []tftypes.Value{}
	err := plan.As(&planVals)
	if err != nil {
		diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nError converting plan: "+err.Error())
		return plan, diags
	}
	elemTypes := wets.ElementTypes()
	for index, elemPlan := range planVals {
		if index >= len(elemTypes) {
			diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nPlan has a value at "+path.WithElementKeyInt(index).String()+" that the tuple has no type for.")
			return plan, diags
		}
		elemType := elemTypes[index]
		if index >= len(stateVals) {
			diags.AddAttributeError(path, "Error generating plan", "An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nPlan has a value at "+path.WithElementKeyInt(index).String()+" that has no state equivalent, which is not allowed in tuples.")
			return plan, diags
		}
		elemState := tftypes.NewValue(elemType.TerraformType(ctx), nil)
		if index < len(stateVals) {
			elemState = stateVals[index]
		}
		newPlan, ds := attributeTypeModifyPlan(ctx, elemType, elemState, elemPlan, path.WithElementKeyInt(index))
		diags.Append(ds...)
		if diags.HasError() {
			return plan, diags
		}
		rawNewPlan, err := attr.ValueToTerraform(ctx, newPlan)
		if err != nil {
			diags.AddAttributeError(path, "Error generating plan",
				fmt.Sprintf("An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nCouldn't convert the attr.Value at %s back to a tftypes.Value: %s", path.WithElementKeyInt(index), err),
			)
			return plan, diags
		}
		planVals[index] = rawNewPlan
	}
	err = tftypes.ValidateValue(typ.TerraformType(ctx), planVals)
	if err != nil {
		diags.AddAttributeError(path, "Error generating plan",
			fmt.Sprintf("An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nThe modified plan was no longer compatible with the Terraform type: %s", err),
		)
		return plan, diags
	}
	return tftypes.NewValue(typ.TerraformType(ctx), planVals), diags
}

func attributeTypeModifyPlan(ctx context.Context, typ attr.Type, state, plan tftypes.Value, path *tftypes.AttributePath) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	if plan.IsKnown() && !plan.IsNull() {
		if typ.TerraformType(ctx).Is(tftypes.Object{}) {
			newPlan, ds := attributeTypeModifyPlanObject(ctx, typ, state, plan, path)
			diags.Append(ds...)
			if diags.HasError() {
				return nil, diags
			}
			plan = newPlan
		} else if typ.TerraformType(ctx).Is(tftypes.Map{}) {
			newPlan, ds := attributeTypeModifyPlanMap(ctx, typ, state, plan, path)
			diags.Append(ds...)
			if diags.HasError() {
				return nil, diags
			}
			plan = newPlan
		} else if typ.TerraformType(ctx).Is(tftypes.List{}) {
			newPlan, ds := attributeTypeModifyPlanList(ctx, typ, state, plan, path)
			diags.Append(ds...)
			if diags.HasError() {
				return nil, diags
			}
			plan = newPlan
		} else if typ.TerraformType(ctx).Is(tftypes.Set{}) {
			// modifying the plan for each element in a set isn't
			// supported at the moment, because there's no way to
			// correlate the new value in the plan with the old
			// value in the state; sets can only be reliably
			// compared by the identity of the elements, and the
			// identity of the element changed.
			//
			// as such, sets must have plan modification applied at
			// the set level, because anything else doesn't really
			// make much sense.
			//
			// providers unhappy with this can always implement the
			// logic to call each element's ModifyPlan inside their
			// set type's ModifyPlan, but I can't see a way to do
			// it that's not rife with weird behaviors.
		} else if typ.TerraformType(ctx).Is(tftypes.Tuple{}) {
			newPlan, ds := attributeTypeModifyPlanTuple(ctx, typ, state, plan, path)
			diags.Append(ds...)
			if diags.HasError() {
				return nil, diags
			}
			plan = newPlan
		}
	}
	planModifier, ok := typ.(attr.TypeWithModifyPlan)
	if !ok {
		planVal, err := typ.ValueFromTerraform(ctx, plan)
		if err != nil {
			diags.AddAttributeError(path, "Error generating plan",
				fmt.Sprintf("An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nThe modified plan was no longer compatible with the Terraform type: %s", err),
			)
			return nil, diags
		}
		return planVal, nil
	}
	stateVal, err := typ.ValueFromTerraform(ctx, state)
	if err != nil {
		diags.AddAttributeError(path, "Error generating plan",
			fmt.Sprintf("An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nError creating state tftypes.Value from attr.Value: %s", err),
		)
		return nil, diags
	}
	planVal, err := typ.ValueFromTerraform(ctx, plan)
	if err != nil {
		diags.AddAttributeError(path, "Error generating plan",
			fmt.Sprintf("An unexpected error was encountered while trying to generate the plan. Please report the following to the provider developer:\n\nError creating plan tftypes.Value from attr.Value: %s", err),
		)
		return nil, diags
	}
	return planModifier.ModifyPlan(ctx, stateVal, planVal, path)
}
