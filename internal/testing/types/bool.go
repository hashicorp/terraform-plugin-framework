package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.Type = BoolType{}
)

// BoolType is a reimplementation of types.BoolType that can be used as a base
// for other extension types in testing.
type BoolType struct{}

func (t BoolType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, t.String())
}

func (t BoolType) Equal(o attr.Type) bool {
	other, ok := o.(BoolType)
	if !ok {
		return false
	}
	return t == other
}

func (t BoolType) String() string {
	return "testtypes.BoolType"
}

func (t BoolType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.Bool
}

func (t BoolType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.IsNull() {
		return types.Bool{
			Null: true,
		}, nil
	}
	if !in.IsKnown() {
		return types.Bool{
			Unknown: true,
		}, nil
	}
	var b bool
	err := in.As(&b)
	if err != nil {
		return nil, err
	}
	return types.Bool{Value: b}, nil
}
