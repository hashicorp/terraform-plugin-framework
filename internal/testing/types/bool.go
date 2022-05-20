package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ attr.Type  = BoolType{}
	_ attr.Value = Bool{}
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
		return Bool{
			Bool:      types.Bool{Null: true},
			CreatedBy: t,
		}, nil
	}
	if !in.IsKnown() {
		return Bool{
			Bool:      types.Bool{Unknown: true},
			CreatedBy: t,
		}, nil
	}
	var b bool
	err := in.As(&b)
	if err != nil {
		return nil, err
	}
	return Bool{Bool: types.Bool{Value: b}, CreatedBy: t}, nil
}

type Bool struct {
	types.Bool

	CreatedBy attr.Type
}

func (b Bool) Type(_ context.Context) attr.Type {
	return b.CreatedBy
}

func (b Bool) Equal(o attr.Value) bool {
	ob, ok := o.(Bool)
	if !ok {
		return false
	}
	return b.Bool.Equal(ob.Bool)
}

func (b Bool) IsNull() bool {
	return b.Bool.IsNull()
}

func (b Bool) IsUnknown() bool {
	return b.Bool.IsUnknown()
}
