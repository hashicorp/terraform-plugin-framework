// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromtftypes"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type ProposeNewStateRequest struct {
	PriorState tfsdk.State
	Config     tfsdk.Config
}

type ProposeNewStateResponse struct {
	ProposedNewState tfsdk.Plan
	Diagnostics      diag.Diagnostics
}

func SchemaProposeNewState(ctx context.Context, s fwschema.Schema, req ProposeNewStateRequest, resp *ProposeNewStateResponse) {
	if req.PriorState.Raw.IsNull() {
		// Populate prior state with a top-level round of nulls from the schema
		req.PriorState = tfsdk.State{
			Raw:    s.EmptyValue(ctx),
			Schema: s,
		}
	}

	proposedNewState, diags := proposedNew(ctx, s, tftypes.NewAttributePath(), req.PriorState.Raw, req.Config.Raw)
	resp.Diagnostics.Append(diags...)

	resp.ProposedNewState = tfsdk.Plan{
		Raw:    proposedNewState,
		Schema: s,
	}
}

func proposedNew(ctx context.Context, s fwschema.Schema, path *tftypes.AttributePath, prior, config tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if config.IsNull() {
		return config, diags
	}

	if !config.IsKnown() {
		return prior, diags
	}

	if (!prior.Type().Is(tftypes.Object{})) || (!config.Type().Is(tftypes.Object{})) {
		diags.Append(diag.NewErrorDiagnostic(
			"Invalid Value Type",
			"An unexpected error occurred while trying to create the proposed new state. "+
				"This is an error in terraform-plugin-framework used by the provider. "+
				"Please report the following to the provider developers.\n\n"+
				fmt.Sprintf("Original Error: %s", "proposedNew only supports object-typed values"),
		))
		return tftypes.Value{}, diags
	}

	newAttrs, newAttrDiags := proposedNewAttributes(ctx, s, s.GetAttributes(), path, prior, config)
	diags.Append(newAttrDiags...)
	if diags.HasError() {
		return tftypes.Value{}, diags
	}

	for name, blockType := range s.GetBlocks() {
		attrVal, err := prior.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Invalid Prior State Attribute Path",
				"An unexpected error occurred while trying to retrieve a value from prior state. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}

		priorVal := attrVal.(tftypes.Value) //nolint

		attrVal, err = config.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Invalid Config Attribute Path",
				"An unexpected error occurred while trying to retrieve a value from config. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
		configVal := attrVal.(tftypes.Value) //nolint

		nestedBlockDiags := diag.Diagnostics{} //nolint
		newAttrs[name], nestedBlockDiags = proposeNewNestedBlock(ctx, s, blockType, path.WithAttributeName(name), priorVal, configVal)
		diags.Append(nestedBlockDiags...)
		if diags.HasError() {
			return tftypes.Value{}, diags
		}
	}

	err := tftypes.ValidateValue(s.Type().TerraformType(ctx), newAttrs)
	if err != nil {
		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
		diags.Append(fwPathDiags...)

		diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
			"Invalid Value Type",
			"An unexpected error occurred while trying to create the proposed new state. "+
				"This is an error in terraform-plugin-framework used by the provider. "+
				"Please report the following to the provider developers.\n\n"+
				fmt.Sprintf("Original Error: %s", err),
		))
		return tftypes.Value{}, diags
	}
	return tftypes.NewValue(s.Type().TerraformType(ctx), newAttrs), diags
}

