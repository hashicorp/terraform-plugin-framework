// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"maps"
	"sort"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// GenerateResourceConfigRequest is the framework server request for the
// GenerateResourceConfig RPC.
type GenerateResourceConfigRequest struct {
	// State is the resource's state value.
	State *tfsdk.State

	// ResourceSchema is the resource's schema.
	ResourceSchema fwschema.Schema
}

// GenerateResourceConfigResponse is the framework server response for the
// GenerateResourceConfig RPC.
type GenerateResourceConfigResponse struct {
	// GeneratedConfig contains the resource's generated config value.
	GeneratedConfig *tfsdk.Config

	Diagnostics diag.Diagnostics
}

// GenerateResourceConfig implements the framework server GenerateResourceConfig RPC.
func (s *Server) GenerateResourceConfig(ctx context.Context, req *GenerateResourceConfigRequest, resp *GenerateResourceConfigResponse) {
	if req == nil {
		return
	}

	if req.State == nil {
		resp.Diagnostics.AddError(
			"Unexpected Generate Config Request",
			"An unexpected error was encountered when generating resource configuration. The current state was missing.\n\n"+
				"This is always a problem with Terraform or terraform-plugin-framework. Please report this to the provider developer.",
		)
		return
	}

	// copy as we'll modify in place
	config := req.State.Raw.Copy()
	var diags diag.Diagnostics
	// TODO: make sure all error cases are reflected in diags and not just ignored, maybe some need to be caught?

	resp.GeneratedConfig = stateToConfig(*req.State)

	// smarter algorithm steps:
	// 1) Set top level properties named id and timeouts to null
	idPath := tftypes.NewAttributePath().WithAttributeName("id")

	// id
	idAttribute, err := req.State.Schema.AttributeAtTerraformPath(ctx, idPath)
	if err == nil && idAttribute.IsComputed() && idAttribute.IsOptional() {
		config = nullValueAtPath(config, idPath)
	} // else: no id attribute, ignoring

	// timeouts
	timeoutPath := tftypes.NewAttributePath().WithAttributeName("timeouts")
	config, err = tftypes.Transform(config, func(path *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
		if path.Equal(timeoutPath) {
			return tftypes.NewValue(value.Type(), nil), nil
		}
		return value, nil
	})

	// 2) Set Computed only properties to null
	config, err = tftypes.Transform(config, func(path *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
		if value.IsNull() || len(path.Steps()) == 0 {
			return value, nil
		}
		attr, err := req.ResourceSchema.AttributeAtTerraformPath(ctx, path)
		if err == nil && attr.IsComputed() && !attr.IsOptional() {
			return tftypes.NewValue(value.Type(), nil), nil
		}
		return value, nil
	})

	// 3) Set empty Optional properties to null.
	// Unlike the SDKv2 implementation, we do not set schema-defined default values
	// into the generated config. In the Framework, defaults can only be set on
	// Optional+Computed attributes, and if we placed a default value in the config
	// it would be treated as a practitioner-set Optional value rather than a
	// provider-set Computed value, which could have implications during planning.
	// Instead, we null out empty values and let the Framework apply defaults
	// during the plan phase as usual.
	// (Note that for boolean properties the empty value false will be kept instead of setting it to null)
	config, diags = nullEmptyOptionalValues(ctx, config, req.ResourceSchema, diags)

	// 4) Construct a mapping of ConflictsWith properties to iterate over alphabetically.
	// If a group of ConflictsWith properties has more than one value set, the property
	// names are sorted and the first non-null value is retained while the others are
	// set to null
	config, diags = resolveConflictsWithGroups(ctx, config, req.ResourceSchema, diags)
	// TODO: also handle resource level validators!

	// 5) Construct a mapping of ExactlyOneOf properties to iterate over alphabetically.
	// If a group of ExactlyOneOf properties has more than one value set, the property
	// names are sorted and the first non-null value is retained while the others are
	// set to null. If all are null, we attempt to set one in the group by checking
	// for a default value
	config, diags = resolveExactlyOneOfGroups(ctx, config, req.ResourceSchema, diags)

	// 6) Construct a mapping of AlsoRequires (RequiredWith in SDKv2) properties to
	// iterate over alphabetically. If a group of AlsoRequires properties are not
	// all set, we set all properties in the group to null
	config, diags = resolveAlsoRequiresGroups(ctx, config, req.ResourceSchema, diags)

	resp.GeneratedConfig.Raw = config
	resp.Diagnostics = diags
}

