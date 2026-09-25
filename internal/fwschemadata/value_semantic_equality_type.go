// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package fwschemadata

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// typeMightHaveSemanticEquals reports whether a value of the given type, or any
// value nested anywhere within its type tree, could implement one of the
// basetypes semantic equality interfaces.
//
// It is a conservative, static pre-check used to skip the recursive per-element
// semantic equality walk performed by the collection reconciliation functions.
// When it returns false, no value the collection can contain implements
// semantic equality, so the walk is guaranteed to be a no-op and can be
// skipped. It only returns false when the entire type tree is concrete and
// provably free of semantic-equality implementers; anything uncertain (most
// importantly dynamic types, whose concrete type is only known at runtime)
// returns true so the normal walk still runs.
func typeMightHaveSemanticEquals(ctx context.Context, t attr.Type) bool {
	if t == nil {
		return false
	}

	// A representative value of the type reveals whether the value type
	// implements any semantic equality interface.
	switch t.ValueType(ctx).(type) {
	case basetypes.BoolValuableWithSemanticEquals,
		basetypes.Float32ValuableWithSemanticEquals,
		basetypes.Float64ValuableWithSemanticEquals,
		basetypes.Int32ValuableWithSemanticEquals,
		basetypes.Int64ValuableWithSemanticEquals,
		basetypes.ListValuableWithSemanticEquals,
		basetypes.MapValuableWithSemanticEquals,
		basetypes.NumberValuableWithSemanticEquals,
		basetypes.ObjectValuableWithSemanticEquals,
		basetypes.SetValuableWithSemanticEquals,
		basetypes.StringValuableWithSemanticEquals,
		basetypes.DynamicValuableWithSemanticEquals:
		return true
	case basetypes.DynamicValuable:
		// The concrete underlying type of a dynamic value is only known at
		// runtime and may itself implement semantic equality, so be
		// conservative and assume it might.
		return true
	}

	// Recurse into nested element/attribute types. Terraform type trees are
	// always finite, so no explicit cycle guard is required.
	switch nested := t.(type) {
	case attr.TypeWithElementType:
		return typeMightHaveSemanticEquals(ctx, nested.ElementType())
	case attr.TypeWithElementTypes:
		for _, elementType := range nested.ElementTypes() {
			if typeMightHaveSemanticEquals(ctx, elementType) {
				return true
			}
		}
	case attr.TypeWithAttributeTypes:
		for _, attributeType := range nested.AttributeTypes() {
			if typeMightHaveSemanticEquals(ctx, attributeType) {
				return true
			}
		}
	}

	return false
}
