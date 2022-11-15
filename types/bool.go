package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ BoolValuable = Bool{}
)

// BoolValuable extends attr.Value for boolean value types.
// Implement this interface to create a custom Bool value type.
type BoolValuable interface {
	attr.Value

	// ToBoolValue should convert the value type to a Bool.
	ToBoolValue(ctx context.Context) (Bool, diag.Diagnostics)
}

// BoolNull creates a Bool with a null value. Determine whether the value is
// null via the Bool type IsNull method.
//
// Setting the deprecated Bool type Null, Unknown, or Value fields after
// creating a Bool with this function has no effect.
func BoolNull() Bool {
	return Bool{
		state: attr.ValueStateNull,
	}
}

// BoolUnknown creates a Bool with an unknown value. Determine whether the
// value is unknown via the Bool type IsUnknown method.
//
// Setting the deprecated Bool type Null, Unknown, or Value fields after
// creating a Bool with this function has no effect.
func BoolUnknown() Bool {
	return Bool{
		state: attr.ValueStateUnknown,
	}
}

// BoolValue creates a Bool with a known value. Access the value via the Bool
// type ValueBool method.
//
// Setting the deprecated Bool type Null, Unknown, or Value fields after
// creating a Bool with this function has no effect.
func BoolValue(value bool) Bool {
	return Bool{
		state: attr.ValueStateKnown,
		value: value,
	}
}

func boolValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.IsNull() {
		return BoolNull(), nil
	}
	if !in.IsKnown() {
		return BoolUnknown(), nil
	}
	var b bool
	err := in.As(&b)
	if err != nil {
		return nil, err
	}
	return BoolValue(b), nil
}

// Bool represents a boolean value.
type Bool struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value bool
}

// Type returns a BoolType.
func (b Bool) Type(_ context.Context) attr.Type {
	return BoolType
}

// ToTerraformValue returns the data contained in the Bool as a tftypes.Value.
func (b Bool) ToTerraformValue(_ context.Context) (tftypes.Value, error) {
	switch b.state {
	case attr.ValueStateKnown:
		if err := tftypes.ValidateValue(tftypes.Bool, b.value); err != nil {
			return tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.Bool, b.value), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(tftypes.Bool, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Bool state in ToTerraformValue: %s", b.state))
	}
}

// Equal returns true if `other` is a *Bool and has the same value as `b`.
func (b Bool) Equal(other attr.Value) bool {
	o, ok := other.(Bool)

	if !ok {
		return false
	}

	if b.state != o.state {
		return false
	}

	if b.state != attr.ValueStateKnown {
		return true
	}

	return b.value == o.value
}

// IsNull returns true if the Bool represents a null value.
func (b Bool) IsNull() bool {
	return b.state == attr.ValueStateNull
}

// IsUnknown returns true if the Bool represents a currently unknown value.
func (b Bool) IsUnknown() bool {
	return b.state == attr.ValueStateUnknown
}

// String returns a human-readable representation of the Bool value.
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (b Bool) String() string {
	if b.IsUnknown() {
		return attr.UnknownValueString
	}

	if b.IsNull() {
		return attr.NullValueString
	}

	return fmt.Sprintf("%t", b.value)
}

// ValueBool returns the known bool value. If Bool is null or unknown, returns
// false.
func (b Bool) ValueBool() bool {
	return b.value
}

// ToBoolValue returns Bool.
func (b Bool) ToBoolValue(context.Context) (Bool, diag.Diagnostics) {
	return b, nil
}
