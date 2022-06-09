package toproto5

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Schema returns the *tfprotov5.Schema equivalent of a Schema.
func Schema(ctx context.Context, s *tfsdk.Schema) (*tfprotov5.Schema, error) {
	if s == nil {
		return nil, nil
	}

	result := &tfprotov5.Schema{
		Version: s.Version,
	}

	var attrs []*tfprotov5.SchemaAttribute
	var blocks []*tfprotov5.SchemaNestedBlock

	for name, attr := range s.Attributes {
		a, err := SchemaAttribute(ctx, name, tftypes.NewAttributePath().WithAttributeName(name), attr)

		if err != nil {
			return nil, err
		}

		attrs = append(attrs, a)
	}

	for name, block := range s.Blocks {
		proto5, err := Block(ctx, name, tftypes.NewAttributePath().WithAttributeName(name), block)

		if err != nil {
			return nil, err
		}

		blocks = append(blocks, proto5)
	}

	sort.Slice(attrs, func(i, j int) bool {
		if attrs[i] == nil {
			return true
		}

		if attrs[j] == nil {
			return false
		}

		return attrs[i].Name < attrs[j].Name
	})

	sort.Slice(blocks, func(i, j int) bool {
		if blocks[i] == nil {
			return true
		}

		if blocks[j] == nil {
			return false
		}

		return blocks[i].TypeName < blocks[j].TypeName
	})

	result.Block = &tfprotov5.SchemaBlock{
		// core doesn't do anything with version, as far as I can tell,
		// so let's not set it.
		Attributes: attrs,
		BlockTypes: blocks,
		Deprecated: s.DeprecationMessage != "",
	}

	if s.Description != "" {
		result.Block.Description = s.Description
		result.Block.DescriptionKind = tfprotov5.StringKindPlain
	}

	if s.MarkdownDescription != "" {
		result.Block.Description = s.MarkdownDescription
		result.Block.DescriptionKind = tfprotov5.StringKindMarkdown
	}

	return result, nil
}
