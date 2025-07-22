package fwserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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

	proposedNewState := proposedNew(ctx, s, tftypes.NewAttributePath(), req.PriorState.Raw, req.Config.Raw)

	resp.ProposedNewState = tfsdk.Plan{
		Raw:    proposedNewState,
		Schema: s,
	}
}

func proposedNew(ctx context.Context, s fwschema.Schema, path *tftypes.AttributePath, prior, config tftypes.Value) tftypes.Value {
	if config.IsNull() {
		return config
	}

	if !config.IsKnown() {
		return prior
	}

	if (!prior.Type().Is(tftypes.Object{})) || (!config.Type().Is(tftypes.Object{})) {
		// TODO: switch to non-panics
		panic("proposedNew only supports object-typed values")
	}

	newAttrs := proposedNewAttributes(ctx, s, s.GetAttributes(), path, prior, config)

	for name, blockType := range s.GetBlocks() {
		attrVal, _ := prior.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
		priorVal := attrVal.(tftypes.Value)

		attrVal, _ = config.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
		configVal := attrVal.(tftypes.Value)
		newAttrs[name] = proposeNewNestedBlock(ctx, s, blockType, path.WithAttributeName(name), priorVal, configVal)
	}

	err := tftypes.ValidateValue(s.Type().TerraformType(ctx), newAttrs)
	if err != nil {
		// TODO: throw diag
		return tftypes.Value{}
	}
	return tftypes.NewValue(s.Type().TerraformType(ctx), newAttrs)
}

func proposedNewAttributes(ctx context.Context, s fwschema.Schema, attrs map[string]fwschema.Attribute, path *tftypes.AttributePath, priorObj, configObj tftypes.Value) map[string]tftypes.Value {
	newAttrs := make(map[string]tftypes.Value, len(attrs))
	for name, attr := range attrs {
		attrPath := path.WithAttributeName(name)

		var priorVal tftypes.Value
		switch {
		case priorObj.IsNull():
			priorObjType := priorObj.Type().(tftypes.Object) //nolint

			err := tftypes.ValidateValue(priorObjType.AttributeTypes[name], nil)
			if err != nil {
				// TODO: handle error
				return nil
			}

			priorVal = tftypes.NewValue(priorObjType.AttributeTypes[name], nil)
		case !priorObj.IsKnown():
			priorObjType := priorObj.Type().(tftypes.Object) //nolint

			err := tftypes.ValidateValue(priorObjType.AttributeTypes[name], tftypes.UnknownValue)
			if err != nil {
				// TODO: handle error
				return nil
			}

			priorVal = tftypes.NewValue(priorObjType.AttributeTypes[name], tftypes.UnknownValue)
		default:
			// TODO: handle error
			attrVal, err := priorObj.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
			if err != nil {
				panic(err)
			}
			priorVal = attrVal.(tftypes.Value) //nolint

		}

		var configVal tftypes.Value
		switch {
		case configObj.IsNull():
			configObjType := configObj.Type().(tftypes.Object) //nolint

			err := tftypes.ValidateValue(configObjType.AttributeTypes[name], nil)
			if err != nil {
				// TODO: handle error
				return nil
			}

			configVal = tftypes.NewValue(configObjType.AttributeTypes[name], nil)
		case !configObj.IsKnown():
			configObjType := configObj.Type().(tftypes.Object) //nolint

			err := tftypes.ValidateValue(configObjType.AttributeTypes[name], tftypes.UnknownValue)
			if err != nil {
				// TODO: handle error
				return nil
			}

			configVal = tftypes.NewValue(configObjType.AttributeTypes[name], tftypes.UnknownValue)
		default:
			// TODO: handle error
			configIface, err := configObj.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
			if err != nil {
				panic(err)
			}
			configVal = configIface.(tftypes.Value) //nolint

		}

		var newVal tftypes.Value
		if attr.IsComputed() && configVal.IsNull() {
			newVal = priorVal

			if optionalValueNotComputable(ctx, s, attrPath, priorVal) {
				newVal = configVal
			}
		} else if nestedAttr, isNested := attr.(fwschema.NestedAttribute); isNested {
			newVal = proposeNewNestedAttribute(ctx, s, nestedAttr, attrPath, priorVal, configVal)
		} else {
			newVal = configVal
		}

		newAttrs[name] = newVal
	}

	return newAttrs
}

