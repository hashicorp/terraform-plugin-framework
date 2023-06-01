// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ fwschema.NestedAttribute                 = NestedAttributeWithObjectDefaultValue{}
	_ fwschema.AttributeWithObjectDefaultValue = NestedAttributeWithObjectDefaultValue{}
)

type NestedAttributeWithObjectDefaultValue struct {
	Attributes          map[string]schema.Attribute
	Computed            bool
	Default             defaults.Object
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	Type                attr.Type
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectDefaultValue) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectDefaultValue) Equal(o fwschema.Attribute) bool {
	_, ok := o.(NestedAttributeWithObjectDefaultValue)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectDefaultValue) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectDefaultValue) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectDefaultValue) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithObjectDefaultValue) GetNestedObject() fwschema.NestedAttributeObject {
	return nil
}

// GetNestingMode satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithObjectDefaultValue) GetNestingMode() fwschema.NestingMode {
	return fwschema.NestingModeSingle
}

// GetType satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectDefaultValue) GetType() attr.Type {
	attrTypes := make(map[string]attr.Type, len(a.Attributes))

	for name, attribute := range a.Attributes {
		attrTypes[name] = attribute.GetType()
	}

	return types.ObjectType{
		AttrTypes: attrTypes,
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectDefaultValue) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectDefaultValue) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectDefaultValue) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectDefaultValue) IsSensitive() bool {
	return a.Sensitive
}

// ObjectDefaultValue satisfies the fwschema.AttributeWithListDefaultValue interface.
func (a NestedAttributeWithObjectDefaultValue) ObjectDefaultValue() defaults.Object {
	return a.Default
}
