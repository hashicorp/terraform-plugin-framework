package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.Value = Bool{}
)

// BoolNull creates a Bool with a null value. Determine whether the value is
// null via the Bool type IsNull method.
//
// Setting the deprecated Bool type Null, Unknown, or Value fields after
// creating a Bool with this function has no effect.
func BoolNull() Bool {
	return Bool{
		state: valueStateNull,
	}
}

// BoolUnknown creates a Bool with an unknown value. Determine whether the
// value is unknown via the Bool type IsUnknown method.
//
// Setting the deprecated Bool type Null, Unknown, or Value fields after
// creating a Bool with this function has no effect.
func BoolUnknown() Bool {
	return Bool{
		state: valueStateUnknown,
	}
}

// BoolValue creates a Bool with a known value. Access the value via the Bool
// type ValueBool method.
//
// Setting the deprecated Bool type Null, Unknown, or Value fields after
// creating a Bool with this function has no effect.
func BoolValue(value bool) Bool {
	return Bool{
		state: valueStateKnown,
		value: value,
	}
}

func boolValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.IsNull() {
		return Bool{
			Null:  true,
			state: valueStateDeprecated,
		}, nil
	}
	if !in.IsKnown() {
		return Bool{
			Unknown: true,
			state:   valueStateDeprecated,
		}, nil
	}
	var b bool
	err := in.As(&b)
	if err != nil {
		return nil, err
	}
	return Bool{
		Value: b,
		state: valueStateDeprecated,
	}, nil
}

// Bool represents a boolean value.
type Bool struct {
	// Unknown will be true if the value is not yet known.
	//
	// If the Bool was created with the BoolValue, BoolNull, or BoolUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the BoolUnknown function to create an unknown Bool
	// value or use the IsUnknown method to determine whether the Bool value
	// is unknown instead.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	//
	// If the Bool was created with the BoolValue, BoolNull, or BoolUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the BoolNull function to create a null Bool value or
	// use the IsNull method to determine whether the Bool value is null
	// instead.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	//
	// If the Bool was created with the BoolValue, BoolNull, or BoolUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the BoolValue function to create a known Bool value or
	// use the ValueBool method to retrieve the Bool value instead.
	Value bool

	// state represents whether the Bool is null, unknown, or known.
	state valueState

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
	case valueStateDeprecated:
		if b.Null {
			return tftypes.NewValue(tftypes.Bool, nil), nil
		}
		if b.Unknown {
			return tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue), nil
		}
		if err := tftypes.ValidateValue(tftypes.Bool, b.Value); err != nil {
			return tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue), err
		}
		return tftypes.NewValue(tftypes.Bool, b.Value), nil
	case valueStateKnown:
		if err := tftypes.ValidateValue(tftypes.Bool, b.value); err != nil {
			return tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.Bool, b.value), nil
	case valueStateNull:
		return tftypes.NewValue(tftypes.Bool, nil), nil
	case valueStateUnknown:
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
	if b.state == valueStateKnown {
		return b.value == o.value
	}
	if b.Unknown != o.Unknown {
		return false
	}
	if b.Null != o.Null {
		return false
	}
	return b.Value == o.Value
}

// IsNull returns true if the Bool represents a null value.
func (b Bool) IsNull() bool {
	if b.state == valueStateNull {
		return true
	}

	return b.state == valueStateDeprecated && b.Null
}

// IsUnknown returns true if the Bool represents a currently unknown value.
func (b Bool) IsUnknown() bool {
	if b.state == valueStateUnknown {
		return true
	}

	return b.state == valueStateDeprecated && b.Unknown
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

	if b.state == valueStateKnown {
		return fmt.Sprintf("%t", b.value)
	}

	return fmt.Sprintf("%t", b.Value)
}

// ValueBool returns the known bool value. If Bool is null or unknown, returns
// false.
func (b Bool) ValueBool() bool {
	if b.state == valueStateDeprecated {
		return b.Value
	}

	return b.value
}
