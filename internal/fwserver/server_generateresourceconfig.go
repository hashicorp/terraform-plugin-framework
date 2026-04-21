// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"errors"
	"math/big"
	"sort"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fromtftypes"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschemadata"
	"github.com/hashicorp/terraform-plugin-framework/internal/logging"
	"github.com/hashicorp/terraform-plugin-framework/internal/totftypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	schemavalidator "github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

// GenerateResourceConfigRequest is the framework server request for the
// GenerateResourceConfig RPC.
type GenerateResourceConfigRequest struct {
	// Resource is the resource implementation.
	Resource resource.Resource

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

	var diags diag.Diagnostics
	resp.GeneratedConfig = stateToConfig(*req.State)

	var resourceConfigValidators []resource.ConfigValidator

	if resourceWithConfigValidators, ok := req.Resource.(resource.ResourceWithConfigValidators); ok {
		resourceConfigValidators = resourceWithConfigValidators.ConfigValidators(ctx)
	}

	config := req.State.Raw
	markedForNullification := path.Paths{}
	resourceValidatorGroups := resolveResourceValidatorGroups(ctx, resp.GeneratedConfig, resourceConfigValidators, &diags)

	// First pass: drop values that should never appear in generated config and
	// record any validator-driven paths that must be nulled as a group.
	config, err := tftypes.Transform(config, func(tfPath *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
		if len(tfPath.Steps()) == 0 {
			return value, nil
		}

		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, tfPath, req.ResourceSchema)
		diags.Append(fwPathDiags...)

		if fwPathDiags.HasError() {
			return value, nil
		}

		var attribute fwschema.Attribute
		attribute, err := req.ResourceSchema.AttributeAtTerraformPath(ctx, tfPath)
		if err != nil {
			if !errors.Is(err, fwschema.ErrPathIsBlock) && !errors.Is(err, fwschema.ErrPathInsideDynamicAttribute) && !errors.Is(err, fwschema.ErrPathInsideAtomicAttribute) {
				logging.FrameworkError(ctx, "couldn't find attribute in resource schema")

				diags.AddAttributeError(
					fwPath,
					"Generate Resource Config Error",
					"An unexpected error was encountered trying to generate the resource config for import. "+
						"This likely indicates a bug in the Terraform provider framework or Terraform Core. Please report the following to the provider developer:\n\n"+err.Error(),
				)

				return value, err
			}
		}

		block, blockErr := fwschema.SchemaBlockAtTerraformPath(ctx, req.ResourceSchema, tfPath)
		if blockErr != nil {
			if !errors.Is(blockErr, fwschema.ErrPathIsAttribute) && !errors.Is(blockErr, fwschema.ErrPathInsideDynamicAttribute) && !errors.Is(blockErr, fwschema.ErrPathInsideAtomicAttribute) {
				logging.FrameworkError(ctx, "couldn't find block in resource schema")

				diags.AddAttributeError(
					fwPath,
					"Generate Resource Config Error",
					"An unexpected error was encountered trying to generate the resource config for import. "+
						"This likely indicates a bug in the Terraform provider framework or Terraform Core. Please report the following to the provider developer:\n\n"+blockErr.Error(),
				)

				return value, blockErr
			}
		}

		if value.IsNull() {
			return value, nil
		}

		null := tftypes.NewValue(value.Type(), nil)

		if fwPath.Equal(path.Root("timeouts")) {
			return null, nil
		}

		if attribute != nil {
			if attribute.GetDeprecationMessage() != "" {
				return null, nil
			}

			if attribute.IsComputed() && !attribute.IsOptional() {
				return null, nil
			}

			if fwPath.Equal(path.Root("id")) && attribute.IsComputed() && attribute.IsOptional() {
				return null, nil
			}

			if value.Type().Equal(tftypes.String) && attribute.IsOptional() {
				var stringValue string

				if err := value.As(&stringValue); err != nil {
					diags.AddAttributeError(
						fwPath,
						"Generate Resource Config Error",
						"An unexpected error was encountered trying to generate the resource config for import. "+
							"This likely indicates a bug in the Terraform provider framework or Terraform Core. Please report the following to the provider developer:\n\n"+err.Error(),
					)

					return value, err
				}

				if len(stringValue) == 0 {
					return null, nil
				}
			}

			if value.Type().Equal(tftypes.Number) && attribute.IsOptional() {
				var numberValue big.Float

				if err := value.As(&numberValue); err != nil {
					diags.AddAttributeError(
						fwPath,
						"Generate Resource Config Error",
						"An unexpected error was encountered trying to generate the resource config for import. "+
							"This likely indicates a bug in the Terraform provider framework or Terraform Core. Please report the following to the provider developer:\n\n"+err.Error(),
					)

					return value, err
				}

				if numberValue.Sign() == 0 {
					return null, nil
				}
			}
		}

		if block != nil && block.GetDeprecationMessage() != "" {
			return null, nil
		}

		if attribute != nil {
			attributeValidatorGroups := resolveAttributeValidatorGroups(ctx, resp.GeneratedConfig, fwPath, attribute, &diags)

			if applyValidatorGroups(ctx, req.ResourceSchema, attributeValidatorGroups, config, fwPath, false, &markedForNullification, &diags) {
				return null, nil
			}
		}

		if block != nil {
			blockValidatorGroups := resolveBlockValidatorGroups(ctx, resp.GeneratedConfig, fwPath, block, &diags)

			if applyValidatorGroups(ctx, req.ResourceSchema, blockValidatorGroups, config, fwPath, false, &markedForNullification, &diags) {
				return null, nil
			}
		}

		if applyValidatorGroups(ctx, req.ResourceSchema, resourceValidatorGroups, config, fwPath, true, &markedForNullification, &diags) {
			return null, nil
		}

		return value, nil
	})

	if err != nil {
		logging.FrameworkError(ctx,
			"Error transforming state value during resource config generation",
			map[string]any{
				logging.KeyError: err.Error(),
			},
		)
	}

	// Second pass: apply the accumulated group nullifications after every path has
	// had a chance to contribute to its validator group.
	config, err = tftypes.Transform(config, func(tfPath *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
		if len(tfPath.Steps()) == 0 || value.IsNull() {
			return value, nil
		}

		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, tfPath, req.ResourceSchema)
		diags.Append(fwPathDiags...)

		if fwPathDiags.HasError() {
			return value, nil
		}

		if markedForNullification.Contains(fwPath) {
			return tftypes.NewValue(value.Type(), nil), nil
		}

		return value, nil
	})

	if err != nil {
		logging.FrameworkError(ctx,
			"Error nullifying generated resource config values",
			map[string]any{
				logging.KeyError: err.Error(),
			},
		)
	}

	resp.GeneratedConfig.Raw = config
	resp.Diagnostics = diags
}

