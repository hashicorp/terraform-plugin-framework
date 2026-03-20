// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwserver

import (
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// buildAttributeValidatorGroups builds deduplicated top-level validator groups.
// Relative expressions are resolved against the current attribute path and only
// exact top-level attribute paths are included.
func buildAttributeValidatorGroups(schema fwschema.Schema, expressionFunc func(fwschema.Attribute) path.Expressions) map[string][]string {
	groups := map[string][]string{}

	for name, attr := range schema.GetAttributes() {
		expressions := expressionFunc(attr)

		if len(expressions) == 0 {
			continue
		}

		members := []string{name}
		memberSet := map[string]struct{}{
			name: {},
		}
		currentExpression := path.MatchRoot(name)

		for _, expression := range expressions {
			memberName, ok := topLevelAttributeName(currentExpression.Merge(expression).Resolve())

			if !ok {
				continue
			}

			if _, ok := memberSet[memberName]; ok {
				continue
			}

			memberSet[memberName] = struct{}{}
			members = append(members, memberName)
		}

		if len(members) <= 1 {
			continue
		}

		sort.Strings(members)
		groups[strings.Join(members, ",")] = members
	}

	return groups
}

func topLevelAttributeName(expression path.Expression) (string, bool) {
	steps := expression.Steps()

	if len(steps) != 1 {
		return "", false
	}

	step, ok := steps[0].(path.ExpressionStepAttributeNameExact)
	if !ok {
		return "", false
	}

	return string(step), true
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
