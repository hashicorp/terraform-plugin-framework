package types

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func int64Validate(_ context.Context, in tftypes.Value, path *tftypes.AttributePath) diag.Diagnostics {
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
		return Int64{Unknown: true}, nil
	}

	if in.IsNull() {
		return Int64{Null: true}, nil
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

	return Int64{Value: i}, nil
}

var _ attr.Value = Int64{}

// Int64 represents a 64-bit integer value, exposed as an int64.
type Int64 struct {
	// Unknown will be true if the value is not yet known.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	Value int64
}

// Equal returns true if `other` is an Int64 and has the same value as `i`.
func (i Int64) Equal(other attr.Value) bool {
	o, ok := other.(Int64)

	if !ok {
		return false
	}

	if i.Unknown != o.Unknown {
		return false
	}

	if i.Null != o.Null {
		return false
	}

	return i.Value == o.Value
}

// ToTerraformValue returns the data contained in the Int64 as a tftypes.Value.
func (i Int64) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	if i.Null {
		return tftypes.NewValue(tftypes.Number, nil), nil
	}

	if i.Unknown {
		return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), nil
	}

	bf := new(big.Float).SetInt64(i.Value)
	if err := tftypes.ValidateValue(tftypes.Number, bf); err != nil {
		return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), err
	}
	return tftypes.NewValue(tftypes.Number, bf), nil
}

// Type returns a NumberType.
func (i Int64) Type(ctx context.Context) attr.Type {
	return Int64Type
}
