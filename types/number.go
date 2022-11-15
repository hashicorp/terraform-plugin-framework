package types

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
)

var (
	_ NumberValuable = Number{}
)

// NumberValuable extends attr.Value for number value types.
// Implement this interface to create a custom Number value type.
type NumberValuable interface {
	attr.Value

	// ToNumberValue should convert the value type to a Number.
	ToNumberValue(ctx context.Context) (Number, diag.Diagnostics)
}

// NumberNull creates a Number with a null value. Determine whether the value is
// null via the Number type IsNull method.
//
// Setting the deprecated Number type Null, Unknown, or Value fields after
// creating a Number with this function has no effect.
func NumberNull() Number {
	return Number{
		state: attr.ValueStateNull,
	}
}

// NumberUnknown creates a Number with an unknown value. Determine whether the
// value is unknown via the Number type IsUnknown method.
//
// Setting the deprecated Number type Null, Unknown, or Value fields after
// creating a Number with this function has no effect.
func NumberUnknown() Number {
	return Number{
		state: attr.ValueStateUnknown,
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
		state: attr.ValueStateKnown,
		value: value,
	}
}

func numberValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return NumberUnknown(), nil
	}
	if in.IsNull() {
		return NumberNull(), nil
	}
	n := big.NewFloat(0)
	err := in.As(&n)
	if err != nil {
		return nil, err
	}
	return NumberValue(n), nil
}

// Number represents a number value, exposed as a *big.Float. Numbers can be
// floats or integers.
type Number struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

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
	case attr.ValueStateKnown:
		if n.value == nil {
			return tftypes.NewValue(tftypes.Number, nil), nil
		}

		if err := tftypes.ValidateValue(tftypes.Number, n.value); err != nil {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.Number, n.value), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(tftypes.Number, nil), nil
	case attr.ValueStateUnknown:
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

	if n.state != attr.ValueStateKnown {
		return true
	}

	return n.value.Cmp(o.value) == 0
}

// IsNull returns true if the Number represents a null value.
func (n Number) IsNull() bool {
	return n.state == attr.ValueStateNull
}

// IsUnknown returns true if the Number represents a currently unknown value.
func (n Number) IsUnknown() bool {
	return n.state == attr.ValueStateUnknown
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

	return n.value.String()
}

// ValueBigFloat returns the known *big.Float value. If Number is null or unknown, returns
// 0.0.
func (n Number) ValueBigFloat() *big.Float {
	return n.value
}

// ToNumberValue returns Number.
func (n Number) ToNumberValue(context.Context) (Number, diag.Diagnostics) {
	return n, nil
}
