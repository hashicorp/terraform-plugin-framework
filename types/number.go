package types

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.Value = Number{}
)

// NumberNull creates a Number with a null value. Determine whether the value is
// null via the Number type IsNull method.
//
// Setting the deprecated Number type Null, Unknown, or Value fields after
// creating a Number with this function has no effect.
func NumberNull() Number {
	return Number{
		state: valueStateNull,
	}
}

// NumberUnknown creates a Number with an unknown value. Determine whether the
// value is unknown via the Number type IsUnknown method.
//
// Setting the deprecated Number type Null, Unknown, or Value fields after
// creating a Number with this function has no effect.
func NumberUnknown() Number {
	return Number{
		state: valueStateUnknown,
	}
}

// NumberValue creates a Number with a known value. Access the value via the Number
// type ValueBigFloat method. If the given value is nil, a null Number is created.
//
// Setting the deprecated Number type Null, Unknown, or Value fields after
// creating a Number with this function has no effect.
func NumberValue(value *big.Float) Number {
	if value == nil {
		return NumberNull()
	}

	return Number{
		state: valueStateKnown,
		value: value,
	}
}

func numberValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return Number{
			Unknown: true,
			state:   valueStateDeprecated,
		}, nil
	}
	if in.IsNull() {
		return Number{
			Null:  true,
			state: valueStateDeprecated,
		}, nil
	}
	n := big.NewFloat(0)
	err := in.As(&n)
	if err != nil {
		return nil, err
	}
	return Number{
		Value: n,
		state: valueStateDeprecated,
	}, nil
}

// Number represents a number value, exposed as a *big.Float. Numbers can be
// floats or integers.
type Number struct {
	// Unknown will be true if the value is not yet known.
	//
	// If the Number was created with the NumberValue, NumberNull, or NumberUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the NumberUnknown function to create an unknown Number
	// value or use the IsUnknown method to determine whether the Number value
	// is unknown instead.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	//
	// If the Number was created with the NumberValue, NumberNull, or NumberUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the NumberNull function to create a null Number value or
	// use the IsNull method to determine whether the Number value is null
	// instead.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	//
	// If the Number was created with the NumberValue, NumberNull, or NumberUnknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the NumberValue function to create a known Number value or
	// use the ValueBigFloat method to retrieve the Number value instead.
	Value *big.Float

	// state represents whether the Number is null, unknown, or known. During the
	// exported field deprecation period, this state can also be "deprecated",
	// which remains the zero-value for compatibility to ensure exported field
	// updates take effect. The zero-value will be changed to null in a future
	// version.
	state valueState

	// value contains the known value, if not null or unknown.
	value *big.Float
}

// Type returns a NumberType.
func (n Number) Type(_ context.Context) attr.Type {
	return NumberType
}

// ToTerraformValue returns the data contained in the Number as a tftypes.Value.
func (n Number) ToTerraformValue(_ context.Context) (tftypes.Value, error) {
	switch n.state {
	case valueStateDeprecated:
		if n.Null {
			return tftypes.NewValue(tftypes.Number, nil), nil
		}
		if n.Unknown {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), nil
		}
		if n.Value == nil {
			return tftypes.NewValue(tftypes.Number, nil), nil
		}
		if err := tftypes.ValidateValue(tftypes.Number, n.Value); err != nil {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), err
		}
		return tftypes.NewValue(tftypes.Number, n.Value), nil
	case valueStateKnown:
		if n.value == nil {
			return tftypes.NewValue(tftypes.Number, nil), nil
		}

		if err := tftypes.ValidateValue(tftypes.Number, n.value); err != nil {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.Number, n.value), nil
	case valueStateNull:
		return tftypes.NewValue(tftypes.Number, nil), nil
	case valueStateUnknown:
		return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Number state in ToTerraformValue: %s", n.state))
	}
}

// Equal returns true if `other` is a Number and has the same value as `n`.
func (n Number) Equal(other attr.Value) bool {
	o, ok := other.(Number)
	if !ok {
		return false
	}
	if n.state != o.state {
		return false
	}
	if n.state == valueStateKnown {
		return n.value.Cmp(o.value) == 0
	}
	if n.Unknown != o.Unknown {
		return false
	}
	if n.Null != o.Null {
		return false
	}
	if n.Value == nil && o.Value == nil {
		return true
	}
	if n.Value == nil || o.Value == nil {
		return false
	}
	return n.Value.Cmp(o.Value) == 0
}

// IsNull returns true if the Number represents a null value.
func (n Number) IsNull() bool {
	if n.state == valueStateNull {
		return true
	}

	if n.state == valueStateDeprecated && n.Null {
		return true
	}

	return n.state == valueStateDeprecated && (!n.Unknown && n.Value == nil)
}

// IsUnknown returns true if the Number represents a currently unknown value.
func (n Number) IsUnknown() bool {
	if n.state == valueStateUnknown {
		return true
	}

	return n.state == valueStateDeprecated && n.Unknown
}

// String returns a human-readable representation of the Number value.
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (n Number) String() string {
	if n.IsUnknown() {
		return attr.UnknownValueString
	}

	if n.IsNull() {
		return attr.NullValueString
	}

	if n.state == valueStateKnown {
		return n.value.String()
	}

	return n.Value.String()
}

// ValueBigFloat returns the known *big.Float value. If Number is null or unknown, returns
// 0.0.
func (n Number) ValueBigFloat() *big.Float {
	if n.state == valueStateDeprecated {
		return n.Value
	}

	return n.value
}
