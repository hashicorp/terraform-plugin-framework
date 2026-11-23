// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package toproto5

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Block returns the *tfprotov5.SchemaNestedBlock equivalent of a Block.
// Errors will be tftypes.AttributePathErrors based on `path`. `name` is the
// name of the attribute.
func Block(ctx context.Context, name string, path *tftypes.AttributePath, b fwschema.Block) (*tfprotov5.SchemaNestedBlock, error) {
	result, err := block(ctx, name, b)
	if err != nil {
		return nil, err.attributePathError(path)
	}
	return result, nil
}

func block(ctx context.Context, name string, b fwschema.Block) (*tfprotov5.SchemaNestedBlock, *schemaError) {
	schemaNestedBlock := &tfprotov5.SchemaNestedBlock{
		Block: &tfprotov5.SchemaBlock{
			Deprecated:         b.GetDeprecationMessage() != "",
			DeprecationMessage: b.GetDeprecationMessage(),
		},
		TypeName: name,
	}

	if b.GetDescription() != "" {
		schemaNestedBlock.Block.Description = b.GetDescription()
		schemaNestedBlock.Block.DescriptionKind = tfprotov5.StringKindPlain
	}

	if b.GetMarkdownDescription() != "" {
		schemaNestedBlock.Block.Description = b.GetMarkdownDescription()
		schemaNestedBlock.Block.DescriptionKind = tfprotov5.StringKindMarkdown
	}

	nm := b.GetNestingMode()
	switch nm {
	case fwschema.BlockNestingModeList:
		schemaNestedBlock.Nesting = tfprotov5.SchemaNestedBlockNestingModeList
	case fwschema.BlockNestingModeSet:
		schemaNestedBlock.Nesting = tfprotov5.SchemaNestedBlockNestingModeSet
	case fwschema.BlockNestingModeSingle:
		schemaNestedBlock.Nesting = tfprotov5.SchemaNestedBlockNestingModeSingle
	default:
		return nil, &schemaError{message: fmt.Sprintf("unrecognized nesting mode %v", nm)}
	}

	nestedBlockObject := b.GetNestedObject()

	if nestedBlockObject.LenAttributes() > 0 {
		var conversionErr *schemaError
		nestedBlockObject.RangeAttributes(func(attrName string, attr fwschema.Attribute) bool {
			attrProto5, err := schemaAttribute(ctx, attrName, attr)

			if err != nil {
				conversionErr = err.withAttributeName(attrName)
				return false
			}

			schemaNestedBlock.Block.Attributes = append(schemaNestedBlock.Block.Attributes, attrProto5)
			return true
		})
		if conversionErr != nil {
			return nil, conversionErr
		}
	}

	if nestedBlockObject.LenBlocks() > 0 {
		var conversionErr *schemaError
		nestedBlockObject.RangeBlocks(func(blockName string, nestedBlock fwschema.Block) bool {
			blockProto5, err := block(ctx, blockName, nestedBlock)

			if err != nil {
				conversionErr = err.withAttributeName(blockName)
				return false
			}

			schemaNestedBlock.Block.BlockTypes = append(schemaNestedBlock.Block.BlockTypes, blockProto5)
			return true
		})
		if conversionErr != nil {
			return nil, conversionErr
		}
	}

	slices.SortFunc(schemaNestedBlock.Block.Attributes, func(a, b *tfprotov5.SchemaAttribute) int {
		if a == nil {
			if b == nil {
				return 0
			}

			return -1
		}

		if b == nil {
			return 1
		}

		return strings.Compare(a.Name, b.Name)
	})

	slices.SortFunc(schemaNestedBlock.Block.BlockTypes, func(a, b *tfprotov5.SchemaNestedBlock) int {
		if a == nil {
			if b == nil {
				return 0
			}

			return -1
		}

		if b == nil {
			return 1
		}

		return strings.Compare(a.TypeName, b.TypeName)
	})

	return schemaNestedBlock, nil
}
