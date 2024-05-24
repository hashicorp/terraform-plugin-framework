package fwserver

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
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
	// TODO: This is in core's logic, but I'm not sure what how this scenario would be triggered
	// Need to verify if it's relevant...
	if req.Config.Raw.IsNull() && req.PriorState.Raw.IsNull() {
		resp.ProposedNewState = stateToPlan(req.PriorState)
		return
	}

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
	// TODO: This is in core's logic, but I'm not sure what how this scenario would be triggered
	// Need to verify if it's relevant...
	if config.IsNull() || !config.IsKnown() {
		return prior
	}

	if (!prior.Type().Is(tftypes.Object{})) || (!config.Type().Is(tftypes.Object{})) {
		// TODO: switch to non-panics
		panic("proposedNew only supports object-typed values")
	}

	newAttrs := proposedNewAttributes(ctx, s, s.GetAttributes(), path, prior, config)

	// TODO: add block logic

	// TODO: validate before doing this? To avoid panic
	return tftypes.NewValue(s.Type().TerraformType(ctx), newAttrs)
}

func proposedNewAttributes(ctx context.Context, s fwschema.Schema, attrs map[string]fwschema.Attribute, path *tftypes.AttributePath, priorObj, configObj tftypes.Value) map[string]tftypes.Value {
	newAttrs := make(map[string]tftypes.Value, len(attrs))
	for name, attr := range attrs {
		attrPath := path.WithAttributeName(name)

		var priorVal tftypes.Value
		if priorObj.IsNull() {
			priorObjType := priorObj.Type().(tftypes.Object) //nolint
			// TODO: validate before doing this? To avoid panic
			priorVal = tftypes.NewValue(priorObjType.AttributeTypes[name], nil)
		} else {
			// TODO: handle error
			attrVal, _ := priorObj.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
			priorVal = attrVal.(tftypes.Value) //nolint
		}

		// TODO: handle error
		configIface, _ := configObj.ApplyTerraform5AttributePathStep(tftypes.AttributeName(name))
		configVal := configIface.(tftypes.Value) //nolint

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
		// TODO: handle map
	case fwschema.NestingModeSet:
		// TODO: handle set
	default:
		// TODO: Shouldn't happen, return diag
		panic(fmt.Sprintf("unsupported attribute nesting mode %d", attr.GetNestingMode()))
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

		// TODO: should work for tuples + lists
		newVal = tftypes.NewValue(config.Type(), newVals)
	}

	return newVal
}

func proposedNewObjectAttributes(ctx context.Context, s fwschema.Schema, attr fwschema.NestedAttribute, path *tftypes.AttributePath, prior, config tftypes.Value) tftypes.Value {
	if config.IsNull() {
		return config
	}

	// TODO: validate before doing this? To avoid panic
	return tftypes.NewValue(
		attr.GetNestedObject().Type().TerraformType(ctx),
		proposedNewAttributes(ctx, s, attr.GetNestedObject().GetAttributes(), path, prior, config),
	)
}

func optionalValueNotComputable(ctx context.Context, s fwschema.Schema, absPath *tftypes.AttributePath, val tftypes.Value) bool {
	// TODO: handle error
	attr, _ := s.AttributeAtTerraformPath(ctx, absPath)

	if !attr.IsOptional() { //nolint
		return false
	}

	_, nested := attr.(fwschema.NestedAttribute)
	if !nested {
		return false
	}

	foundNonComputedAttr := false
	tftypes.Walk(val, func(path *tftypes.AttributePath, v tftypes.Value) (bool, error) { //nolint
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
			return false, nil //nolint
		}

		if !attrSchema.IsComputed() {
			foundNonComputedAttr = true
			return false, nil
		}

		return true, nil
	})

	return foundNonComputedAttr
}