func proposedNewAttributes(ctx context.Context, s fwschema.Schema, attrs map[string]fwschema.Attribute, path *tftypes.AttributePath, priorObj, configObj tftypes.Value) (map[string]tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	newAttrs := make(map[string]tftypes.Value, len(attrs))

	for name, attr := range attrs {
		attrPath := path.WithAttributeName(name)

		var priorVal tftypes.Value
		switch {
		case priorObj.IsNull():
			priorObjType := priorObj.Type().(tftypes.Object) //nolint

			err := tftypes.ValidateValue(priorObjType.AttributeTypes[name], nil)
			if err != nil {
				fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, attrPath, s)
				diags.Append(fwPathDiags...)

				diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
					"Invalid Prior State Value Type",
					"An unexpected error occurred while trying to validate a value from prior state. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						fmt.Sprintf("Original Error: %s", err),
				))
				return nil, diags
			}

			priorVal = tftypes.NewValue(priorObjType.AttributeTypes[name], nil)
		case !priorObj.IsKnown():
			priorObjType := priorObj.Type().(tftypes.Object) //nolint

			err := tftypes.ValidateValue(priorObjType.AttributeTypes[name], tftypes.UnknownValue)
			if err != nil {
				fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, attrPath, s)
				diags.Append(fwPathDiags...)

				diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
					"Invalid Prior State Value Type",
					"An unexpected error occurred while trying to validate a value from prior state. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						fmt.Sprintf("Original Error: %s", err),
				))
				return nil, diags
			}

			priorVal = tftypes.NewValue(priorObjType.AttributeTypes[name], tftypes.UnknownValue)
		default:
			attrVal, err := priorObj.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
			if err != nil {
				fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, attrPath, s)
				diags.Append(fwPathDiags...)

				diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
					"Invalid Prior State Attribute Path",
					"An unexpected error occurred while trying to retrieve a value from prior state. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						fmt.Sprintf("Original Error: %s", err),
				))
				return nil, diags
			}
			priorVal = attrVal.(tftypes.Value) //nolint

		}

		var configVal tftypes.Value
		switch {
		case configObj.IsNull():
			configObjType := configObj.Type().(tftypes.Object) //nolint

			err := tftypes.ValidateValue(configObjType.AttributeTypes[name], nil)
			if err != nil {
				fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, attrPath, s)
				diags.Append(fwPathDiags...)

				diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
					"Invalid Config Value Type",
					"An unexpected error occurred while trying to validate a value from config. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						fmt.Sprintf("Original Error: %s", err),
				))
				return nil, diags
			}

			configVal = tftypes.NewValue(configObjType.AttributeTypes[name], nil)
		case !configObj.IsKnown():
			configObjType := configObj.Type().(tftypes.Object) //nolint

			err := tftypes.ValidateValue(configObjType.AttributeTypes[name], tftypes.UnknownValue)
			if err != nil {
				fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, attrPath, s)
				diags.Append(fwPathDiags...)

				diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
					"Invalid Config Value Type",
					"An unexpected error occurred while trying to validate a value from config. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						fmt.Sprintf("Original Error: %s", err),
				))
				return nil, diags
			}

			configVal = tftypes.NewValue(configObjType.AttributeTypes[name], tftypes.UnknownValue)
		default:
			configIface, err := configObj.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
			if err != nil {
				fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, attrPath, s)
				diags.Append(fwPathDiags...)

				diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
					"Invalid Config Attribute Path",
					"An unexpected error occurred while trying to retrieve a value from config. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						fmt.Sprintf("Original Error: %s", err),
				))
				return nil, diags
			}
			configVal = configIface.(tftypes.Value) //nolint

		}

		var newVal tftypes.Value
		if attr.IsComputed() && configVal.IsNull() {
			newVal = priorVal

			notComputable, notComputableDiags := optionalValueNotComputable(ctx, s, attrPath, priorVal)
			diags.Append(notComputableDiags...)
			if diags.HasError() {
				return map[string]tftypes.Value{}, diags
			}

			if notComputable {
				newVal = configVal
			}
		} else if nestedAttr, isNested := attr.(fwschema.NestedAttribute); isNested {
			nestedAttrDiags := diag.Diagnostics{} //nolint

			newVal, nestedAttrDiags = proposeNewNestedAttribute(ctx, s, nestedAttr, attrPath, priorVal, configVal)
			diags.Append(nestedAttrDiags...)
			if diags.HasError() {
				return nil, diags
			}
		} else {
			newVal = configVal
		}

		newAttrs[name] = newVal
	}

	return newAttrs, diags
}

