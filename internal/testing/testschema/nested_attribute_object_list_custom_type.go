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

var _ basetypes.ObjectTypable = ListNestedObjectCustomType{}

type ListNestedObjectCustomType struct {
	basetypes.ObjectType
}

func (c ListNestedObjectCustomType) Equal(o attr.Type) bool {
	other, ok := o.(ListNestedObjectCustomType)

	if !ok {
		return false
	}

	return c.ObjectType.Equal(other.ObjectType)
}

func (c ListNestedObjectCustomType) String() string {
	return "ListNestedObjectCustomType"
}

func (c ListNestedObjectCustomType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	value := ListNestedObjectCustomValue{
		ObjectValue: in,
	}

	return value, nil
}

func (c ListNestedObjectCustomType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (c ListNestedObjectCustomType) ValueType(ctx context.Context) attr.Value {
	return ListNestedObjectCustomValue{}
}

var _ basetypes.ObjectValuable = ListNestedObjectCustomValue{}

type ListNestedObjectCustomValue struct {
	basetypes.ObjectValue
}

func (c ListNestedObjectCustomValue) Equal(o attr.Value) bool {
	other, ok := o.(ListNestedObjectCustomValue)

	if !ok {
		return false
	}

	return c.ObjectValue.Equal(other.ObjectValue)
}

func (c ListNestedObjectCustomValue) Type(ctx context.Context) attr.Type {
	return ListNestedObjectCustomType{
		basetypes.ObjectType{
			AttrTypes: map[string]attr.Type{
				"nested_list_nested_attribute": types.ListType{
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
