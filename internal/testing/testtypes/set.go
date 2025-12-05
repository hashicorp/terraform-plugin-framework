// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.SetTypable  = SetType{}
	_ basetypes.SetValuable = SetValue{}
)

type SetType struct {
	basetypes.SetType
}

func (t SetType) Equal(o attr.Type) bool {
	other, ok := o.(SetType)

	if !ok {
		return false
	}

	return t.SetType.Equal(other.SetType)
}

type SetValue struct {
	basetypes.SetValue
}

func (v SetValue) Equal(o attr.Value) bool {
	other, ok := o.(SetValue)

	if !ok {
		return false
	}

	return v.SetValue.Equal(other.SetValue)
}
