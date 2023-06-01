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
	_ basetypes.MapTypable  = MapNestedAttributesCustomTypeType{}
	_ basetypes.MapValuable = &MapNestedAttributesCustomValue{}
)

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
