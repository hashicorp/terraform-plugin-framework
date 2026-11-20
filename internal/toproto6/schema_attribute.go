// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package toproto6

import (
	"context"
	"fmt"
	"sort"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
)

// SchemaAttribute returns the *tfprotov6.SchemaAttribute equivalent of an
// Attribute. Errors will be tftypes.AttributePathErrors based on `path`.
// `name` is the name of the attribute.
func SchemaAttribute(ctx context.Context, name string, path *tftypes.AttributePath, a fwschema.Attribute) (*tfprotov6.SchemaAttribute, error) {
	result, err := schemaAttribute(ctx, name, a)
	if err != nil {
		return nil, err.attributePathError(path)
	}
	return result, nil
}

func schemaAttribute(ctx context.Context, name string, a fwschema.Attribute) (*tfprotov6.SchemaAttribute, *schemaError) {
	if !a.IsRequired() && !a.IsOptional() && !a.IsComputed() {
		return nil, &schemaError{message: "must have Required, Optional, or Computed set"}
	}

	result := &tfprotov6.SchemaAttribute{
		Name:      name,
		Required:  a.IsRequired(),
		Optional:  a.IsOptional(),
		Computed:  a.IsComputed(),
		Sensitive: a.IsSensitive(),
		Type:      a.GetType().TerraformType(ctx),
		WriteOnly: a.IsWriteOnly(),
	}

	if a.GetDeprecationMessage() != "" {
		result.Deprecated = true
		result.DeprecationMessage = a.GetDeprecationMessage()
	}

	if a.GetDescription() != "" {
		result.Description = a.GetDescription()
		result.DescriptionKind = tfprotov6.StringKindPlain
	}

	if a.GetMarkdownDescription() != "" {
		result.Description = a.GetMarkdownDescription()
		result.DescriptionKind = tfprotov6.StringKindMarkdown
	}

	nestedAttribute, ok := a.(fwschema.NestedAttribute)

	if !ok {
		return result, nil
	}

	object := &tfprotov6.SchemaObject{}
	nm := nestedAttribute.GetNestingMode()
	switch nm {
	case fwschema.NestingModeSingle:
		object.Nesting = tfprotov6.SchemaObjectNestingModeSingle
	case fwschema.NestingModeList:
		object.Nesting = tfprotov6.SchemaObjectNestingModeList
	case fwschema.NestingModeSet:
		object.Nesting = tfprotov6.SchemaObjectNestingModeSet
	case fwschema.NestingModeMap:
		object.Nesting = tfprotov6.SchemaObjectNestingModeMap
	default:
		return nil, &schemaError{message: fmt.Sprintf("unrecognized nesting mode %v", nm)}
	}

	for nestedName, nestedA := range nestedAttribute.GetNestedObject().GetAttributes() {
		nestedSchemaAttribute, err := schemaAttribute(ctx, nestedName, nestedA)

		if err != nil {
			return nil, err.withAttributeName(nestedName)
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

	result.NestedType = object
	result.Type = nil

	return result, nil
}
