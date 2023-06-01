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
	_ fwschema.NestedAttribute                 = NestedAttributeWithListPlanModifiers{}
	_ fwxschema.AttributeWithListPlanModifiers = NestedAttributeWithListPlanModifiers{}
)

type NestedAttributeWithListPlanModifiers struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	NestedObject        fwschema.NestedAttributeObject
	Optional            bool
	PlanModifiers       []planmodifier.List
	Required            bool
	Sensitive           bool
	Type                attr.Type
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListPlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListPlanModifiers) Equal(o fwschema.Attribute) bool {
	_, ok := o.(NestedAttributeWithListPlanModifiers)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListPlanModifiers) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListPlanModifiers) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListPlanModifiers) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithListPlanModifiers) GetNestedObject() fwschema.NestedAttributeObject {
	return a.NestedObject
}

// GetNestingMode satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithListPlanModifiers) GetNestingMode() fwschema.NestingMode {
	return fwschema.NestingModeList
}

// GetType satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListPlanModifiers) GetType() attr.Type {
	if a.Type != nil {
		return a.Type
	}

	return types.ListType{
		ElemType: a.GetNestedObject().Type(),
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListPlanModifiers) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListPlanModifiers) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListPlanModifiers) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithListPlanModifiers) IsSensitive() bool {
	return a.Sensitive
}

// ListPlanModifiers satisfies the fwxschema.AttributeWithListPlanModifiers interface.
func (a NestedAttributeWithListPlanModifiers) ListPlanModifiers() []planmodifier.List {
	return a.PlanModifiers
}
