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

var _ fwxschema.BlockWithListPlanModifiers = BlockWithListPlanModifiers{}

type BlockWithListPlanModifiers struct {
	Attributes          map[string]fwschema.Attribute
	Blocks              map[string]fwschema.Block
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	PlanModifiers       []planmodifier.List
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Block interface.
func (b BlockWithListPlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return b.Type().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Block interface.
func (b BlockWithListPlanModifiers) Equal(o fwschema.Block) bool {
	_, ok := o.(BlockWithListPlanModifiers)

	if !ok {
		return false
	}

	return fwschema.BlocksEqual(b, o)
}

// GetDeprecationMessage satisfies the fwschema.Block interface.
func (b BlockWithListPlanModifiers) GetDeprecationMessage() string {
	return b.DeprecationMessage
}

// GetDescription satisfies the fwschema.Block interface.
func (b BlockWithListPlanModifiers) GetDescription() string {
	return b.Description
}

// GetMarkdownDescription satisfies the fwschema.Block interface.
func (b BlockWithListPlanModifiers) GetMarkdownDescription() string {
	return b.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.Block interface.
func (b BlockWithListPlanModifiers) GetNestedObject() fwschema.NestedBlockObject {
	return NestedBlockObject{
		Attributes: b.Attributes,
		Blocks:     b.Blocks,
	}
}

// GetNestingMode satisfies the fwschema.Block interface.
func (b BlockWithListPlanModifiers) GetNestingMode() fwschema.BlockNestingMode {
	return fwschema.BlockNestingModeList
}

// ListPlanModifiers satisfies the fwxschema.BlockWithListPlanModifiers interface.
func (b BlockWithListPlanModifiers) ListPlanModifiers() []planmodifier.List {
	return b.PlanModifiers
}

// Type satisfies the fwschema.Block interface.
func (b BlockWithListPlanModifiers) Type() attr.Type {
	return types.ListType{
		ElemType: b.GetNestedObject().Type(),
	}
}
