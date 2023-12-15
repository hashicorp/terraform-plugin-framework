// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.ObjectTypable  = ObjectType{}
	_ basetypes.ObjectValuable = ObjectValue{}
)

type ObjectType struct {
	basetypes.ObjectType
}

func (t ObjectType) Equal(o attr.Type) bool {
	other, ok := o.(ObjectType)

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

type ObjectValue struct {
	basetypes.ObjectValue
}

func (v ObjectValue) Equal(o attr.Value) bool {
	other, ok := o.(ObjectValue)

	if !ok {
		return false
	}

	return v.ObjectValue.Equal(other.ObjectValue)
}
