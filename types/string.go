package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func stringValueFromTerraform(_ context.Context, in tftypes.Value) (attr.Value, error) {
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

// String represents a UTF-8 string value.
type String struct {
	// Unknown will be true if the value is not yet known.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	Value string
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
func (s String) Equal(other attr.Value) bool {
	o, ok := other.(String)
	if !ok {
		return false
	}
	return s.Value == o.Value
}