func proposeNewNestedBlock(ctx context.Context, s fwschema.Schema, block fwschema.Block, path *tftypes.AttributePath, prior, config tftypes.Value) tftypes.Value {
	// if the config isn't known at all, then we must use that value
	if !config.IsKnown() {
		return config
	}

	newVal := config

	switch block.GetNestingMode() {
	case fwschema.BlockNestingModeSingle:
		if config.IsNull() {
			break
		}
		newVal = proposedNewBlockObjectAttributes(ctx, s, block, path, prior, config)
	case fwschema.BlockNestingModeList:
		newVal = proposedNewBlockListNested(ctx, s, block, path, prior, config)
	case fwschema.BlockNestingModeSet:
		newVal = proposedNewBlockSetNested(ctx, s, block, path, prior, config)
	default:
		// TODO: Shouldn't happen, return diag
		panic(fmt.Sprintf("unsupported attribute nesting mode %d", block.GetNestingMode()))
	}

	return newVal
}

func proposeNewNestedBlockObject(ctx context.Context, s fwschema.Schema, nestedBlock fwschema.NestedBlockObject, path *tftypes.AttributePath, prior, config tftypes.Value) tftypes.Value {
	if config.IsNull() {
		return config
	}
	valuesMap := proposedNewAttributes(ctx, s, nestedBlock.GetAttributes(), path, prior, config)

	for name, blockType := range nestedBlock.GetBlocks() {
		var priorVal tftypes.Value
		if prior.IsNull() {
			priorObjType := prior.Type().(tftypes.Object) //nolint

			err := tftypes.ValidateValue(priorObjType.AttributeTypes[name], nil)
			if err != nil {
				// TODO: handle error
				return tftypes.Value{}
			}

			priorVal = tftypes.NewValue(priorObjType.AttributeTypes[name], nil)
		} else {
			// TODO: handle error
			attrVal, err := prior.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
			if err != nil {
				panic(err)
			}
			priorVal = attrVal.(tftypes.Value) //nolint
		}

		attrVal, _ := config.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
		configVal := attrVal.(tftypes.Value)
		valuesMap[name] = proposeNewNestedBlock(ctx, s, blockType, path.WithAttributeName(name), priorVal, configVal)
	}

	err := tftypes.ValidateValue(nestedBlock.Type().TerraformType(ctx), valuesMap)
	if err != nil {
		// TODO: handle error
		return tftypes.Value{}
	}

	return tftypes.NewValue(
		nestedBlock.Type().TerraformType(ctx),
		valuesMap,
	)
}

func proposeNewNestedAttribute(ctx context.Context, s fwschema.Schema, attr fwschema.NestedAttribute, path *tftypes.AttributePath, prior, config tftypes.Value) tftypes.Value {
	// if the config isn't known at all, then we must use that value
	if !config.IsKnown() {
		return config
	}

	newVal := config

	switch attr.GetNestingMode() {
	case fwschema.NestingModeSingle:
		if config.IsNull() {
			break
		}
		newVal = proposedNewObjectAttributes(ctx, s, attr, path, prior, config)
	case fwschema.NestingModeList:
		newVal = proposedNewListNested(ctx, s, attr, path, prior, config)
	case fwschema.NestingModeMap:
		newVal = proposedNewMapNested(ctx, s, attr, path, prior, config)
	case fwschema.NestingModeSet:
		newVal = proposedNewSetNested(ctx, s, attr, path, prior, config)
	default:
		// TODO: Shouldn't happen, return diag
		panic(fmt.Sprintf("unsupported attribute nesting mode %d", attr.GetNestingMode()))
	}

	return newVal
}

