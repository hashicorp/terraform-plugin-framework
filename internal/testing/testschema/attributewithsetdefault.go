// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ fwschema.AttributeWithSetDefaultValue = AttributeWithSetDefaultValue{}

type AttributeWithSetDefaultValue struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	ElementType         attr.Type
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	WriteOnly           bool
	RequiredForImport   bool
	OptionalForImport   bool
	Default             defaults.Set
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// SetDefaultValue satisfies the fwschema.AttributeWithSetDefaultValue interface.
func (a AttributeWithSetDefaultValue) SetDefaultValue() defaults.Set {
	return a.Default
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithSetDefaultValue)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) GetType() attr.Type {
	return types.SetType{
		ElemType: a.ElementType,
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) IsSensitive() bool {
	return a.Sensitive
}

// IsWriteOnly satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) IsWriteOnly() bool {
	return a.WriteOnly
}

// IsRequiredForImport satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) IsRequiredForImport() bool {
	return a.RequiredForImport
}

// IsOptionalForImport satisfies the fwschema.Attribute interface.
func (a AttributeWithSetDefaultValue) IsOptionalForImport() bool {
	return a.OptionalForImport
}
