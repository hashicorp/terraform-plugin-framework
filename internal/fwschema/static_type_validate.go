// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// TODO: Both of these functions should likely move to a different package, but not sure which? attr?

// StructuralTypeContainsDynamic will return true if an attr.Type is a structural type (object or tuple) that contains
// any collection types with dynamic types, which are not supported by the framework type system.
//
// Unsupported collection types include:
//   - Lists that contain a dynamic type
//   - Maps that contain a dynamic type
//   - Sets that contain a dynamic type
func StructuralTypeContainsDynamic(typ attr.Type) bool {
	switch attrType := typ.(type) {
	// Dynamic types in structural types (object or tuple) are allowed
	case attr.TypeWithDynamicValue:
		return false
	// Lists, maps, sets
	case attr.TypeWithElementType:
		return CollectionTypeContainsDynamic(attrType.ElementType())
	// Tuples
	case attr.TypeWithElementTypes:
		for _, elemType := range attrType.ElementTypes() {
			hasDynamic := StructuralTypeContainsDynamic(elemType)
			if hasDynamic {
				return true
			}
		}
		return false
	// Objects
	case attr.TypeWithAttributeTypes:
		for _, objAttrType := range attrType.AttributeTypes() {
			hasDynamic := StructuralTypeContainsDynamic(objAttrType)
			if hasDynamic {
				return true
			}
		}
		return false
	// Missing or unsupported type
	default:
		return false
	}
}

// CollectionTypeContainsDynamic will return true if an attr.Type is a collection type that contains
// any dynamic types, which are not supported by the framework type system.
//
// Unsupported collection types include:
//   - Lists that contain a dynamic type
//   - Maps that contain a dynamic type
//   - Sets that contain a dynamic type
func CollectionTypeContainsDynamic(typ attr.Type) bool {
	switch attrType := typ.(type) {
	// Found a dynamic!
	case attr.TypeWithDynamicValue:
		return true
	// Lists, maps, sets
	case attr.TypeWithElementType:
		return CollectionTypeContainsDynamic(attrType.ElementType())
	// Tuples
	case attr.TypeWithElementTypes:
		for _, elemType := range attrType.ElementTypes() {
			hasDynamic := CollectionTypeContainsDynamic(elemType)
			if hasDynamic {
				return true
			}
		}
		return false
	// Objects
	case attr.TypeWithAttributeTypes:
		for _, objAttrType := range attrType.AttributeTypes() {
			hasDynamic := CollectionTypeContainsDynamic(objAttrType)
			if hasDynamic {
				return true
			}
		}
		return false
	// Missing or unsupported type
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
