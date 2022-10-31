package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ attr.Type  = StringType{}
	_ attr.Value = String{}
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
			InternalString: types.String{Unknown: true},
			CreatedBy:      t,
		}, nil
	}
	if in.IsNull() {
		return String{
			InternalString: types.String{Null: true},
			CreatedBy:      t,
		}, nil
	}
	var s string
	err := in.As(&s)
	if err != nil {
		return nil, err
	}
	return String{
		InternalString: types.String{Value: s},
		CreatedBy:      t,
	}, nil
}

// ValueType returns the Value type.
func (t StringType) ValueType(_ context.Context) attr.Value {
	return String{}
}

type String struct {
	InternalString types.String

	CreatedBy attr.Type
}

func (s String) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	return s.InternalString.ToTerraformValue(ctx)
}

func (s String) ToFrameworkValue() attr.Value {
	return s.InternalString
}

func (s String) Type(_ context.Context) attr.Type {
	return s.CreatedBy
}

func (s String) Equal(o attr.Value) bool {
	os, ok := o.(String)
	if !ok {
		return false
	}
	return s.InternalString.Equal(os.InternalString)
}

func (s String) IsNull() bool {
	return s.InternalString.IsNull()
}

func (s String) IsUnknown() bool {
	return s.InternalString.IsUnknown()
}

func (s String) String() string {
	return s.InternalString.String()
}
