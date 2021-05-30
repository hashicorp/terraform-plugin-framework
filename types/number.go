package types

import (
	"context"
	"math/big"

	tfsdk "github.com/hashicorp/terraform-plugin-framework"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func numberValueFromTerraform(ctx context.Context, in tftypes.Value) (tfsdk.AttributeValue, error) {
	n := new(Number)
	err := n.SetTerraformValue(ctx, in)
	return n, err
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
func (n *Number) ToTerraformValue(_ context.Context) (interface{}, error) {
	if n.Null {
		return nil, nil
	}
	if n.Unknown {
		return tftypes.UnknownValue, nil
	}
	return n.Value, nil
}

// Equal must return true if the AttributeValue is considered
// semantically equal to the AttributeValue passed as an argument.
func (n *Number) Equal(other tfsdk.AttributeValue) bool {
	o, ok := other.(*Number)
	if !ok {
		return false
	}
	return n.Value == o.Value
}

func (n *Number) SetTerraformValue(ctx context.Context, in tftypes.Value) error {
	n.Unknown = false
	n.Null = false
	n.Value = nil
	if !in.IsKnown() {
		n.Unknown = true
		return nil
	}
	if in.IsNull() {
		n.Null = true
		return nil
	}
	err := in.As(&n.Value)
	return err
}
