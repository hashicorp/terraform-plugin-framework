// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwtype

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// ContainsCollectionWithDynamic will return true if an attr.Type is a complex type that either is or contains any
// collection types with dynamic types, which are not supported by the framework type system. Primitives or invalid
// types (missing) will return false.
//
// Unsupported collection types include:
//   - Lists that contain a dynamic type
//   - Maps that contain a dynamic type
//   - Sets that contain a dynamic type
func ContainsCollectionWithDynamic(typ attr.Type) bool {
	switch attrType := typ.(type) {
	// We haven't run into a collection type yet, so it's valid for this to be a dynamic type
	case attr.TypeWithDynamicValue:
		return false
	// Lists, maps, sets
	case attr.TypeWithElementType:
		// We found a collection, need to ensure there are no dynamics from this point on.
		return containsDynamic(attrType.ElementType())
	// Tuples
	case attr.TypeWithElementTypes:
		for _, elemType := range attrType.ElementTypes() {
			hasDynamic := ContainsCollectionWithDynamic(elemType)
			if hasDynamic {
				return true
			}
		}
		return false
	// Objects
	case attr.TypeWithAttributeTypes:
		for _, objAttrType := range attrType.AttributeTypes() {
			hasDynamic := ContainsCollectionWithDynamic(objAttrType)
			if hasDynamic {
				return true
			}
		}
		return false
	// Primitives, missing types, etc.
	default:
		return false
	}
}

// containsDynamic is a helper that ensures that no nested types contain a dynamic type.
func containsDynamic(typ attr.Type) bool {
	switch attrType := typ.(type) {
	// Found a dynamic!
	case attr.TypeWithDynamicValue:
		return true
	// Lists, maps, sets
	case attr.TypeWithElementType:
		return containsDynamic(attrType.ElementType())
	// Tuples
	case attr.TypeWithElementTypes:
		for _, elemType := range attrType.ElementTypes() {
			hasDynamic := containsDynamic(elemType)
			if hasDynamic {
				return true
			}
		}
		return false
	// Objects
	case attr.TypeWithAttributeTypes:
		for _, objAttrType := range attrType.AttributeTypes() {
			hasDynamic := containsDynamic(objAttrType)
			if hasDynamic {
				return true
			}
		}
		return false
	// Primitives, missing types, etc.
	default:
		return false
	}
}

func AttributeCollectionWithDynamicTypeDiag(attributePath path.Path) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Invalid Schema Implementation",
		"When validating the schema, an implementation issue was found. "+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("%q is an attribute that contains a collection type with a nested dynamic type. ", attributePath)+
			"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
	)
}

func BlockCollectionWithDynamicTypeDiag(attributePath path.Path) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Invalid Schema Implementation",
		"When validating the schema, an implementation issue was found. "+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("%q is a block that contains a collection type with a nested dynamic type. ", attributePath)+
			"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
	)
}

func ParameterCollectionWithDynamicTypeDiag(argument int64) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Invalid Function Definition",
		"When validating the function definition, an implementation issue was found. "+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("Parameter at position %d contains a collection type with a nested dynamic type. ", argument)+
			"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
	)
}

func ReturnCollectionWithDynamicTypeDiag() diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Invalid Function Definition",
		"When validating the function definition, an implementation issue was found. "+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			"Return contains a collection type with a nested dynamic type. "+
			"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
	)
}
