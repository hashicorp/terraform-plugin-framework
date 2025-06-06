// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package metaschema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisifies the desired interfaces.
var (
	_ NestedAttribute = MapNestedAttribute{}
)

// MapNestedAttribute represents an attribute that is a map of objects where
// the object attributes can be fully defined, including further nested
// attributes. When retrieving the value for this attribute, use types.Map
// as the value type unless the CustomType field is set. The NestedObject field
// must be set. Nested attributes are only compatible with protocol version 6.
//
// Use MapAttribute if the underlying elements are of a single type and do
// not require definition beyond type information.
//
// Terraform configurations configure this attribute using expressions that
// return a map of objects or directly via curly brace syntax.
//
//	# map of objects
//	example_attribute = {
//		key = {
//			nested_attribute = #...
//		},
//	]
//
// Terraform configurations reference this attribute using expressions that
// accept a map of objects or an element directly via square brace string
// syntax:
//
//	# known object at key
//	.example_attribute["key"]
//	# known object nested_attribute value at key
//	.example_attribute["key"].nested_attribute
type MapNestedAttribute struct {
	// NestedObject is the underlying object that contains nested attributes.
	// This field must be set.
	NestedObject NestedAttributeObject

	// CustomType enables the use of a custom attribute type in place of the
	// default types.MapType of types.ObjectType. When retrieving data, the
	// basetypes.MapValuable associated with this custom type must be used in
	// place of types.Map.
	CustomType basetypes.MapTypable

	// Required indicates whether the practitioner must enter a value for
	// this attribute or not. Required and Optional cannot both be true,
	// and Required and Computed cannot both be true.
	Required bool

	// Optional indicates whether the practitioner can choose to enter a value
	// for this attribute or not. Optional and Required cannot both be true.
	Optional bool

	// Description is used in various tooling, like the language server, to
	// give practitioners more information about what this attribute is,
	// what it's for, and how it should be used. It should be written as
	// plain text, with no special formatting.
	Description string

	// MarkdownDescription is used in various tooling, like the
	// documentation generator, to give practitioners more information
	// about what this attribute is, what it's for, and how it should be
	// used. It should be formatted using Markdown.
	MarkdownDescription string
}

// ApplyTerraform5AttributePathStep returns the Attributes field value if step
// is ElementKeyString, otherwise returns an error.
func (a MapNestedAttribute) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	_, ok := step.(tftypes.ElementKeyString)

	if !ok {
		return nil, fmt.Errorf("cannot apply step %T to MapNestedAttribute", step)
	}

	return a.NestedObject, nil
}

// Equal returns true if the given Attribute is a MapNestedAttribute
// and all fields are equal.
func (a MapNestedAttribute) Equal(o fwschema.Attribute) bool {
	other, ok := o.(MapNestedAttribute)

	if !ok {
		return false
	}

	return fwschema.NestedAttributesEqual(a, other)
}

// GetDeprecationMessage always returns an empty string as there is no
// deprecation validation support for provider meta schemas.
func (a MapNestedAttribute) GetDeprecationMessage() string {
	return ""
}

// GetDescription returns the Description field value.
func (a MapNestedAttribute) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription returns the MarkdownDescription field value.
func (a MapNestedAttribute) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetNestedObject returns the NestedObject field value.
func (a MapNestedAttribute) GetNestedObject() fwschema.NestedAttributeObject {
	return a.NestedObject
}

// GetNestingMode always returns NestingModeMap.
func (a MapNestedAttribute) GetNestingMode() fwschema.NestingMode {
	return fwschema.NestingModeMap
}

// GetType returns MapType of ObjectType or CustomType.
func (a MapNestedAttribute) GetType() attr.Type {
	if a.CustomType != nil {
		return a.CustomType
	}

	return types.MapType{
		ElemType: a.NestedObject.Type(),
	}
}

// IsComputed always returns false as provider schemas cannot be Computed.
func (a MapNestedAttribute) IsComputed() bool {
	return false
}

// IsOptional returns the Optional field value.
func (a MapNestedAttribute) IsOptional() bool {
	return a.Optional
}

// IsRequired returns the Required field value.
func (a MapNestedAttribute) IsRequired() bool {
	return a.Required
}

// IsSensitive always returns false as there is no plan for provider meta
// schema data.
func (a MapNestedAttribute) IsSensitive() bool {
	return false
}

// IsWriteOnly returns false as write-only attributes are not relevant to provider meta schemas,
// as these schemas describe data explicitly not saved to any artifact.
func (a MapNestedAttribute) IsWriteOnly() bool {
	return false
}

// IsRequiredForImport returns false as this behavior is only relevant
// for managed resource identity schema attributes.
func (a MapNestedAttribute) IsRequiredForImport() bool {
	return false
}

// IsOptionalForImport returns false as this behavior is only relevant
// for managed resource identity schema attributes.
func (a MapNestedAttribute) IsOptionalForImport() bool {
	return false
}
