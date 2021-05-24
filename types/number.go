package types

import (
	"context"
	"math/big"

	tfsdk "github.com/hashicorp/terraform-plugin-framework"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func numberValueFromTerraform(_ context.Context, in tftypes.Value) (tfsdk.AttributeValue, error) {
	var val Number
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

// Number represents a number value, exposed as a *big.Float. Numbers can be
// floats or integers.
type Number struct {
	// Unknown will be true if the value is not yet known.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	Value *big.Float
}

// ToTerraformValue returns the data contained in the AttributeValue as
// a Go type that tftypes.NewValue will accept.
func (s Number) ToTerraformValue(_ context.Context) (interface{}, error) {
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
func (s Number) Equal(other tfsdk.AttributeValue) bool {
	o, ok := other.(Number)
	if !ok {
		return false
	}
	return s.Value == o.Value
}
