package tfsdk

import (
	"context"
	"errors"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Attribute defines the constraints and behaviors of a single field in a
// schema. Attributes are the fields that show up in Terraform state files and
// can be used in configuration files.
type Attribute struct {
	// Type indicates what kind of attribute this is. You'll most likely
	// want to use one of the types in the types package.
	//
	// If Type is set, Attributes cannot be.
	Type attr.Type

	// Attributes can have their own, nested attributes. This nested map of
	// attributes behaves exactly like the map of attributes on the Schema
	// type.
	//
	// If Attributes is set, Type cannot be.
	Attributes NestedAttributes

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

	// Required indicates whether the practitioner must enter a value for
	// this attribute or not. Required and Optional cannot both be true,
	// and Required and Computed cannot both be true.
	Required bool

	// Optional indicates whether the practitioner can choose not to enter
	// a value for this attribute or not. Optional and Required cannot both
	// be true.
	Optional bool

	// Computed indicates whether the provider may return its own value for
	// this attribute or not. Required and Computed cannot both be true. If
	// Required and Optional are both false, Computed must be true, and the
	// attribute will be considered "read only" for the practitioner, with
	// only the provider able to set its value.
	Computed bool

	// Sensitive indicates whether the value of this attribute should be
	// considered sensitive data. Setting it to true will obscure the value
	// in CLI output. Sensitive does not impact how values are stored, and
	// practitioners are encouraged to store their state as if the entire
	// file is sensitive.
	Sensitive bool

	// DeprecationMessage defines a message to display to practitioners
	// using this attribute, warning them that it is deprecated and
	// instructing them on what upgrade steps to take.
	DeprecationMessage string
}

// ApplyTerraform5AttributePathStep transparently calls
// ApplyTerraform5AttributePathStep on a.Type or a.Attributes, whichever is
// non-nil. It allows Attributes to be walked using tftypes.Walk and
// tftypes.Transform.
func (a Attribute) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	if a.Type != nil {
		return a.Type.ApplyTerraform5AttributePathStep(step)
	}
	if a.Attributes != nil {
		return a.Attributes.ApplyTerraform5AttributePathStep(step)
	}
	return nil, errors.New("Attribute has no type or nested attributes")
}

// Equal returns true if `a` and `o` should be considered Equal.
func (a Attribute) Equal(o Attribute) bool {
	if a.Type == nil && o.Type != nil {
		return false
	} else if a.Type != nil && o.Type == nil {
		return false
	} else if a.Type != nil && o.Type != nil && !a.Type.Equal(o.Type) {
		return false
	}
	if a.Attributes == nil && o.Attributes != nil {
		return false
	} else if a.Attributes != nil && o.Attributes == nil {
		return false
	} else if a.Attributes != nil && o.Attributes != nil && !a.Attributes.Equal(o.Attributes) {
		return false
	}
	if a.Description != o.Description {
		return false
	}
	if a.MarkdownDescription != o.MarkdownDescription {
		return false
	}
	if a.Required != o.Required {
		return false
	}
	if a.Optional != o.Optional {
		return false
	}
	if a.Computed != o.Computed {
		return false
	}
	if a.Sensitive != o.Sensitive {
		return false
	}
	if a.DeprecationMessage != o.DeprecationMessage {
		return false
	}
	return true
}

// tfprotov6 returns the *tfprotov6.SchemaAttribute equivalent of an
// Attribute. Errors will be tftypes.AttributePathErrors based on
// `path`. `name` is the name of the attribute.
func (a Attribute) tfprotov6SchemaAttribute(ctx context.Context, name string, path *tftypes.AttributePath) (*tfprotov6.SchemaAttribute, error) {
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

	if a.Attributes != nil && len(a.Attributes.GetAttributes()) > 0 && a.Type != nil {
		return nil, path.NewErrorf("can't have both Attributes and Type set")
	}

	if (a.Attributes == nil || len(a.Attributes.GetAttributes()) < 1) && a.Type == nil {
		return nil, path.NewErrorf("must have Attributes or Type set")
	}

	if a.Type != nil {
		schemaAttribute.Type = a.Type.TerraformType(ctx)

		return schemaAttribute, nil
	}

	object := &tfprotov6.SchemaObject{
		MinItems: a.Attributes.GetMinItems(),
		MaxItems: a.Attributes.GetMaxItems(),
	}
	nm := a.Attributes.GetNestingMode()
	switch nm {
	case NestingModeSingle:
		object.Nesting = tfprotov6.SchemaObjectNestingModeSingle
	case NestingModeList:
		object.Nesting = tfprotov6.SchemaObjectNestingModeList
	case NestingModeSet:
		object.Nesting = tfprotov6.SchemaObjectNestingModeSet
	case NestingModeMap:
		object.Nesting = tfprotov6.SchemaObjectNestingModeMap
	default:
		return nil, path.NewErrorf("unrecognized nesting mode %v", nm)
	}

	for nestedName, nestedA := range a.Attributes.GetAttributes() {
		nestedSchemaAttribute, err := nestedA.tfprotov6SchemaAttribute(ctx, nestedName, path.WithAttributeName(nestedName))

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
