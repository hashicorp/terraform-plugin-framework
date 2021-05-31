package types

import (
	"context"

	tfsdk "github.com/hashicorp/terraform-plugin-framework"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func boolValueFromTerraform(ctx context.Context, in tftypes.Value) (tfsdk.AttributeValue, error) {
	val := new(Bool)
	err := val.SetTerraformValue(ctx, in)
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
func (b *Bool) ToTerraformValue(_ context.Context) (interface{}, error) {
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
func (b *Bool) Equal(other tfsdk.AttributeValue) bool {
	o, ok := other.(*Bool)
	if !ok {
		return false
	}
	return b.Value == o.Value
}

// SetTerraformValue updates the Bool to match the contents of `val`.
func (b *Bool) SetTerraformValue(ctx context.Context, val tftypes.Value) error {
	if val.IsNull() {
		b.Unknown = false
		b.Value = false
		b.Null = true
		return nil
	}
	if !val.IsKnown() {
		b.Unknown = true
		b.Value = false
		b.Null = false
		return nil
	}
	return val.As(&b.Value)
}
