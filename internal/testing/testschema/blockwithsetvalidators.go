// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ fwxschema.BlockWithSetValidators = BlockWithSetValidators{}

type BlockWithSetValidators struct {
	Attributes          map[string]fwschema.Attribute
	Blocks              map[string]fwschema.Block
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Validators          []validator.Set
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Block interface.
func (b BlockWithSetValidators) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return b.Type().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Block interface.
func (b BlockWithSetValidators) Equal(o fwschema.Block) bool {
	_, ok := o.(BlockWithSetValidators)

	if !ok {
		return false
	}

	return fwschema.BlocksEqual(b, o)
}

// GetDeprecationMessage satisfies the fwschema.Block interface.
func (b BlockWithSetValidators) GetDeprecationMessage() string {
	return b.DeprecationMessage
}

// GetDescription satisfies the fwschema.Block interface.
func (b BlockWithSetValidators) GetDescription() string {
	return b.Description
}

// GetMarkdownDescription satisfies the fwschema.Block interface.
func (b BlockWithSetValidators) GetMarkdownDescription() string {
	return b.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.Block interface.
func (b BlockWithSetValidators) GetNestedObject() fwschema.NestedBlockObject {
	return NestedBlockObject{
		Attributes: b.Attributes,
		Blocks:     b.Blocks,
	}
}

// GetNestingMode satisfies the fwschema.Block interface.
func (b BlockWithSetValidators) GetNestingMode() fwschema.BlockNestingMode {
	return fwschema.BlockNestingModeSet
}

// SetValidators satisfies the fwxschema.BlockWithSetValidators interface.
func (b BlockWithSetValidators) SetValidators() []validator.Set {
	return b.Validators
}

// Type satisfies the fwschema.Block interface.
func (b BlockWithSetValidators) Type() attr.Type {
	return types.SetType{
		ElemType: b.GetNestedObject().Type(),
	}
}