// readTerraformValue returns the raw Terraform value at the given framework path.
func readTerraformValue(ctx context.Context, schema fwschema.Schema, currentConfig tftypes.Value, targetPath path.Path, diags *diag.Diagnostics) (tftypes.Value, bool) {
	tfPath, pathDiags := totftypes.AttributePath(ctx, targetPath)
	diags.Append(pathDiags...)

	if pathDiags.HasError() {
		return tftypes.Value{}, false
	}

	value, err := fwschemadata.Data{
		Description:    fwschemadata.DataDescriptionConfiguration,
		Schema:         schema,
		TerraformValue: currentConfig,
	}.TerraformValueAtTerraformPath(ctx, tfPath)
	if err != nil {
		return tftypes.Value{}, false
	}

	return value, true
}

// sortedPaths returns a copy ordered by string form so group decisions are deterministic.
func sortedPaths(paths path.Paths) path.Paths {
	result := make(path.Paths, 0, len(paths))
	result.Append(paths...)

	sort.Slice(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})

	return result
}

// resolveExpressions expands validator expressions against the current config.
func resolveExpressions(ctx context.Context, config *tfsdk.Config, baseExpression path.Expression, expressions path.Expressions, diags *diag.Diagnostics) path.Paths {
	var matches path.Paths

	for _, expression := range expressions {
		resolvedExpression := baseExpression.Merge(expression)
		resolvedMatches, matchDiags := config.PathMatches(ctx, resolvedExpression)
		diags.Append(matchDiags...)
		matches.Append(resolvedMatches...)
	}

	return matches
}

type validatorGroups struct {
	ConflictsWith []path.Paths
	ExactlyOneOf  []path.Paths
	AlsoRequires  []path.Paths
}

