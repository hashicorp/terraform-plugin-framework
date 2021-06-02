package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type primitive uint8

const (
	// StringType represents a UTF-8 string type.
	StringType primitive = iota

	// NumberType represents a number type, either an integer or a float.
	NumberType

	// BoolType represents a boolean type.
	BoolType
)

var (
	_ attr.Type = StringType
	_ attr.Type = NumberType
	_ attr.Type = BoolType
)

func (p primitive) String() string {
	switch p {
	case StringType:
		return "types.StringType"
	case NumberType:
		return "types.NumberType"
	case BoolType:
		return "types.BoolType"
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
	case NumberType:
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
	case StringType, NumberType, BoolType:
		return p == other
	default:
		// unrecognized types are never equal to anything.
		return false
	}
}

func (p primitive) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, p.String())
}
