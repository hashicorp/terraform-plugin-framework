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
	_ basetypes.SetTypable  = SetNestedAttributesCustomTypeType{}
	_ basetypes.SetValuable = &SetNestedAttributesCustomValue{}
)

type SingleNestedAttributesCustomTypeType struct {
	types.ObjectType
}

func (tt SingleNestedAttributesCustomTypeType) ValueFromTerraform(ctx context.Context, value tftypes.Value) (attr.Value, error) {
	val, err := tt.ObjectType.ValueFromTerraform(ctx, value)
	if err != nil {
		return nil, err
	}

	s, ok := val.(types.Object)
	if !ok {
		return nil, fmt.Errorf("cannot assert %T as types.Object", val)
	}

	return SingleNestedAttributesCustomValue{
		s,
	}, nil
}

type SingleNestedAttributesCustomValue struct {
	types.Object
}
