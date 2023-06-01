// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure the implementation satisifies the desired interfaces.
var _ fwxschema.NestedBlockObjectWithPlanModifiers = NestedBlockObjectWithPlanModifiers{}

type NestedBlockObjectWithPlanModifiers struct {
	Attributes    map[string]fwschema.Attribute
	Blocks        map[string]fwschema.Block
	PlanModifiers []planmodifier.Object
}

// ApplyTerraform5AttributePathStep performs an AttributeName step on the
// underlying attributes or returns an error.
func (o NestedBlockObjectWithPlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return fwschema.NestedBlockObjectApplyTerraform5AttributePathStep(o, step)
}

// Equal returns true if the given NestedBlockObjectWithPlanModifiers is equivalent.
func (o NestedBlockObjectWithPlanModifiers) Equal(other fwschema.NestedBlockObject) bool {
	if _, ok := other.(NestedBlockObjectWithPlanModifiers); !ok {
		return false
	}

	return fwschema.NestedBlockObjectEqual(o, other)
}

// GetAttributes returns the Attributes field value.
func (o NestedBlockObjectWithPlanModifiers) GetAttributes() fwschema.UnderlyingAttributes {
	return o.Attributes
}

// GetAttributes returns the Blocks field value.
func (o NestedBlockObjectWithPlanModifiers) GetBlocks() map[string]fwschema.Block {
	return o.Blocks
}

// ObjectPlanModifiers returns the PlanModifiers field value.
func (o NestedBlockObjectWithPlanModifiers) ObjectPlanModifiers() []planmodifier.Object {
	return o.PlanModifiers
}

// Type returns the framework type of the NestedBlockObjectWithPlanModifiers.
func (o NestedBlockObjectWithPlanModifiers) Type() basetypes.ObjectTypable {
	return fwschema.NestedBlockObjectType(o)
}
