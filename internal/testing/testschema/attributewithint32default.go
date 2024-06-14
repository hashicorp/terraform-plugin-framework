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

var _ fwschema.AttributeWithInt32DefaultValue = AttributeWithInt32DefaultValue{}

type AttributeWithInt32DefaultValue struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	Default             defaults.Int32
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32DefaultValue) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Int32DefaultValue satisfies the fwxschema.AttributeWithInt32DefaultValue interface.
func (a AttributeWithInt32DefaultValue) Int32DefaultValue() defaults.Int32 {
	return a.Default
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32DefaultValue) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithInt32DefaultValue)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32DefaultValue) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32DefaultValue) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32DefaultValue) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32DefaultValue) GetType() attr.Type {
	return types.Int32Type
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32DefaultValue) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32DefaultValue) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32DefaultValue) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32DefaultValue) IsSensitive() bool {
	return a.Sensitive
}
