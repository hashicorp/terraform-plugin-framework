// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.MapTypable  = MapType{}
	_ basetypes.MapValuable = MapValue{}
)

type MapType struct {
	basetypes.MapType
}

func (t MapType) Equal(o attr.Type) bool {
	other, ok := o.(MapType)

	if !ok {
		return false
	}

	return t.MapType.Equal(other.MapType)
}

type MapValue struct {
	basetypes.MapValue
}

func (v MapValue) Equal(o attr.Value) bool {
	other, ok := o.(MapValue)

	if !ok {
		return false
	}

	return v.MapValue.Equal(other.MapValue)
}
