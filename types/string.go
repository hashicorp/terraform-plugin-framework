package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attribute"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type StringType struct{}

// TerraformType returns the tftypes.Type that should be used to
// represent this type. This constrains what user input will be
// accepted and what kind of data can be set in state. The framework
// will use this to translate the AttributeType to something Terraform
// can understand.
func (s StringType) TerraformType(ctx context.Context) tftypes.Type {
	return tftypes.String
}

// Validate returns any warnings or errors about the value that is
// being used to populate the AttributeType. It is generally used to
// check the data format and ensure that it complies with the
// requirements of the AttributeType.
//
// TODO: don't use tfprotov6.Diagnostic, use our type
func (s StringType) Validate(_ context.Context, _ tftypes.Value) []*tfprotov6.Diagnostic {
	return nil
}

// Description returns a practitioner-friendly explanation of the type
// and the constraints of the data it accepts and returns. It will be
// combined with the Description associated with the Attribute.
func (s StringType) Description(_ context.Context, _ attribute.StringKind) string {
	return ""
}

// ValueFromTerraform returns an AttributeValue given a tftypes.Value.
// This is meant to convert the tftypes.Value into a more convenient Go
// type for the provider to consume the data with.
func (s StringType) ValueFromTerraform(_ context.Context, in tftypes.Value) (attribute.AttributeValue, error) {
	var val String
	if !in.IsKnown() {
		val.Unknown = true
		return val, nil
	}
	if in.IsNull() {
		val.Null = true
		return val, nil
	}
	err := in.As(&val.Value)
	return val, err
}

type String struct {
	Unknown bool
	Null    bool
	Value   string
}

// ToTerraformValue returns the data contained in the AttributeValue as
// a Go type that tftypes.NewValue will accept.
func (s String) ToTerraformValue(_ context.Context) (interface{}, error) {
	if s.Null {
		return nil, nil
	}
	if s.Unknown {
		return tftypes.UnknownValue, nil
	}
	return s.Value, nil
}

// Equal must return true if the AttributeValue is considered
// semantically equal to the AttributeValue passed as an argument.
func (s String) Equal(other attribute.AttributeValue) bool {
	o, ok := other.(String)
	if !ok {
		return false
	}
	return s.Value == o.Value
}
