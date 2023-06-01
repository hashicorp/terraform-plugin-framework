package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.SetTypable                    = SetTypeWithSemanticEquals{}
	_ basetypes.SetValuableWithSemanticEquals = SetValueWithSemanticEquals{}
)

// SetTypeWithSemanticEquals is a SetType associated with
// SetValueWithSemanticEquals, which implements semantic equality logic that
// returns the SemanticEquals boolean for testing.
type SetTypeWithSemanticEquals struct {
	basetypes.SetType

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (t SetTypeWithSemanticEquals) Equal(o attr.Type) bool {
	other, ok := o.(SetTypeWithSemanticEquals)

	if !ok {
		return false
	}

	if t.SemanticEquals != other.SemanticEquals {
		return false
	}

	return t.SetType.Equal(other.SetType)
}

func (t SetTypeWithSemanticEquals) String() string {
	return fmt.Sprintf("SetTypeWithSemanticEquals(%t)", t.SemanticEquals)
}

func (t SetTypeWithSemanticEquals) ValueFromSet(ctx context.Context, in basetypes.SetValue) (basetypes.SetValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := SetValueWithSemanticEquals{
		SetValue:                  in,
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}

	return value, diags
}

func (t SetTypeWithSemanticEquals) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.SetType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.SetValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromSet(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting SetValue to SetValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t SetTypeWithSemanticEquals) ValueType(ctx context.Context) attr.Value {
	return SetValueWithSemanticEquals{
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}
}

type SetValueWithSemanticEquals struct {
	basetypes.SetValue

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (v SetValueWithSemanticEquals) Equal(o attr.Value) bool {
	other, ok := o.(SetValueWithSemanticEquals)

	if !ok {
		return false
	}

	return v.SetValue.Equal(other.SetValue)
}

func (v SetValueWithSemanticEquals) SetSemanticEquals(ctx context.Context, otherV basetypes.SetValuable) (bool, diag.Diagnostics) {
	return v.SemanticEquals, v.SemanticEqualsDiagnostics
}

func (v SetValueWithSemanticEquals) Type(ctx context.Context) attr.Type {
	return SetTypeWithSemanticEquals{
		SemanticEquals:            v.SemanticEquals,
		SemanticEqualsDiagnostics: v.SemanticEqualsDiagnostics,
	}
}
