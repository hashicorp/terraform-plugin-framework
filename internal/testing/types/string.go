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
			Str:       types.String{Unknown: true},
			CreatedBy: t,
		}, nil
	}
	if in.IsNull() {
		return String{
			Str:       types.String{Null: true},
			CreatedBy: t,
		}, nil
	}
	var s string
	err := in.As(&s)
	if err != nil {
		return nil, err
	}
	return String{
		Str:       types.String{Value: s},
		CreatedBy: t,
	}, nil
}

type String struct {
	Str       types.String
	CreatedBy attr.Type
}

func (s String) Type(_ context.Context) attr.Type {
	return s.CreatedBy
}

func (s String) ToTerraformValue(ctx context.Context) (interface{}, error) {
	return s.Str.ToTerraformValue(ctx)
}

func (s String) Equal(o attr.Value) bool {
	os, ok := o.(String)
	if !ok {
		return false
	}
	return s.Str.Equal(os.Str)
}

func (s String) String() string {
	res := "testtypes.String<"
	if s.Str.Unknown {
		res += "unknown"
	} else if s.Str.Null {
		res += "null"
	} else {
		res += "\"" + s.Str.Value + "\""
	}
	res += ">"
	return res
}
