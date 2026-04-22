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
//
// MAINTAINER NOTE:
//
// The algorithm below is a port of the SDKv2 implementation in
// terraform-plugin-sdk/helper/schema/grpc_provider.go (GenerateResourceConfig)
// and terraform-plugin-sdk/helper/schema/resource_config_generation.go
// (processConflictsWith / processExactlyOneOf / processRequiredWith). The intent
// is functional parity for providers migrating between SDKv2 and the framework.
//
// Two things differ from SDKv2 by design:
//
//  1. Group sources. SDKv2 reads ConflictsWith/ExactlyOneOf/RequiredWith as
//     []string fields on the schema. The framework has no such fields; instead,
//     validators may opt in by implementing the marker interfaces in
//     schema/validator (ConflictsWithValidator, ExactlyOneOfValidator,
//     AlsoRequiresValidator). Resource-level resource.ConfigValidator instances
//     can also implement these interfaces and contribute groups; this is a
//     framework-only capability not present in SDKv2.
//
//  2. AtLeastOneOf is intentionally not handled. SDKv2 also omits it from
//     genconfig.
//
// Two passes are required:
//
//   - Pass 1 walks every value, drops values that should never appear in
//     generated config (deprecated/computed/timeouts/etc.), and accumulates
//     paths that must be nulled because of validator group decisions. A group
//     decision can only be made once every member of the group has been seen,
//     so the per-value early returns may add other paths to the
//     markedForNullification set.
//
//   - Pass 2 walks the result and nulls anything in markedForNullification.
//
// MAINTAINER NOTE:
//
// The "optional Number value of zero becomes null" rule (see
// shouldDropAlwaysAbsentValue) has no SDKv2 equivalent. SDKv2 only does the
// zero-becomes-null trick for strings, where it exists to compensate for the
// SDKv2 historical inability to distinguish "" from null. Numbers do not have
// the same ambiguity. The rule is preserved here for behavioural parity with
// the existing test suite, but is a candidate for removal in a future revision.
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
	if r, ok := req.Resource.(resource.ResourceWithConfigValidators); ok {
		resourceConfigValidators = r.ConfigValidators(ctx)
	}

	resourceValidatorGroups := resolveResourceValidatorGroups(ctx, resp.GeneratedConfig, resourceConfigValidators, &diags)

	config := req.State.Raw
	markedForNullification := path.Paths{}

	// Pass 1: per-value drops + group decision accumulation.
	config, err := tftypes.Transform(config, func(tfPath *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
		if len(tfPath.Steps()) == 0 {
			return value, nil
		}

		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, tfPath, req.ResourceSchema)
		diags.Append(fwPathDiags...)
		if fwPathDiags.HasError() {
			return value, nil
		}

		attribute, block, lookupErr := lookupSchemaNode(ctx, req.ResourceSchema, fwPath, tfPath, &diags)
		if lookupErr != nil {
			return value, lookupErr
		}

		if value.IsNull() {
			return value, nil
		}

		nullValue := tftypes.NewValue(value.Type(), nil)

		if drop, dropErr := shouldDropAlwaysAbsentValue(fwPath, attribute, block, value, &diags); drop {
			return nullValue, dropErr
		} else if dropErr != nil {
			return value, dropErr
		}

		if applyValidatorRules(ctx, req.ResourceSchema, resp.GeneratedConfig, fwPath, attribute, block, resourceValidatorGroups, config, &markedForNullification, &diags) {
			return nullValue, nil
		}

		return value, nil
	})
	if err != nil {
		logging.FrameworkError(ctx,
			"Error transforming state value during resource config generation",
			map[string]any{logging.KeyError: err.Error()},
		)
	}

	// Pass 2: apply accumulated group nullifications.
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
			map[string]any{logging.KeyError: err.Error()},
		)
	}

	resp.GeneratedConfig.Raw = config
	resp.Diagnostics = diags
}