// resolveResourceValidatorGroups materializes resource-level validator expressions into concrete path groups.
func resolveResourceValidatorGroups(ctx context.Context, config *tfsdk.Config, validators []resource.ConfigValidator, diags *diag.Diagnostics) validatorGroups {
	var groups validatorGroups

	for _, validator := range validators {
		appendValidatorGroups(ctx, config, path.Expression{}, path.Empty(), false, validator, &groups, diags)
	}

	return groups
}

// resolveAttributeValidatorGroups collects attribute validator groups for the current path.
func resolveAttributeValidatorGroups(ctx context.Context, config *tfsdk.Config, currentPath path.Path, attribute fwschema.Attribute, diags *diag.Diagnostics) validatorGroups {
	var groups validatorGroups

	switch attributeWithValidators := attribute.(type) {
	case fwxschema.AttributeWithBoolValidators:
		for _, validator := range attributeWithValidators.BoolValidators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.AttributeWithFloat32Validators:
		for _, validator := range attributeWithValidators.Float32Validators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.AttributeWithFloat64Validators:
		for _, validator := range attributeWithValidators.Float64Validators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.AttributeWithInt32Validators:
		for _, validator := range attributeWithValidators.Int32Validators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.AttributeWithInt64Validators:
		for _, validator := range attributeWithValidators.Int64Validators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.AttributeWithListValidators:
		for _, validator := range attributeWithValidators.ListValidators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.AttributeWithMapValidators:
		for _, validator := range attributeWithValidators.MapValidators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.AttributeWithNumberValidators:
		for _, validator := range attributeWithValidators.NumberValidators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.AttributeWithObjectValidators:
		for _, validator := range attributeWithValidators.ObjectValidators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.AttributeWithSetValidators:
		for _, validator := range attributeWithValidators.SetValidators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.AttributeWithStringValidators:
		for _, validator := range attributeWithValidators.StringValidators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.AttributeWithDynamicValidators:
		for _, validator := range attributeWithValidators.DynamicValidators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	}

	return groups
}

// resolveBlockValidatorGroups collects block validator groups for the current path.
func resolveBlockValidatorGroups(ctx context.Context, config *tfsdk.Config, currentPath path.Path, block fwschema.Block, diags *diag.Diagnostics) validatorGroups {
	var groups validatorGroups

	switch blockWithValidators := block.(type) {
	case fwxschema.BlockWithListValidators:
		for _, validator := range blockWithValidators.ListValidators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.BlockWithObjectValidators:
		for _, validator := range blockWithValidators.ObjectValidators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	case fwxschema.BlockWithSetValidators:
		for _, validator := range blockWithValidators.SetValidators() {
			appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, validator, &groups, diags)
		}
	}

	return groups
}

// appendValidatorGroups adds any group-oriented validator paths exposed by a validator.
func appendValidatorGroups(ctx context.Context, config *tfsdk.Config, baseExpression path.Expression, currentPath path.Path, includeCurrent bool, validator any, groups *validatorGroups, diags *diag.Diagnostics) {
	if validatorWithConflictsWith, ok := validator.(schemavalidator.ConflictsWithValidator); ok {
		members := resolveValidatorGroupPaths(ctx, config, baseExpression, currentPath, includeCurrent, validatorWithConflictsWith.ConflictsWithPaths(), diags)

		if len(members) > 0 {
			groups.ConflictsWith = append(groups.ConflictsWith, members)
		}
	}

	if validatorWithExactlyOneOf, ok := validator.(schemavalidator.ExactlyOneOfValidator); ok {
		members := resolveValidatorGroupPaths(ctx, config, baseExpression, currentPath, includeCurrent, validatorWithExactlyOneOf.ExactlyOneOfPaths(), diags)

		if len(members) > 0 {
			groups.ExactlyOneOf = append(groups.ExactlyOneOf, members)
		}
	}

	if validatorWithAlsoRequires, ok := validator.(schemavalidator.AlsoRequiresValidator); ok {
		members := resolveValidatorGroupPaths(ctx, config, baseExpression, currentPath, includeCurrent, validatorWithAlsoRequires.AlsoRequiresPaths(), diags)

		if len(members) > 0 {
			groups.AlsoRequires = append(groups.AlsoRequires, members)
		}
	}
}

