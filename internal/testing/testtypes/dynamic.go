// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.DynamicTypable  = DynamicType{}
	_ basetypes.DynamicValuable = DynamicValue{}
)

type DynamicType struct {
	basetypes.DynamicType
}

func (t DynamicType) Equal(o attr.Type) bool {
	other, ok := o.(DynamicType)

	if !ok {
		return false
	}

	return t.DynamicType.Equal(other.DynamicType)
}

type DynamicValue struct {
	basetypes.DynamicValue
}

func (v DynamicValue) Equal(o attr.Value) bool {
	other, ok := o.(DynamicValue)

	if !ok {
		return false
	}

	return v.DynamicValue.Equal(other.DynamicValue)
}
