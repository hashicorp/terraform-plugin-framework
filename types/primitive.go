package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type primitive uint8

const (
	// StringType represents a UTF-8 string type.
	StringType primitive = iota

	// NumberType represents a number type, either an integer or a float.
	NumberType

	// BoolType represents a boolean type.
	BoolType

	// Int64Type represents a 64-bit integer.
	Int64Type

	// Float64Type represents a 64-bit floating point.
	Float64Type
)

var (
	_ StringTypable  = StringType
	_ NumberTypable  = NumberType
	_ BoolTypable    = BoolType
	_ Int64Typable   = Int64Type
	_ Float64Typable = Float64Type
)

type StringTypable interface {
	attr.Type

	ValueFromString(context.Context, String) (attr.Value, diag.Diagnostics)
}

type NumberTypable interface {
	attr.Type

	ValueFromNumber(context.Context, Number) (attr.Value, diag.Diagnostics)
}

type BoolTypable interface {
	attr.Type

	ValueFromBool(context.Context, Bool) (attr.Value, diag.Diagnostics)
}

type Int64Typable interface {
	xattr.TypeWithValidate

	ValueFromInt64(context.Context, Int64) (attr.Value, diag.Diagnostics)
}

type Float64Typable interface {
	xattr.TypeWithValidate

	ValueFromFloat64(context.Context, Float64) (attr.Value, diag.Diagnostics)
}

// ValueFromString returns an attr.Value given a String.
func (p primitive) ValueFromString(_ context.Context, s String) (attr.Value, diag.Diagnostics) {
	return s, nil
}

// ValueFromNumber returns an attr.Value given a Number.
func (p primitive) ValueFromNumber(_ context.Context, n Number) (attr.Value, diag.Diagnostics) {
	return n, nil
}

// ValueFromBool returns an attr.Value given a Bool.
func (p primitive) ValueFromBool(_ context.Context, b Bool) (attr.Value, diag.Diagnostics) {
	return b, nil
}

// ValueFromInt64 returns an attr.Value given a Int64.
func (p primitive) ValueFromInt64(_ context.Context, i Int64) (attr.Value, diag.Diagnostics) {
	return i, nil
}

// ValueFromFloat64 returns an attr.Value given a Float64.
func (p primitive) ValueFromFloat64(_ context.Context, f Float64) (attr.Value, diag.Diagnostics) {
	return f, nil
}

func (p primitive) String() string {
	switch p {
	case StringType:
		return "types.StringType"
	case NumberType:
		return "types.NumberType"
	case BoolType:
		return "types.BoolType"
	case Int64Type:
		return "types.Int64Type"
	case Float64Type:
		return "types.Float64Type"
	default:
		return fmt.Sprintf("unknown primitive %d", p)
	}
}

// TerraformType returns the tftypes.Type that should be used to represent this
// type. This constrains what user input will be accepted and what kind of data
// can be set in state. The framework will use this to translate the Type to
// something Terraform can understand.
func (p primitive) TerraformType(_ context.Context) tftypes.Type {
	switch p {
	case StringType:
		return tftypes.String
	case NumberType, Int64Type, Float64Type:
		return tftypes.Number
	case BoolType:
		return tftypes.Bool
	default:
		panic(fmt.Sprintf("unknown primitive %d", p))
	}
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to
// convert the tftypes.Value into a more convenient Go type for the provider to
// consume the data with.
func (p primitive) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	switch p {
	case StringType:
		return stringValueFromTerraform(ctx, in)
	case NumberType:
		return numberValueFromTerraform(ctx, in)
	case BoolType:
		return boolValueFromTerraform(ctx, in)
	case Int64Type:
		return int64ValueFromTerraform(ctx, in)
	case Float64Type:
		return float64ValueFromTerraform(ctx, in)
	default:
		panic(fmt.Sprintf("unknown primitive %d", p))
	}
}

// ValueType returns the Value type.
func (p primitive) ValueType(_ context.Context) attr.Value {
	// These Value do not need to be valid.
	switch p {
	case BoolType:
		return Bool{}
	case Float64Type:
		return Float64{}
	case Int64Type:
		return Int64{}
	case NumberType:
		return Number{}
	case StringType:
		return String{}
	default:
		panic(fmt.Sprintf("unknown primitive %d", p))
	}
}

// Equal returns true if `o` is also a primitive, and is the same type of
// primitive as `p`.
func (p primitive) Equal(o attr.Type) bool {
	other, ok := o.(primitive)
	if !ok {
		return false
	}
	switch p {
	case StringType, NumberType, BoolType, Int64Type, Float64Type:
		return p == other
	default:
		// unrecognized types are never equal to anything.
		return false
	}
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// type.
func (p primitive) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, p.String())
}

// Validate implements type validation.
func (p primitive) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	switch p {
	case Int64Type:
		diags.Append(int64Validate(ctx, in, path)...)
	case Float64Type:
		diags.Append(float64Validate(ctx, in, path)...)
	}

	return diags
}