// applyValidatorGroups applies all validator group rules for the current path.
func applyValidatorGroups(ctx context.Context, schema fwschema.Schema, groups validatorGroups, configVal tftypes.Value, curPath path.Path, requireMembership bool, markedForNullification *path.Paths, diags *diag.Diagnostics) bool {
	if applyValidatorGroupRule(ctx, schema, groups.ConflictsWith, configVal, curPath, requireMembership, markedForNullification, diags, processKeepFirstSetGroup) {
		return true
	}

	if applyValidatorGroupRule(ctx, schema, groups.ExactlyOneOf, configVal, curPath, requireMembership, markedForNullification, diags, processKeepFirstSetGroup) {
		return true
	}

	if applyValidatorGroupRule(ctx, schema, groups.AlsoRequires, configVal, curPath, requireMembership, markedForNullification, diags, processAlsoRequires) {
		return true
	}

	return false
}

// applyValidatorGroupRule runs one validator rule family across all resolved groups.
// Resource-level groups require explicit membership because they are resolved once up
// front, while attribute/block groups are already scoped to the current path.
func applyValidatorGroupRule(ctx context.Context, schema fwschema.Schema, groups []path.Paths, configVal tftypes.Value, curPath path.Path, requireMembership bool, markedForNullification *path.Paths, diags *diag.Diagnostics, processor func(context.Context, fwschema.Schema, path.Paths, tftypes.Value, path.Path, *diag.Diagnostics) path.Paths) bool {
	for _, groupPaths := range groups {
		if requireMembership && !groupPaths.Contains(curPath) {
			continue
		}

		markedForNullification.Append(processor(ctx, schema, groupPaths, configVal, curPath, diags)...)

		if markedForNullification.Contains(curPath) {
			return true
		}
	}

	return false
}

// resolveValidatorGroupPaths returns the concrete members of a validator group for the current path.
func resolveValidatorGroupPaths(ctx context.Context, config *tfsdk.Config, baseExpression path.Expression, currentPath path.Path, includeCurrent bool, expressions path.Expressions, diags *diag.Diagnostics) path.Paths {
	var members path.Paths

	if includeCurrent {
		members.Append(currentPath)
	}

	members.Append(resolveExpressions(ctx, config, baseExpression, expressions, diags)...)
	members = sortedPaths(members)

	if len(members) < 2 {
		return nil
	}

	return members
}

// processKeepFirstSetGroup preserves the lexicographically first configured path in
// a mutually-exclusive group and marks the rest for nullification.
func processKeepFirstSetGroup(ctx context.Context, schema fwschema.Schema, paths path.Paths, configVal tftypes.Value, curPath path.Path, diags *diag.Diagnostics) path.Paths {
	var markedForNullification path.Paths
	var nonNullKeys path.Paths

	if len(paths) == 0 {
		return markedForNullification
	}

	nonNullKeys.Append(curPath)

	for _, key := range paths {
		if key.Equal(curPath) {
			continue
		}

		val, ok := readTerraformValue(ctx, schema, configVal, key, diags)
		if !ok || val.IsNull() {
			continue
		}

		nonNullKeys.Append(key)
	}

	nonNullKeys = sortedPaths(nonNullKeys)

	for keyIndex, key := range nonNullKeys {
		if keyIndex == 0 {
			continue
		}

		markedForNullification.Append(key)
	}

	return markedForNullification
}

// processAlsoRequires nulls all configured members when the required peer set is incomplete.
func processAlsoRequires(ctx context.Context, schema fwschema.Schema, alsoRequires path.Paths, configVal tftypes.Value, curPath path.Path, diags *diag.Diagnostics) path.Paths {
	var markedForNullification path.Paths
	var nonNullKeys path.Paths

	if len(alsoRequires) == 0 {
		return markedForNullification
	}

	nonNullKeys.Append(curPath)

	for _, key := range alsoRequires {
		if key.Equal(curPath) {
			continue
		}

		val, ok := readTerraformValue(ctx, schema, configVal, key, diags)
		if !ok || val.IsNull() {
			continue
		}

		nonNullKeys.Append(key)
	}

	if len(nonNullKeys) == len(alsoRequires) {
		return markedForNullification
	}

	return sortedPaths(nonNullKeys)
}

// stateToConfig returns a *tfsdk.Config with a copied value from a tfsdk.State.
func stateToConfig(state tfsdk.State) *tfsdk.Config {
	return &tfsdk.Config{
		Raw:    state.Raw.Copy(),
		Schema: state.Schema,
	}
}
