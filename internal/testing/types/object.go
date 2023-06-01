package types

import "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

var (
	_ basetypes.ObjectTypable  = ObjectType{}
	_ basetypes.ObjectValuable = ObjectValue{}
)

type ObjectType struct {
	basetypes.ObjectType
}

type ObjectValue struct {
	basetypes.ObjectValue
}
