// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ fwxschema.AttributeWithListValidators = AttributeWithListValidators{}

type AttributeWithListValidators struct {
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
	Validators          []validator.List
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithListValidators)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) GetType() attr.Type {
	return types.ListType{
		ElemType: a.ElementType,
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) IsSensitive() bool {
	return a.Sensitive
}

// IsWriteOnly satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) IsWriteOnly() bool {
	return a.WriteOnly
}

// ListValidators satisfies the fwxschema.AttributeWithListValidators interface.
func (a AttributeWithListValidators) ListValidators() []validator.List {
	return a.Validators
}

// IsRequiredForImport satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) IsRequiredForImport() bool {
	return a.RequiredForImport
}

// IsOptionalForImport satisfies the fwschema.Attribute interface.
func (a AttributeWithListValidators) IsOptionalForImport() bool {
	return a.OptionalForImport
}
