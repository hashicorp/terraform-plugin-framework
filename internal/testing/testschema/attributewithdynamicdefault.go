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

var _ fwschema.AttributeWithDynamicDefaultValue = AttributeWithDynamicDefaultValue{}

type AttributeWithDynamicDefaultValue struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	Default             defaults.Dynamic
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicDefaultValue) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// DynamicDefaultValue satisfies the fwxschema.AttributeWithDynamicDefaultValue interface.
func (a AttributeWithDynamicDefaultValue) DynamicDefaultValue() defaults.Dynamic {
	return a.Default
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicDefaultValue) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithDynamicDefaultValue)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicDefaultValue) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicDefaultValue) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicDefaultValue) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicDefaultValue) GetType() attr.Type {
	return types.DynamicType
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicDefaultValue) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicDefaultValue) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicDefaultValue) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicDefaultValue) IsSensitive() bool {
	return a.Sensitive
}
