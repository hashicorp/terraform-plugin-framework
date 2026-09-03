// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
)

// typeCanBeStringKeyed reports whether values of the given type can be grouped
// by their String() output as a stand-in for equality.
//
// Grouping is only sound when String() is equal for every pair of values that
// Equal reports as equal. That holds for primitives, objects (String() sorts
// attribute names) and maps, but not for sets: set equality ignores element
// order while SetValue.String() renders elements in slice order, so two equal
// sets can produce different strings. Dynamic types are only concrete at
// runtime, so they are excluded as well.
//
// Callers that get false must fall back to comparing values pairwise.
func typeCanBeStringKeyed(t attr.Type) bool {
	if t == nil {
		return false
	}

	switch nested := t.(type) {
	case SetTypable, DynamicTypable:
		return false
	case attr.TypeWithElementType:
		return typeCanBeStringKeyed(nested.ElementType())
	case attr.TypeWithElementTypes:
		for _, elementType := range nested.ElementTypes() {
			if !typeCanBeStringKeyed(elementType) {
				return false
			}
		}
	case attr.TypeWithAttributeTypes:
		for _, attributeType := range nested.AttributeTypes() {
			if !typeCanBeStringKeyed(attributeType) {
				return false
			}
		}
	}

	return true
}
