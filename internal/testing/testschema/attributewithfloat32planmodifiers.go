// Copyright IBM Corp. 2021, 2026
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

var _ fwxschema.AttributeWithFloat32PlanModifiers = AttributeWithFloat32PlanModifiers{}

type AttributeWithFloat32PlanModifiers struct {
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
	PlanModifiers       []planmodifier.Float32
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithFloat32PlanModifiers)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// Float32PlanModifiers satisfies the fwxschema.AttributeWithFloat32PlanModifiers interface.
func (a AttributeWithFloat32PlanModifiers) Float32PlanModifiers() []planmodifier.Float32 {
	return a.PlanModifiers
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) GetType() attr.Type {
	return types.Float32Type
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) IsSensitive() bool {
	return a.Sensitive
}

// IsWriteOnly satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) IsWriteOnly() bool {
	return a.WriteOnly
}

// IsRequiredForImport satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) IsRequiredForImport() bool {
	return a.RequiredForImport
}

// IsOptionalForImport satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat32PlanModifiers) IsOptionalForImport() bool {
	return a.OptionalForImport
}
