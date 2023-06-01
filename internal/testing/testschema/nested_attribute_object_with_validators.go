// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure the implementation satisifies the desired interfaces.
var _ fwxschema.NestedAttributeObjectWithValidators = NestedAttributeObjectWithValidators{}

type NestedAttributeObjectWithValidators struct {
	Attributes map[string]fwschema.Attribute
	Validators []validator.Object
}

// ApplyTerraform5AttributePathStep performs an AttributeName step on the
// underlying attributes or returns an error.
func (o NestedAttributeObjectWithValidators) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	name, ok := step.(tftypes.AttributeName)

	if !ok {
		return nil, fmt.Errorf("cannot apply AttributePathStep %T to NestedAttributeObjectWithValidators", step)
	}

	attribute, ok := o.GetAttributes()[string(name)]

	if ok {
		return attribute, nil
	}

	return nil, fmt.Errorf("no attribute %q on NestedAttributeObjectWithValidators", name)

}

// Equal returns true if the given NestedAttributeObjectWithValidators is equivalent.
func (o NestedAttributeObjectWithValidators) Equal(other fwschema.NestedAttributeObject) bool {
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
func (o NestedAttributeObjectWithValidators) GetAttributes() fwschema.UnderlyingAttributes {
	return o.Attributes
}

// ObjectValidators returns the Validators field value.
func (o NestedAttributeObjectWithValidators) ObjectValidators() []validator.Object {
	return o.Validators
}

// Type returns the framework type of the NestedAttributeObjectWithValidators.
func (o NestedAttributeObjectWithValidators) Type() basetypes.ObjectTypable {
	attrTypes := make(map[string]attr.Type, len(o.Attributes))

	for name, attribute := range o.Attributes {
		attrTypes[name] = attribute.GetType()
	}

	return types.ObjectType{
		AttrTypes: attrTypes,
	}
}
