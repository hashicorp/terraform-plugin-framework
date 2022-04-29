package tfsdk

import (
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var _ tftypes.AttributePathStepper = Block{}

// Block defines the constraints and behaviors of a single structural field in a
// schema.
type Block struct {
	// Attributes are value fields inside the block. This map of attributes
	// behaves exactly like the map of attributes on the Schema type.
	Attributes map[string]Attribute

	// Blocks can have their own nested blocks. This nested map of blocks
	// behaves exactly like the map of blocks on the Schema type.
	Blocks map[string]Block

	// DeprecationMessage defines a message to display to practitioners
	// using this block, warning them that it is deprecated and
	// instructing them on what upgrade steps to take.
	DeprecationMessage string

	// Description is used in various tooling, like the language server, to
	// give practitioners more information about what this attribute is,
	// what it's for, and how it should be used. It should be written as
	// plain text, with no special formatting.
	Description string

	// MarkdownDescription is used in various tooling, like the
	// documentation generator, to give practitioners more information
	// about what this attribute is, what it's for, and how it should be
	// used. It should be formatted using Markdown.
	MarkdownDescription string

	// MaxItems is the maximum number of blocks that can be present in a
	// practitioner configuration.
	MaxItems int64

	// MinItems is the minimum number of blocks that must be present in a
	// practitioner configuration. Setting to 1 or above effectively marks
	// this configuration as required.
	MinItems int64

	// NestingMode indicates the block kind.
	NestingMode BlockNestingMode

	// PlanModifiers defines a sequence of modifiers for this block at
	// plan time. Block-level plan modifications occur before any
	// resource-level plan modifications.
	//
	// Any errors will prevent further execution of this sequence
	// of modifiers and modifiers associated with any nested Attribute or
	// Block, but will not prevent execution of PlanModifiers on any
	// other Attribute or Block in the Schema.
	//
	// Plan modification only applies to resources, not data sources or
	// providers. Setting PlanModifiers on a data source or provider attribute
	// will have no effect.
	PlanModifiers AttributePlanModifiers

	// Validators defines validation functionality for the block.
	Validators []AttributeValidator
}

// ApplyTerraform5AttributePathStep allows Blocks to be walked using
// tftypes.Walk and tftypes.Transform.
func (b Block) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	switch b.NestingMode {
	case BlockNestingModeList:
		_, ok := step.(tftypes.ElementKeyInt)

		if !ok {
			return nil, fmt.Errorf("can't apply %T to block NestingModeList", step)
		}

		return nestedBlock{Block: b}, nil
	case BlockNestingModeSet:
		_, ok := step.(tftypes.ElementKeyValue)

		if !ok {
			return nil, fmt.Errorf("can't apply %T to block NestingModeSet", step)
		}

		return nestedBlock{Block: b}, nil
	default:
		return nil, fmt.Errorf("unsupported block nesting mode: %v", b.NestingMode)
	}
}

// Equal returns true if `b` and `o` should be considered Equal.
func (b Block) Equal(o Block) bool {
	if !cmp.Equal(b.Attributes, o.Attributes) {
		return false
	}
	if !cmp.Equal(b.Blocks, o.Blocks) {
		return false
	}
	if b.DeprecationMessage != o.DeprecationMessage {
		return false
	}
	if b.Description != o.Description {
		return false
	}
	if b.MarkdownDescription != o.MarkdownDescription {
		return false
	}
	if b.MaxItems != o.MaxItems {
		return false
	}
	if b.MinItems != o.MinItems {
		return false
	}
	if b.NestingMode != o.NestingMode {
		return false
	}
	return true
}

// attributeType returns an attr.Type corresponding to the block.
func (b Block) attributeType() attr.Type {
	attrType := types.ObjectType{
		AttrTypes: map[string]attr.Type{},
	}

	for attrName, attr := range b.Attributes {
		attrType.AttrTypes[attrName] = attr.attributeType()
	}

	for blockName, block := range b.Attributes {
		attrType.AttrTypes[blockName] = block.attributeType()
	}

	switch b.NestingMode {
	case BlockNestingModeList:
		return types.ListType{
			ElemType: attrType,
		}
	case BlockNestingModeSet:
		return types.SetType{
			ElemType: attrType,
		}
	default:
		panic(fmt.Sprintf("unsupported block nesting mode: %v", b.NestingMode))
	}
}

// terraformType returns an tftypes.Type corresponding to the block.
func (b Block) terraformType(ctx context.Context) tftypes.Type {
	return b.attributeType().TerraformType(ctx)
}

type nestedBlock struct {
	Block
}

// ApplyTerraform5AttributePathStep allows Blocks to be walked using
// tftypes.Walk and tftypes.Transform.
func (b nestedBlock) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	a, ok := step.(tftypes.AttributeName)

	if !ok {
		return nil, fmt.Errorf("can't apply %T to block", step)
	}

	attrName := string(a)

	if attr, ok := b.Block.Attributes[attrName]; ok {
		return attr, nil
	}

	if block, ok := b.Block.Blocks[attrName]; ok {
		return block, nil
	}

	return nil, fmt.Errorf("no attribute %q on Attributes or Blocks", a)
}
