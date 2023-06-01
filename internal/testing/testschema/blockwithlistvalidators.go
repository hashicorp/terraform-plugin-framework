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

var _ fwxschema.BlockWithListValidators = BlockWithListValidators{}

type BlockWithListValidators struct {
	Attributes          map[string]fwschema.Attribute
	Blocks              map[string]fwschema.Block
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Validators          []validator.List
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Block interface.
func (b BlockWithListValidators) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return b.Type().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Block interface.
func (b BlockWithListValidators) Equal(o fwschema.Block) bool {
	_, ok := o.(BlockWithListValidators)

	if !ok {
		return false
	}

	return fwschema.BlocksEqual(b, o)
}

// GetDeprecationMessage satisfies the fwschema.Block interface.
func (b BlockWithListValidators) GetDeprecationMessage() string {
	return b.DeprecationMessage
}

// GetDescription satisfies the fwschema.Block interface.
func (b BlockWithListValidators) GetDescription() string {
	return b.Description
}

// GetMarkdownDescription satisfies the fwschema.Block interface.
func (b BlockWithListValidators) GetMarkdownDescription() string {
	return b.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.Block interface.
func (b BlockWithListValidators) GetNestedObject() fwschema.NestedBlockObject {
	return NestedBlockObject{
		Attributes: b.Attributes,
		Blocks:     b.Blocks,
	}
}

// GetNestingMode satisfies the fwschema.Block interface.
func (b BlockWithListValidators) GetNestingMode() fwschema.BlockNestingMode {
	return fwschema.BlockNestingModeList
}

// ListValidators satisfies the fwxschema.BlockWithListValidators interface.
func (b BlockWithListValidators) ListValidators() []validator.List {
	return b.Validators
}

// Type satisfies the fwschema.Block interface.
func (b BlockWithListValidators) Type() attr.Type {
	return types.ListType{
		ElemType: b.GetNestedObject().Type(),
	}
}
