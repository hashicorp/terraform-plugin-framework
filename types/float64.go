package types

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ Float64Valuable = Float64{}
)

// Float64Valuable extends attr.Value for float64 value types.
// Implement this interface to create a custom Float64 value type.
type Float64Valuable interface {
	attr.Value

	// ToFloat64Value should convert the value type to a Float64.
	ToFloat64Value(ctx context.Context) (Float64, diag.Diagnostics)
}

// Float64Null creates a Float64 with a null value. Determine whether the value is
// null via the Float64 type IsNull method.
//
// Setting the deprecated Float64 type Null, Unknown, or Value fields after
// creating a Float64 with this function has no effect.
func Float64Null() Float64 {
	return Float64{
		state: attr.ValueStateNull,
	}
}

// Float64Unknown creates a Float64 with an unknown value. Determine whether the
// value is unknown via the Float64 type IsUnknown method.
//
// Setting the deprecated Float64 type Null, Unknown, or Value fields after
// creating a Float64 with this function has no effect.
func Float64Unknown() Float64 {
	return Float64{
		state: attr.ValueStateUnknown,
	}
}

// Float64Value creates a Float64 with a known value. Access the value via the Float64
// type ValueFloat64 method.
//
// Setting the deprecated Float64 type Null, Unknown, or Value fields after
// creating a Float64 with this function has no effect.
func Float64Value(value float64) Float64 {
	return Float64{
		state: attr.ValueStateKnown,
		value: value,
	}
}

func float64Validate(_ context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if in.Type() == nil {
		return diags
	}

	if !in.Type().Equal(tftypes.Number) {
		diags.AddAttributeError(
			path,
			"Float64 Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Expected Number value, received %T with value: %v", in, in),
		)
		return diags
	}

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value *big.Float
	err := in.As(&value)

	if err != nil {
		diags.AddAttributeError(
			path,
			"Float64 Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Cannot convert value to big.Float: %s", err),
		)
		return diags
	}

	_, accuracy := value.Float64()

	if accuracy != 0 {
		diags.AddAttributeError(
			path,
			"Float64 Type Validation Error",
			fmt.Sprintf("Value %s cannot be represented as a 64-bit floating point.", value),
		)
		return diags
	}

	return diags
}

func float64ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return Float64Unknown(), nil
	}

	if in.IsNull() {
		return Float64Null(), nil
	}

	var bigF *big.Float
	err := in.As(&bigF)

	if err != nil {
		return nil, err
	}

	f, accuracy := bigF.Float64()

	if accuracy != 0 {
		return nil, fmt.Errorf("Value %s cannot be represented as a 64-bit floating point.", bigF)
	}

	return Float64Value(f), nil
}

// Float64 represents a 64-bit floating point value, exposed as a float64.
type Float64 struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value float64
}

// Equal returns true if `other` is a Float64 and has the same value as `f`.
func (f Float64) Equal(other attr.Value) bool {
	o, ok := other.(Float64)

	if !ok {
		return false
	}

	if f.state != o.state {
		return false
	}

	if f.state != attr.ValueStateKnown {
		return true
	}

	return f.value == o.value
}

// ToTerraformValue returns the data contained in the Float64 as a tftypes.Value.
func (f Float64) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	switch f.state {
	case attr.ValueStateKnown:
		if err := tftypes.ValidateValue(tftypes.Number, f.value); err != nil {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.Number, f.value), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(tftypes.Number, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Float64 state in ToTerraformValue: %s", f.state))
	}
}

// Type returns a Float64Type.
func (f Float64) Type(ctx context.Context) attr.Type {
	return Float64Type
}

// IsNull returns true if the Float64 represents a null value.
func (f Float64) IsNull() bool {
	return f.state == attr.ValueStateNull
}

// IsUnknown returns true if the Float64 represents a currently unknown value.
func (f Float64) IsUnknown() bool {
	return f.state == attr.ValueStateUnknown
}

// String returns a human-readable representation of the Float64 value.
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (f Float64) String() string {
	if f.IsUnknown() {
		return attr.UnknownValueString
	}

	if f.IsNull() {
		return attr.NullValueString
	}

	return fmt.Sprintf("%f", f.value)
}

// ValueFloat64 returns the known float64 value. If Float64 is null or unknown, returns
// 0.0.
func (f Float64) ValueFloat64() float64 {
	return f.value
}

// ToFloat64Value returns Float64.
func (f Float64) ToFloat64Value(context.Context) (Float64, diag.Diagnostics) {
	return f, nil
}
