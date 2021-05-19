package attr

import (
	"context"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Type defines an interface for describing a kind of attribute. Types are
// collections of constraints and behaviors such that they can be reused on
// multiple attributes easily.
type Type interface {
	// TerraformType returns the tftypes.Type that should be used to
	// represent this type. This constrains what user input will be
	// accepted and what kind of data can be set in state. The framework
	// will use this to translate the Type to something Terraform
	// can understand.
	TerraformType(context.Context) tftypes.Type

	// Validate returns any warnings or errors about the value that is
	// being used to populate the Type. It is generally used to
	// check the data format and ensure that it complies with the
	// requirements of the Type.
	//
	// TODO: don't use tfprotov6.Diagnostic, use our type
	Validate(context.Context, tftypes.Value) []*tfprotov6.Diagnostic

	// Description returns a practitioner-friendly explanation of the type
	// and the constraints of the data it accepts and returns. It will be
	// combined with the Description associated with the Attribute.
	Description(context.Context) string

	// ValueFromTerraform returns an Value given a tftypes.Value. This is
	// meant to convert the tftypes.Value into a more convenient Go type
	// for the provider to consume the data with.
	ValueFromTerraform(context.Context, tftypes.Value) (Value, error)
}
