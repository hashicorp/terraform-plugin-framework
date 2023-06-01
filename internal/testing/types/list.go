package types

import "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

var (
	_ basetypes.ListTypable  = ListType{}
	_ basetypes.ListValuable = ListValue{}
)

type ListType struct {
	basetypes.ListType
}

type ListValue struct {
	basetypes.ListValue
}
