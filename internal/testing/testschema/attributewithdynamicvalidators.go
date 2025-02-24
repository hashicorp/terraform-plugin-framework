// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ fwxschema.AttributeWithDynamicValidators = AttributeWithDynamicValidators{}

type AttributeWithDynamicValidators struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	WriteOnly           bool
	RequiredForImport   bool
	OptionalForImport   bool
	Validators          []validator.Dynamic
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithDynamicValidators)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) GetType() attr.Type {
	return types.DynamicType
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) IsSensitive() bool {
	return a.Sensitive
}

// DynamicValidators satisfies the fwxschema.AttributeWithDynamicValidators interface.
func (a AttributeWithDynamicValidators) DynamicValidators() []validator.Dynamic {
	return a.Validators
}

// IsWriteOnly satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) IsWriteOnly() bool {
	return a.WriteOnly
}

// IsRequiredForImport satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) IsRequiredForImport() bool {
	return a.RequiredForImport
}

// IsOptionalForImport satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicValidators) IsOptionalForImport() bool {
	return a.OptionalForImport
}
