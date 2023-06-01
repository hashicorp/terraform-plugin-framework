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
	_ basetypes.Int64Typable                    = Int64TypeWithSemanticEquals{}
	_ basetypes.Int64ValuableWithSemanticEquals = Int64ValueWithSemanticEquals{}
)

// Int64TypeWithSemanticEquals is a Int64Type associated with
// Int64ValueWithSemanticEquals, which implements semantic equality logic that
// returns the SemanticEquals boolean for testing.
type Int64TypeWithSemanticEquals struct {
	basetypes.Int64Type

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (t Int64TypeWithSemanticEquals) Equal(o attr.Type) bool {
	other, ok := o.(Int64TypeWithSemanticEquals)

	if !ok {
		return false
	}

	if t.SemanticEquals != other.SemanticEquals {
		return false
	}

	return t.Int64Type.Equal(other.Int64Type)
}

func (t Int64TypeWithSemanticEquals) String() string {
	return fmt.Sprintf("Int64TypeWithSemanticEquals(%t)", t.SemanticEquals)
}

func (t Int64TypeWithSemanticEquals) ValueFromInt64(ctx context.Context, in basetypes.Int64Value) (basetypes.Int64Valuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := Int64ValueWithSemanticEquals{
		Int64Value:                in,
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}

	return value, diags
}

func (t Int64TypeWithSemanticEquals) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.Int64Type.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.Int64Value)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromInt64(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting Int64Value to Int64Valuable: %v", diags)
	}

	return stringValuable, nil
}

func (t Int64TypeWithSemanticEquals) ValueType(ctx context.Context) attr.Value {
	return Int64ValueWithSemanticEquals{
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}
}

type Int64ValueWithSemanticEquals struct {
	basetypes.Int64Value

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (v Int64ValueWithSemanticEquals) Equal(o attr.Value) bool {
	other, ok := o.(Int64ValueWithSemanticEquals)

	if !ok {
		return false
	}

	return v.Int64Value.Equal(other.Int64Value)
}

func (v Int64ValueWithSemanticEquals) Int64SemanticEquals(ctx context.Context, otherV basetypes.Int64Valuable) (bool, diag.Diagnostics) {
	return v.SemanticEquals, v.SemanticEqualsDiagnostics
}

func (v Int64ValueWithSemanticEquals) Type(ctx context.Context) attr.Type {
	return Int64TypeWithSemanticEquals{
		SemanticEquals:            v.SemanticEquals,
		SemanticEqualsDiagnostics: v.SemanticEqualsDiagnostics,
	}
}