// stateToConfig returns a *tfsdk.Config with a copied value from a tfsdk.State.
func stateToConfig(state tfsdk.State) *tfsdk.Config {
	return &tfsdk.Config{
		Raw:    state.Raw.Copy(),
		Schema: state.Schema,
	}
}

// nullEmptyOptionalValues transforms a config value by replacing empty optional
// attribute values with null. Boolean false values are kept as-is since false
// is the empty value for booleans.
func nullEmptyOptionalValues(ctx context.Context, config tftypes.Value, schema fwschema.Schema, diags diag.Diagnostics) (tftypes.Value, diag.Diagnostics) {
	newConfig, err := tftypes.Transform(config, func(attrPath *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
		if value.IsNull() || len(attrPath.Steps()) == 0 {
			return value, nil
		}

		attr, err := schema.AttributeAtTerraformPath(ctx, attrPath)
		if err != nil || !attr.IsOptional() {
			return value, nil
		}

		tfType := attr.GetType().TerraformType(ctx)
		if tfType == nil {
			return value, nil
		}

		null := tftypes.NewValue(tfType, nil)

		switch {
		case tfType.Equal(tftypes.Bool):
			var boolVal bool
			if err := value.As(&boolVal); err != nil {
				return value, err
			}
			// Keep false (empty bool) as-is rather than nulling it out
			if !boolVal {
				return value, nil
			}
		case tfType.Equal(tftypes.String):
			var stringVal string
			if err := value.As(&stringVal); err != nil {
				return value, err
			}
			if len(stringVal) == 0 {
				return null, nil
			}
		case tfType.Equal(tftypes.Number):
			var numVal float64
			if err := value.As(&numVal); err != nil {
				return value, err
			}
			if numVal == 0 {
				return null, nil
			}
		}

		return value, nil
	})

	if err != nil {
		diags.AddError(
			"Generate Resource Config Error",
			"An unexpected error was encountered replacing empty optional values: "+err.Error(),
		)
		return config, diags
	}

	return newConfig, diags
}

// resolveConflictsWithGroups finds all ConflictsWith validator groups in the schema.
// For each group where more than one attribute has a non-null value, the attribute
// names are sorted alphabetically and the first non-null value is retained while
// the others are set to null.
func resolveConflictsWithGroups(ctx context.Context, config tftypes.Value, schema fwschema.Schema, diags diag.Diagnostics) (tftypes.Value, diag.Diagnostics) {
	groups := buildAttributeValidatorGroups(schema, getConflictsWithExpressions)

	for _, key := range sortedGroupKeys(groups) {
		members := groups[key]
		nonNullMembers := nonNullGroupMembers(config, members)

		if len(nonNullMembers) <= 1 {
			continue
		}

		for _, memberName := range nonNullMembers[1:] {
			config = nullValueAtPath(config, rootAttributePath(memberName))
		}
	}

	return config, diags
}

// resolveExactlyOneOfGroups finds all ExactlyOneOf validator groups in the schema.
// For each group with more than one non-null value, the first non-null value is
// retained and the others are set to null. If all values are null, the first
// attribute in alphabetical order with a default value is set.
func resolveExactlyOneOfGroups(ctx context.Context, config tftypes.Value, schema fwschema.Schema, diags diag.Diagnostics) (tftypes.Value, diag.Diagnostics) {
	groups := buildAttributeValidatorGroups(schema, getExactlyOneOfExpressions)

	for _, key := range sortedGroupKeys(groups) {
		members := groups[key]
		nonNullMembers := nonNullGroupMembers(config, members)

		switch len(nonNullMembers) {
		case 0:
			for _, memberName := range members {
				var applied bool

				config, applied, diags = setDefaultValueAtAttribute(ctx, config, schema, memberName, diags)

				if applied {
					break
				}
			}
		case 1:
			continue
		default:
			for _, memberName := range nonNullMembers[1:] {
				config = nullValueAtPath(config, rootAttributePath(memberName))
			}
		}
	}

	return config, diags
}

