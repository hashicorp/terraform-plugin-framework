package attribute

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

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

// NestingMode represents a specific way a group of attributes can be nested.
type NestingMode uint8

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