// lookupSchemaNode resolves the schema attribute and/or block at the given path.
// Either return value may be nil; an error is reported only when the underlying
// schema lookup returned an unexpected error (not the routine "this path is a
// block, not an attribute" sentinels). Diagnostics are appended on unexpected
// errors and the same error is returned so the caller can abort the transform.
func lookupSchemaNode(ctx context.Context, schema fwschema.Schema, fwPath path.Path, tfPath *tftypes.AttributePath, diags *diag.Diagnostics) (fwschema.Attribute, fwschema.Block, error) {
	attribute, attrErr := schema.AttributeAtTerraformPath(ctx, tfPath)
	if attrErr != nil {
		if !errors.Is(attrErr, fwschema.ErrPathIsBlock) &&
			!errors.Is(attrErr, fwschema.ErrPathInsideDynamicAttribute) &&
			!errors.Is(attrErr, fwschema.ErrPathInsideAtomicAttribute) {
			logging.FrameworkError(ctx, "couldn't find attribute in resource schema")
			diags.AddAttributeError(fwPath, "Generate Resource Config Error", genConfigErrDetail(attrErr))
			return nil, nil, attrErr
		}
		attribute = nil
	}

	block, blockErr := fwschema.SchemaBlockAtTerraformPath(ctx, schema, tfPath)
	if blockErr != nil {
		if !errors.Is(blockErr, fwschema.ErrPathIsAttribute) &&
			!errors.Is(blockErr, fwschema.ErrPathInsideDynamicAttribute) &&
			!errors.Is(blockErr, fwschema.ErrPathInsideAtomicAttribute) {
			logging.FrameworkError(ctx, "couldn't find block in resource schema")
			diags.AddAttributeError(fwPath, "Generate Resource Config Error", genConfigErrDetail(blockErr))
			return attribute, nil, blockErr
		}
		block = nil
	}

	return attribute, block, nil
}

func genConfigErrDetail(err error) string {
	return "An unexpected error was encountered trying to generate the resource config for import. " +
		"This likely indicates a bug in the Terraform provider framework or Terraform Core. Please report the following to the provider developer:\n\n" +
		err.Error()
}

// shouldDropAlwaysAbsentValue returns (drop, err) for the rules that are
// independent of validator groups: timeouts, deprecation, computed-only,
// SDKv2-style id quirk, empty optional strings, and zero optional numbers.
//
// The empty-string rule mirrors SDKv2 (compensating for SDKv2's inability to
// distinguish "" from null). The zero-number rule has no SDKv2 equivalent; see
// the maintainer note on GenerateResourceConfig.
func shouldDropAlwaysAbsentValue(fwPath path.Path, attribute fwschema.Attribute, block fwschema.Block, value tftypes.Value, diags *diag.Diagnostics) (bool, error) {
	if fwPath.Equal(path.Root("timeouts")) {
		return true, nil
	}

	if attribute != nil {
		if attribute.GetDeprecationMessage() != "" {
			return true, nil
		}

		if attribute.IsComputed() && !attribute.IsOptional() {
			return true, nil
		}

		// SDKv2 compatibility: the SDKv2 adds an Optional+Computed "id" attribute
		// even when not declared in provider code, which then trips Core
		// validation. Drop it from generated config.
		if fwPath.Equal(path.Root("id")) && attribute.IsComputed() && attribute.IsOptional() {
			return true, nil
		}

		// SDKv2 compatibility: empty optional string is treated as null because
		// SDKv2 cannot distinguish "" from null.
		if value.Type().Equal(tftypes.String) && attribute.IsOptional() {
			var s string
			if err := value.As(&s); err != nil {
				diags.AddAttributeError(fwPath, "Generate Resource Config Error", genConfigErrDetail(err))
				return false, err
			}
			if len(s) == 0 {
				return true, nil
			}
		}

		// Note: no SDKv2 equivalent. Preserved for parity with existing tests.
		if value.Type().Equal(tftypes.Number) && attribute.IsOptional() {
			var n big.Float
			if err := value.As(&n); err != nil {
				diags.AddAttributeError(fwPath, "Generate Resource Config Error", genConfigErrDetail(err))
				return false, err
			}
			if n.Sign() == 0 {
				return true, nil
			}
		}
	}

	if block != nil && block.GetDeprecationMessage() != "" {
		return true, nil
	}

	return false, nil
}

