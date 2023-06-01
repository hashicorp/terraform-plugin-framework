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

var _ fwxschema.AttributeWithMapValidators = AttributeWithMapValidators{}

type AttributeWithMapValidators struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	ElementType         attr.Type
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	Validators          []validator.Map
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithMapValidators) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithMapValidators) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithMapValidators)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithMapValidators) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithMapValidators) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithMapValidators) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithMapValidators) GetType() attr.Type {
	return types.MapType{
		ElemType: a.ElementType,
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithMapValidators) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithMapValidators) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithMapValidators) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithMapValidators) IsSensitive() bool {
	return a.Sensitive
}

// MapValidators satisfies the fwxschema.AttributeWithMapValidators interface.
func (a AttributeWithMapValidators) MapValidators() []validator.Map {
	return a.Validators
}
