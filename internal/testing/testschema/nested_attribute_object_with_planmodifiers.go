// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure the implementation satisifies the desired interfaces.
var _ fwxschema.NestedAttributeObjectWithPlanModifiers = NestedAttributeObjectWithPlanModifiers{}

type NestedAttributeObjectWithPlanModifiers struct {
	Attributes    map[string]fwschema.Attribute
	PlanModifiers []planmodifier.Object
}

// ApplyTerraform5AttributePathStep performs an AttributeName step on the
// underlying attributes or returns an error.
func (o NestedAttributeObjectWithPlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	name, ok := step.(tftypes.AttributeName)

	if !ok {
		return nil, fmt.Errorf("cannot apply AttributePathStep %T to NestedAttributeObjectWithPlanModifiers", step)
	}

	attribute, ok := o.GetAttributes()[string(name)]

	if ok {
		return attribute, nil
	}

	return nil, fmt.Errorf("no attribute %q on NestedAttributeObjectWithPlanModifiers", name)

}

// Equal returns true if the given NestedAttributeObjectWithPlanModifiers is equivalent.
func (o NestedAttributeObjectWithPlanModifiers) Equal(other fwschema.NestedAttributeObject) bool {
	if !o.Type().Equal(other.Type()) {
		return false
	}

	if len(o.GetAttributes()) != len(other.GetAttributes()) {
		return false
	}

	for name, oAttribute := range o.GetAttributes() {
		otherAttribute, ok := other.GetAttributes()[name]

		if !ok {
			return false
		}

		if !oAttribute.Equal(otherAttribute) {
			return false
		}
	}

	return true
}

// GetAttributes returns the Attributes field value.
func (o NestedAttributeObjectWithPlanModifiers) GetAttributes() fwschema.UnderlyingAttributes {
	return o.Attributes
}

// ObjectPlanModifiers returns the PlanModifiers field value.
func (o NestedAttributeObjectWithPlanModifiers) ObjectPlanModifiers() []planmodifier.Object {
	return o.PlanModifiers
}

// Type returns the framework type of the NestedAttributeObjectWithPlanModifiers.
func (o NestedAttributeObjectWithPlanModifiers) Type() basetypes.ObjectTypable {
	attrTypes := make(map[string]attr.Type, len(o.Attributes))

	for name, attribute := range o.Attributes {
		attrTypes[name] = attribute.GetType()
	}

	return types.ObjectType{
		AttrTypes: attrTypes,
	}
}
