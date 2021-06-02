package schema

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

func (a Attribute) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	if a.Type != nil {
		return a.Type.ApplyTerraform5AttributePathStep(step)
	}
	if a.Attributes != nil {
		return a.Attributes.ApplyTerraform5AttributePathStep(step)
	}
}
