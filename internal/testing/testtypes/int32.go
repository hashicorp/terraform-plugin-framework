// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.Int32Typable  = Int32Type{}
	_ basetypes.Int32Valuable = Int32Value{}
)

type Int32Type struct {
	basetypes.Int32Type
}

func (t Int32Type) Equal(o attr.Type) bool {
	other, ok := o.(Int32Type)

	if !ok {
		return false
	}

	return t.Int32Type.Equal(other.Int32Type)
}

type Int32Value struct {
	basetypes.Int32Value
}

func (v Int32Value) Equal(o attr.Value) bool {
	other, ok := o.(Int32Value)

	if !ok {
		return false
	}

	return v.Int32Value.Equal(other.Int32Value)
}
