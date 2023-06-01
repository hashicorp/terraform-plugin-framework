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
	_ basetypes.NumberTypable                    = NumberTypeWithSemanticEquals{}
	_ basetypes.NumberValuableWithSemanticEquals = NumberValueWithSemanticEquals{}
)

// NumberTypeWithSemanticEquals is a NumberType associated with
// NumberValueWithSemanticEquals, which implements semantic equality logic that
// returns the SemanticEquals boolean for testing.
type NumberTypeWithSemanticEquals struct {
	basetypes.NumberType

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (t NumberTypeWithSemanticEquals) Equal(o attr.Type) bool {
	other, ok := o.(NumberTypeWithSemanticEquals)

	if !ok {
		return false
	}

	if t.SemanticEquals != other.SemanticEquals {
		return false
	}

	return t.NumberType.Equal(other.NumberType)
}

func (t NumberTypeWithSemanticEquals) String() string {
	return fmt.Sprintf("NumberTypeWithSemanticEquals(%t)", t.SemanticEquals)
}

func (t NumberTypeWithSemanticEquals) ValueFromNumber(ctx context.Context, in basetypes.NumberValue) (basetypes.NumberValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := NumberValueWithSemanticEquals{
		NumberValue:               in,
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}

	return value, diags
}

func (t NumberTypeWithSemanticEquals) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.NumberType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.NumberValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromNumber(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting NumberValue to NumberValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t NumberTypeWithSemanticEquals) ValueType(ctx context.Context) attr.Value {
	return NumberValueWithSemanticEquals{
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}
}

type NumberValueWithSemanticEquals struct {
	basetypes.NumberValue

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (v NumberValueWithSemanticEquals) Equal(o attr.Value) bool {
	other, ok := o.(NumberValueWithSemanticEquals)

	if !ok {
		return false
	}

	return v.NumberValue.Equal(other.NumberValue)
}

func (v NumberValueWithSemanticEquals) NumberSemanticEquals(ctx context.Context, otherV basetypes.NumberValuable) (bool, diag.Diagnostics) {
	return v.SemanticEquals, v.SemanticEqualsDiagnostics
}

func (v NumberValueWithSemanticEquals) Type(ctx context.Context) attr.Type {
	return NumberTypeWithSemanticEquals{
		SemanticEquals:            v.SemanticEquals,
		SemanticEqualsDiagnostics: v.SemanticEqualsDiagnostics,
	}
}
