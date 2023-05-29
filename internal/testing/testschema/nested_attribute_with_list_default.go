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
	_ fwschema.NestedAttribute               = NestedAttributeWithListDefaultValue{}
	_ fwschema.AttributeWithListDefaultValue = NestedAttributeWithListDefaultValue{}
)

type NestedAttributeWithListDefaultValue struct {
	Computed            bool
	Default             defaults.List
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
func (a NestedAttributeWithListDefaultValue) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListDefaultValue) Equal(o fwschema.Attribute) bool {
	_, ok := o.(NestedAttributeWithListDefaultValue)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListDefaultValue) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListDefaultValue) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListDefaultValue) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithListDefaultValue) GetNestedObject() fwschema.NestedAttributeObject {
	return a.NestedObject
}

// GetNestingMode satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithListDefaultValue) GetNestingMode() fwschema.NestingMode {
	return fwschema.NestingModeList
}

// GetType satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListDefaultValue) GetType() attr.Type {
	if a.Type != nil {
		return a.Type
	}

	return types.ListType{
		ElemType: a.GetNestedObject().Type(),
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListDefaultValue) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListDefaultValue) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListDefaultValue) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListDefaultValue) IsSensitive() bool {
	return a.Sensitive
}

// ListDefaultValue satisfies the fwschema.AttributeWithListDefaultValue interface.
func (a NestedAttributeWithListDefaultValue) ListDefaultValue() defaults.List {
	return a.Default
}
