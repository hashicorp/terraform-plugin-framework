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
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// GenerateResourceConfigRequest is the framework server request for the
// GenerateResourceConfig RPC.
type GenerateResourceConfigRequest struct {
	// State is the resource's state value.
	State *tfsdk.State
	// Resource is the resource implementation.
	Resource resource.Resource

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

	resp.GeneratedConfig = &tfsdk.Config{
		Raw:    req.State.Raw,
		Schema: req.State.Schema,
	}

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
	if err != nil {
		diags.AddError(
			"Generate Resource Config Error",
			"An unexpected error was encountered setting the top-level timeouts property to null: "+err.Error()+"\n\n"+
				"This is always an issue with the provider and should be reported to the provider developers.",
		)
	}

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
	if err != nil {
		diags.AddError(
			"Generate Resource Config Error",
			"An unexpected error was encountered setting computed-only properties to null: "+err.Error()+"\n\n"+
				"This is always an issue with the provider and should be reported to the provider developers.",
		)
	}

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
	config, diags = resolveConflictsWithGroups(ctx, config, req.ResourceSchema, req.Resource, diags)

	// 5) Construct a mapping of ExactlyOneOf properties to iterate over alphabetically.
	// If a group of ExactlyOneOf properties has more than one value set, the property
	// names are sorted and the first non-null value is retained while the others are
	// set to null. If all are null, we attempt to set one in the group by checking
	// for a default value
	config, diags = resolveExactlyOneOfGroups(ctx, config, req.ResourceSchema, req.Resource, diags)

	// 6) Construct a mapping of AlsoRequires (RequiredWith in SDKv2) properties to
	// iterate over alphabetically. If a group of AlsoRequires properties are not
	// all set, we set all properties in the group to null
	config, diags = resolveAlsoRequiresGroups(ctx, config, req.ResourceSchema, req.Resource, diags)

	resp.GeneratedConfig.Raw = config
	resp.Diagnostics = diags
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
			"An unexpected error was encountered replacing empty optional values with null: "+err.Error()+"\n\n"+
				"This is always an issue with the provider and should be reported to the provider developers.",
		)
		return config, diags
	}

	return newConfig, diags
}

// resolveConflictsWithGroups finds all ConflictsWith validator groups in the schema.
// For each group where more than one attribute has a non-null value, the attribute
// names are sorted alphabetically and the first non-null value is retained while
// the others are set to null.
func resolveConflictsWithGroups(ctx context.Context, config tftypes.Value, schema fwschema.Schema, res resource.Resource, diags diag.Diagnostics) (tftypes.Value, diag.Diagnostics) {
	groups := buildValidatorGroups(ctx, config, schema, res, getConflictsWithPaths)

	for _, key := range sortedGroupKeys(groups) {
		members := groups[key]
		nonNullMembers := nonNullGroupMembers(ctx, config, members)

		// nothing to do if there aren't multiple non-null values that conflict with each other
		if len(nonNullMembers) <= 1 {
			continue
		}

		// null out all but the first non-null value in the group
		for _, memberPath := range nonNullMembers[1:] {
			config = nullValueAtPath(config, terraformPathFromPath(ctx, memberPath))
		}
	}

	return config, diags
}

// resolveExactlyOneOfGroups finds all ExactlyOneOf validator groups in the schema.
// For each group with more than one non-null value, the first non-null value is
// retained and the others are set to null. If all values are null, the first
// attribute in alphabetical order with a default value is set.
func resolveExactlyOneOfGroups(ctx context.Context, config tftypes.Value, schema fwschema.Schema, res resource.Resource, diags diag.Diagnostics) (tftypes.Value, diag.Diagnostics) {
	groups := buildValidatorGroups(ctx, config, schema, res, getExactlyOneOfPaths)

	for _, key := range sortedGroupKeys(groups) {
		members := groups[key]
		nonNullMembers := nonNullGroupMembers(ctx, config, members)

		switch len(nonNullMembers) {
		case 0:
			// if there are no non-null values, attempt to set the first attribute in alphabetical order with a default value
			// if there's no default value, try the next
			for _, memberPath := range members {
				var defaultApplied bool

				config, defaultApplied, diags = setDefaultValueAtPath(ctx, config, schema, memberPath, diags)

				if defaultApplied {
					break
				}
			}
		case 1:
			// there is exactly one non-null value, do nothing, keep it in place
			continue
		default:
			// if there are multiple non-null values, null out all but the first non-null value in the group
			for _, memberPath := range nonNullMembers[1:] {
				config = nullValueAtPath(config, terraformPathFromPath(ctx, memberPath))
			}
		}
	}

	return config, diags
}

// resolveAlsoRequiresGroups finds all AlsoRequires validator groups in the schema.
// If some, but not all, values in a group are set then the set values are nulled.
// The process is repeated until no group changes so transitive requirements are
// also cleared.
func resolveAlsoRequiresGroups(ctx context.Context, config tftypes.Value, schema fwschema.Schema, res resource.Resource, diags diag.Diagnostics) (tftypes.Value, diag.Diagnostics) {
	groups := buildValidatorGroups(ctx, config, schema, res, getAlsoRequiresPaths)
	groupKeys := sortedGroupKeys(groups)

	for {
		before := config

		for _, key := range groupKeys {
			members := groups[key]
			nonNullMembers := nonNullGroupMembers(ctx, config, members)

			if len(nonNullMembers) == 0 || len(nonNullMembers) == len(members) {
				continue
			}

			for _, memberPath := range nonNullMembers {
				config = nullValueAtPath(config, terraformPathFromPath(ctx, memberPath))
			}
		}

		if config.Equal(before) {
			break // exit loop if there were no changes in the pass
		}
	}

	return config, diags
}

// sortedGroupKeys returns the keys of the given groups mapping sorted alphabetically.
func sortedGroupKeys(groups map[string]path.Paths) []string {
	keys := make([]string, 0, len(groups))

	for key := range groups {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

// nonNullGroupMembers returns the subset of the given group member paths that have non-null values in the config.
func nonNullGroupMembers(ctx context.Context, config tftypes.Value, members path.Paths) path.Paths {
	result := make(path.Paths, 0, len(members))

	for _, memberPath := range members {
		value, ok := valueAtPath(config, terraformPathFromPath(ctx, memberPath))

		if !ok || value.IsNull() {
			continue
		}

		result = append(result, memberPath)
	}

	return result
}

// valueAtPath returns the value at the given path in the config. If the path does not exist, a null value and false are returned.
func valueAtPath(value tftypes.Value, attrPath *tftypes.AttributePath) (tftypes.Value, bool) {
	rawValue, _, err := tftypes.WalkAttributePath(value, attrPath)
	if err != nil {
		return tftypes.Value{}, false
	}

	tfValue, ok := rawValue.(tftypes.Value)

	return tfValue, ok
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