// applyValidatorRules consults all attribute, block, and resource validator
// groups for the current path and updates markedForNullification accordingly.
// Returns true when the current path itself was marked, signalling the caller
// to return a null value immediately.
//
// Mirrors SDKv2's three sequential rule applications inside the cty.Transform
// callback (grpc_provider.go:1867-1912), one rule kind at a time.
func applyValidatorRules(ctx context.Context, schema fwschema.Schema, config *tfsdk.Config, fwPath path.Path, attribute fwschema.Attribute, block fwschema.Block, resourceGroups validatorGroups, configVal tftypes.Value, markedForNullification *path.Paths, diags *diag.Diagnostics) bool {
	if attribute != nil {
		groups := resolveAttributeValidatorGroups(ctx, config, fwPath, attribute, diags)
		if applyGroups(ctx, schema, groups, configVal, fwPath, false, markedForNullification, diags) {
			return true
		}
	}

	if block != nil {
		groups := resolveBlockValidatorGroups(ctx, config, fwPath, block, diags)
		if applyGroups(ctx, schema, groups, configVal, fwPath, false, markedForNullification, diags) {
			return true
		}
	}

	if applyGroups(ctx, schema, resourceGroups, configVal, fwPath, true, markedForNullification, diags) {
		return true
	}

	return false
}

// applyGroups runs the three rule families against the resolved groups. The
// requireMembership flag is true for resource-level groups (which are resolved
// up front against any path in the schema, so we must skip groups that don't
// contain the current path) and false for attribute/block groups (which are
// already scoped to the current path by construction).
func applyGroups(ctx context.Context, schema fwschema.Schema, groups validatorGroups, configVal tftypes.Value, curPath path.Path, requireMembership bool, markedForNullification *path.Paths, diags *diag.Diagnostics) bool {
	for _, members := range groups.ConflictsWith {
		if requireMembership && !members.Contains(curPath) {
			continue
		}
		markedForNullification.Append(applyKeepFirst(ctx, schema, members, configVal, curPath, diags)...)
		if markedForNullification.Contains(curPath) {
			return true
		}
	}

	for _, members := range groups.ExactlyOneOf {
		if requireMembership && !members.Contains(curPath) {
			continue
		}
		markedForNullification.Append(applyKeepFirst(ctx, schema, members, configVal, curPath, diags)...)
		if markedForNullification.Contains(curPath) {
			return true
		}
	}

	for _, members := range groups.AlsoRequires {
		if requireMembership && !members.Contains(curPath) {
			continue
		}
		markedForNullification.Append(applyAlsoRequires(ctx, schema, members, configVal, curPath, diags)...)
		if markedForNullification.Contains(curPath) {
			return true
		}
	}

	return false
}

// applyKeepFirst implements the SDKv2 ConflictsWith / ExactlyOneOf semantics:
// among all configured (non-null) members of the group, keep the
// lexicographically first one and mark the rest for nullification. Equivalent
// to processConflictsWith / processExactlyOneOf in SDKv2.
func applyKeepFirst(ctx context.Context, schema fwschema.Schema, members path.Paths, configVal tftypes.Value, curPath path.Path, diags *diag.Diagnostics) path.Paths {
	if len(members) == 0 {
		return nil
	}

	nonNull := configuredMembers(ctx, schema, members, configVal, curPath, diags)
	if len(nonNull) <= 1 {
		return nil
	}

	var marked path.Paths
	marked.Append(nonNull[1:]...)
	return marked
}

