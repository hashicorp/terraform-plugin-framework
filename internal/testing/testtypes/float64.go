// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.Float64Typable  = Float64Type{}
	_ basetypes.Float64Valuable = Float64Value{}
)

type Float64Type struct {
	basetypes.Float64Type
}

func (t Float64Type) Equal(o attr.Type) bool {
	other, ok := o.(Float64Type)

	if !ok {
		return false
	}

	return t.Float64Type.Equal(other.Float64Type)
}

type Float64Value struct {
	basetypes.Float64Value
}

func (v Float64Value) Equal(o attr.Value) bool {
	other, ok := o.(Float64Value)

	if !ok {
		return false
	}

	return v.Float64Value.Equal(other.Float64Value)
}
