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

var _ fwxschema.BlockWithSetPlanModifiers = BlockWithSetPlanModifiers{}

type BlockWithSetPlanModifiers struct {
	Attributes          map[string]fwschema.Attribute
	Blocks              map[string]fwschema.Block
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	PlanModifiers       []planmodifier.Set
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Block interface.
func (b BlockWithSetPlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return b.Type().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Block interface.
func (b BlockWithSetPlanModifiers) Equal(o fwschema.Block) bool {
	_, ok := o.(BlockWithSetPlanModifiers)

	if !ok {
		return false
	}

	return fwschema.BlocksEqual(b, o)
}

// GetDeprecationMessage satisfies the fwschema.Block interface.
func (b BlockWithSetPlanModifiers) GetDeprecationMessage() string {
	return b.DeprecationMessage
}

// GetDescription satisfies the fwschema.Block interface.
func (b BlockWithSetPlanModifiers) GetDescription() string {
	return b.Description
}

// GetMarkdownDescription satisfies the fwschema.Block interface.
func (b BlockWithSetPlanModifiers) GetMarkdownDescription() string {
	return b.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.Block interface.
func (b BlockWithSetPlanModifiers) GetNestedObject() fwschema.NestedBlockObject {
	return NestedBlockObject{
		Attributes: b.Attributes,
		Blocks:     b.Blocks,
	}
}

// GetNestingMode satisfies the fwschema.Block interface.
func (b BlockWithSetPlanModifiers) GetNestingMode() fwschema.BlockNestingMode {
	return fwschema.BlockNestingModeSet
}

// SetPlanModifiers satisfies the fwxschema.BlockWithSetPlanModifiers interface.
func (b BlockWithSetPlanModifiers) SetPlanModifiers() []planmodifier.Set {
	return b.PlanModifiers
}

// Type satisfies the fwschema.Block interface.
func (b BlockWithSetPlanModifiers) Type() attr.Type {
	return types.SetType{
		ElemType: b.GetNestedObject().Type(),
	}
}
