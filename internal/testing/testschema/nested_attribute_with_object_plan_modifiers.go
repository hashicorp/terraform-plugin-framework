// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ fwschema.NestedAttribute                   = NestedAttributeWithObjectPlanModifiers{}
	_ fwxschema.AttributeWithObjectPlanModifiers = NestedAttributeWithObjectPlanModifiers{}
)

type NestedAttributeWithObjectPlanModifiers struct {
	Computed            bool
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	NestedObject        fwschema.NestedAttributeObject
	Optional            bool
	PlanModifiers       []planmodifier.Object
	Required            bool
	Sensitive           bool
	Type                attr.Type
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectPlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return a.GetType().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectPlanModifiers) Equal(o fwschema.Attribute) bool {
	_, ok := o.(NestedAttributeWithObjectPlanModifiers)

	if !ok {
		return false
	}

	return fwschema.AttributesEqual(a, o)
}

// GetDeprecationMessage satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectPlanModifiers) GetDeprecationMessage() string {
	return a.DeprecationMessage
}

// GetDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectPlanModifiers) GetDescription() string {
	return a.Description
}

// GetMarkdownDescription satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectPlanModifiers) GetMarkdownDescription() string {
	return a.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithObjectPlanModifiers) GetNestedObject() fwschema.NestedAttributeObject {
	return a.NestedObject
}

// GetNestingMode satisfies the fwschema.NestedAttribute interface.
func (a NestedAttributeWithObjectPlanModifiers) GetNestingMode() fwschema.NestingMode {
	return fwschema.NestingModeSingle
}

// GetType satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectPlanModifiers) GetType() attr.Type {
	if a.Type != nil {
		return a.Type
	}

	return a.GetNestedObject().Type()
}

// IsComputed satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectPlanModifiers) IsComputed() bool {
	return a.Computed
}

// IsOptional satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectPlanModifiers) IsOptional() bool {
	return a.Optional
}

// IsRequired satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectPlanModifiers) IsRequired() bool {
	return a.Required
}

// IsSensitive satisfies the fwschema.Attribute interface.
func (a NestedAttributeWithObjectPlanModifiers) IsSensitive() bool {
	return a.Sensitive
}

// ObjectPlanModifiers satisfies the fwxschema.AttributeWithObjectPlanModifiers interface.
func (a NestedAttributeWithObjectPlanModifiers) ObjectPlanModifiers() []planmodifier.Object {
	return a.PlanModifiers
}