func proposeNewNestedAttribute(ctx context.Context, s fwschema.Schema, attr fwschema.NestedAttribute, path *tftypes.AttributePath, prior, config tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// if the config isn't known at all, then we must use that value
	if !config.IsKnown() {
		return config, diags
	}

	newVal := config

	switch attr.GetNestingMode() {
	case fwschema.NestingModeSingle:
		if config.IsNull() {
			break
		}
		nestedDiags := diag.Diagnostics{} //nolint
		newVal, nestedDiags = proposedNewNestedObjectAttributes(ctx, s, attr, path, prior, config)
		diags.Append(nestedDiags...)
		if nestedDiags.HasError() {
			return tftypes.Value{}, diags
		}
	case fwschema.NestingModeList:
		nestedDiags := diag.Diagnostics{} //nolint
		newVal, nestedDiags = proposedNewListNested(ctx, s, attr, path, prior, config)
		diags.Append(nestedDiags...)
		if nestedDiags.HasError() {
			return tftypes.Value{}, diags
		}
	case fwschema.NestingModeMap:
		nestedDiags := diag.Diagnostics{} //nolint
		newVal, nestedDiags = proposedNewMapNested(ctx, s, attr, path, prior, config)
		diags.Append(nestedDiags...)
		if nestedDiags.HasError() {
			return tftypes.Value{}, diags
		}
	case fwschema.NestingModeSet:
		nestedDiags := diag.Diagnostics{} //nolint
		newVal, nestedDiags = proposedNewSetNested(ctx, s, attr, path, prior, config)
		diags.Append(nestedDiags...)
		if nestedDiags.HasError() {
			return tftypes.Value{}, diags
		}
	default:
		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
		diags.Append(fwPathDiags...)

		diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
			"Invalid Attribute Nesting Mode",
			"An unexpected error occurred while trying to construct the proposed new state. "+
				"This is an error in terraform-plugin-framework used by the provider. "+
				"Please report the following to the provider developers.\n\n"+
				fmt.Sprintf("Original Error: %s", fmt.Sprintf("unsupported attribute nesting mode %d", attr.GetNestingMode()))))

		return tftypes.Value{}, diags
	}

	return newVal, diags
}

func proposedNewMapNested(ctx context.Context, s fwschema.Schema, attr fwschema.NestedAttribute, path *tftypes.AttributePath, prior, config tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	newVal := config

	configMap := make(map[string]tftypes.Value)
	priorMap := make(map[string]tftypes.Value)

	configValLen := 0
	if !config.IsNull() {
		err := config.As(&configMap)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Error Converting Config Value",
				"An unexpected error occurred while trying to convert the config value to a go map. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
		configValLen = len(configMap)
	}

	if !prior.IsNull() {
		err := prior.As(&priorMap)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Error Converting Prior State Value",
				"An unexpected error occurred while trying to convert the prior state value to a go map. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
	}

	if configValLen > 0 {
		newVals := make(map[string]tftypes.Value, configValLen)
		for name, configEV := range configMap {
			priorEV, inPrior := priorMap[name]
			if !inPrior {
				// if the prior value was unknown the map won't have any
				// keys, so generate an unknown value.
				if !prior.IsKnown() {
					err := tftypes.ValidateValue(configEV.Type(), tftypes.UnknownValue)
					if err != nil {
						fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
						diags.Append(fwPathDiags...)

						diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
							"Invalid Config Value Type",
							"An unexpected error occurred while trying to create an unknown config value. "+
								"This is an error in terraform-plugin-framework used by the provider. "+
								"Please report the following to the provider developers.\n\n"+
								fmt.Sprintf("Original Error: %s", err),
						))
						return tftypes.Value{}, diags
					}

					priorEV = tftypes.NewValue(configEV.Type(), tftypes.UnknownValue)
				} else {
					err := tftypes.ValidateValue(configEV.Type(), nil)
					if err != nil {
						fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
						diags.Append(fwPathDiags...)

						diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
							"Invalid Config Value Type",
							"An unexpected error occurred while trying to create a null config value. "+
								"This is an error in terraform-plugin-framework used by the provider. "+
								"Please report the following to the provider developers.\n\n"+
								fmt.Sprintf("Original Error: %s", err),
						))
						return tftypes.Value{}, diags
					}

					priorEV = tftypes.NewValue(configEV.Type(), nil)

				}
			}

			nestedDiags := diag.Diagnostics{} //nolint
			newVals[name], nestedDiags = proposedNewNestedObjectAttributes(ctx, s, attr, path.WithElementKeyString(name), priorEV, configEV)
			diags.Append(nestedDiags...)
			if diags.HasError() {
				return tftypes.Value{}, diags
			}
		}

		err := tftypes.ValidateValue(config.Type(), newVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Invalid Config Value Type",
				"An unexpected error occurred while trying to create new value of the config type. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))

			return tftypes.Value{}, diags
		}

		newVal = tftypes.NewValue(config.Type(), newVals)
	}

	return newVal, diags
}

