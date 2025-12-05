// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.ListTypable  = ListType{}
	_ basetypes.ListValuable = ListValue{}
)

type ListType struct {
	basetypes.ListType
}

func (t ListType) Equal(o attr.Type) bool {
	other, ok := o.(ListType)

	if !ok {
		return false
	}

	return t.ListType.Equal(other.ListType)
}

type ListValue struct {
	basetypes.ListValue
}

func (v ListValue) Equal(o attr.Value) bool {
	other, ok := o.(ListValue)

	if !ok {
		return false
	}

	return v.ListValue.Equal(other.ListValue)
}
