package metadata

import (
	"context"
	"encoding/json"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-go/metadata"
)

// 1. fwschema.Schema -> metadata.SchemaBlock
//    - looping through attributes, for each attribute
// 			2. fwschema.Attribute -> metadata.Attribute
//			  - for each attribute that "is" a nested attribute (type assertion)
// 					3. fwschema.NestedAttribute -> metadata.NestedAttribute (?)
//
// walk, transform => traversing a tree recursively with different goals (walk = visit each node in a tree, transform = change some nodes in a tree)
// - function in a package, that accepts a callback function

// 1 resource schema with 2 attributes (string, boolean), stubbed type

func MetadataSchemaAttribute(ctx context.Context, attr fwschema.Attribute) (*metadata.Attribute, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	/*nestedAttr, ok := attr.(schema.NestedAttribute)

	if ok {
		nestedAttr.GetNestedObject().GetAttributes() // more attributes to loop through
		// nestedAttr.GetNestingMode()                  // nested mode (list, set, map, single)
		// some examples of recursive type logic
		// attribute validation
		// nestedMetadataObject := newFunc() // eventual recursion

		object := &tfprotov6.SchemaObject{}
		nm := nestedAttr.GetNestingMode()
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
			diags.Append(diag.NewErrorDiagnostic("error", "unrecognized nesting mode"))
		}

		for nestedName, nestedA := range nestedAttr.GetNestedObject().GetAttributes() {
			nestedSchemaAttribute, err := MetadataSchemaAttribute(ctx, nestedName, nestedA)

			if err != nil {
				return nil, err
			}

			return nestedSchemaAttribute, err
		}

	}*/

	desc := attr.GetDescription()
	kind := metadata.Plain
	var depMessage *string
	dep := true

	if attr.GetDeprecationMessage() == "" {
		dep = false
	}

	if attr.GetDescription() != "" {
		desc = attr.GetDescription()
		kind = metadata.Plain
	} else if attr.GetMarkdownDescription() != "" {
		desc = attr.GetMarkdownDescription()
		kind = metadata.Markdown
	}

	if attr.GetDeprecationMessage() != "" {
		depMessage = metadata.StringPointer(attr.GetDeprecationMessage())
	}

	ty, _ := attr.GetType().TerraformType(ctx).MarshalJSON() // unmarshal to get to metadata.Type?

	typ := json.RawMessage(ty)

	this := &metadata.Attribute{
		Computed:           metadata.BoolPointer(attr.IsComputed()),
		Deprecated:         metadata.BoolPointer(dep),
		DeprecationMessage: depMessage,
		Description:        metadata.StringPointer(desc),
		DescriptionKind:    metadata.DescriptionPointer(kind), // call both to see if it is plain txt or markdown
		NestedType:         nil,                               // make flag to check if true or false and negate for Type
		Optional:           metadata.BoolPointer(attr.IsOptional()),
		PlanModifications:  nil, // TODO: there some type assertion -> resource/schema.Attribute => plan modifier function
		Required:           metadata.BoolPointer(attr.IsRequired()),
		SDKType:            nil, // TODO: Net new logic, ignore for now
		Sensitive:          metadata.BoolPointer(attr.IsSensitive()),
		Type:               metadata.TypePointer(typ), // make flag to check if true or false
		Validations:        nil,                       // list
	}

	return this, diags
}

func MetadataSchemaAttributes(ctx context.Context, attrs map[string]fwschema.Attribute) (map[string]metadata.Attribute, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	out := make(map[string]metadata.Attribute, len(attrs)) // making empty map to return

	for name, attr := range attrs { // top level function checks for nesting
		if nestedAttr, isNested := attr.(fwschema.NestedAttribute); isNested {
			nestedAttrMeta, nestedDiags := MetadataNestedSchemaAttribute(ctx, nestedAttr) // fills out not nested attributes
			if nestedDiags.HasError() {
				// TODO: handle diags later
				return nil, nestedDiags
			}
			out[name] = *nestedAttrMeta
		} else {
			attrMeta, attrDiags := MetadataSchemaAttribute(ctx, attr)
			if attrDiags.HasError() {
				return nil, attrDiags
			}
			out[name] = *attrMeta
		}
	}

	return out, diags
}

func MetadataNestedSchemaAttribute(ctx context.Context, attr fwschema.NestedAttribute) (*metadata.Attribute, diag.Diagnostics) {
	diags := diag.Diagnostics{}

	nestedAttr := attr.GetNestedObject().GetAttributes() // more attributes to loop through
	// some examples of recursive type logic
	// attribute validation
	// nestedMetadataObject := newFunc() // eventual recursion

	object := &metadata.NestedAttributeType{}
	nm := attr.GetNestingMode() // nested mode (list, set, map, single)
	switch nm {
	case fwschema.NestingModeSingle:
		object.NestingMode = metadata.AnyPointer(metadata.PurpleSingle)
	case fwschema.NestingModeList:
		object.NestingMode = metadata.AnyPointer(metadata.PurpleList)
	case fwschema.NestingModeSet:
		object.NestingMode = metadata.AnyPointer(metadata.PurpleSet)
	case fwschema.NestingModeMap:
		object.NestingMode = metadata.AnyPointer(metadata.PurpleMap)

	default:
		diags.Append(diag.NewErrorDiagnostic("error", "unrecognized nesting mode"))
	}

	attr.GetNestedObject()

	otherStuff, _ := MetadataSchemaAttributes(ctx, nestedAttr)

	object.Attributes = otherStuff

	desc := attr.GetDescription()
	kind := metadata.Plain
	var depMessage *string
	dep := true

	if attr.GetDeprecationMessage() == "" {
		dep = false
	}

	if attr.GetDescription() != "" {
		desc = attr.GetDescription()
		kind = metadata.Plain
	} else if attr.GetMarkdownDescription() != "" {
		desc = attr.GetMarkdownDescription()
		kind = metadata.Markdown
	}

	if attr.GetDeprecationMessage() != "" {
		depMessage = metadata.StringPointer(attr.GetDeprecationMessage())
	}

	ty, _ := attr.GetType().TerraformType(ctx).MarshalJSON() // unmarshal to get to metadata.Type?

	typ := json.RawMessage(ty)

	this := &metadata.Attribute{
		Computed:           metadata.BoolPointer(attr.IsComputed()),
		Deprecated:         metadata.BoolPointer(dep),
		DeprecationMessage: depMessage,
		Description:        metadata.StringPointer(desc),
		DescriptionKind:    metadata.DescriptionPointer(kind), // call both to see if it is plain txt or markdown
		NestedType:         object,                            // make flag to check if true or false and negate for Type
		Optional:           metadata.BoolPointer(attr.IsOptional()),
		PlanModifications:  nil, // TODO: there some type assertion -> resource/schema.Attribute => plan modifier function
		Required:           metadata.BoolPointer(attr.IsRequired()),
		SDKType:            nil, // TODO: Net new logic, ignore for now
		Sensitive:          metadata.BoolPointer(attr.IsSensitive()),
		Type:               metadata.TypePointer(typ), // make flag to check if true or false
		Validations:        nil,                       // list
	}

	return this, diags
}
