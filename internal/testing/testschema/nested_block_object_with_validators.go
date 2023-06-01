// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Ensure the implementation satisifies the desired interfaces.
var _ fwxschema.NestedBlockObjectWithValidators = NestedBlockObjectWithValidators{}

type NestedBlockObjectWithValidators struct {
	Attributes map[string]fwschema.Attribute
	Blocks     map[string]fwschema.Block
	Validators []validator.Object
}

// ApplyTerraform5AttributePathStep performs an AttributeName step on the
// underlying attributes or returns an error.
func (o NestedBlockObjectWithValidators) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return fwschema.NestedBlockObjectApplyTerraform5AttributePathStep(o, step)
}

// Equal returns true if the given NestedBlockObjectWithValidators is equivalent.
func (o NestedBlockObjectWithValidators) Equal(other fwschema.NestedBlockObject) bool {
	if _, ok := other.(NestedBlockObjectWithValidators); !ok {
		return false
	}

	return fwschema.NestedBlockObjectEqual(o, other)
}

// GetAttributes returns the Attributes field value.
func (o NestedBlockObjectWithValidators) GetAttributes() fwschema.UnderlyingAttributes {
	return o.Attributes
}

// GetAttributes returns the Blocks field value.
func (o NestedBlockObjectWithValidators) GetBlocks() map[string]fwschema.Block {
	return o.Blocks
}

// ObjectValidators returns the Validators field value.
func (o NestedBlockObjectWithValidators) ObjectValidators() []validator.Object {
	return o.Validators
}

// Type returns the framework type of the NestedBlockObjectWithValidators.
func (o NestedBlockObjectWithValidators) Type() basetypes.ObjectTypable {
	return fwschema.NestedBlockObjectType(o)
}