func proposedNewListNested(ctx context.Context, s fwschema.Schema, attr fwschema.NestedAttribute, path *tftypes.AttributePath, prior, config tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	newVal := config

	configVals := make([]tftypes.Value, 0)
	priorVals := make([]tftypes.Value, 0)

	configValLen := 0
	if !config.IsNull() {
		err := config.As(&configVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Error Converting Config Value",
				"An unexpected error occurred while trying to convert the config value to a go list. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
		configValLen = len(configVals)
	}

	if !prior.IsNull() {
		err := prior.As(&priorVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Error Converting Prior State Value",
				"An unexpected error occurred while trying to convert the prior state value to a go list. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
	}

	if configValLen > 0 {
		newVals := make([]tftypes.Value, 0, configValLen)
		for idx, configEV := range configVals {
			if prior.IsKnown() && (prior.IsNull() || idx >= len(priorVals)) {
				// No corresponding prior element, take config val
				newVals = append(newVals, configEV)
				continue
			}

			priorEV := priorVals[idx]
			newNestedVals, newNestedValDiags := proposedNewNestedObjectAttributes(ctx, s, attr, path.WithElementKeyInt(idx), priorEV, configEV)
			diags.Append(newNestedValDiags...)
			if diags.HasError() {
				return tftypes.Value{}, diags
			}
			newVals = append(newVals, newNestedVals)
		}

		err := tftypes.ValidateValue(config.Type(), newVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Invalid List Nested Attribute Value Type",
				"An unexpected error occurred while trying to create a list nested attribute value. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}

		newVal = tftypes.NewValue(config.Type(), newVals)
	}

	return newVal, diags
}

func proposedNewSetNested(ctx context.Context, s fwschema.Schema, attr fwschema.NestedAttribute, path *tftypes.AttributePath, prior, config tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	newVal := config

	configVals := make([]tftypes.Value, 0)
	priorVals := make([]tftypes.Value, 0)

	configValLen := 0
	if !config.IsNull() {
		err := config.As(&configVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Error Converting Config Value",
				"An unexpected error occurred while trying to convert the config value to a go list. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
		configValLen = len(configVals)
	}

	if !prior.IsNull() {
		err := prior.As(&priorVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Error Converting Prior State Value",
				"An unexpected error occurred while trying to convert the prior state value to a go list. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
	}

	if configValLen > 0 {
		// track which prior elements have been used
		used := make([]bool, len(priorVals))
		newVals := make([]tftypes.Value, 0, configValLen)
		for _, configEV := range configVals {
			var priorEV tftypes.Value
			for i, priorCmp := range priorVals {
				if used[i] {
					continue
				}

				// It is possible that multiple prior elements could be valid
				// matches for a configuration value, in which case we will end up
				// picking the first match encountered (but it will always be
				// consistent due to cty's iteration order). Because configured set
				// elements must also be entirely unique in order to be included in
				// the set, these matches either will not matter because they only
				// differ by computed values, or could not have come from a valid
				// config with all unique set elements.
				if validPriorFromConfig(ctx, s, path, priorCmp, configEV) {
					priorEV = priorCmp
					used[i] = true
					break
				}
			}

			if priorEV.IsNull() {
				err := tftypes.ValidateValue(attr.GetNestedObject().Type().TerraformType(ctx), nil)
				if err != nil {
					fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
					diags.Append(fwPathDiags...)

					diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
						"Invalid Prior State Value Type",
						"An unexpected error occurred while trying to create an null prior state value. "+
							"This is an error in terraform-plugin-framework used by the provider. "+
							"Please report the following to the provider developers.\n\n"+
							fmt.Sprintf("Original Error: %s", err),
					))
					return tftypes.Value{}, diags
				}

				priorEV = tftypes.NewValue(attr.GetNestedObject().Type().TerraformType(ctx), nil)
			}
			newNestedVals, newNestedValDiags := proposedNewNestedObjectAttributes(ctx, s, attr, path.WithElementKeyValue(priorEV), priorEV, configEV)
			diags.Append(newNestedValDiags...)
			if diags.HasError() {
				return tftypes.Value{}, diags
			}
			newVals = append(newVals, newNestedVals)
		}
		err := tftypes.ValidateValue(config.Type(), newVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Invalid Set Nested Attribute Value Type",
				"An unexpected error occurred while trying to create a set nested attribute value. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}

		newVal = tftypes.NewValue(config.Type(), newVals)
	}

	return newVal, diags
}

func proposedNewNestedObjectAttributes(ctx context.Context, s fwschema.Schema, attr fwschema.NestedAttribute, path *tftypes.AttributePath, prior, config tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if config.IsNull() {
		return config, diags
	}

	objType := attr.GetNestedObject().Type().TerraformType(ctx)
	newAttrs, newAttrsDiags := proposedNewAttributes(ctx, s, attr.GetNestedObject().GetAttributes(), path, prior, config)
	diags.Append(newAttrsDiags...)
	if diags.HasError() {
		return tftypes.Value{}, diags
	}

	err := tftypes.ValidateValue(
		objType,
		newAttrs,
	)
	if err != nil {
		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
		diags.Append(fwPathDiags...)

		diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
			"Invalid Nested Attribute Value Type",
			"An unexpected error occurred while trying to create a nested attribute value. "+
				"This is an error in terraform-plugin-framework used by the provider. "+
				"Please report the following to the provider developers.\n\n"+
				fmt.Sprintf("Original Error: %s", err),
		))
		return tftypes.Value{}, diags
	}

	return tftypes.NewValue(
		objType,
		newAttrs,
	), diags
}

func proposeNewNestedBlock(ctx context.Context, s fwschema.Schema, block fwschema.Block, path *tftypes.AttributePath, prior, config tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	// if the config isn't known at all, then we must use that value
	if !config.IsKnown() {
		return config, diags
	}

	newVal := config

	switch block.GetNestingMode() {
	case fwschema.BlockNestingModeSingle:
		if config.IsNull() {
			break
		}
		blockDiags := diag.Diagnostics{} //nolint
		newVal, blockDiags = proposedNewNestedBlockObjectAttributes(ctx, s, block, path, prior, config)
		diags.Append(blockDiags...)
		if blockDiags.HasError() {
			return tftypes.Value{}, diags
		}
	case fwschema.BlockNestingModeList:
		blockDiags := diag.Diagnostics{} //nolint
		newVal, blockDiags = proposedNewBlockListNested(ctx, s, block, path, prior, config)
		diags.Append(blockDiags...)
		if blockDiags.HasError() {
			return tftypes.Value{}, diags
		}
	case fwschema.BlockNestingModeSet:
		blockDiags := diag.Diagnostics{} //nolint
		newVal, blockDiags = proposedNewBlockSetNested(ctx, s, block, path, prior, config)
		diags.Append(blockDiags...)
		if blockDiags.HasError() {
			return tftypes.Value{}, diags
		}
	default:
		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
		diags.Append(fwPathDiags...)

		diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
			"Invalid Block Nesting Mode",
			"An unexpected error occurred while trying to construct the proposed new state. "+
				"This is an error in terraform-plugin-framework used by the provider. "+
				"Please report the following to the provider developers.\n\n"+
				fmt.Sprintf("Original Error: %s", fmt.Sprintf("unsupported attribute nesting mode %d", block.GetNestingMode()))))

		return tftypes.Value{}, diags
	}

	return newVal, diags
}

