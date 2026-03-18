// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	// 6) Construct a mapping of AlsoRequires (RequiredWith in SDKv2) properties to
	// iterate over alphabetically. If a group of AlsoRequires properties are not
	// all set, we set all properties in the group to null

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
	// Phase 1: Build conflict groups from schema attributes.
	// Each ConflictsWith validator defines a group: the attribute itself + the paths it conflicts with.
	// Groups are deduplicated by sorting member names and using the joined string as a key.
	groups := map[string][]string{}

	for name, attr := range schema.GetAttributes() {
		conflictExprs := getConflictsWithExpressions(attr)
		if len(conflictExprs) == 0 {
			continue
		}

		members := []string{name}
		for _, expr := range conflictExprs {
			members = append(members, expr.String())
		}
		sort.Strings(members)

		key := strings.Join(members, ",")
		groups[key] = members
	}

	// Phase 2: For each group, find which attributes have non-null values.
	// If more than one is set, keep the first (alphabetically) and null the rest.
	for _, members := range groups {
		var nonNullMembers []string
		for _, memberName := range members {
			attrPath := tftypes.NewAttributePath().WithAttributeName(memberName)
			rawVal, _, err := tftypes.WalkAttributePath(config, attrPath)
			if err != nil {
				continue
			}
			if tfVal, ok := rawVal.(tftypes.Value); ok && !tfVal.IsNull() {
				nonNullMembers = append(nonNullMembers, memberName)
			}
		}

		if len(nonNullMembers) <= 1 {
			continue
		}

		// Members are already sorted; null all but the first non-null
		for _, memberName := range nonNullMembers[1:] {
			attrPath := tftypes.NewAttributePath().WithAttributeName(memberName)
			config = nullValueAtPath(config, attrPath)
		}
	}

	return config, diags
}

// getConflictsWithExpressions extracts ConflictsWith path expressions from an attribute's
// validators. It checks all typed validator interfaces (String, Bool, Int64, etc.) and
// returns the paths from any validator that implements validator.ConflictsWithValidator.
func getConflictsWithExpressions(attr fwschema.Attribute) path.Expressions {
	var result path.Expressions

	checkValidator := func(v interface{}) {
		if cv, ok := v.(validator.ConflictsWithValidator); ok {
			result = append(result, cv.Paths()...)
		}
	}

	if a, ok := attr.(fwxschema.AttributeWithBoolValidators); ok {
		for _, v := range a.BoolValidators() {
			checkValidator(v)
		}
	}
	if a, ok := attr.(fwxschema.AttributeWithFloat32Validators); ok {
		for _, v := range a.Float32Validators() {
			checkValidator(v)
		}
	}
	if a, ok := attr.(fwxschema.AttributeWithFloat64Validators); ok {
		for _, v := range a.Float64Validators() {
			checkValidator(v)
		}
	}
	if a, ok := attr.(fwxschema.AttributeWithInt32Validators); ok {
		for _, v := range a.Int32Validators() {
			checkValidator(v)
		}
	}
	if a, ok := attr.(fwxschema.AttributeWithInt64Validators); ok {
		for _, v := range a.Int64Validators() {
			checkValidator(v)
		}
	}
	if a, ok := attr.(fwxschema.AttributeWithListValidators); ok {
		for _, v := range a.ListValidators() {
			checkValidator(v)
		}
	}
	if a, ok := attr.(fwxschema.AttributeWithMapValidators); ok {
		for _, v := range a.MapValidators() {
			checkValidator(v)
		}
	}
	if a, ok := attr.(fwxschema.AttributeWithNumberValidators); ok {
		for _, v := range a.NumberValidators() {
			checkValidator(v)
		}
	}
	if a, ok := attr.(fwxschema.AttributeWithObjectValidators); ok {
		for _, v := range a.ObjectValidators() {
			checkValidator(v)
		}
	}
	if a, ok := attr.(fwxschema.AttributeWithSetValidators); ok {
		for _, v := range a.SetValidators() {
			checkValidator(v)
		}
	}
	if a, ok := attr.(fwxschema.AttributeWithStringValidators); ok {
		for _, v := range a.StringValidators() {
			checkValidator(v)
		}
	}
	if a, ok := attr.(fwxschema.AttributeWithDynamicValidators); ok {
		for _, v := range a.DynamicValidators() {
			checkValidator(v)
		}
	}

	return result
}

// nulls the value at the given path in the value and returns the modified value. If the path does not exist, the original value is returned unmodified.
func nullValueAtPath(value tftypes.Value, path *tftypes.AttributePath) tftypes.Value {
	newValue, err := tftypes.Transform(value, func(p *tftypes.AttributePath, v tftypes.Value) (tftypes.Value, error) {
		if p.Equal(path) {
			return tftypes.NewValue(v.Type(), nil), nil
		}
		return v, nil
	})

	if err != nil {
		return value
	}

	return newValue
}
