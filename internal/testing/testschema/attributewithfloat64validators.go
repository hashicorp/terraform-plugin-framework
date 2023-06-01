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

var _ fwxschema.AttributeWithFloat64Validators = AttributeWithFloat64Validators{}

type AttributeWithFloat64Validators struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	Validators          []validator.Float64
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64Validators) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64Validators) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithFloat64Validators)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// Float64Validators satisfies the fwxschema.AttributeWithFloat64Validators interface.
func (a AttributeWithFloat64Validators) Float64Validators() []validator.Float64 {
	return a.Validators
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64Validators) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64Validators) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64Validators) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64Validators) GetType() attr.Type {
	return types.Float64Type
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64Validators) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64Validators) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64Validators) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64Validators) IsSensitive() bool {
	return a.Sensitive
}