func proposedNewNestedBlockObjectAttributes(ctx context.Context, s fwschema.Schema, block fwschema.Block, path *tftypes.AttributePath, prior, config tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	if config.IsNull() {
		return config, diags
	}
	valuesMap, attrDiags := proposedNewAttributes(ctx, s, block.GetNestedObject().GetAttributes(), path, prior, config)
	diags.Append(attrDiags...)
	if diags.HasError() {
		return tftypes.Value{}, diags
	}

	for name, blockType := range block.GetNestedObject().GetBlocks() {
		attrVal, err := prior.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Invalid Prior State Attribute Path",
				"An unexpected error occurred while trying to retrieve a value from prior state. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
		priorVal := attrVal.(tftypes.Value) //nolint

		attrVal, err = config.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Invalid Config Attribute Path",
				"An unexpected error occurred while trying to retrieve a value from config. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
		configVal := attrVal.(tftypes.Value) //nolint

		nestedBlockDiags := diag.Diagnostics{} //nolint
		valuesMap[name], nestedBlockDiags = proposeNewNestedBlock(ctx, s, blockType, tftypes.NewAttributePath().WithAttributeName(name).WithElementKeyInt(0), priorVal, configVal)
		diags.Append(nestedBlockDiags...)
		if diags.HasError() {
			return tftypes.Value{}, diags
		}
	}

	objType := block.GetNestedObject().Type().TerraformType(ctx)

	err := tftypes.ValidateValue(
		objType,
		valuesMap,
	)
	if err != nil {
		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
		diags.Append(fwPathDiags...)

		diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
			"Invalid Nested Block Object Value Type",
			"An unexpected error occurred while trying to create a nested block object value. "+
				"This is an error in terraform-plugin-framework used by the provider. "+
				"Please report the following to the provider developers.\n\n"+
				fmt.Sprintf("Original Error: %s", err),
		))
		return tftypes.Value{}, diags
	}

	return tftypes.NewValue(
		objType,
		valuesMap,
	), diags
}

