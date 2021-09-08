package types

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func boolValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.IsNull() {
		return Bool{
			Null: true,
		}, nil
	}
	if !in.IsKnown() {
		return Bool{
			Unknown: true,
		}, nil
	}
	var b bool
	err := in.As(&b)
	if err != nil {
		return nil, err
	}
	return Bool{Value: b}, nil
}

var _ attr.Value = Bool{}

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

// Type returns a BoolType.
func (b Bool) Type(_ context.Context) attr.Type {
	return BoolType
}

// ToTerraformValue returns the data contained in the *Bool as a bool. If
// Unknown is true, it returns a tftypes.UnknownValue. If Null is true, it
// returns nil.
func (b Bool) ToTerraformValue(_ context.Context) (interface{}, error) {
	if b.Null {
		return nil, nil
	}
	if b.Unknown {
		return tftypes.UnknownValue, nil
	}
	return b.Value, nil
}

// Equal returns true if `other` is a *Bool and has the same value as `b`.
func (b Bool) Equal(other attr.Value) bool {
	o, ok := other.(Bool)
	if !ok {
		return false
	}
	if b.Unknown != o.Unknown {
		return false
	}
	if b.Null != o.Null {
		return false
	}
	return b.Value == o.Value
}
