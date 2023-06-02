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

var _ fwschema.AttributeWithObjectDefaultValue = AttributeWithObjectDefaultValue{}

type AttributeWithObjectDefaultValue struct {
	AttributeTypes      map[string]attr.Type
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	Default             defaults.Object
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithObjectDefaultValue) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// ObjectDefaultValue satisfies the fwschema.AttributeWithObjectDefaultValue interface.
func (a AttributeWithObjectDefaultValue) ObjectDefaultValue() defaults.Object {
	return a.Default
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithObjectDefaultValue) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithObjectDefaultValue)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithObjectDefaultValue) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithObjectDefaultValue) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithObjectDefaultValue) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithObjectDefaultValue) GetType() attr.Type {
	return types.ObjectType{
		AttrTypes: a.AttributeTypes,
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithObjectDefaultValue) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithObjectDefaultValue) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithObjectDefaultValue) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithObjectDefaultValue) IsSensitive() bool {
	return a.Sensitive
}
