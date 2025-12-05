// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.Float32Typable  = Float32Type{}
	_ basetypes.Float32Valuable = Float32Value{}
)

type Float32Type struct {
	basetypes.Float32Type
}

func (t Float32Type) Equal(o attr.Type) bool {
	other, ok := o.(Float32Type)

	if !ok {
		return false
	}

	return t.Float32Type.Equal(other.Float32Type)
}

type Float32Value struct {
	basetypes.Float32Value
}

func (v Float32Value) Equal(o attr.Value) bool {
	other, ok := o.(Float32Value)

	if !ok {
		return false
	}

	return v.Float32Value.Equal(other.Float32Value)
}
