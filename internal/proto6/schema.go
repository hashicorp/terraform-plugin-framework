package proto6

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/schema"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Schema returns the *tfprotov6.Schema equivalent of a schema.Schema. At least
// one attribute must be set in the schema, or an error will be returned.
func Schema(ctx context.Context, s schema.Schema) (*tfprotov6.Schema, error) {
	result := &tfprotov6.Schema{
		Version: s.Version,
	}
	var attrs []*tfprotov6.SchemaAttribute
	for name, attr := range s.Attributes {
		a, err := Attribute(ctx, name, attr, tftypes.NewAttributePath().WithAttributeName(name))
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, a)
	}
	if len(attrs) < 1 {
		return nil, errors.New("must have at least one attribute in the schema")
	}
	result.Block = &tfprotov6.SchemaBlock{
		// core doesn't do anything with version, as far as I can tell,
		// so let's not set it.
		Attributes: attrs,
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

// Attribute returns the *tfprotov6.SchemaAttribute equivalent of a
// schema.Attribute. Errors will be tftypes.AttributePathErrors based on
// `path`. `name` is the name of the attribute.
func Attribute(ctx context.Context, name string, attr schema.Attribute, path *tftypes.AttributePath) (*tfprotov6.SchemaAttribute, error) {
	a := &tfprotov6.SchemaAttribute{
		Name:      name,
		Required:  attr.Required,
		Optional:  attr.Optional,
		Computed:  attr.Computed,
		Sensitive: attr.Sensitive,
	}
	if attr.DeprecationMessage != "" {
		a.Deprecated = true
	}
	if attr.Description != "" {
		a.Description = attr.Description
		a.DescriptionKind = tfprotov6.StringKindPlain
	}
	if attr.MarkdownDescription != "" {
		a.Description = attr.MarkdownDescription
		a.DescriptionKind = tfprotov6.StringKindMarkdown
	}
	if attr.Type != nil && attr.Attributes == nil {
		a.Type = attr.Type.TerraformType(ctx)
	} else if attr.Attributes != nil && attr.Type == nil {
		object := &tfprotov6.SchemaObject{
			MinItems: attr.Attributes.GetMinItems(),
			MaxItems: attr.Attributes.GetMaxItems(),
		}
		nm := attr.Attributes.GetNestingMode()
		switch nm {
		case schema.NestingModeSingle:
			object.Nesting = tfprotov6.SchemaObjectNestingModeSingle
		case schema.NestingModeList:
			object.Nesting = tfprotov6.SchemaObjectNestingModeList
		case schema.NestingModeSet:
			object.Nesting = tfprotov6.SchemaObjectNestingModeSet
		case schema.NestingModeMap:
			object.Nesting = tfprotov6.SchemaObjectNestingModeMap
		default:
			return nil, path.NewErrorf("unrecognized nesting mode %v", nm)
		}
		attrs := attr.Attributes.GetAttributes()
		for nestedName, nestedAttr := range attrs {
			nestedA, err := Attribute(ctx, name, nestedAttr, path.WithAttributeName(nestedName))
			if err != nil {
				return nil, err
			}
			object.Attributes = append(object.Attributes, nestedA)
		}
		a.NestedType = object
	} else if attr.Attributes != nil && attr.Type != nil {
		return nil, path.NewErrorf("can't have both Attributes and Type set")
	} else if attr.Attributes == nil && attr.Type == nil {
		return nil, path.NewErrorf("must have Attributes or Type set")
	}
	return a, nil
}
