package toproto5

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// SchemaAttribute returns the *tfprotov5.SchemaAttribute equivalent of an
// Attribute. Errors will be tftypes.AttributePathErrors based on `path`.
// `name` is the name of the attribute.
func SchemaAttribute(ctx context.Context, name string, path *tftypes.AttributePath, a tfsdk.Attribute) (*tfprotov5.SchemaAttribute, error) {
	if a.Attributes != nil && len(a.Attributes.GetAttributes()) > 0 {
		return nil, path.NewErrorf("protocol version 5 cannot have Attributes set")
	}

	if a.Type == nil {
		return nil, path.NewErrorf("must have Type set")
	}

	if !a.Required && !a.Optional && !a.Computed {
		return nil, path.NewErrorf("must have Required, Optional, or Computed set")
	}

	schemaAttribute := &tfprotov5.SchemaAttribute{
		Name:      name,
		Required:  a.Required,
		Optional:  a.Optional,
		Computed:  a.Computed,
		Sensitive: a.Sensitive,
		Type:      a.Type.TerraformType(ctx),
	}

	if a.DeprecationMessage != "" {
		schemaAttribute.Deprecated = true
	}

	if a.Description != "" {
		schemaAttribute.Description = a.Description
		schemaAttribute.DescriptionKind = tfprotov5.StringKindPlain
	}

	if a.MarkdownDescription != "" {
		schemaAttribute.Description = a.MarkdownDescription
		schemaAttribute.DescriptionKind = tfprotov5.StringKindMarkdown
	}

	return schemaAttribute, nil
}
