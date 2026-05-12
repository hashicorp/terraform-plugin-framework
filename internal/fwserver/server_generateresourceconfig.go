// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"errors"
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
// MAINTAINER NOTE: // TODO: reword
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
	resp.GeneratedConfig = &tfsdk.Config{
		Raw:    req.State.Raw.Copy(),
		Schema: req.State.Schema,
	}

	var resourceConfigValidators []resource.ConfigValidator
	if r, ok := req.Resource.(resource.ResourceWithConfigValidators); ok {
		resourceConfigValidators = r.ConfigValidators(ctx)
	}

	// we'll collect all validators into these groups in the first Transform pass
	var conflictsWith []path.Paths
	var exactlyOneOf []path.Paths
	var alsoRequires []path.Paths

	// resource level ones aren't part of Transform, so add them here ahead of time
	for _, validator := range resourceConfigValidators {
		if v, ok := validator.(schemavalidator.ConflictsWithValidator); ok {
			if members := resolvePathExpressions(ctx, resp.GeneratedConfig, path.Expression{}, path.Empty(), v.ConflictsWithPaths(), &diags); len(members) > 1 {
				conflictsWith = append(conflictsWith, members)
			}
		}
		if v, ok := validator.(schemavalidator.ExactlyOneOfValidator); ok {
			if members := resolvePathExpressions(ctx, resp.GeneratedConfig, path.Expression{}, path.Empty(), v.ExactlyOneOfPaths(), &diags); len(members) > 1 {
				exactlyOneOf = append(exactlyOneOf, members)
			}
		}
		if v, ok := validator.(schemavalidator.AlsoRequiresValidator); ok {
			if members := resolvePathExpressions(ctx, resp.GeneratedConfig, path.Expression{}, path.Empty(), v.AlsoRequiresPaths(), &diags); len(members) > 1 {
				alsoRequires = append(alsoRequires, members)
			}
		}
	}

	config := req.State.Raw

	// Pass 1: per-value drops + group decision accumulation.
	config, err := tftypes.Transform(config, func(tfPath *tftypes.AttributePath, value tftypes.Value) (tftypes.Value, error) {
		if len(tfPath.Steps()) == 0 {
			return value, nil
		}

		if value.IsNull() {
			return value, nil
		}

		fwPath, fwPathDiags := fromtftypes.AttributePath(ctx, tfPath, req.ResourceSchema)
		diags.Append(fwPathDiags...)
		if fwPathDiags.HasError() {
			return value, nil
		}

		attribute, attrErr := req.ResourceSchema.AttributeAtTerraformPath(ctx, tfPath)
		if attrErr != nil {
			if !errors.Is(attrErr, fwschema.ErrPathIsBlock) &&
				!errors.Is(attrErr, fwschema.ErrPathInsideDynamicAttribute) &&
				!errors.Is(attrErr, fwschema.ErrPathInsideAtomicAttribute) {
				logging.FrameworkError(ctx, "couldn't find attribute in resource schema")
				diags.AddAttributeError(fwPath, "Generate Resource Config Error", genConfigErrDetail(attrErr))
				return value, attrErr
			}
			attribute = nil
		}

		block, blockErr := fwschema.SchemaBlockAtTerraformPath(ctx, req.ResourceSchema, tfPath)
		if blockErr != nil {
			if !errors.Is(blockErr, fwschema.ErrPathIsAttribute) &&
				!errors.Is(blockErr, fwschema.ErrPathInsideDynamicAttribute) &&
				!errors.Is(blockErr, fwschema.ErrPathInsideAtomicAttribute) {
				logging.FrameworkError(ctx, "couldn't find block in resource schema")
				diags.AddAttributeError(fwPath, "Generate Resource Config Error", genConfigErrDetail(blockErr))
				return value, blockErr
			}
			block = nil
		}

		nullValue := tftypes.NewValue(value.Type(), nil)

		if fwPath.Equal(path.Root("timeouts")) {
			return nullValue, nil
		}

		if attribute != nil {
			if attribute.GetDeprecationMessage() != "" {
				return nullValue, nil
			}

			if attribute.IsComputed() && !attribute.IsOptional() {
				return nullValue, nil
			}
		}

		if block != nil && block.GetDeprecationMessage() != "" {
			return nullValue, nil
		}

		if attribute != nil {
			validators := getValidatorsFromAttribute(attribute)
			for _, validator := range validators {

				if v, ok := validator.(schemavalidator.ConflictsWithValidator); ok {
					if members := resolvePathExpressions(ctx, resp.GeneratedConfig, fwPath.Expression(), fwPath, v.ConflictsWithPaths(), &diags); len(members) > 1 {
						conflictsWith = append(conflictsWith, members)
					}
				}
				if v, ok := validator.(schemavalidator.ExactlyOneOfValidator); ok {
					if members := resolvePathExpressions(ctx, resp.GeneratedConfig, fwPath.Expression(), fwPath, v.ExactlyOneOfPaths(), &diags); len(members) > 1 {
						exactlyOneOf = append(exactlyOneOf, members)
					}
				}
				if v, ok := validator.(schemavalidator.AlsoRequiresValidator); ok {
					if members := resolvePathExpressions(ctx, resp.GeneratedConfig, fwPath.Expression(), fwPath, v.AlsoRequiresPaths(), &diags); len(members) > 1 {
						alsoRequires = append(alsoRequires, members)
					}
				}

			}
		}

		if block != nil {
			validators := getValidatorsFromBlock(block)
			for _, validator := range validators {

				if v, ok := validator.(schemavalidator.ConflictsWithValidator); ok {
					if members := resolvePathExpressions(ctx, resp.GeneratedConfig, fwPath.Expression(), fwPath, v.ConflictsWithPaths(), &diags); len(members) > 1 {
						conflictsWith = append(conflictsWith, members)
					}
				}
				if v, ok := validator.(schemavalidator.ExactlyOneOfValidator); ok {
					if members := resolvePathExpressions(ctx, resp.GeneratedConfig, fwPath.Expression(), fwPath, v.ExactlyOneOfPaths(), &diags); len(members) > 1 {
						exactlyOneOf = append(exactlyOneOf, members)
					}
				}
				if v, ok := validator.(schemavalidator.AlsoRequiresValidator); ok {
					if members := resolvePathExpressions(ctx, resp.GeneratedConfig, fwPath.Expression(), fwPath, v.AlsoRequiresPaths(), &diags); len(members) > 1 {
						alsoRequires = append(alsoRequires, members)
					}
				}

			}

		}

		return value, nil
	}) // end first Transform()

	// collect paths that will be set to null in the second Transform() pass, depending on the behaviour of the validator type
	markedForNullification := path.Paths{}
	for _, members := range conflictsWith {
		nonNull := getNonNullMembers(ctx, req.ResourceSchema, config, members, &diags)

		// more than one member with a value? mark all but the lexicographically first for nullification
		if len(nonNull) > 1 {
			sort.Slice(nonNull, func(i, j int) bool { return nonNull[i].String() < nonNull[j].String() })
			markedForNullification.Append(nonNull[1:]...)
		}
	}

	// yes, this is the same as conflictsWith because we let Terraform
	// handle defaults which means these two groups behave the same
	for _, members := range exactlyOneOf {
		nonNull := getNonNullMembers(ctx, req.ResourceSchema, config, members, &diags)

		// more than one member with a value? mark all but the lexicographically first for nullification
		if len(nonNull) > 1 {
			sort.Slice(nonNull, func(i, j int) bool { return nonNull[i].String() < nonNull[j].String() })
			markedForNullification.Append(nonNull[1:]...)
		}
	}

	for _, members := range alsoRequires {
		nonNull := getNonNullMembers(ctx, req.ResourceSchema, config, members, &diags)

		// if not all members have a value mark all for nullification
		if len(nonNull) != len(members) {
			markedForNullification.Append(nonNull...)
		}
	}

	if err != nil {
		logging.FrameworkError(ctx,
			"Error transforming state value during resource config generation",
			map[string]any{logging.KeyError: err.Error()},
		)
	}

	// Pass 2: set values to null that are marked for nullification
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