// resolveAlsoRequiresGroups finds all AlsoRequires validator groups in the schema.
// If some, but not all, values in a group are set then the set values are nulled.
// The process is repeated until no group changes so transitive requirements are
// also cleared.
func resolveAlsoRequiresGroups(_ context.Context, config tftypes.Value, schema fwschema.Schema, diags diag.Diagnostics) (tftypes.Value, diag.Diagnostics) {
	groups := buildAttributeValidatorGroups(schema, getAlsoRequiresExpressions)
	groupKeys := sortedGroupKeys(groups)

	for {
		changed := false

		for _, key := range groupKeys {
			members := groups[key]
			nonNullMembers := nonNullGroupMembers(config, members)

			if len(nonNullMembers) == 0 || len(nonNullMembers) == len(members) {
				continue
			}

			for _, memberName := range nonNullMembers {
				config = nullValueAtPath(config, rootAttributePath(memberName))
			}

			changed = true
		}

		if !changed {
			break
		}
	}

	return config, diags
}

func sortedGroupKeys(groups map[string][]string) []string {
	keys := make([]string, 0, len(groups))

	for key := range groups {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func nonNullGroupMembers(config tftypes.Value, members []string) []string {
	result := make([]string, 0, len(members))

	for _, memberName := range members {
		value, ok := valueAtPath(config, rootAttributePath(memberName))

		if !ok || value.IsNull() {
			continue
		}

		result = append(result, memberName)
	}

	return result
}

func valueAtPath(value tftypes.Value, attrPath *tftypes.AttributePath) (tftypes.Value, bool) {
	rawValue, _, err := tftypes.WalkAttributePath(value, attrPath)
	if err != nil {
		return tftypes.Value{}, false
	}

	tfValue, ok := rawValue.(tftypes.Value)

	return tfValue, ok
}

func rootAttributePath(name string) *tftypes.AttributePath {
	return tftypes.NewAttributePath().WithAttributeName(name)
}

func setDefaultValueAtAttribute(ctx context.Context, config tftypes.Value, schema fwschema.Schema, attributeName string, diags diag.Diagnostics) (tftypes.Value, bool, diag.Diagnostics) {
	attr, ok := schema.GetAttributes()[attributeName]
	if !ok {
		return config, false, diags
	}

	defaultValue, hasDefault, diags := attributeDefaultValue(ctx, attr, path.Root(attributeName), diags)
	if !hasDefault || defaultValue.IsNull() {
		return config, false, diags
	}

	config, err := replaceValueAtPath(config, rootAttributePath(attributeName), defaultValue)
	if err != nil {
		diags.AddError(
			"Generate Resource Config Error",
			"An unexpected error was encountered setting a default value for attribute "+attributeName+": "+err.Error(),
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

// nulls the value at the given path in the value and returns the modified value. If the path does not exist, the original value is returned unmodified.
func nullValueAtPath(value tftypes.Value, path *tftypes.AttributePath) tftypes.Value {
	currentValue, ok := valueAtPath(value, path)
	if !ok {
		return value
	}

	newValue, err := replaceValueAtPath(value, path, tftypes.NewValue(currentValue.Type(), nil))

	if err != nil {
		return value
	}

	return newValue
}

func replaceValueAtPath(value tftypes.Value, path *tftypes.AttributePath, replaceWith tftypes.Value) (tftypes.Value, error) {
	steps := path.Steps()

	// Top-level attribute replacement needs special handling. During transform the
	// root object is visited at the empty path, so replacing a single-step path
	// means rebuilding that root object with one field swapped out.
	if len(steps) == 1 {
		if attributeName, ok := steps[0].(tftypes.AttributeName); ok {
			var objectValue map[string]tftypes.Value
			if err := value.As(&objectValue); err != nil {
				return value, err
			}

			copiedObjectValue := maps.Clone(objectValue)
			copiedObjectValue[string(attributeName)] = replaceWith

			return tftypes.NewValue(value.Type(), copiedObjectValue), nil
		}
	}

	return tftypes.Transform(value, func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if p.Equal(path) {
			return replaceWith, nil
		}

		return v, nil
	})
}
