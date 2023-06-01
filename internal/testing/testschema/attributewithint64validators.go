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

var _ fwxschema.AttributeWithInt64Validators = AttributeWithInt64Validators{}

type AttributeWithInt64Validators struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	Validators          []validator.Int64
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithInt64Validators) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithInt64Validators) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithInt64Validators)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithInt64Validators) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithInt64Validators) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithInt64Validators) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithInt64Validators) GetType() attr.Type {
	return types.Int64Type
}

// Int64Validators satisfies the fwxschema.AttributeWithInt64Validators interface.
func (a AttributeWithInt64Validators) Int64Validators() []validator.Int64 {
	return a.Validators
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithInt64Validators) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithInt64Validators) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithInt64Validators) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithInt64Validators) IsSensitive() bool {
	return a.Sensitive
}
