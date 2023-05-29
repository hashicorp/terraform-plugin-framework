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

var _ fwschema.AttributeWithFloat64DefaultValue = AttributeWithFloat64DefaultValue{}

type AttributeWithFloat64DefaultValue struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	Default             defaults.Float64
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64DefaultValue) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Float64DefaultValue satisfies the fwxschema.AttributeWithFloat64DefaultValue interface.
func (a AttributeWithFloat64DefaultValue) Float64DefaultValue() defaults.Float64 {
	return a.Default
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64DefaultValue) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithFloat64DefaultValue)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64DefaultValue) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64DefaultValue) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64DefaultValue) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64DefaultValue) GetType() attr.Type {
	return types.Float64Type
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64DefaultValue) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64DefaultValue) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64DefaultValue) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64DefaultValue) IsSensitive() bool {
	return a.Sensitive
}
