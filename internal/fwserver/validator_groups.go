// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"context"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/totftypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// buildValidatorGroups builds deduplicated validator groups from schema
// attributes, blocks, and nested objects. Matching uses the current config so
// wildcard expressions expand into concrete paths when values are present.
func buildValidatorGroups(ctx context.Context, config tftypes.Value, schema fwschema.Schema, res resource.Resource, attributeExpressionFunc func(fwschema.Attribute) path.Expressions) map[string]path.Paths {
	groups := map[string]path.Paths{}
	configView := tfsdk.Config{Raw: config, Schema: schema}

	for name, attr := range schema.GetAttributes() {
		attributePath := path.Root(name)
		addAttributeValidatorGroup(ctx, configView, groups, attributePath, attributeExpressionFunc(attr))

		collectNestedAttributeValidatorGroups(ctx, configView, attributePath, attr, attributeExpressionFunc, groups)
	}

	for name, block := range schema.GetBlocks() {
		blockPath := path.Root(name)
		addResolvedValidatorGroup(ctx, configView, groups, blockPath.Expression(), getBlockValidatorPathExpressions(block))

		collectNestedBlockValidatorGroups(ctx, configView, blockPath, block, attributeExpressionFunc, groups)
	}

	if resourceWithConfigValidators, ok := res.(resource.ResourceWithConfigValidators); ok {
		for _, configValidator := range resourceWithConfigValidators.ConfigValidators(ctx) {
			addResolvedValidatorGroup(ctx, configView, groups, path.MatchRelative(), getGenericValidatorPathExpressions(configValidator))
		}
	}

	return groups
}

func addAttributeValidatorGroup(ctx context.Context, config tfsdk.Config, groups map[string]path.Paths, attributePath path.Path, expressions path.Expressions) {
	members := path.Paths{attributePath}
	members = appendResolvedMembers(ctx, config, attributePath.Expression(), members, expressions)
	addValidatorGroup(groups, members)
}

func addResolvedValidatorGroup(ctx context.Context, config tfsdk.Config, groups map[string]path.Paths, baseExpression path.Expression, expressions path.Expressions) {
	members := appendResolvedMembers(ctx, config, baseExpression, nil, expressions)
	addValidatorGroup(groups, members)
}

func addValidatorGroup(groups map[string]path.Paths, members path.Paths) {
	if len(members) <= 1 {
		return
	}

	sort.Slice(members, func(i, j int) bool {
		return members[i].String() < members[j].String()
	})

	keyParts := make([]string, 0, len(members))
	for _, member := range members {
		keyParts = append(keyParts, member.String())
	}

	groups[strings.Join(keyParts, ",")] = members
}

func getExactlyOneOfExpressions(attr fwschema.Attribute) path.Expressions {
	return getValidatorPathExpressions(attr, func(v interface{}) path.Expressions {
		if exactlyOneOfValidator, ok := v.(validator.ExactlyOneOfValidator); ok {
			return exactlyOneOfValidator.Paths()
		}

		return nil
	})
}

func getAlsoRequiresExpressions(attr fwschema.Attribute) path.Expressions {
	return getValidatorPathExpressions(attr, func(v interface{}) path.Expressions {
		if alsoRequiresValidator, ok := v.(validator.AlsoRequiresValidator); ok {
			return alsoRequiresValidator.Paths()
		}

		return nil
	})
}

// getConflictsWithExpressions extracts ConflictsWith path expressions from an attribute's
// validators. It checks all typed validator interfaces (String, Bool, Int64, etc.) and
// returns the paths from any validator that implements validator.ConflictsWithValidator.
func getConflictsWithExpressions(attr fwschema.Attribute) path.Expressions {
	return getValidatorPathExpressions(attr, func(v interface{}) path.Expressions {
		if conflictsWithValidator, ok := v.(validator.ConflictsWithValidator); ok {
			return conflictsWithValidator.Paths()
		}

		return nil
	})
}

