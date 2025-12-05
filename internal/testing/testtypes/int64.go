// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.Int64Typable  = Int64Type{}
	_ basetypes.Int64Valuable = Int64Value{}
)

type Int64Type struct {
	basetypes.Int64Type
}

func (t Int64Type) Equal(o attr.Type) bool {
	other, ok := o.(Int64Type)

	if !ok {
		return false
	}

	return t.Int64Type.Equal(other.Int64Type)
}

type Int64Value struct {
	basetypes.Int64Value
}

func (v Int64Value) Equal(o attr.Value) bool {
	other, ok := o.(Int64Value)

	if !ok {
		return false
	}

	return v.Int64Value.Equal(other.Int64Value)
}
