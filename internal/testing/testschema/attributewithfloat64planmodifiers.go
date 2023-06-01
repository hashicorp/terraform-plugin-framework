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

var _ fwxschema.AttributeWithFloat64PlanModifiers = AttributeWithFloat64PlanModifiers{}

type AttributeWithFloat64PlanModifiers struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Optional            bool
	Required            bool
	Sensitive           bool
	PlanModifiers       []planmodifier.Float64
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64PlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64PlanModifiers) Equal(o fwschema.Attribute) bool {
	_, ok := o.(AttributeWithFloat64PlanModifiers)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// Float64PlanModifiers satisfies the fwxschema.AttributeWithFloat64PlanModifiers interface.
func (a AttributeWithFloat64PlanModifiers) Float64PlanModifiers() []planmodifier.Float64 {
	return a.PlanModifiers
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64PlanModifiers) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64PlanModifiers) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64PlanModifiers) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetType satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64PlanModifiers) GetType() attr.Type {
	return types.Float64Type
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64PlanModifiers) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64PlanModifiers) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64PlanModifiers) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a AttributeWithFloat64PlanModifiers) IsSensitive() bool {
	return a.Sensitive
}