func proposedNewMapNested(ctx context.Context, s fwschema.Schema, attr fwschema.NestedAttribute, path *tftypes.AttributePath, prior, config tftypes.Value) tftypes.Value {
	newVal := config

	configMap := make(map[string]tftypes.Value)
	priorMap := make(map[string]tftypes.Value)

	configValLen := 0
	if !config.IsNull() {
		err := config.As(&priorMap)
		// TODO: handle err
		if err != nil {
			panic(err)
		}
		configValLen = len(configMap)
	}

	if !prior.IsNull() {
		err := prior.As(&priorMap)
		// TODO: handle err
		if err != nil {
			panic(err)
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
						// TODO: handle error
						return tftypes.Value{}
					}

					priorEV = tftypes.NewValue(configEV.Type(), tftypes.UnknownValue)
				} else {
					err := tftypes.ValidateValue(configEV.Type(), nil)
					if err != nil {
						// TODO: handle error
						return tftypes.Value{}
					}

					priorEV = tftypes.NewValue(configEV.Type(), nil)

				}
			}

			newVals[name] = proposedNewObjectAttributes(ctx, s, attr, path.WithElementKeyString(name), priorEV, configEV)
		}

		err := tftypes.ValidateValue(config.Type(), newVals)
		if err != nil {
			// TODO: handle error
			return tftypes.Value{}
		}

		newVal = tftypes.NewValue(config.Type(), newVals)
	}

	return newVal
}

func proposedNewBlockListNested(ctx context.Context, s fwschema.Schema, block fwschema.Block, path *tftypes.AttributePath, prior, config tftypes.Value) tftypes.Value {
	newVal := config

	configVals := make([]tftypes.Value, 0)
	priorVals := make([]tftypes.Value, 0)

	configValLen := 0
	if !config.IsNull() {
		err := config.As(&configVals)
		// TODO: handle err
		if err != nil {
			panic(err)
		}
		configValLen = len(configVals)
	}

	if !prior.IsNull() {
		err := prior.As(&priorVals)
		// TODO: handle err
		if err != nil {
			panic(err)
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
			newVals = append(newVals, proposedNewBlockObjectAttributes(ctx, s, block, path.WithElementKeyInt(idx), priorEV, configEV))
		}

		err := tftypes.ValidateValue(config.Type(), newVals)
		if err != nil {
			// TODO: handle error
			return tftypes.Value{}
		}

		newVal = tftypes.NewValue(config.Type(), newVals)
	}

	return newVal
}

func proposedNewBlockSetNested(ctx context.Context, s fwschema.Schema, block fwschema.Block, path *tftypes.AttributePath, prior, config tftypes.Value) tftypes.Value {
	newVal := config

	configVals := make([]tftypes.Value, 0)
	priorVals := make([]tftypes.Value, 0)

	configValLen := 0
	if !config.IsNull() {
		err := config.As(&configVals)
		// TODO: handle err
		if err != nil {
			panic(err)
		}
		configValLen = len(configVals)
	}

	if !prior.IsNull() {
		err := prior.As(&priorVals)
		// TODO: handle err
		if err != nil {
			panic(err)
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
					// TODO: handle error
					return tftypes.Value{}
				}

				priorEV = tftypes.NewValue(block.GetNestedObject().Type().TerraformType(ctx), nil)
			}
			newVals = append(newVals, proposeNewNestedBlockObject(ctx, s, block.GetNestedObject(), path.WithElementKeyValue(priorEV), priorEV, configEV))
		}

		err := tftypes.ValidateValue(config.Type(), newVals)
		if err != nil {
			// TODO: handle error
			return tftypes.Value{}
		}
		newVal = tftypes.NewValue(config.Type(), newVals)
	}

	return newVal
}

func proposedNewListNested(ctx context.Context, s fwschema.Schema, attr fwschema.NestedAttribute, path *tftypes.AttributePath, prior, config tftypes.Value) tftypes.Value {
	newVal := config

	configVals := make([]tftypes.Value, 0)
	priorVals := make([]tftypes.Value, 0)

	configValLen := 0
	if !config.IsNull() {
		err := config.As(&configVals)
		// TODO: handle err
		if err != nil {
			panic(err)
		}
		configValLen = len(configVals)
	}

	if !prior.IsNull() {
		err := prior.As(&priorVals)
		// TODO: handle err
		if err != nil {
			panic(err)
		}
	}

	if configValLen > 0 {
		newVals := make([]tftypes.Value, 0, configValLen)
		for idx, configEV := range configVals {
			if prior.IsKnown() && (prior.IsNull() || idx > len(priorVals)) {
				// No corresponding prior element, take config val
				newVals = append(newVals, configEV)
				continue
			}

			priorEV := priorVals[idx]
			newVals = append(newVals, proposedNewObjectAttributes(ctx, s, attr, path.WithElementKeyInt(idx), priorEV, configEV))
		}

		err := tftypes.ValidateValue(config.Type(), newVals)
		if err != nil {
			// TODO: handle error
			return tftypes.Value{}
		}

		newVal = tftypes.NewValue(config.Type(), newVals)
	}

	return newVal
}

