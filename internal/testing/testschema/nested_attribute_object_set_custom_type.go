// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testschema

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ basetypes.ObjectTypable = SetNestedObjectCustomType{}

type SetNestedObjectCustomType struct {
	basetypes.ObjectType
}

func (c SetNestedObjectCustomType) Equal(o attr.Type) bool {
	other, ok := o.(SetNestedObjectCustomType)

	if !ok {
		return false
	}

	return c.ObjectType.Equal(other.ObjectType)
}

func (c SetNestedObjectCustomType) String() string {
	return "SetNestedObjectCustomType"
}

func (c SetNestedObjectCustomType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	value := SetNestedObjectCustomValue{
		ObjectValue: in,
	}

	return value, nil
}

func (c SetNestedObjectCustomType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (c SetNestedObjectCustomType) ValueType(ctx context.Context) attr.Value {
	return SetNestedObjectCustomValue{}
}

var _ basetypes.ObjectValuable = SetNestedObjectCustomValue{}

type SetNestedObjectCustomValue struct {
	basetypes.ObjectValue
}

func (c SetNestedObjectCustomValue) Equal(o attr.Value) bool {
	other, ok := o.(SetNestedObjectCustomValue)

	if !ok {
		return false
	}

	return c.ObjectValue.Equal(other.ObjectValue)
}

func (c SetNestedObjectCustomValue) Type(ctx context.Context) attr.Type {
	return SetNestedObjectCustomType{
		basetypes.ObjectType{
			AttrTypes: c.AttributeTypes(ctx),
		},
	}
}
