package toproto6

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Schema returns the *tfprotov6.Schema equivalent of a Schema.
func Schema(ctx context.Context, s *tfsdk.Schema) (*tfprotov6.Schema, error) {
	if s == nil {
		return nil, nil
	}

	result := &tfprotov6.Schema{
		Version: s.Version,
	}

	var attrs []*tfprotov6.SchemaAttribute
	var blocks []*tfprotov6.SchemaNestedBlock

	for name, attr := range s.Attributes {
		a, err := SchemaAttribute(ctx, name, tftypes.NewAttributePath().WithAttributeName(name), attr)

		if err != nil {
			return nil, err
		}

		attrs = append(attrs, a)
	}

	//nolint:staticcheck // Block support is required within the framework.
	for name, block := range s.Blocks {
		proto6, err := Block(ctx, name, tftypes.NewAttributePath().WithAttributeName(name), block)

		if err != nil {
			return nil, err
		}

		blocks = append(blocks, proto6)
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

	result.Block = &tfprotov6.SchemaBlock{
		// core doesn't do anything with version, as far as I can tell,
		// so let's not set it.
		Attributes: attrs,
		BlockTypes: blocks,
		Deprecated: s.DeprecationMessage != "",
	}

	if s.Description != "" {
		result.Block.Description = s.Description
		result.Block.DescriptionKind = tfprotov6.StringKindPlain
	}

	if s.MarkdownDescription != "" {
		result.Block.Description = s.MarkdownDescription
		result.Block.DescriptionKind = tfprotov6.StringKindMarkdown
	}

	return result, nil
}
