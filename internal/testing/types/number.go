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
		return Number{
			Number:    types.Number{Unknown: true},
			CreatedBy: t,
		}, nil
	}
	if in.IsNull() {
		return Number{
			Number:    types.Number{Null: true},
			CreatedBy: t,
		}, nil
	}
	n := big.NewFloat(0)
	err := in.As(&n)
	if err != nil {
		return nil, err
	}
	return Number{
		Number:    types.Number{Value: n},
		CreatedBy: t,
	}, nil
}

type Number struct {
	types.Number

	CreatedBy attr.Type
}

func (n Number) Type(_ context.Context) attr.Type {
	return n.CreatedBy
}

func (n Number) Equal(o attr.Value) bool {
	on, ok := o.(Number)
	if !ok {
		return false
	}
	return n.Number.Equal(on.Number)
}
func (n Number) String() string {
	res := "testtypes.Number<"
	if n.Number.Unknown {
		res += "unknown"
	} else if n.Number.Null {
		res += "null"
	} else {
		res += n.Number.Value.String()
	}
	res += ">"
	return res
}