// applyAlsoRequires implements the SDKv2 RequiredWith semantics: if every
// member of the group is configured, do nothing; otherwise null all configured
// members. Equivalent to processRequiredWith in SDKv2.
func applyAlsoRequires(ctx context.Context, schema fwschema.Schema, members path.Paths, configVal tftypes.Value, curPath path.Path, diags *diag.Diagnostics) path.Paths {
	if len(members) == 0 {
		return nil
	}

	nonNull := configuredMembers(ctx, schema, members, configVal, curPath, diags)
	if len(nonNull) == len(members) {
		return nil
	}

	return nonNull
}

// configuredMembers returns the members of a validator group whose current
// configured value is non-null, sorted lexicographically. The current path is
// always included (the surrounding tftypes.Transform only invokes the callback
// for non-null values, so curPath is non-null by definition).
func configuredMembers(ctx context.Context, schema fwschema.Schema, members path.Paths, configVal tftypes.Value, curPath path.Path, diags *diag.Diagnostics) path.Paths {
	var nonNull path.Paths
	nonNull.Append(curPath)

	for _, member := range members {
		if member.Equal(curPath) {
			continue
		}

		val, ok := readTerraformValue(ctx, schema, configVal, member, diags)
		if !ok || val.IsNull() {
			continue
		}

		nonNull.Append(member)
	}

	return sortedPaths(nonNull)
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
	sort.Slice(result, func(i, j int) bool { return result[i].String() < result[j].String() })
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

// validatorGroups holds resolved members for each rule family. Members are
// concrete paths, already deduplicated and sorted. Each rule family is a slice
// of groups so a single attribute/block/resource may contribute multiple
// independent groups (e.g. two ConflictsWith validators on the same attribute).
type validatorGroups struct {
	ConflictsWith []path.Paths
	ExactlyOneOf  []path.Paths
	AlsoRequires  []path.Paths
}

// resolveResourceValidatorGroups materializes resource-level validator
// expressions into concrete path groups.
func resolveResourceValidatorGroups(ctx context.Context, config *tfsdk.Config, validators []resource.ConfigValidator, diags *diag.Diagnostics) validatorGroups {
	var groups validatorGroups
	for _, v := range validators {
		appendValidatorGroups(ctx, config, path.Expression{}, path.Empty(), false, v, &groups, diags)
	}
	return groups
}

// resolveAttributeValidatorGroups collects attribute validator groups for the current path.
//
// MAINTAINER NOTE: the 12-case switch is unavoidable: fwxschema models each
// element type's validators as a separate interface. Every case body is
// identical except for the validators accessor.
func resolveAttributeValidatorGroups(ctx context.Context, config *tfsdk.Config, currentPath path.Path, attribute fwschema.Attribute, diags *diag.Diagnostics) validatorGroups {
	var groups validatorGroups
	collect := func(v any) {
		appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, v, &groups, diags)
	}

	switch a := attribute.(type) {
	case fwxschema.AttributeWithBoolValidators:
		for _, v := range a.BoolValidators() {
			collect(v)
		}
	case fwxschema.AttributeWithFloat32Validators:
		for _, v := range a.Float32Validators() {
			collect(v)
		}
	case fwxschema.AttributeWithFloat64Validators:
		for _, v := range a.Float64Validators() {
			collect(v)
		}
	case fwxschema.AttributeWithInt32Validators:
		for _, v := range a.Int32Validators() {
			collect(v)
		}
	case fwxschema.AttributeWithInt64Validators:
		for _, v := range a.Int64Validators() {
			collect(v)
		}
	case fwxschema.AttributeWithListValidators:
		for _, v := range a.ListValidators() {
			collect(v)
		}
	case fwxschema.AttributeWithMapValidators:
		for _, v := range a.MapValidators() {
			collect(v)
		}
	case fwxschema.AttributeWithNumberValidators:
		for _, v := range a.NumberValidators() {
			collect(v)
		}
	case fwxschema.AttributeWithObjectValidators:
		for _, v := range a.ObjectValidators() {
			collect(v)
		}
	case fwxschema.AttributeWithSetValidators:
		for _, v := range a.SetValidators() {
			collect(v)
		}
	case fwxschema.AttributeWithStringValidators:
		for _, v := range a.StringValidators() {
			collect(v)
		}
	case fwxschema.AttributeWithDynamicValidators:
		for _, v := range a.DynamicValidators() {
			collect(v)
		}
	}

	return groups
}

