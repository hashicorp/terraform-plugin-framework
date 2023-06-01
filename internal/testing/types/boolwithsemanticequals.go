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
	_ basetypes.BoolTypable                    = BoolTypeWithSemanticEquals{}
	_ basetypes.BoolValuableWithSemanticEquals = BoolValueWithSemanticEquals{}
)

// BoolTypeWithSemanticEquals is a BoolType associated with
// BoolValueWithSemanticEquals, which implements semantic equality logic that
// returns the SemanticEquals boolean for testing.
type BoolTypeWithSemanticEquals struct {
	basetypes.BoolType

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (t BoolTypeWithSemanticEquals) Equal(o attr.Type) bool {
	other, ok := o.(BoolTypeWithSemanticEquals)

	if !ok {
		return false
	}

	if t.SemanticEquals != other.SemanticEquals {
		return false
	}

	return t.BoolType.Equal(other.BoolType)
}

func (t BoolTypeWithSemanticEquals) String() string {
	return fmt.Sprintf("BoolTypeWithSemanticEquals(%t)", t.SemanticEquals)
}

func (t BoolTypeWithSemanticEquals) ValueFromBool(ctx context.Context, in basetypes.BoolValue) (basetypes.BoolValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := BoolValueWithSemanticEquals{
		BoolValue:                 in,
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}

	return value, diags
}

func (t BoolTypeWithSemanticEquals) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.BoolType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.BoolValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromBool(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting BoolValue to BoolValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t BoolTypeWithSemanticEquals) ValueType(ctx context.Context) attr.Value {
	return BoolValueWithSemanticEquals{
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}
}

type BoolValueWithSemanticEquals struct {
	basetypes.BoolValue

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (v BoolValueWithSemanticEquals) Equal(o attr.Value) bool {
	other, ok := o.(BoolValueWithSemanticEquals)

	if !ok {
		return false
	}

	return v.BoolValue.Equal(other.BoolValue)
}

func (v BoolValueWithSemanticEquals) BoolSemanticEquals(ctx context.Context, otherV basetypes.BoolValuable) (bool, diag.Diagnostics) {
	return v.SemanticEquals, v.SemanticEqualsDiagnostics
}

func (v BoolValueWithSemanticEquals) Type(ctx context.Context) attr.Type {
	return BoolTypeWithSemanticEquals{
		SemanticEquals:            v.SemanticEquals,
		SemanticEqualsDiagnostics: v.SemanticEqualsDiagnostics,
	}
}
