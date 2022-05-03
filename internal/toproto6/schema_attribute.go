package toproto6

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// SchemaAttribute returns the *tfprotov6.SchemaAttribute equivalent of an
// Attribute. Errors will be tftypes.AttributePathErrors based on `path`.
// `name` is the name of the attribute.
func SchemaAttribute(ctx context.Context, name string, path *tftypes.AttributePath, a tfsdk.Attribute) (*tfprotov6.SchemaAttribute, error) {
	if a.Attributes != nil && len(a.Attributes.GetAttributes()) > 0 && a.Type != nil {
		return nil, path.NewErrorf("cannot have both Attributes and Type set")
	}

	if (a.Attributes == nil || len(a.Attributes.GetAttributes()) == 0) && a.Type == nil {
		return nil, path.NewErrorf("must have Attributes or Type set")
	}

	if !a.Required && !a.Optional && !a.Computed {
		return nil, path.NewErrorf("must have Required, Optional, or Computed set")
	}

	schemaAttribute := &tfprotov6.SchemaAttribute{
		Name:      name,
		Required:  a.Required,
		Optional:  a.Optional,
		Computed:  a.Computed,
		Sensitive: a.Sensitive,
	}

	if a.DeprecationMessage != "" {
		schemaAttribute.Deprecated = true
	}

	if a.Description != "" {
		schemaAttribute.Description = a.Description
		schemaAttribute.DescriptionKind = tfprotov6.StringKindPlain
	}

	if a.MarkdownDescription != "" {
		schemaAttribute.Description = a.MarkdownDescription
		schemaAttribute.DescriptionKind = tfprotov6.StringKindMarkdown
	}

	if a.Type != nil {
		schemaAttribute.Type = a.Type.TerraformType(ctx)

		return schemaAttribute, nil
	}

	object := &tfprotov6.SchemaObject{}
	nm := a.Attributes.GetNestingMode()
	switch nm {
	case tfsdk.NestingModeSingle:
		object.Nesting = tfprotov6.SchemaObjectNestingModeSingle
	case tfsdk.NestingModeList:
		object.Nesting = tfprotov6.SchemaObjectNestingModeList
	case tfsdk.NestingModeSet:
		object.Nesting = tfprotov6.SchemaObjectNestingModeSet
	case tfsdk.NestingModeMap:
		object.Nesting = tfprotov6.SchemaObjectNestingModeMap
	default:
		return nil, path.NewErrorf("unrecognized nesting mode %v", nm)
	}

	for nestedName, nestedA := range a.Attributes.GetAttributes() {
		nestedSchemaAttribute, err := SchemaAttribute(ctx, nestedName, path.WithAttributeName(nestedName), nestedA)

		if err != nil {
			return nil, err
		}

		object.Attributes = append(object.Attributes, nestedSchemaAttribute)
	}

	sort.Slice(object.Attributes, func(i, j int) bool {
		if object.Attributes[i] == nil {
			return true
		}

		if object.Attributes[j] == nil {
			return false
		}

		return object.Attributes[i].Name < object.Attributes[j].Name
	})

	schemaAttribute.NestedType = object

	return schemaAttribute, nil
}
