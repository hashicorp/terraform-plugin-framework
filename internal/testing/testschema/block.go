// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ fwschema.Block = Block{}

type Block struct {
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	NestedObject        fwschema.NestedBlockObject
	NestingMode         fwschema.BlockNestingMode
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Block interface.
func (b Block) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return b.Type().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Block interface.
func (b Block) Equal(o fwschema.Block) bool {
	_, ok := o.(Block)

	if !ok {
		return false
	}

	return fwschema.BlocksEqual(b, o)
}

// GetDeprecationMessage satisfies the fwschema.Block interface.
func (b Block) GetDeprecationMessage() string {
	return b.DeprecationMessage
}

// GetDescription satisfies the fwschema.Block interface.
func (b Block) GetDescription() string {
	return b.Description
}

// GetMarkdownDescription satisfies the fwschema.Block interface.
func (b Block) GetMarkdownDescription() string {
	return b.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.Block interface.
func (b Block) GetNestedObject() fwschema.NestedBlockObject {
	return b.NestedObject
}

// GetNestingMode satisfies the fwschema.Block interface.
func (b Block) GetNestingMode() fwschema.BlockNestingMode {
	return b.NestingMode
}

// Type satisfies the fwschema.Block interface.
func (b Block) Type() attr.Type {
	switch b.GetNestingMode() {
	case fwschema.BlockNestingModeList:
		return types.ListType{
			ElemType: b.GetNestedObject().Type(),
		}
	case fwschema.BlockNestingModeSet:
		return types.SetType{
			ElemType: b.GetNestedObject().Type(),
		}
	case fwschema.BlockNestingModeSingle:
		return b.GetNestedObject().Type()
	default:
		return nil
	}
}
