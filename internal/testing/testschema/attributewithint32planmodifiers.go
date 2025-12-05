// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ fwxschema.AttributeWithInt32PlanModifiers = AttributeWithInt32PlanModifiers{}

type AttributeWithInt32PlanModifiers struct {
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
	PlanModifiers       []planmodifier.Int32
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithInt32PlanModifiers)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) GetType() attr.Type {
	return types.Int32Type
}

// Int32PlanModifiers satisfies the fwxschema.AttributeWithInt32PlanModifiers interface.
func (a AttributeWithInt32PlanModifiers) Int32PlanModifiers() []planmodifier.Int32 {
	return a.PlanModifiers
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) IsSensitive() bool {
	return a.Sensitive
}

// IsWriteOnly satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) IsWriteOnly() bool {
	return a.WriteOnly
}

// IsRequiredForImport satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) IsRequiredForImport() bool {
	return a.RequiredForImport
}

// IsOptionalForImport satisfies the fwschema.Attribute interface.
func (a AttributeWithInt32PlanModifiers) IsOptionalForImport() bool {
	return a.OptionalForImport
}
