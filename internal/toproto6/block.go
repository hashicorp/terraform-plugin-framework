// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Block returns the *tfprotov6.SchemaNestedBlock equivalent of a Block.
// Errors will be tftypes.AttributePathErrors based on `path`. `name` is the
// name of the attribute.
func Block(ctx context.Context, name string, path *tftypes.AttributePath, b fwschema.Block) (*tfprotov6.SchemaNestedBlock, error) {
	result, err := block(ctx, name, b)
	if err != nil {
		return nil, err.attributePathError(path)
	}
	return result, nil
}

func block(ctx context.Context, name string, b fwschema.Block) (*tfprotov6.SchemaNestedBlock, *schemaError) {
	schemaNestedBlock := &tfprotov6.SchemaNestedBlock{
		Block: &tfprotov6.SchemaBlock{
			Deprecated:         b.GetDeprecationMessage() != "",
			DeprecationMessage: b.GetDeprecationMessage(),
		},
		TypeName: name,
	}

	if b.GetDescription() != "" {
		schemaNestedBlock.Block.Description = b.GetDescription()
		schemaNestedBlock.Block.DescriptionKind = tfprotov6.StringKindPlain
	}

	if b.GetMarkdownDescription() != "" {
		schemaNestedBlock.Block.Description = b.GetMarkdownDescription()
		schemaNestedBlock.Block.DescriptionKind = tfprotov6.StringKindMarkdown
	}

	nm := b.GetNestingMode()
	switch nm {
	case fwschema.BlockNestingModeList:
		schemaNestedBlock.Nesting = tfprotov6.SchemaNestedBlockNestingModeList
	case fwschema.BlockNestingModeSet:
		schemaNestedBlock.Nesting = tfprotov6.SchemaNestedBlockNestingModeSet
	case fwschema.BlockNestingModeSingle:
		schemaNestedBlock.Nesting = tfprotov6.SchemaNestedBlockNestingModeSingle
	default:
		return nil, &schemaError{message: fmt.Sprintf("unrecognized nesting mode %v", nm)}
	}

	nestedBlockObject := b.GetNestedObject()

	for attrName, attr := range nestedBlockObject.GetAttributes() {
		attrProto6, err := schemaAttribute(ctx, attrName, attr)

		if err != nil {
			return nil, err.withAttributeName(attrName)
		}

		schemaNestedBlock.Block.Attributes = append(schemaNestedBlock.Block.Attributes, attrProto6)
	}

	for blockName, nestedBlock := range nestedBlockObject.GetBlocks() {
		blockProto6, err := block(ctx, blockName, nestedBlock)

		if err != nil {
			return nil, err.withAttributeName(blockName)
		}

		schemaNestedBlock.Block.BlockTypes = append(schemaNestedBlock.Block.BlockTypes, blockProto6)
	}

	sort.Slice(schemaNestedBlock.Block.Attributes, func(i, j int) bool {
		if schemaNestedBlock.Block.Attributes[i] == nil {
			return true
		}

		if schemaNestedBlock.Block.Attributes[j] == nil {
			return false
		}

		return schemaNestedBlock.Block.Attributes[i].Name < schemaNestedBlock.Block.Attributes[j].Name
	})

	sort.Slice(schemaNestedBlock.Block.BlockTypes, func(i, j int) bool {
		if schemaNestedBlock.Block.BlockTypes[i] == nil {
			return true
		}

		if schemaNestedBlock.Block.BlockTypes[j] == nil {
			return false
		}

		return schemaNestedBlock.Block.BlockTypes[i].TypeName < schemaNestedBlock.Block.BlockTypes[j].TypeName
	})

	return schemaNestedBlock, nil
}
