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
	_ attr.Value = Float64{}
)

// Float64Null creates a Float64 with a null value. Determine whether the value is
// null via the Float64 type IsNull method.
//
// Setting the deprecated Float64 type Null, Unknown, or Value fields after
// creating a Float64 with this function has no effect.
func Float64Null() Float64 {
	return Float64{
		state: valueStateNull,
	}
}

// Float64Unknown creates a Float64 with an unknown value. Determine whether the
// value is unknown via the Float64 type IsUnknown method.
//
// Setting the deprecated Float64 type Null, Unknown, or Value fields after
// creating a Float64 with this function has no effect.
func Float64Unknown() Float64 {
	return Float64{
		state: valueStateUnknown,
	}
}

// Float64Value creates a Float64 with a known value. Access the value via the Float64
// type ValueFloat64 method.
//
// Setting the deprecated Float64 type Null, Unknown, or Value fields after
// creating a Float64 with this function has no effect.
func Float64Value(value float64) Float64 {
	return Float64{
		state: valueStateKnown,
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
		return Float64{
			Unknown: true,
			state:   valueStateDeprecated,
		}, nil
	}

	if in.IsNull() {
		return Float64{
			Null:  true,
			state: valueStateDeprecated,
		}, nil
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

	return Float64{
		Value: f,
		state: valueStateDeprecated,
	}, nil
}

// Float64 represents a 64-bit floating point value, exposed as a float64.
type Float64 struct {
	// Unknown will be true if the value is not yet known.
	//
	// If the Float64 was created with the Float64Value, Float64Null, or Float64Unknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the Float64Unknown function to create an unknown Float64
	// value or use the IsUnknown method to determine whether the Float64 value
	// is unknown instead.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	//
	// If the Float64 was created with the Float64Value, Float64Null, or Float64Unknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the Float64Null function to create a null Float64 value or
	// use the IsNull method to determine whether the Float64 value is null
	// instead.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	//
	// If the Float64 was created with the Float64Value, Float64Null, or Float64Unknown
	// functions, changing this field has no effect.
	//
	// Deprecated: Use the Float64Value function to create a known Float64 value or
	// use the ValueFloat64 method to retrieve the Float64 value instead.
	Value float64

	// state represents whether the Float64 is null, unknown, or known. During the
	// exported field deprecation period, this state can also be "deprecated",
	// which remains the zero-value for compatibility to ensure exported field
	// updates take effect. The zero-value will be changed to null in a future
	// version.
	state valueState

	// value contains the known value, if not null or unknown.
	value float64
}

func (f Float64) ToFrameworkValue() attr.Value {
	return f
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

	if f.state == valueStateKnown {
		return f.value == o.value
	}

	if f.Unknown != o.Unknown {
		return false
	}

	if f.Null != o.Null {
		return false
	}

	return f.Value == o.Value
}

// ToTerraformValue returns the data contained in the Float64 as a tftypes.Value.
func (f Float64) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	switch f.state {
	case valueStateDeprecated:
		if f.Null {
			return tftypes.NewValue(tftypes.Number, nil), nil
		}
		if f.Unknown {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), nil
		}
		if err := tftypes.ValidateValue(tftypes.Number, f.Value); err != nil {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), err
		}
		return tftypes.NewValue(tftypes.Number, f.Value), nil
	case valueStateKnown:
		if err := tftypes.ValidateValue(tftypes.Number, f.value); err != nil {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.Number, f.value), nil
	case valueStateNull:
		return tftypes.NewValue(tftypes.Number, nil), nil
	case valueStateUnknown:
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
	if f.state == valueStateNull {
		return true
	}

	return f.state == valueStateDeprecated && f.Null
}

// IsUnknown returns true if the Float64 represents a currently unknown value.
func (f Float64) IsUnknown() bool {
	if f.state == valueStateUnknown {
		return true
	}

	return f.state == valueStateDeprecated && f.Unknown
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

	if f.state == valueStateKnown {
		return fmt.Sprintf("%f", f.value)
	}

	return fmt.Sprintf("%f", f.Value)
}

// ValueFloat64 returns the known float64 value. If Float64 is null or unknown, returns
// 0.0.
func (f Float64) ValueFloat64() float64 {
	if f.state == valueStateDeprecated {
		return f.Value
	}

	return f.value
}