func getValidatorPathExpressions(attr fwschema.Attribute, expressionFunc func(interface{}) path.Expressions) path.Expressions {
	var result path.Expressions

	checkValidator := func(v interface{}) {
		result = append(result, expressionFunc(v)...)
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

func getBlockValidatorPathExpressions(block fwschema.Block) path.Expressions {
	var result path.Expressions

	appendValidatorPaths := func(v interface{}) {
		switch validatorWithPaths := v.(type) {
		case validator.List:
			result = append(result, getGenericValidatorPathExpressions(validatorWithPaths)...)
		case validator.Object:
			result = append(result, getGenericValidatorPathExpressions(validatorWithPaths)...)
		case validator.Set:
			result = append(result, getGenericValidatorPathExpressions(validatorWithPaths)...)
		}
	}

	if b, ok := block.(fwxschema.BlockWithListValidators); ok {
		for _, v := range b.ListValidators() {
			appendValidatorPaths(v)
		}
	}
	if b, ok := block.(fwxschema.BlockWithObjectValidators); ok {
		for _, v := range b.ObjectValidators() {
			appendValidatorPaths(v)
		}
	}
	if b, ok := block.(fwxschema.BlockWithSetValidators); ok {
		for _, v := range b.SetValidators() {
			appendValidatorPaths(v)
		}
	}

	return result
}

func getNestedAttributeObjectValidatorPathExpressions(object fwschema.NestedAttributeObject) path.Expressions {
	if o, ok := object.(fwxschema.NestedAttributeObjectWithValidators); ok {
		return getObjectValidatorsPathExpressions(o.ObjectValidators())
	}

	return nil
}

func getNestedBlockObjectValidatorPathExpressions(object fwschema.NestedBlockObject) path.Expressions {
	if o, ok := object.(fwxschema.NestedBlockObjectWithValidators); ok {
		return getObjectValidatorsPathExpressions(o.ObjectValidators())
	}

	return nil
}

func getObjectValidatorsPathExpressions(validators []validator.Object) path.Expressions {
	var result path.Expressions

	for _, v := range validators {
		result = append(result, getGenericValidatorPathExpressions(v)...)
	}

	return result
}

func getGenericValidatorPathExpressions(v interface{}) path.Expressions {
	if exactlyOneOfValidator, ok := v.(validator.ExactlyOneOfValidator); ok {
		return exactlyOneOfValidator.Paths()
	}

	if alsoRequiresValidator, ok := v.(validator.AlsoRequiresValidator); ok {
		return alsoRequiresValidator.Paths()
	}

	if conflictsWithValidator, ok := v.(validator.ConflictsWithValidator); ok {
		return conflictsWithValidator.Paths()
	}

	return nil
}

func collectNestedAttributeValidatorGroups(ctx context.Context, config tfsdk.Config, currentPath path.Path, attr fwschema.Attribute, attributeExpressionFunc func(fwschema.Attribute) path.Expressions, groups map[string]path.Paths) {
	nestedAttribute, ok := attr.(fwschema.NestedAttribute)
	if !ok {
		return
	}

	for _, instancePath := range nestedInstancePaths(ctx, config, currentPath, nestedAttribute.GetNestingMode()) {
		collectNestedAttributeObjectValidatorGroups(ctx, config, instancePath, nestedAttribute.GetNestedObject(), attributeExpressionFunc, groups)
	}
}

func collectNestedAttributeObjectValidatorGroups(ctx context.Context, config tfsdk.Config, currentPath path.Path, object fwschema.NestedAttributeObject, attributeExpressionFunc func(fwschema.Attribute) path.Expressions, groups map[string]path.Paths) {
	addResolvedValidatorGroup(ctx, config, groups, currentPath.Expression(), getNestedAttributeObjectValidatorPathExpressions(object))

	for name, attr := range object.GetAttributes() {
		nextPath := currentPath.AtName(name)
		addAttributeValidatorGroup(ctx, config, groups, nextPath, attributeExpressionFunc(attr))

		collectNestedAttributeValidatorGroups(ctx, config, nextPath, attr, attributeExpressionFunc, groups)
	}
}

func collectNestedBlockValidatorGroups(ctx context.Context, config tfsdk.Config, currentPath path.Path, block fwschema.Block, attributeExpressionFunc func(fwschema.Attribute) path.Expressions, groups map[string]path.Paths) {
	addResolvedValidatorGroup(ctx, config, groups, currentPath.Expression(), getBlockValidatorPathExpressions(block))

	for _, instancePath := range nestedBlockInstancePaths(ctx, config, currentPath, block.GetNestingMode()) {
		collectNestedBlockObjectValidatorGroups(ctx, config, instancePath, block.GetNestedObject(), attributeExpressionFunc, groups)
	}
}

func collectNestedBlockObjectValidatorGroups(ctx context.Context, config tfsdk.Config, currentPath path.Path, object fwschema.NestedBlockObject, attributeExpressionFunc func(fwschema.Attribute) path.Expressions, groups map[string]path.Paths) {
	addResolvedValidatorGroup(ctx, config, groups, currentPath.Expression(), getNestedBlockObjectValidatorPathExpressions(object))

	for name, attr := range object.GetAttributes() {
		nextPath := currentPath.AtName(name)
		addAttributeValidatorGroup(ctx, config, groups, nextPath, attributeExpressionFunc(attr))

		collectNestedAttributeValidatorGroups(ctx, config, nextPath, attr, attributeExpressionFunc, groups)
	}

	for name, block := range object.GetBlocks() {
		nextPath := currentPath.AtName(name)
		collectNestedBlockValidatorGroups(ctx, config, nextPath, block, attributeExpressionFunc, groups)
	}
}

func appendResolvedMembers(ctx context.Context, config tfsdk.Config, baseExpression path.Expression, members path.Paths, expressions path.Expressions) path.Paths {
	if len(expressions) == 0 {
		return members
	}

	memberSet := map[string]struct{}{}
	for _, member := range members {
		memberSet[member.String()] = struct{}{}
	}

	for _, expression := range expressions {
		resolvedExpression := baseExpression.Merge(expression).Resolve()
		matchedPaths, diags := config.PathMatches(ctx, resolvedExpression)
		if diags.HasError() {
			continue
		}

		for _, matchedPath := range matchedPaths {
			if _, ok := memberSet[matchedPath.String()]; ok {
				continue
			}

			memberSet[matchedPath.String()] = struct{}{}
			members = append(members, matchedPath)
		}
	}

	return members
}

func nestedInstancePaths(ctx context.Context, config tfsdk.Config, currentPath path.Path, nestingMode fwschema.NestingMode) path.Paths {
	switch nestingMode {
	case fwschema.NestingModeSingle:
		return path.Paths{currentPath}
	case fwschema.NestingModeList:
		return collectionInstancePaths(ctx, config, currentPath.Expression().AtAnyListIndex(), len(currentPath.Steps()))
	case fwschema.NestingModeMap:
		return collectionInstancePaths(ctx, config, currentPath.Expression().AtAnyMapKey(), len(currentPath.Steps()))
	case fwschema.NestingModeSet:
		return collectionInstancePaths(ctx, config, currentPath.Expression().AtAnySetValue(), len(currentPath.Steps()))
	default:
		return nil
	}
}

func nestedBlockInstancePaths(ctx context.Context, config tfsdk.Config, currentPath path.Path, nestingMode fwschema.BlockNestingMode) path.Paths {
	switch nestingMode {
	case fwschema.BlockNestingModeSingle:
		return path.Paths{currentPath}
	case fwschema.BlockNestingModeList:
		return collectionInstancePaths(ctx, config, currentPath.Expression().AtAnyListIndex(), len(currentPath.Steps()))
	case fwschema.BlockNestingModeSet:
		return collectionInstancePaths(ctx, config, currentPath.Expression().AtAnySetValue(), len(currentPath.Steps()))
	default:
		return nil
	}
}

func collectionInstancePaths(ctx context.Context, config tfsdk.Config, expression path.Expression, parentDepth int) path.Paths {
	paths, diags := config.PathMatches(ctx, expression)
	if diags.HasError() {
		return nil
	}

	var result path.Paths
	for _, matchedPath := range paths {
		if len(matchedPath.Steps()) <= parentDepth {
			continue
		}

		result = append(result, matchedPath)
	}

	return result
}

func terraformPathFromPath(ctx context.Context, fwPath path.Path) *tftypes.AttributePath {
	tfPath, diags := totftypes.AttributePath(ctx, fwPath)
	if diags.HasError() {
		return nil
	}

	return tfPath
}
