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

var _ fwxschema.AttributeWithListPlanModifiers = AttributeWithListPlanModifiers{}

type AttributeWithListPlanModifiers struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	ElementType         attr.Type
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	PlanModifiers       []planmodifier.List
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithListPlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithListPlanModifiers) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithListPlanModifiers)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithListPlanModifiers) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithListPlanModifiers) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithListPlanModifiers) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithListPlanModifiers) GetType() attr.Type {
	return types.ListType{
		ElemType: a.ElementType,
	}
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithListPlanModifiers) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithListPlanModifiers) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithListPlanModifiers) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithListPlanModifiers) IsSensitive() bool {
	return a.Sensitive
}

// ListPlanModifiers satisfies the fwxschema.AttributeWithListPlanModifiers interface.
func (a AttributeWithListPlanModifiers) ListPlanModifiers() []planmodifier.List {
	return a.PlanModifiers
}
