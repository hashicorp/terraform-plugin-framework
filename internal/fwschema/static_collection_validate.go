// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// ValidateStaticCollectionType will return diagnostics if an attr.Type is a collection type that contains
// any dynamic types, which are not supported by the framework type system.
//
// Unsupported collection types include:
//   - Lists that contain a dynamic type
//   - Maps that contain a dynamic type
//   - Sets that contain a dynamic type
func ValidateStaticCollectionType(attrPath path.Path, typ attr.Type) diag.Diagnostic {
	switch attrType := typ.(type) {
	case attr.TypeWithDynamicValue:
		return collectionWithDynamicTypeDiag(attrPath)
	// Lists, maps, sets
	case attr.TypeWithElementType:
		return ValidateStaticCollectionType(attrPath, attrType.ElementType())
	// Tuples
	case attr.TypeWithElementTypes:
		for _, elemType := range attrType.ElementTypes() {
			diag := ValidateStaticCollectionType(attrPath, elemType)
			if diag != nil {
				return diag
			}
		}
		return nil
	// Objects
	case attr.TypeWithAttributeTypes:
		for _, objAttrType := range attrType.AttributeTypes() {
			diag := ValidateStaticCollectionType(attrPath, objAttrType)
			if diag != nil {
				return diag
			}
		}
		return nil
	// Missing or unsupported type
	default:
		return nil
	}
}

func collectionWithDynamicTypeDiag(attributePath path.Path) diag.Diagnostic {
	// The diagnostic path is intentionally omitted as it is invalid in this
	// context. Diagnostic paths are intended to be mapped to actual data,
	// while this path information must be synthesized.
	return diag.NewErrorDiagnostic(
		"Invalid Schema Implementation",
		"When validating the schema, an implementation issue was found. "+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("%q is a collection type that contains a dynamic type. ", attributePath)+
			"Dynamic types inside of collections are not currently supported in terraform-plugin-framework.",
	)
}
