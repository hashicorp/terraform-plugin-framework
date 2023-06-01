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
	_ basetypes.MapTypable                    = MapTypeWithSemanticEquals{}
	_ basetypes.MapValuableWithSemanticEquals = MapValueWithSemanticEquals{}
)

// MapTypeWithSemanticEquals is a MapType associated with
// MapValueWithSemanticEquals, which implements semantic equality logic that
// returns the SemanticEquals boolean for testing.
type MapTypeWithSemanticEquals struct {
	basetypes.MapType

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (t MapTypeWithSemanticEquals) Equal(o attr.Type) bool {
	other, ok := o.(MapTypeWithSemanticEquals)

	if !ok {
		return false
	}

	if t.SemanticEquals != other.SemanticEquals {
		return false
	}

	return t.MapType.Equal(other.MapType)
}

func (t MapTypeWithSemanticEquals) String() string {
	return fmt.Sprintf("MapTypeWithSemanticEquals(%t)", t.SemanticEquals)
}

func (t MapTypeWithSemanticEquals) ValueFromMap(ctx context.Context, in basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := MapValueWithSemanticEquals{
		MapValue:                  in,
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}

	return value, diags
}

func (t MapTypeWithSemanticEquals) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.MapType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.MapValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromMap(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting MapValue to MapValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t MapTypeWithSemanticEquals) ValueType(ctx context.Context) attr.Value {
	return MapValueWithSemanticEquals{
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}
}

type MapValueWithSemanticEquals struct {
	basetypes.MapValue

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (v MapValueWithSemanticEquals) Equal(o attr.Value) bool {
	other, ok := o.(MapValueWithSemanticEquals)

	if !ok {
		return false
	}

	return v.MapValue.Equal(other.MapValue)
}

func (v MapValueWithSemanticEquals) MapSemanticEquals(ctx context.Context, otherV basetypes.MapValuable) (bool, diag.Diagnostics) {
	return v.SemanticEquals, v.SemanticEqualsDiagnostics
}

func (v MapValueWithSemanticEquals) Type(ctx context.Context) attr.Type {
	return MapTypeWithSemanticEquals{
		SemanticEquals:            v.SemanticEquals,
		SemanticEqualsDiagnostics: v.SemanticEqualsDiagnostics,
	}
}
