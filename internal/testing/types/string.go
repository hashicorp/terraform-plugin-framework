package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ attr.Type = StringType{}
)

// StringType is a reimplementation of types.StringType that can be used as a base
// for other extension types in testing.
type StringType struct{}

func (t StringType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, t.String())
}

func (t StringType) Equal(o attr.Type) bool {
	other, ok := o.(StringType)
	if !ok {
		return false
	}
	return t == other
}

func (t StringType) String() string {
	return "testtypes.StringType"
}

func (t StringType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.String
}

func (t StringType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return String{
			String:    types.String{Unknown: true},
			CreatedBy: t,
		}, nil
	}
	if in.IsNull() {
		return String{
			String:    types.String{Null: true},
			CreatedBy: t,
		}, nil
	}
	var s string
	err := in.As(&s)
	if err != nil {
		return nil, err
	}
	return String{
		String:    types.String{Value: s},
		CreatedBy: t,
	}, nil
}

type String struct {
	types.String
	CreatedBy attr.Type
}

func (s String) Type(_ context.Context) attr.Type {
	return s.CreatedBy
}

func (s String) Equal(o attr.Value) bool {
	os, ok := o.(String)
	if !ok {
		return false
	}
	return s.String.Equal(os.String)
}