func proposedNewBlockListNested(ctx context.Context, s fwschema.Schema, block fwschema.Block, path *tftypes.AttributePath, prior, config tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	newVal := config

	configVals := make([]tftypes.Value, 0)
	priorVals := make([]tftypes.Value, 0)

	configValLen := 0
	if !config.IsNull() {
		err := config.As(&configVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Error Converting Config Value",
				"An unexpected error occurred while trying to convert the config value to a go list. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
		configValLen = len(configVals)
	}

	if !prior.IsNull() {
		err := prior.As(&priorVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Error Converting Prior State Value",
				"An unexpected error occurred while trying to convert the prior state value to a go list. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
	}

	if configValLen > 0 {
		newVals := make([]tftypes.Value, 0, configValLen)
		for idx, configEV := range configVals {
			if prior.IsKnown() && (prior.IsNull() || idx >= len(priorVals)) {
				// No corresponding prior element, take config val
				newVals = append(newVals, configEV)
				continue
			}

			priorEV := priorVals[idx]
			newNestedVal, newNestedValDiags := proposedNewNestedBlockObjectAttributes(ctx, s, block, path.WithElementKeyInt(idx), priorEV, configEV)
			diags.Append(newNestedValDiags...)
			if diags.HasError() {
				return tftypes.Value{}, nil
			}
			newVals = append(newVals, newNestedVal)
		}

		err := tftypes.ValidateValue(config.Type(), newVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Invalid List Nested Block Value Type",
				"An unexpected error occurred while trying to create a list nested block value. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}

		newVal = tftypes.NewValue(config.Type(), newVals)
	}

	return newVal, diags
}

func proposedNewBlockSetNested(ctx context.Context, s fwschema.Schema, block fwschema.Block, path *tftypes.AttributePath, prior, config tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	newVal := config

	configVals := make([]tftypes.Value, 0)
	priorVals := make([]tftypes.Value, 0)

	configValLen := 0
	if !config.IsNull() {
		err := config.As(&configVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Error Converting Config Value",
				"An unexpected error occurred while trying to convert the config value to a go list. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
		configValLen = len(configVals)
	}

	if !prior.IsNull() {
		err := prior.As(&priorVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Error Converting Prior State Value",
				"An unexpected error occurred while trying to convert the prior state value to a go list. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
	}

	if configValLen > 0 {
		// track which prior elements have been used
		used := make([]bool, len(priorVals))
		newVals := make([]tftypes.Value, 0, configValLen)
		for _, configEV := range configVals {
			var priorEV tftypes.Value
			for i, priorCmp := range priorVals {
				if used[i] {
					continue
				}

				// It is possible that multiple prior elements could be valid
				// matches for a configuration value, in which case we will end up
				// picking the first match encountered (but it will always be
				// consistent due to cty's iteration order). Because configured set
				// elements must also be entirely unique in order to be included in
				// the set, these matches either will not matter because they only
				// differ by computed values, or could not have come from a valid
				// config with all unique set elements.
				if validPriorFromConfig(ctx, s, path, priorCmp, configEV) {
					priorEV = priorCmp
					used[i] = true
					break
				}
			}

			if priorEV.IsNull() {
				err := tftypes.ValidateValue(block.GetNestedObject().Type().TerraformType(ctx), nil)
				if err != nil {
					fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
					diags.Append(fwPathDiags...)

					diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
						"Invalid Prior State Value Type",
						"An unexpected error occurred while trying to create an null prior state value. "+
							"This is an error in terraform-plugin-framework used by the provider. "+
							"Please report the following to the provider developers.\n\n"+
							fmt.Sprintf("Original Error: %s", err),
					))
					return tftypes.Value{}, diags
				}

				priorEV = tftypes.NewValue(block.GetNestedObject().Type().TerraformType(ctx), nil)
			}
			newNestedVal, newNestedValDiags := proposeNewNestedBlockObject(ctx, s, block.GetNestedObject(), path.WithElementKeyValue(priorEV), priorEV, configEV)
			diags.Append(newNestedValDiags...)
			if diags.HasError() {
				return tftypes.Value{}, nil
			}

			newVals = append(newVals, newNestedVal)
		}

		err := tftypes.ValidateValue(config.Type(), newVals)
		if err != nil {
			fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
			diags.Append(fwPathDiags...)

			diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
				"Invalid Set Nested Block Value Type",
				"An unexpected error occurred while trying to create a set nested block value. "+
					"This is an error in terraform-plugin-framework used by the provider. "+
					"Please report the following to the provider developers.\n\n"+
					fmt.Sprintf("Original Error: %s", err),
			))
			return tftypes.Value{}, diags
		}
		newVal = tftypes.NewValue(config.Type(), newVals)
	}

	return newVal, diags
}

func proposeNewNestedBlockObject(ctx context.Context, s fwschema.Schema, nestedBlock fwschema.NestedBlockObject, path *tftypes.AttributePath, prior, config tftypes.Value) (tftypes.Value, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	if config.IsNull() {
		return config, diags
	}
	valuesMap, attrDiags := proposedNewAttributes(ctx, s, nestedBlock.GetAttributes(), path, prior, config)
	diags.Append(attrDiags...)
	if diags.HasError() {
		return tftypes.Value{}, diags
	}

	for name, blockType := range nestedBlock.GetBlocks() {
		var priorVal tftypes.Value
		if prior.IsNull() {
			priorObjType := prior.Type().(tftypes.Object) //nolint

			err := tftypes.ValidateValue(priorObjType.AttributeTypes[name], nil)
			if err != nil {
				fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
				diags.Append(fwPathDiags...)

				diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
					"Invalid Prior State Value Type",
					"An unexpected error occurred while trying to validate a value from prior state. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						fmt.Sprintf("Original Error: %s", err),
				))
				return tftypes.Value{}, diags
			}

			priorVal = tftypes.NewValue(priorObjType.AttributeTypes[name], nil)
		} else {
			attrVal, err := prior.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
			if err != nil {
				fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
				diags.Append(fwPathDiags...)

				diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
					"Invalid Prior State Attribute Path",
					"An unexpected error occurred while trying to retrieve a value from prior state. "+
						"This is an error in terraform-plugin-framework used by the provider. "+
						"Please report the following to the provider developers.\n\n"+
						fmt.Sprintf("Original Error: %s", err),
				))
				return tftypes.Value{}, diags
			}
			priorVal = attrVal.(tftypes.Value) //nolint
		}

		attrVal, _ := config.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
		configVal := attrVal.(tftypes.Value) //nolint

		nestedBlockDiags := diag.Diagnostics{} //nolint
		valuesMap[name], nestedBlockDiags = proposeNewNestedBlock(ctx, s, blockType, path.WithAttributeName(name), priorVal, configVal)
		diags.Append(nestedBlockDiags...)
		if nestedBlockDiags.HasError() {
			return tftypes.Value{}, diags
		}

	}

	err := tftypes.ValidateValue(nestedBlock.Type().TerraformType(ctx), valuesMap)
	if err != nil {
		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, path, s)
		diags.Append(fwPathDiags...)

		diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
			"Invalid Nested Block Value Type",
			"An unexpected error occurred while trying to create a nested block value. "+
				"This is an error in terraform-plugin-framework used by the provider. "+
				"Please report the following to the provider developers.\n\n"+
				fmt.Sprintf("Original Error: %s", err),
		))
		return tftypes.Value{}, diags
	}

	return tftypes.NewValue(
		nestedBlock.Type().TerraformType(ctx),
		valuesMap,
	), diags
}

