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
	_ basetypes.Float64Typable                    = Float64TypeWithSemanticEquals{}
	_ basetypes.Float64ValuableWithSemanticEquals = Float64ValueWithSemanticEquals{}
)

// Float64TypeWithSemanticEquals is a Float64Type associated with
// Float64ValueWithSemanticEquals, which implements semantic equality logic that
// returns the SemanticEquals boolean for testing.
type Float64TypeWithSemanticEquals struct {
	basetypes.Float64Type

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (t Float64TypeWithSemanticEquals) Equal(o attr.Type) bool {
	other, ok := o.(Float64TypeWithSemanticEquals)

	if !ok {
		return false
	}

	if t.SemanticEquals != other.SemanticEquals {
		return false
	}

	return t.Float64Type.Equal(other.Float64Type)
}

func (t Float64TypeWithSemanticEquals) String() string {
	return fmt.Sprintf("Float64TypeWithSemanticEquals(%t)", t.SemanticEquals)
}

func (t Float64TypeWithSemanticEquals) ValueFromFloat64(ctx context.Context, in basetypes.Float64Value) (basetypes.Float64Valuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := Float64ValueWithSemanticEquals{
		Float64Value:              in,
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}

	return value, diags
}

func (t Float64TypeWithSemanticEquals) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.Float64Type.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.Float64Value)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromFloat64(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting Float64Value to Float64Valuable: %v", diags)
	}

	return stringValuable, nil
}

func (t Float64TypeWithSemanticEquals) ValueType(ctx context.Context) attr.Value {
	return Float64ValueWithSemanticEquals{
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}
}

type Float64ValueWithSemanticEquals struct {
	basetypes.Float64Value

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (v Float64ValueWithSemanticEquals) Equal(o attr.Value) bool {
	other, ok := o.(Float64ValueWithSemanticEquals)

	if !ok {
		return false
	}

	return v.Float64Value.Equal(other.Float64Value)
}

func (v Float64ValueWithSemanticEquals) Float64SemanticEquals(ctx context.Context, otherV basetypes.Float64Valuable) (bool, diag.Diagnostics) {
	return v.SemanticEquals, v.SemanticEqualsDiagnostics
}

func (v Float64ValueWithSemanticEquals) Type(ctx context.Context) attr.Type {
	return Float64TypeWithSemanticEquals{
		SemanticEquals:            v.SemanticEquals,
		SemanticEqualsDiagnostics: v.SemanticEqualsDiagnostics,
	}
}
