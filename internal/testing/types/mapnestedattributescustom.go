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
	_ types.MapTypable  = MapNestedAttributesCustomTypeType{}
	_ types.MapValuable = &MapNestedAttributesCustomValue{}
)

type MapNestedAttributesCustomType struct {
	fwschema.NestedAttributes
}

func (t MapNestedAttributesCustomType) Type() attr.Type {
	return MapNestedAttributesCustomTypeType{
		t.NestedAttributes.Type().(types.MapType),
	}
}

type MapNestedAttributesCustomTypeType struct {
	types.MapType
}

func (tt MapNestedAttributesCustomTypeType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	val, err := tt.MapType.ValueFromTerraform(ctx, value)
	if err != nil {
		return nil, err
	}

	m, ok := val.(types.Map)
	if !ok {
		return nil, fmt.Errorf("cannot assert %T as types.Map", val)
	}

	return MapNestedAttributesCustomValue{
		m,
	}, nil
}

type MapNestedAttributesCustomValue struct {
	types.Map
}
