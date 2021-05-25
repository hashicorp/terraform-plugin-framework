package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func boolValueFromTerraform(_ context.Context, in tftypes.Value) (attr.Value, error) {
	var val Bool
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

// Bool represents a boolean value.
type Bool struct {
	// Unknown will be true if the value is not yet known.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	Value bool
}

// ToTerraformValue returns the data contained in the AttributeValue as
// a Go type that tftypes.NewValue will accept.
func (b Bool) ToTerraformValue(_ context.Context) (interface{}, error) {
	if b.Null {
		return nil, nil
	}
	if b.Unknown {
		return tftypes.UnknownValue, nil
	}
	return b.Value, nil
}

// Equal must return true if the AttributeValue is considered
// semantically equal to the AttributeValue passed as an argument.
func (b Bool) Equal(other attr.Value) bool {
	o, ok := other.(Bool)
	if !ok {
		return false
	}
	return b.Value == o.Value
}