func optionalValueNotComputable(ctx context.Context, s fwschema.Schema, absPath *tftypes.AttributePath, val tftypes.Value) (bool, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attr, err := s.AttributeAtTerraformPath(ctx, absPath)
	if err != nil {
		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, absPath, s)
		diags.Append(fwPathDiags...)

		diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
			"Invalid Attribute Path",
			"An unexpected error occurred while trying to retrieve attribute at path. "+
				"This is an error in terraform-plugin-framework used by the provider. "+
				"Please report the following to the provider developers.\n\n"+
				fmt.Sprintf("Original Error: %s", err),
		))

		return false, diags
	}

	if !attr.IsOptional() { //nolint
		return false, diags
	}

	_, nested := attr.(fwschema.NestedAttribute)
	if !nested {
		return false, diags
	}

	foundNonComputedAttr := false
	err = tftypes.Walk(val, func(path *tftypes.AttributePath, v tftypes.Value) (bool, error) { //nolint
		if v.IsNull() {
			return true, nil
		}

		// Continue past the root
		if len(path.Steps()) < 1 {
			return true, nil
		}

		attrPath := tftypes.NewAttributePathWithSteps(append(absPath.Steps(), path.Steps()...))
		attrSchema, err := s.AttributeAtTerraformPath(ctx, attrPath)
		if err != nil {
			return true, nil //nolint
		}

		if !attrSchema.IsComputed() {
			foundNonComputedAttr = true
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, absPath, s)
		diags.Append(fwPathDiags...)

		diags.Append(diag.NewAttributeErrorDiagnostic(fwPath,
			"Invalid Attribute Path",
			"An unexpected error occurred while trying to walk the value at path. "+
				"This is an error in terraform-plugin-framework used by the provider. "+
				"Please report the following to the provider developers.\n\n"+
				fmt.Sprintf("Original Error: %s", err),
		))

	}

	return foundNonComputedAttr, diags
}

