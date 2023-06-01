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
	_ fwschema.NestedAttribute                = NestedAttributeWithMapPlanModifiers{}
	_ fwxschema.AttributeWithMapPlanModifiers = NestedAttributeWithMapPlanModifiers{}
)

type NestedAttributeWithMapPlanModifiers struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	NestedObject        fwschema.NestedAttributeObject
	Optional            bool
	PlanModifiers       []planmodifier.Map
	Required            bool
	Sensitive           bool
	Type                attr.Type
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithMapPlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithMapPlanModifiers) Equal(o fwschema.Attribute) bool {
	_, ok := o.(NestedAttributeWithMapPlanModifiers)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithMapPlanModifiers) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithMapPlanModifiers) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithMapPlanModifiers) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithMapPlanModifiers) GetNestedObject() fwschema.NestedAttributeObject {
	return a.NestedObject
}

// GetNestingMode satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithMapPlanModifiers) GetNestingMode() fwschema.NestingMode {
	return fwschema.NestingModeMap
}

// GetType satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithMapPlanModifiers) GetType() attr.Type {
	if a.Type != nil {
		return a.Type
	}

	return types.MapType{
		ElemType: a.GetNestedObject().Type(),
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithMapPlanModifiers) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithMapPlanModifiers) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithMapPlanModifiers) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithMapPlanModifiers) IsSensitive() bool {
	return a.Sensitive
}

// MapPlanModifiers satisfies the fwxschema.AttributeWithMapPlanModifiers interface.
func (a NestedAttributeWithMapPlanModifiers) MapPlanModifiers() []planmodifier.Map {
	return a.PlanModifiers
}
