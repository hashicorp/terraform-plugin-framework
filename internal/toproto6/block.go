package toproto6

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Block returns the *tfprotov6.SchemaNestedBlock equivalent of a Block.
// Errors will be tftypes.AttributePathErrors based on `path`. `name` is the
// name of the attribute.
func Block(ctx context.Context, name string, path *tftypes.AttributePath, b tfsdk.Block) (*tfprotov6.SchemaNestedBlock, error) {
	schemaNestedBlock := &tfprotov6.SchemaNestedBlock{
		Block: &tfprotov6.SchemaBlock{
			Deprecated: b.DeprecationMessage != "",
		},
		MinItems: b.MinItems,
		MaxItems: b.MaxItems,
		TypeName: name,
	}

	if b.Description != "" {
		schemaNestedBlock.Block.Description = b.Description
		schemaNestedBlock.Block.DescriptionKind = tfprotov6.StringKindPlain
	}

	if b.MarkdownDescription != "" {
		schemaNestedBlock.Block.Description = b.MarkdownDescription
		schemaNestedBlock.Block.DescriptionKind = tfprotov6.StringKindMarkdown
	}

	nm := b.NestingMode
	switch nm {
	case tfsdk.BlockNestingModeList:
		schemaNestedBlock.Nesting = tfprotov6.SchemaNestedBlockNestingModeList
	case tfsdk.BlockNestingModeSet:
		schemaNestedBlock.Nesting = tfprotov6.SchemaNestedBlockNestingModeSet
	default:
		return nil, path.NewErrorf("unrecognized nesting mode %v", nm)
	}

	for attrName, attr := range b.Attributes {
		attrPath := path.WithAttributeName(attrName)
		attrProto6, err := SchemaAttribute(ctx, attrName, attrPath, attr)

		if err != nil {
			return nil, err
		}

		schemaNestedBlock.Block.Attributes = append(schemaNestedBlock.Block.Attributes, attrProto6)
	}

	for blockName, block := range b.Blocks {
		blockPath := path.WithAttributeName(blockName)
		blockProto6, err := Block(ctx, blockName, blockPath, block)

		if err != nil {
			return nil, err
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
