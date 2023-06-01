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

var (
	_ fwschema.NestedAttribute                = NestedAttributeWithSetPlanModifiers{}
	_ fwxschema.AttributeWithSetPlanModifiers = NestedAttributeWithSetPlanModifiers{}
)

type NestedAttributeWithSetPlanModifiers struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	NestedObject        fwschema.NestedAttributeObject
	Optional            bool
	PlanModifiers       []planmodifier.Set
	Required            bool
	Sensitive           bool
	Type                attr.Type
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetPlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetPlanModifiers) Equal(o fwschema.Attribute) bool {
	_, ok := o.(NestedAttributeWithSetPlanModifiers)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetPlanModifiers) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetPlanModifiers) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetPlanModifiers) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithSetPlanModifiers) GetNestedObject() fwschema.NestedAttributeObject {
	return a.NestedObject
}

// GetNestingMode satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithSetPlanModifiers) GetNestingMode() fwschema.NestingMode {
	return fwschema.NestingModeSet
}

// GetType satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetPlanModifiers) GetType() attr.Type {
	if a.Type != nil {
		return a.Type
	}

	return types.SetType{
		ElemType: a.GetNestedObject().Type(),
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetPlanModifiers) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetPlanModifiers) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetPlanModifiers) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithSetPlanModifiers) IsSensitive() bool {
	return a.Sensitive
}

// SetPlanModifiers satisfies the fwxschema.AttributeWithSetPlanModifiers interface.
func (a NestedAttributeWithSetPlanModifiers) SetPlanModifiers() []planmodifier.Set {
	return a.PlanModifiers
}
