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

var _ fwxschema.AttributeWithFloat32Validators = AttributeWithFloat32Validators{}

type AttributeWithFloat32Validators struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	Validators          []validator.Float32
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32Validators) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32Validators) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithFloat32Validators)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// Float32Validators satisfies the fwxschema.AttributeWithFloat32Validators interface.
func (a AttributeWithFloat32Validators) Float32Validators() []validator.Float32 {
	return a.Validators
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32Validators) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32Validators) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32Validators) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32Validators) GetType() attr.Type {
	return types.Float32Type
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32Validators) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32Validators) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32Validators) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32Validators) IsSensitive() bool {
	return a.Sensitive
}