func proposedNewSetNested(ctx context.Context, s fwschema.Schema, attr fwschema.NestedAttribute, path *tftypes.AttributePath, prior, config tftypes.Value) tftypes.Value {
	newVal := config

	configVals := make([]tftypes.Value, 0)
	priorVals := make([]tftypes.Value, 0)

	configValLen := 0
	if !config.IsNull() {
		err := config.As(&configVals)
		// TODO: handle err
		if err != nil {
			panic(err)
		}
		configValLen = len(configVals)
	}

	if !prior.IsNull() {
		err := prior.As(&priorVals)
		// TODO: handle err
		if err != nil {
			panic(err)
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
					// TODO: handle error
					return tftypes.Value{}
				}

				priorEV = tftypes.NewValue(attr.GetNestedObject().Type().TerraformType(ctx), nil)
			}
			newVals = append(newVals, proposedNewObjectAttributes(ctx, s, attr, path.WithElementKeyValue(priorEV), priorEV, configEV))
		}
		err := tftypes.ValidateValue(config.Type(), newVals)
		if err != nil {
			// TODO: handle error
			return tftypes.Value{}
		}

		newVal = tftypes.NewValue(config.Type(), newVals)
	}

	return newVal
}

func proposedNewObjectAttributes(ctx context.Context, s fwschema.Schema, attr fwschema.NestedAttribute, path *tftypes.AttributePath, prior, config tftypes.Value) tftypes.Value {
	if config.IsNull() {
		return config
	}

	objType := attr.GetNestedObject().Type().TerraformType(ctx)
	newAttrs := proposedNewAttributes(ctx, s, attr.GetNestedObject().GetAttributes(), path, prior, config)

	err := tftypes.ValidateValue(
		objType,
		newAttrs,
	)
	if err != nil {
		// TODO: handle error
		return tftypes.Value{}
	}

	return tftypes.NewValue(
		objType,
		newAttrs,
	)
}

func proposedNewBlockObjectAttributes(ctx context.Context, s fwschema.Schema, block fwschema.Block, path *tftypes.AttributePath, prior, config tftypes.Value) tftypes.Value {
	if config.IsNull() {
		return config
	}
	valuesMap := proposedNewAttributes(ctx, s, block.GetNestedObject().GetAttributes(), path, prior, config)

	for name, blockType := range block.GetNestedObject().GetBlocks() {
		attrVal, err := prior.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
		//TODO handle panic
		if err != nil {
			panic(err)
		}
		priorVal := attrVal.(tftypes.Value)

		attrVal, _ = config.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
		configVal := attrVal.(tftypes.Value)
		valuesMap[name] = proposeNewNestedBlock(ctx, s, blockType, tftypes.NewAttributePath().WithAttributeName(name).WithElementKeyInt(0), priorVal, configVal)
	}

	objType := block.GetNestedObject().Type().TerraformType(ctx)

	err := tftypes.ValidateValue(
		objType,
		valuesMap,
	)
	if err != nil {
		// TODO: handle error
		return tftypes.Value{}
	}

	return tftypes.NewValue(
		objType,
		valuesMap,
	)
}

func optionalValueNotComputable(ctx context.Context, s fwschema.Schema, absPath *tftypes.AttributePath, val tftypes.Value) bool {
	// TODO: handle error
	attr, err := s.AttributeAtTerraformPath(ctx, absPath)
	if err != nil {
		panic(err)
	}

	if !attr.IsOptional() { //nolint
		return false
	}

	_, nested := attr.(fwschema.NestedAttribute)
	if !nested {
		return false
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
		//TODO handle panic
		panic(err)
	}

	return foundNonComputedAttr
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
		configV := configIface.(tftypes.Value)

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
