// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ fwschema.NestedAttribute              = NestedAttributeWithSetDefaultValue{}
	_ fwschema.AttributeWithSetDefaultValue = NestedAttributeWithSetDefaultValue{}
)

type NestedAttributeWithSetDefaultValue struct {
	Computed            bool
	Default             defaults.Set
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	NestedObject        fwschema.NestedAttributeObject
	Optional            bool
	Required            bool
	Sensitive           bool
	Type                attr.Type
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetDefaultValue) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetDefaultValue) Equal(o fwschema.Attribute) bool {
	_, ok := o.(NestedAttributeWithSetDefaultValue)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetDefaultValue) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetDefaultValue) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetDefaultValue) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithSetDefaultValue) GetNestedObject() fwschema.NestedAttributeObject {
	return a.NestedObject
}

// GetNestingMode satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithSetDefaultValue) GetNestingMode() fwschema.NestingMode {
	return fwschema.NestingModeSet
}

// GetType satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetDefaultValue) GetType() attr.Type {
	if a.Type != nil {
		return a.Type
	}

	return types.SetType{
		ElemType: a.GetNestedObject().Type(),
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetDefaultValue) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetDefaultValue) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetDefaultValue) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetDefaultValue) IsSensitive() bool {
	return a.Sensitive
}

// MapDefaultValue satisfies the fwschema.AttributeWithMapDefaultValue interface.
func (a NestedAttributeWithSetDefaultValue) SetDefaultValue() defaults.Set {
	return a.Default
}
