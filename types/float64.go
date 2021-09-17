package types

import (
	"context"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func float64ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return Float64{Unknown: true}, nil
	}

	if in.IsNull() {
		return Float64{Null: true}, nil
	}

	var bigF *big.Float
	err := in.As(&bigF)

	if err != nil {
		return nil, err
	}

	// Validation is handled by Validate method.
	f, _ := bigF.Float64()

	return Float64{Value: f}, nil
}

var _ attr.Value = Float64{}

// Float64 represents a 64-bit floating point value, exposed as an float64.
type Float64 struct {
	// Unknown will be true if the value is not yet known.
	Unknown bool

	// Null will be true if the value was not set, or was explicitly set to
	// null.
	Null bool

	// Value contains the set value, as long as Unknown and Null are both
	// false.
	Value float64
}

// Equal returns true if `other` is an Float64 and has the same value as `i`.
func (f Float64) Equal(other attr.Value) bool {
	o, ok := other.(Float64)

	if !ok {
		return false
	}

	if f.Unknown != o.Unknown {
		return false
	}

	if f.Null != o.Null {
		return false
	}

	return f.Value == o.Value
}

// ToTerraformValue returns the data contained in the Float64 as a float64.
// If Unknown is true, it returns a tftypes.UnknownValue. If Null is true, it
// returns nil.
func (f Float64) ToTerraformValue(ctx context.Context) (interface{}, error) {
	if f.Null {
		return nil, nil
	}

	if f.Unknown {
		return tftypes.UnknownValue, nil
	}

	return big.NewFloat(f.Value), nil
}

// Type returns a NumberType.
func (f Float64) Type(ctx context.Context) attr.Type {
	return Int64Type
}
