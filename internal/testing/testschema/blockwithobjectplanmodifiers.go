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

var _ fwxschema.BlockWithObjectPlanModifiers = BlockWithObjectPlanModifiers{}

type BlockWithObjectPlanModifiers struct {
	Attributes          map[string]fwschema.Attribute
	Blocks              map[string]fwschema.Block
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	PlanModifiers       []planmodifier.Object
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Block interface.
func (b BlockWithObjectPlanModifiers) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return b.Type().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Block interface.
func (b BlockWithObjectPlanModifiers) Equal(o fwschema.Block) bool {
	_, ok := o.(BlockWithObjectPlanModifiers)

	if !ok {
		return false
	}

	return fwschema.BlocksEqual(b, o)
}

// GetDeprecationMessage satisfies the fwschema.Block interface.
func (b BlockWithObjectPlanModifiers) GetDeprecationMessage() string {
	return b.DeprecationMessage
}

// GetDescription satisfies the fwschema.Block interface.
func (b BlockWithObjectPlanModifiers) GetDescription() string {
	return b.Description
}

// GetMarkdownDescription satisfies the fwschema.Block interface.
func (b BlockWithObjectPlanModifiers) GetMarkdownDescription() string {
	return b.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.Block interface.
func (b BlockWithObjectPlanModifiers) GetNestedObject() fwschema.NestedBlockObject {
	return NestedBlockObjectWithPlanModifiers{
		Attributes:    b.Attributes,
		Blocks:        b.Blocks,
		PlanModifiers: b.PlanModifiers,
	}
}

// GetNestingMode satisfies the fwschema.Block interface.
func (b BlockWithObjectPlanModifiers) GetNestingMode() fwschema.BlockNestingMode {
	return fwschema.BlockNestingModeSingle
}

// ObjectPlanModifiers satisfies the fwxschema.BlockWithObjectPlanModifiers interface.
func (b BlockWithObjectPlanModifiers) ObjectPlanModifiers() []planmodifier.Object {
	return b.PlanModifiers
}

// Type satisfies the fwschema.Block interface.
func (b BlockWithObjectPlanModifiers) Type() attr.Type {
	return b.GetNestedObject().Type()
}
