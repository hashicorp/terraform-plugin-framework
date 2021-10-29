package tfsdk

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Block defines the constraints and behaviors of a single structural field in a
// schema.
type Block struct {
	tftypes.AttributePathStepper

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
	NestingMode NestingMode
}

// ApplyTerraform5AttributePathStep allows Blocks to be walked using
// tftypes.Walk and tftypes.Transform.
func (b Block) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	switch b.NestingMode {
	case NestingModeList:
		_, ok := step.(tftypes.ElementKeyInt)

		if !ok {
			return nil, fmt.Errorf("can't apply %T to block NestingModeList", step)
		}

		return nestedBlock{Block: b}, nil
	case NestingModeSet:
		_, ok := step.(tftypes.ElementKeyValue)

		if !ok {
			return nil, fmt.Errorf("can't apply %T to block NestingModeSet", step)
		}

		return nestedBlock{Block: b}, nil
	default:
		panic(fmt.Sprintf("unsupported block nesting mode: %v", b.NestingMode))
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
	case NestingModeList:
		return types.ListType{
			ElemType: attrType,
		}
	case NestingModeSet:
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

// tfprotov6 returns the *tfprotov6.SchemaNestedBlock equivalent of a Block.
// Errors will be tftypes.AttributePathErrors based on `path`. `name` is the
// name of the attribute.
func (b Block) tfprotov6(ctx context.Context, name string, path *tftypes.AttributePath) (*tfprotov6.SchemaNestedBlock, error) {
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
	case NestingModeList:
		schemaNestedBlock.Nesting = tfprotov6.SchemaNestedBlockNestingModeList
	case NestingModeSet:
		schemaNestedBlock.Nesting = tfprotov6.SchemaNestedBlockNestingModeSet
	case NestingModeMap, NestingModeSingle:
		// This is intentional to only maintain the previous Terraform Plugin SDK support.
		return nil, path.NewErrorf("unsupported block nesting mode: %v", nm)
	default:
		return nil, path.NewErrorf("unrecognized nesting mode %v", nm)
	}

	for attrName, attr := range b.Attributes {
		attrPath := path.WithAttributeName(attrName)
		attrProto6, err := attr.tfprotov6(ctx, attrName, attrPath)

		if err != nil {
			return nil, err
		}

		schemaNestedBlock.Block.Attributes = append(schemaNestedBlock.Block.Attributes, attrProto6)
	}

	for blockName, block := range b.Blocks {
		blockPath := path.WithAttributeName(blockName)
		blockProto6, err := block.tfprotov6(ctx, blockName, blockPath)

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

// validate performs all Block validation.
func (b Block) validate(ctx context.Context, req ValidateAttributeRequest, resp *ValidateAttributeResponse) {
	attributeConfig, diags := req.Config.getAttributeValue(ctx, req.AttributePath)
	resp.Diagnostics.Append(diags...)

	if diags.HasError() {
		return
	}

	req.AttributeConfig = attributeConfig

	nm := b.NestingMode
	switch nm {
	case NestingModeList:
		l, ok := req.AttributeConfig.(types.List)

		if !ok {
			err := fmt.Errorf("unknown block value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Block Validation Error",
				"Block validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for idx := range l.Elems {
			for name, attr := range b.Attributes {
				nestedAttrReq := ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyInt(idx).WithAttributeName(name),
					Config:        req.Config,
				}
				nestedAttrResp := &ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				attr.validate(ctx, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}

			for name, block := range b.Blocks {
				nestedAttrReq := ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyInt(idx).WithAttributeName(name),
					Config:        req.Config,
				}
				nestedAttrResp := &ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				block.validate(ctx, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case NestingModeSet:
		s, ok := req.AttributeConfig.(types.Set)

		if !ok {
			err := fmt.Errorf("unknown block value type (%T) for nesting mode (%T) at path: %s", req.AttributeConfig, nm, req.AttributePath)
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Block Validation Error",
				"Block validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		for _, value := range s.Elems {
			tfValueRaw, err := value.ToTerraformValue(ctx)

			if err != nil {
				err := fmt.Errorf("error running ToTerraformValue on element value: %v", value)
				resp.Diagnostics.AddAttributeError(
					req.AttributePath,
					"Block Validation Error",
					"Block validation cannot convert element into a Terraform value. Report this to the provider developer:\n\n"+err.Error(),
				)

				return
			}

			tfValue := tftypes.NewValue(s.ElemType.TerraformType(ctx), tfValueRaw)

			for name, attr := range b.Attributes {
				nestedAttrReq := ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyValue(tfValue).WithAttributeName(name),
					Config:        req.Config,
				}
				nestedAttrResp := &ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				attr.validate(ctx, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}

			for name, block := range b.Blocks {
				nestedAttrReq := ValidateAttributeRequest{
					AttributePath: req.AttributePath.WithElementKeyValue(tfValue).WithAttributeName(name),
					Config:        req.Config,
				}
				nestedAttrResp := &ValidateAttributeResponse{
					Diagnostics: resp.Diagnostics,
				}

				block.validate(ctx, nestedAttrReq, nestedAttrResp)

				resp.Diagnostics = nestedAttrResp.Diagnostics
			}
		}
	case NestingModeMap, NestingModeSingle:
		// This is intentional to only maintain the previous Terraform Plugin SDK support.
		err := fmt.Errorf("unsupported block validation nesting mode (%T: %v) at path: %s", nm, nm, req.AttributePath)
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Block Validation Error",
			"Block validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
		)

		return
	default:
		err := fmt.Errorf("unknown block validation nesting mode (%T: %v) at path: %s", nm, nm, req.AttributePath)
		resp.Diagnostics.AddAttributeError(
			req.AttributePath,
			"Block Validation Error",
			"Block validation cannot walk schema. Report this to the provider developer:\n\n"+err.Error(),
		)

		return
	}

	if b.DeprecationMessage != "" && attributeConfig != nil {
		tfValue, err := attributeConfig.ToTerraformValue(ctx)

		if err != nil {
			resp.Diagnostics.AddAttributeError(
				req.AttributePath,
				"Block Validation Error",
				"Block validation cannot convert value. Report this to the provider developer:\n\n"+err.Error(),
			)

			return
		}

		if tfValue != nil {
			resp.Diagnostics.AddAttributeWarning(
				req.AttributePath,
				"Block Deprecated",
				b.DeprecationMessage,
			)
		}
	}
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
