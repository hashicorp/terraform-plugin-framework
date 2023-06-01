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

var _ fwschema.AttributeWithListDefaultValue = AttributeWithListDefaultValue{}

type AttributeWithListDefaultValue struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	ElementType         attr.Type
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	Default             defaults.List
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithListDefaultValue) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// ListDefaultValue satisfies the fwschema.AttributeWithInt64DefaultValue interface.
func (a AttributeWithListDefaultValue) ListDefaultValue() defaults.List {
	return a.Default
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithListDefaultValue) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithListDefaultValue)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithListDefaultValue) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithListDefaultValue) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithListDefaultValue) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithListDefaultValue) GetType() attr.Type {
	return types.ListType{
		ElemType: a.ElementType,
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithListDefaultValue) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithListDefaultValue) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithListDefaultValue) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithListDefaultValue) IsSensitive() bool {
	return a.Sensitive
}
