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

var _ fwxschema.AttributeWithNumberPlanModifiers = AttributeWithNumberPlanModifiers{}

type AttributeWithNumberPlanModifiers struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	PlanModifiers       []planmodifier.Number
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithNumberPlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithNumberPlanModifiers) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithNumberPlanModifiers)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithNumberPlanModifiers) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithNumberPlanModifiers) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithNumberPlanModifiers) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithNumberPlanModifiers) GetType() attr.Type {
	return types.NumberType
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithNumberPlanModifiers) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithNumberPlanModifiers) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithNumberPlanModifiers) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithNumberPlanModifiers) IsSensitive() bool {
	return a.Sensitive
}

// NumberPlanModifiers satisfies the fwxschema.AttributeWithNumberPlanModifiers interface.
func (a AttributeWithNumberPlanModifiers) NumberPlanModifiers() []planmodifier.Number {
	return a.PlanModifiers
}
