package types

import "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

var (
	_ basetypes.SetTypable  = SetType{}
	_ basetypes.SetValuable = SetValue{}
)

type SetType struct {
	basetypes.SetType
}

type SetValue struct {
	basetypes.SetValue
}
