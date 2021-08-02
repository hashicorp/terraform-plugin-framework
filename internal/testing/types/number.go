package types

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.Type = NumberType{}
)

// NumberType is a reimplementation of types.NumberType that can be used as a base
// for other extension types in testing.
type NumberType struct{}

func (t NumberType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, t.String())
}

func (t NumberType) Equal(o attr.Type) bool {
	other, ok := o.(NumberType)
	if !ok {
		return false
	}
	return t == other
}

func (t NumberType) String() string {
	return "testtypes.NumberType"
}

func (t NumberType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.Number
}

func (t NumberType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return types.Number{Unknown: true}, nil
	}
	if in.IsNull() {
		return types.Number{Null: true}, nil
	}
	n := big.NewFloat(0)
	err := in.As(&n)
	if err != nil {
		return nil, err
	}
	return types.Number{Value: n}, nil
}