// resolveBlockValidatorGroups collects block validator groups for the current path.
func resolveBlockValidatorGroups(ctx context.Context, config *tfsdk.Config, currentPath path.Path, block fwschema.Block, diags *diag.Diagnostics) validatorGroups {
	var groups validatorGroups
	collect := func(v any) {
		appendValidatorGroups(ctx, config, currentPath.Expression(), currentPath, true, v, &groups, diags)
	}

	switch b := block.(type) {
	case fwxschema.BlockWithListValidators:
		for _, v := range b.ListValidators() {
			collect(v)
		}
	case fwxschema.BlockWithObjectValidators:
		for _, v := range b.ObjectValidators() {
			collect(v)
		}
	case fwxschema.BlockWithSetValidators:
		for _, v := range b.SetValidators() {
			collect(v)
		}
	}

	return groups
}

// appendValidatorGroups inspects validator for each marker interface
// (ConflictsWithValidator / ExactlyOneOfValidator / AlsoRequiresValidator) and
// appends the resolved member paths to the corresponding bucket on groups.
func appendValidatorGroups(ctx context.Context, config *tfsdk.Config, baseExpression path.Expression, currentPath path.Path, includeCurrent bool, validator any, groups *validatorGroups, diags *diag.Diagnostics) {
	if v, ok := validator.(schemavalidator.ConflictsWithValidator); ok {
		if members := resolveValidatorGroupPaths(ctx, config, baseExpression, currentPath, includeCurrent, v.ConflictsWithPaths(), diags); len(members) > 0 {
			groups.ConflictsWith = append(groups.ConflictsWith, members)
		}
	}
	if v, ok := validator.(schemavalidator.ExactlyOneOfValidator); ok {
		if members := resolveValidatorGroupPaths(ctx, config, baseExpression, currentPath, includeCurrent, v.ExactlyOneOfPaths(), diags); len(members) > 0 {
			groups.ExactlyOneOf = append(groups.ExactlyOneOf, members)
		}
	}
	if v, ok := validator.(schemavalidator.AlsoRequiresValidator); ok {
		if members := resolveValidatorGroupPaths(ctx, config, baseExpression, currentPath, includeCurrent, v.AlsoRequiresPaths(), diags); len(members) > 0 {
			groups.AlsoRequires = append(groups.AlsoRequires, members)
		}
	}
}

// resolveValidatorGroupPaths returns the concrete members of a validator group
// for the current path. If includeCurrent is true the current path is added to
// the group (the framework-validators convention is to list peer paths only,
// excluding self). Returns nil for groups smaller than two members because
// no rule has anything to act on.
func resolveValidatorGroupPaths(ctx context.Context, config *tfsdk.Config, baseExpression path.Expression, currentPath path.Path, includeCurrent bool, expressions path.Expressions, diags *diag.Diagnostics) path.Paths {
	var members path.Paths

	if includeCurrent {
		members.Append(currentPath)
	}

	// path.Paths.Append deduplicates, so an expression that resolves to
	// currentPath (e.g. self-references in resource-level groups) will not
	// produce a duplicate entry.
	members.Append(resolveExpressions(ctx, config, baseExpression, expressions, diags)...)
	members = sortedPaths(members)

	if len(members) < 2 {
		return nil
	}

	return members
}

// stateToConfig returns a *tfsdk.Config with a copied value from a tfsdk.State.
func stateToConfig(state tfsdk.State) *tfsdk.Config {
	return &tfsdk.Config{
		Raw:    state.Raw.Copy(),
		Schema: state.Schema,
	}
}
