// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure the implementation satisifies the desired interfaces.
var _ fwschema.NestedBlockObject = NestedBlockObject{}

type NestedBlockObject struct {
	Attributes map[string]fwschema.Attribute
	Blocks     map[string]fwschema.Block
}

// ApplyTerraform5AttributePathStep performs an AttributeName step on the
// underlying attributes or returns an error.
func (o NestedBlockObject) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return fwschema.NestedBlockObjectApplyTerraform5AttributePathStep(o, step)
}

// Equal returns true if the given NestedBlockObject is equivalent.
func (o NestedBlockObject) Equal(other fwschema.NestedBlockObject) bool {
	if _, ok := other.(NestedBlockObject); !ok {
		return false
	}

	return fwschema.NestedBlockObjectEqual(o, other)
}

// GetAttributes returns the Attributes field value.
func (o NestedBlockObject) GetAttributes() fwschema.UnderlyingAttributes {
	return o.Attributes
}

// GetAttributes returns the Blocks field value.
func (o NestedBlockObject) GetBlocks() map[string]fwschema.Block {
	return o.Blocks
}

// Type returns the framework type of the NestedBlockObject.
func (o NestedBlockObject) Type() basetypes.ObjectTypable {
	return fwschema.NestedBlockObjectType(o)
}
