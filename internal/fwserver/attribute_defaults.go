// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
)

func setDefaultValueAtPath(ctx context.Context, config tftypes.Value, schema fwschema.Schema, attributePath path.Path, diags diag.Diagnostics) (tftypes.Value, bool, diag.Diagnostics) {
	tfPath := terraformPathFromPath(ctx, attributePath)
	if tfPath == nil {
		return config, false, diags
	}

	attr, err := schema.AttributeAtTerraformPath(ctx, tfPath)
	if err != nil {
		return config, false, diags
	}

	defaultValue, hasDefault, diags := attributeDefaultValue(ctx, attr, attributePath, diags)
	if !hasDefault || defaultValue.IsNull() {
		return config, false, diags
	}

	config, err = replaceValueAtPath(config, tfPath, defaultValue)
	if err != nil {
		diags.AddError(
			"Generate Resource Config Error",
			"An unexpected error was encountered setting a default value for attribute "+attributePath.String()+": "+err.Error(),
		)
		return config, false, diags
	}

	return config, true, diags
}

func attributeDefaultValue(ctx context.Context, attr fwschema.Attribute, fwPath path.Path, diags diag.Diagnostics) (tftypes.Value, bool, diag.Diagnostics) {
	var (
		result tftypes.Value
		err    error
	)

	switch a := attr.(type) {
	case fwschema.AttributeWithBoolDefaultValue:
		defaultValue := a.BoolDefaultValue()
		if defaultValue == nil {
			return result, false, diags
		}

		resp := defaults.BoolResponse{}
		defaultValue.DefaultBool(ctx, defaults.BoolRequest{Path: fwPath}, &resp)
		diags.Append(resp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return result, false, diags
		}

		result, err = resp.PlanValue.ToTerraformValue(ctx)
	case fwschema.AttributeWithFloat32DefaultValue:
		defaultValue := a.Float32DefaultValue()
		if defaultValue == nil {
			return result, false, diags
		}

		resp := defaults.Float32Response{}
		defaultValue.DefaultFloat32(ctx, defaults.Float32Request{Path: fwPath}, &resp)
		diags.Append(resp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return result, false, diags
		}

		result, err = resp.PlanValue.ToTerraformValue(ctx)
	case fwschema.AttributeWithFloat64DefaultValue:
		defaultValue := a.Float64DefaultValue()
		if defaultValue == nil {
			return result, false, diags
		}

		resp := defaults.Float64Response{}
		defaultValue.DefaultFloat64(ctx, defaults.Float64Request{Path: fwPath}, &resp)
		diags.Append(resp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return result, false, diags
		}

		result, err = resp.PlanValue.ToTerraformValue(ctx)
	case fwschema.AttributeWithInt32DefaultValue:
		defaultValue := a.Int32DefaultValue()
		if defaultValue == nil {
			return result, false, diags
		}

		resp := defaults.Int32Response{}
		defaultValue.DefaultInt32(ctx, defaults.Int32Request{Path: fwPath}, &resp)
		diags.Append(resp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return result, false, diags
		}

		result, err = resp.PlanValue.ToTerraformValue(ctx)
	case fwschema.AttributeWithInt64DefaultValue:
		defaultValue := a.Int64DefaultValue()
		if defaultValue == nil {
			return result, false, diags
		}

		resp := defaults.Int64Response{}
		defaultValue.DefaultInt64(ctx, defaults.Int64Request{Path: fwPath}, &resp)
		diags.Append(resp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return result, false, diags
		}

		result, err = resp.PlanValue.ToTerraformValue(ctx)
	case fwschema.AttributeWithListDefaultValue:
		defaultValue := a.ListDefaultValue()
		if defaultValue == nil {
			return result, false, diags
		}

		resp := defaults.ListResponse{}
		defaultValue.DefaultList(ctx, defaults.ListRequest{Path: fwPath}, &resp)
		diags.Append(resp.Diagnostics...)

		if resp.Diagnostics.HasError() || resp.PlanValue.ElementType(ctx) == nil {
			return result, false, diags
		}

		result, err = resp.PlanValue.ToTerraformValue(ctx)
	case fwschema.AttributeWithMapDefaultValue:
		defaultValue := a.MapDefaultValue()
		if defaultValue == nil {
			return result, false, diags
		}

		resp := defaults.MapResponse{}
		defaultValue.DefaultMap(ctx, defaults.MapRequest{Path: fwPath}, &resp)
		diags.Append(resp.Diagnostics...)

		if resp.Diagnostics.HasError() || resp.PlanValue.ElementType(ctx) == nil {
			return result, false, diags
		}

		result, err = resp.PlanValue.ToTerraformValue(ctx)
	case fwschema.AttributeWithNumberDefaultValue:
		defaultValue := a.NumberDefaultValue()
		if defaultValue == nil {
			return result, false, diags
		}

		resp := defaults.NumberResponse{}
		defaultValue.DefaultNumber(ctx, defaults.NumberRequest{Path: fwPath}, &resp)
		diags.Append(resp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return result, false, diags
		}

		result, err = resp.PlanValue.ToTerraformValue(ctx)
	case fwschema.AttributeWithObjectDefaultValue:
		defaultValue := a.ObjectDefaultValue()
		if defaultValue == nil {
			return result, false, diags
		}

		resp := defaults.ObjectResponse{}
		defaultValue.DefaultObject(ctx, defaults.ObjectRequest{Path: fwPath}, &resp)
		diags.Append(resp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return result, false, diags
		}

		result, err = resp.PlanValue.ToTerraformValue(ctx)
	case fwschema.AttributeWithSetDefaultValue:
		defaultValue := a.SetDefaultValue()
		if defaultValue == nil {
			return result, false, diags
		}

		resp := defaults.SetResponse{}
		defaultValue.DefaultSet(ctx, defaults.SetRequest{Path: fwPath}, &resp)
		diags.Append(resp.Diagnostics...)

		if resp.Diagnostics.HasError() || resp.PlanValue.ElementType(ctx) == nil {
			return result, false, diags
		}

		result, err = resp.PlanValue.ToTerraformValue(ctx)
	case fwschema.AttributeWithStringDefaultValue:
		defaultValue := a.StringDefaultValue()
		if defaultValue == nil {
			return result, false, diags
		}

		resp := defaults.StringResponse{}
		defaultValue.DefaultString(ctx, defaults.StringRequest{Path: fwPath}, &resp)
		diags.Append(resp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return result, false, diags
		}

		result, err = resp.PlanValue.ToTerraformValue(ctx)
	case fwschema.AttributeWithDynamicDefaultValue:
		defaultValue := a.DynamicDefaultValue()
		if defaultValue == nil {
			return result, false, diags
		}

		resp := defaults.DynamicResponse{}
		defaultValue.DefaultDynamic(ctx, defaults.DynamicRequest{Path: fwPath}, &resp)
		diags.Append(resp.Diagnostics...)

		if resp.Diagnostics.HasError() {
			return result, false, diags
		}

		result, err = resp.PlanValue.ToTerraformValue(ctx)
	default:
		return result, false, diags
	}

	if err != nil {
		diags.AddError(
			"Generate Resource Config Error",
			"An unexpected error was encountered converting a default value at "+fwPath.String()+": "+err.Error(),
		)
		return tftypes.Value{}, false, diags
	}

	return result, true, diags
}
