// Copyright IBM Corp. 2021, 2025
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

var _ basetypes.ObjectTypable = NestedObjectCustomType{}

type NestedObjectCustomType struct {
	basetypes.ObjectType
}

func (c NestedObjectCustomType) Equal(o attr.Type) bool {
	other, ok := o.(NestedObjectCustomType)

	if !ok {
		return false
	}

	return c.ObjectType.Equal(other.ObjectType)
}

func (c NestedObjectCustomType) String() string {
	return "NestedObjectCustomType"
}

func (c NestedObjectCustomType) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	value := NestedObjectCustomValue{
		ObjectValue: in,
	}

	return value, nil
}

func (c NestedObjectCustomType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (c NestedObjectCustomType) ValueType(ctx context.Context) attr.Value {
	return NestedObjectCustomValue{}
}

var _ basetypes.ObjectValuable = NestedObjectCustomValue{}

type NestedObjectCustomValue struct {
	basetypes.ObjectValue
}

func (c NestedObjectCustomValue) Equal(o attr.Value) bool {
	other, ok := o.(NestedObjectCustomValue)

	if !ok {
		return false
	}

	return c.ObjectValue.Equal(other.ObjectValue)
}

func (c NestedObjectCustomValue) Type(ctx context.Context) attr.Type {
	return NestedObjectCustomType{
		basetypes.ObjectType{
			AttrTypes: c.AttributeTypes(ctx),
		},
	}
}
