// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ basetypes.ObjectTypable = MapNestedObjectCustomType{}

type MapNestedObjectCustomType struct {
	basetypes.ObjectType
}

func (c MapNestedObjectCustomType) Equal(o attr.Type) bool {
	other, ok := o.(MapNestedObjectCustomType)

	if !ok {
		return false
	}

	return c.ObjectType.Equal(other.ObjectType)
}

func (c MapNestedObjectCustomType) String() string {
	return "MapNestedObjectCustomType"
}

func (c MapNestedObjectCustomType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	value := MapNestedObjectCustomValue{
		ObjectValue: in,
	}

	return value, nil
}

func (c MapNestedObjectCustomType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := c.ObjectType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	objectValue, ok := attrValue.(basetypes.ObjectValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	objectValuable, diags := c.ValueFromObject(ctx, objectValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting ObjectValue to ObjectValuable: %v", diags)
	}

	return objectValuable, nil
}

func (c MapNestedObjectCustomType) ValueType(ctx context.Context) attr.Value {
	return MapNestedObjectCustomValue{}
}

var _ basetypes.ObjectValuable = MapNestedObjectCustomValue{}

type MapNestedObjectCustomValue struct {
	basetypes.ObjectValue
}

func (c MapNestedObjectCustomValue) Equal(o attr.Value) bool {
	other, ok := o.(MapNestedObjectCustomValue)

	if !ok {
		return false
	}

	return c.ObjectValue.Equal(other.ObjectValue)
}

func (c MapNestedObjectCustomValue) Type(ctx context.Context) attr.Type {
	return MapNestedObjectCustomType{
		basetypes.ObjectType{
			AttrTypes: map[string]attr.Type{
				"nested_map_nested": types.MapType{
					ElemType: NestedObjectCustomType{
						ObjectType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"nested_attr": types.StringType,
							},
						},
					},
				},
			},
		},
	}
}
