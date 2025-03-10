// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ fwxschema.AttributeWithDynamicPlanModifiers = AttributeWithDynamicPlanModifiers{}

type AttributeWithDynamicPlanModifiers struct {
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
	PlanModifiers       []planmodifier.Dynamic
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithDynamicPlanModifiers)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) GetType() attr.Type {
	return types.DynamicType
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) IsSensitive() bool {
	return a.Sensitive
}

// DynamicPlanModifiers satisfies the fwxschema.AttributeWithDynamicPlanModifiers interface.
func (a AttributeWithDynamicPlanModifiers) DynamicPlanModifiers() []planmodifier.Dynamic {
	return a.PlanModifiers
}

// IsWriteOnly satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) IsWriteOnly() bool {
	return a.WriteOnly
}

// IsRequiredForImport satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) IsRequiredForImport() bool {
	return a.RequiredForImport
}

// IsOptionalForImport satisfies the fwschema.Attribute interface.
func (a AttributeWithDynamicPlanModifiers) IsOptionalForImport() bool {
	return a.OptionalForImport
}
