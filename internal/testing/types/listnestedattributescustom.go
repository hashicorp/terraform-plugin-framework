// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.ListTypable  = ListNestedAttributesCustomTypeType{}
	_ basetypes.ListValuable = &ListNestedAttributesCustomValue{}
)

type ListNestedAttributesCustomTypeType struct {
	types.ListType
}

func (tt ListNestedAttributesCustomTypeType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	val, err := tt.ListType.ValueFromTerraform(ctx, value)
	if err != nil {
		return nil, err
	}

	list, ok := val.(types.List)
	if !ok {
		return nil, fmt.Errorf("cannot assert %T as types.List", val)
	}

	return ListNestedAttributesCustomValue{
		list,
	}, nil
}

type ListNestedAttributesCustomValue struct {
	types.List
}
