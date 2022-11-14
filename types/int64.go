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
	_ Int64Valuable = Int64{}
)

// Int64Valuable extends attr.Value for int64 value types.
// Implement this interface to create a custom Int64 value type.
type Int64Valuable interface {
	attr.Value

	// ToInt64Value should convert the value type to an Int64.
	ToInt64Value(ctx context.Context) (Int64, diag.Diagnostics)
}

// Int64Null creates a Int64 with a null value. Determine whether the value is
// null via the Int64 type IsNull method.
//
// Setting the deprecated Int64 type Null, Unknown, or Value fields after
// creating a Int64 with this function has no effect.
func Int64Null() Int64 {
	return Int64{
		state: attr.ValueStateNull,
	}
}

// Int64Unknown creates a Int64 with an unknown value. Determine whether the
// value is unknown via the Int64 type IsUnknown method.
//
// Setting the deprecated Int64 type Null, Unknown, or Value fields after
// creating a Int64 with this function has no effect.
func Int64Unknown() Int64 {
	return Int64{
		state: attr.ValueStateUnknown,
	}
}

// Int64Value creates a Int64 with a known value. Access the value via the Int64
// type ValueInt64 method.
//
// Setting the deprecated Int64 type Null, Unknown, or Value fields after
// creating a Int64 with this function has no effect.
func Int64Value(value int64) Int64 {
	return Int64{
		state: attr.ValueStateKnown,
		value: value,
	}
}

func int64Validate(_ context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if in.Type() == nil {
		return diags
	}

	if !in.Type().Equal(tftypes.Number) {
		diags.AddAttributeError(
			path,
			"Int64 Type Validation Error",
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
			"Int64 Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Cannot convert value to big.Float: %s", err),
		)
		return diags
	}

	if !value.IsInt() {
		diags.AddAttributeError(
			path,
			"Int64 Type Validation Error",
			fmt.Sprintf("Value %s is not an integer.", value),
		)
		return diags
	}

	_, accuracy := value.Int64()

	if accuracy != 0 {
		diags.AddAttributeError(
			path,
			"Int64 Type Validation Error",
			fmt.Sprintf("Value %s cannot be represented as a 64-bit integer.", value),
		)
		return diags
	}

	return diags
}

func int64ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return Int64Unknown(), nil
	}

	if in.IsNull() {
		return Int64Null(), nil
	}

	var bigF *big.Float
	err := in.As(&bigF)

	if err != nil {
		return nil, err
	}

	if !bigF.IsInt() {
		return nil, fmt.Errorf("Value %s is not an integer.", bigF)
	}

	i, accuracy := bigF.Int64()

	if accuracy != 0 {
		return nil, fmt.Errorf("Value %s cannot be represented as a 64-bit integer.", bigF)
	}

	return Int64Value(i), nil
}

// Int64 represents a 64-bit integer value, exposed as an int64.
type Int64 struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value int64
}

// Equal returns true if `other` is an Int64 and has the same value as `i`.
func (i Int64) Equal(other attr.Value) bool {
	o, ok := other.(Int64)

	if !ok {
		return false
	}

	if i.state != o.state {
		return false
	}

	if i.state != attr.ValueStateKnown {
		return true
	}

	return i.value == o.value
}

// ToTerraformValue returns the data contained in the Int64 as a tftypes.Value.
func (i Int64) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	switch i.state {
	case attr.ValueStateKnown:
		if err := tftypes.ValidateValue(tftypes.Number, i.value); err != nil {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.Number, i.value), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(tftypes.Number, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), nil
	default:
		panic(fmt.Sprintf("unhandled Int64 state in ToTerraformValue: %s", i.state))
	}
}

// Type returns a Int64Type.
func (i Int64) Type(ctx context.Context) attr.Type {
	return Int64Type
}

// IsNull returns true if the Int64 represents a null value.
func (i Int64) IsNull() bool {
	return i.state == attr.ValueStateNull
}

// IsUnknown returns true if the Int64 represents a currently unknown value.
func (i Int64) IsUnknown() bool {
	return i.state == attr.ValueStateUnknown
}

// String returns a human-readable representation of the Int64 value.
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (i Int64) String() string {
	if i.IsUnknown() {
		return attr.UnknownValueString
	}

	if i.IsNull() {
		return attr.NullValueString
	}

	return fmt.Sprintf("%d", i.value)
}

// ValueInt64 returns the known float64 value. If Int64 is null or unknown, returns
// 0.0.
func (i Int64) ValueInt64() int64 {
	return i.value
}

// ToInt64Value returns Int64.
func (i Int64) ToInt64Value(context.Context) (Int64, diag.Diagnostics) {
	return i, nil
}