// validPriorFromConfig returns true if the prior object could have been
// derived from the configuration. We do this by walking the prior value to
// determine if it is a valid superset of the config, and only computable
// values have been added. This function is only used to correlated
// configuration with possible valid prior values within sets.
func validPriorFromConfig(ctx context.Context, s fwschema.Schema, absPath *tftypes.AttributePath, prior, config tftypes.Value) bool {
	if config.Equal(prior) {
		return true
	}

	// error value to halt the walk
	stop := errors.New("stop")

	valid := true
	_ = tftypes.Walk(prior, func(path *tftypes.AttributePath, priorV tftypes.Value) (bool, error) {
		if priorV.IsNull() {
			return true, nil
		}

		// Continue past the root
		if len(path.Steps()) < 1 {
			return true, nil
		}

		configIface, _, err := tftypes.WalkAttributePath(config, path)
		if err != nil {
			// most likely dynamic objects with different types
			valid = false
			return false, stop
		}
		configV := configIface.(tftypes.Value) //nolint

		// we don't need to know the schema if both are equal
		if configV.Equal(priorV) {
			// we know they are equal, so no need to descend further
			return false, nil
		}

		// We can't descend into nested sets to correlate configuration, so the
		// overall values must be equal.
		if configV.Type().Is(tftypes.Set{}) {
			valid = false
			return false, stop
		}
		setValPath := tftypes.NewAttributePath().WithElementKeyValue(prior)

		attrPath := tftypes.NewAttributePathWithSteps(append(absPath.Steps(), append(setValPath.Steps(), path.Steps()...)...))
		attrSchema, err := s.AttributeAtTerraformPath(ctx, attrPath)
		if err != nil {
			// Not at a schema attribute, so we can continue until we find leaf
			// attributes.
			return true, nil //nolint
		}

		// If we have nested object attributes we'll be descending into those
		// to compare the individual values and determine why this level is not
		// equal
		_, isNestedType := attrSchema.GetType().(attr.TypeWithAttributeTypes)
		if isNestedType {
			return true, nil
		}

		// This is a leaf attribute, so it must be computed in order to differ
		// from config.
		if !attrSchema.IsComputed() {
			valid = false
			return false, stop
		}

		// And if it is computed, the config must be null to allow a change.
		if !configV.IsNull() {
			valid = false
			return false, stop
		}

		// We sill stop here. The cty value could be far larger, but this was
		// the last level of prescribed schema.
		return false, nil
	})

	return valid
}
