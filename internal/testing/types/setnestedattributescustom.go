package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/fwschema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ types.SetTypable  = SetNestedAttributesCustomTypeType{}
	_ types.SetValuable = &SetNestedAttributesCustomValue{}
)

type SetNestedAttributesCustomType struct {
	fwschema.NestedAttributes
}

func (t SetNestedAttributesCustomType) Type() attr.Type {
	return SetNestedAttributesCustomTypeType{
		t.NestedAttributes.Type().(types.SetType),
	}
}

type SetNestedAttributesCustomTypeType struct {
	types.SetType
}

func (tt SetNestedAttributesCustomTypeType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	val, err := tt.SetType.ValueFromTerraform(ctx, value)
	if err != nil {
		return nil, err
	}

	s, ok := val.(types.Set)
	if !ok {
		return nil, fmt.Errorf("cannot assert %T as types.Set", val)
	}

	return SetNestedAttributesCustomValue{
		s,
	}, nil
}

type SetNestedAttributesCustomValue struct {
	types.Set
}
