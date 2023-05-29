// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

var (
	_ basetypes.MapTypable  = MapType{}
	_ basetypes.MapValuable = MapValue{}
)

type MapType struct {
	basetypes.MapType
}

type MapValue struct {
	basetypes.MapValue
}
