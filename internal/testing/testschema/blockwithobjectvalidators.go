// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema/fwxschema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ fwxschema.BlockWithObjectValidators = BlockWithObjectValidators{}

type BlockWithObjectValidators struct {
	Attributes          map[string]fwschema.Attribute
	Blocks              map[string]fwschema.Block
	DeprecationMessage  string
	Description         string
	MarkdownDescription string
	Validators          []validator.Object
}

// ApplyTerraform5AttributePathStep satisfies the fwschema.Block interface.
func (b BlockWithObjectValidators) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (any, error) {
	return b.Type().ApplyTerraform5AttributePathStep(step)
}

// Equal satisfies the fwschema.Block interface.
func (b BlockWithObjectValidators) Equal(o fwschema.Block) bool {
	_, ok := o.(BlockWithObjectValidators)

	if !ok {
		return false
	}

	return fwschema.BlocksEqual(b, o)
}

// GetDeprecationMessage satisfies the fwschema.Block interface.
func (b BlockWithObjectValidators) GetDeprecationMessage() string {
	return b.DeprecationMessage
}

// GetDescription satisfies the fwschema.Block interface.
func (b BlockWithObjectValidators) GetDescription() string {
	return b.Description
}

// GetMarkdownDescription satisfies the fwschema.Block interface.
func (b BlockWithObjectValidators) GetMarkdownDescription() string {
	return b.MarkdownDescription
}

// GetNestedObject satisfies the fwschema.Block interface.
func (b BlockWithObjectValidators) GetNestedObject() fwschema.NestedBlockObject {
	return NestedBlockObjectWithValidators{
		Attributes: b.Attributes,
		Blocks:     b.Blocks,
		Validators: b.Validators,
	}
}

// GetNestingMode satisfies the fwschema.Block interface.
func (b BlockWithObjectValidators) GetNestingMode() fwschema.BlockNestingMode {
	return fwschema.BlockNestingModeSingle
}

// ObjectValidators satisfies the fwxschema.BlockWithObjectValidators interface.
func (b BlockWithObjectValidators) ObjectValidators() []validator.Object {
	return b.Validators
}

// Type satisfies the fwschema.Block interface.
func (b BlockWithObjectValidators) Type() attr.Type {
	return b.GetNestedObject().Type()
}
