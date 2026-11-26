// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// Schema returns the *tfprotov6.Schema equivalent of a Schema.
func Schema(ctx context.Context, s fwschema.Schema) (*tfprotov6.Schema, error) {
	if s == nil {
		return nil, nil
	}

	result := &tfprotov6.Schema{
		Version: s.GetVersion(),
	}

	var attrs []*tfprotov6.SchemaAttribute
	var blocks []*tfprotov6.SchemaNestedBlock

	for name, attr := range s.GetAttributes() {
		a, err := schemaAttribute(ctx, name, attr)

		if err != nil {
			return nil, err.withAttributeName(name).attributePathError(nil)
		}

		attrs = append(attrs, a)
	}

	for name, nestedBlock := range s.GetBlocks() {
		proto6, err := block(ctx, name, nestedBlock)

		if err != nil {
			return nil, err.withAttributeName(name).attributePathError(nil)
		}

		blocks = append(blocks, proto6)
	}

	slices.SortFunc(attrs, func(a, b *tfprotov6.SchemaAttribute) int {
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

	slices.SortFunc(blocks, func(a, b *tfprotov6.SchemaNestedBlock) int {
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

	result.Block = &tfprotov6.SchemaBlock{
		// core doesn't do anything with version, as far as I can tell,
		// so let's not set it.
		Attributes:         attrs,
		BlockTypes:         blocks,
		Deprecated:         s.GetDeprecationMessage() != "",
		DeprecationMessage: s.GetDeprecationMessage(),
	}

	if s.GetDescription() != "" {
		result.Block.Description = s.GetDescription()
		result.Block.DescriptionKind = tfprotov6.StringKindPlain
	}

	if s.GetMarkdownDescription() != "" {
		result.Block.Description = s.GetMarkdownDescription()
		result.Block.DescriptionKind = tfprotov6.StringKindMarkdown
	}

	return result, nil
}
