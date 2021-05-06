package tf

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Schema is used to define the shape of practitioner-provider information,
// like resources, data sources, and providers. Think of it as a type
// definition, but for Terraform.
type Schema struct {
	// Attributes are the fields inside the resource, provider, or data
	// source that the schema is defining. The map key should be the name
	// of the attribute, and the body defines how it behaves. Names must
	// only contain lowercase letters, numbers, and underscores.
	Attributes map[string]Attribute

	// Version indicates the current version of the schema. Schemas are
	// versioned to help with automatic upgrade process. Whenever you have
	// a change in the schema you'd like to provide a manual migration for,
	// you should increment that schema's version by one.
	Version int64
}

// Attribute defines the constraints and behaviors of a single field in a
// schema. Attributes are the fields that show up in Terraform state files and
// can be used in configuration files.
type Attribute struct {
	// Type indicates what kind of attribute this is. You'll most likely
	// want to use one of the types in the types package.
	//
	// If Type is set, Attributes cannot be.
	Type AttributeType

	// Attributes can have their own, nested attributes. This nested map of
	// attributes behaves exactly like the map of attributes on the Schema
	// type.
	//
	// If Attributes is set, Type cannot be.
	//
	// TODO: support different nesting modes
	// TODO: do we need MaxItems/MinItems? Can we just make those weird
	// validation helpers?
	Attributes map[string]Attribute

	// Description is used in various tooling, like the documentation
	// generator and the language server, to give practitioners more
	// information about what this attribute is, what it's for, and how it
	// should be used.
	Description string

	// DescriptionKind indicates the type of text formatting that
	// Description uses. It should be Markdown or PlainText.
	//
	// TODO: come up with a better interface for this, this is weird.
	DescriptionKind StringKind

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

// AttributeType defines an interface for describing a kind of attribute.
// AttributeTypes are collections of constraints and behaviors such that they
// can be reused on multiple attributes easily.
type AttributeType interface {
	// TerraformType returns the tftypes.Type that should be used to
	// represent this type. This constrains what user input will be
	// accepted and what kind of data can be set in state. The framework
	// will use this to translate the AttributeType to something Terraform
	// can understand.
	TerraformType(context.Context) tftypes.Type

	// Validate returns any warnings or errors about the value that is
	// being used to populate the AttributeType. It is generally used to
	// check the data format and ensure that it complies with the
	// requirements of the AttributeType.
	//
	// TODO: don't use tfprotov6.Diagnostic, use our type
	Validate(context.Context, tftypes.Value) []*tfprotov6.Diagnostic

	// Description returns a practitioner-friendly explanation of the type
	// and the constraints of the data it accepts and returns. It will be
	// combined with the Description associated with the Attribute.
	Description(context.Context, StringKind) string

	// ValueFromTerraform returns an AttributeValue given a tftypes.Value.
	// This is meant to convert the tftypes.Value into a more convenient Go
	// type for the provider to consume the data with.
	ValueFromTerraform(context.Context, tftypes.Value) (AttributeValue, error)
}

// StringKind represents a kind of string formatting.
type StringKind uint8

// AttributeValue defines an interface for describing data associated with an
// attribute. AttributeValues allow provider developers to specify data in a
// convenient format, and have it transparently be converted to formats
// Terraform understands.
type AttributeValue interface {
	// ToTerraformValue returns the data contained in the AttributeValue as
	// a Go type that tftypes.NewValue will accept.
	ToTerraformValue(context.Context) (interface{}, error)

	// Equal must return true if the AttributeValue is considered
	// semantically equal to the AttributeValue passed as an argument.
	Equal(AttributeValue) bool
}