func genConfigErrDetail(err error) string {
	return "An unexpected error was encountered trying to generate the resource config for import. " +
		"This likely indicates a bug in the Terraform provider framework or Terraform Core. Please report the following to the provider developer:\n\n" +
		err.Error()
}

func getNonNullMembers(ctx context.Context, schema fwschema.Schema, currentConfig tftypes.Value, members path.Paths, diags *diag.Diagnostics) path.Paths {
	var nonNull path.Paths
	for _, member := range members {
		val, ok := readTerraformValue(ctx, schema, currentConfig, member, diags)
		if !ok || val.IsNull() {
			continue
		}

		nonNull.Append(member)
	}

	return nonNull
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

func getValidatorsFromAttribute(attribute fwschema.Attribute) []any {
	var validators []any

	switch a := attribute.(type) {
	case fwxschema.AttributeWithBoolValidators:
		for _, v := range a.BoolValidators() {
			validators = append(validators, v)
		}
	case fwxschema.AttributeWithFloat32Validators:
		for _, v := range a.Float32Validators() {
			validators = append(validators, v)
		}
	case fwxschema.AttributeWithFloat64Validators:
		for _, v := range a.Float64Validators() {
			validators = append(validators, v)
		}
	case fwxschema.AttributeWithInt32Validators:
		for _, v := range a.Int32Validators() {
			validators = append(validators, v)
		}
	case fwxschema.AttributeWithInt64Validators:
		for _, v := range a.Int64Validators() {
			validators = append(validators, v)
		}
	case fwxschema.AttributeWithListValidators:
		for _, v := range a.ListValidators() {
			validators = append(validators, v)
		}
	case fwxschema.AttributeWithMapValidators:
		for _, v := range a.MapValidators() {
			validators = append(validators, v)
		}
	case fwxschema.AttributeWithNumberValidators:
		for _, v := range a.NumberValidators() {
			validators = append(validators, v)
		}
	case fwxschema.AttributeWithObjectValidators:
		for _, v := range a.ObjectValidators() {
			validators = append(validators, v)
		}
	case fwxschema.AttributeWithSetValidators:
		for _, v := range a.SetValidators() {
			validators = append(validators, v)
		}
	case fwxschema.AttributeWithStringValidators:
		for _, v := range a.StringValidators() {
			validators = append(validators, v)
		}
	case fwxschema.AttributeWithDynamicValidators:
		for _, v := range a.DynamicValidators() {
			validators = append(validators, v)
		}
	}

	return validators
}

func getValidatorsFromBlock(block fwschema.Block) []any {
	var validators []any

	switch b := block.(type) {
	case fwxschema.BlockWithListValidators:
		for _, v := range b.ListValidators() {
			validators = append(validators, v)
		}
	case fwxschema.BlockWithObjectValidators:
		for _, v := range b.ObjectValidators() {
			validators = append(validators, v)
		}
	case fwxschema.BlockWithSetValidators:
		for _, v := range b.SetValidators() {
			validators = append(validators, v)
		}
	}
	// TODO: case fwxschema.NestedBlockObjectWithValidators? -> it's not passed in here i think, but it should be handled somewhere: NestedBlockObjectWithValidators
	// TODO: -> failing test case name: response-nested-block-object-conflicts-with-group

	return validators
}

func resolvePathExpressions(ctx context.Context, config *tfsdk.Config, baseExpression path.Expression, currentPath path.Path, expressions path.Expressions, diags *diag.Diagnostics) path.Paths {
	var resolvedPaths path.Paths

	// add the currentPath because some validators might not include their
	// own path in the expressions, if it exists path.Paths.Append will deduplicate it
	// skip empty paths, as they are not valid for validators
	if len(currentPath.Steps()) > 0 {
		resolvedPaths.Append(currentPath)
	}

	for _, expression := range expressions {
		mergedExpression := baseExpression.Merge(expression)
		// turn expressions into concrete paths, e.g. resolve "foo.*.id" to ["foo.0.id", "foo.1.id", etc.] based on the current config
		resolvedMatches, matchDiags := config.PathMatches(ctx, mergedExpression)
		diags.Append(matchDiags...)
		resolvedPaths.Append(resolvedMatches...)
	}
	sort.Slice(resolvedPaths, func(i, j int) bool { return resolvedPaths[i].String() < resolvedPaths[j].String() })
	return resolvedPaths
}
