// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwschema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

// ContainsAllWriteOnlyChildAttributes will return true if all child attributes for the
// given nested attribute have WriteOnly set to true.
func ContainsAllWriteOnlyChildAttributes(nestedAttr NestedAttribute) bool {
	nestedObjAttrs := nestedAttr.GetNestedObject().GetAttributes()

	for _, childAttr := range nestedObjAttrs {
		if !childAttr.IsWriteOnly() {
			return false
		}

		nestedAttribute, ok := childAttr.(NestedAttribute)
		if ok {
			if !ContainsAllWriteOnlyChildAttributes(nestedAttribute) {
				return false
			}
		}
	}

	return true
}

// ContainsAnyWriteOnlyChildAttributes will return true if any child attribute for the
// given nested attribute has WriteOnly set to true.
func ContainsAnyWriteOnlyChildAttributes(nestedAttr NestedAttribute) bool {
	nestedObjAttrs := nestedAttr.GetNestedObject().GetAttributes()

	for _, childAttr := range nestedObjAttrs {
		if childAttr.IsWriteOnly() {
			return true
		}

		nestedAttribute, ok := childAttr.(NestedAttribute)
		if ok {
			if ContainsAnyWriteOnlyChildAttributes(nestedAttribute) {
				return true
			}
		}
	}

	return false
}

func InvalidWriteOnlyNestedAttributeDiag(attributePath path.Path) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Invalid Schema Implementation",
		"When validating the schema, an implementation issue was found. "+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("%q is a WriteOnly nested attribute that contains a non-WriteOnly child attribute.\n\n", attributePath)+
			"Every child attribute of a WriteOnly nested attribute must also have WriteOnly set to true.",
	)
}

func InvalidComputedNestedAttributeWithWriteOnlyDiag(attributePath path.Path) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		"Invalid Schema Implementation",
		"When validating the schema, an implementation issue was found. "+
			"This is always an issue with the provider and should be reported to the provider developers.\n\n"+
			fmt.Sprintf("%q is a Computed nested attribute that contains a WriteOnly child attribute.\n\n", attributePath)+
			"Every child attribute of a Computed nested attribute must have WriteOnly set to false.",
	)
}
