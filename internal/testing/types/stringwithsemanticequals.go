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
	_ basetypes.StringTypable                    = StringTypeWithSemanticEquals{}
	_ basetypes.StringValuableWithSemanticEquals = StringValueWithSemanticEquals{}
)

// StringTypeWithSemanticEquals is a StringType associated with
// StringValueWithSemanticEquals, which implements semantic equality logic that
// returns the SemanticEquals boolean for testing.
type StringTypeWithSemanticEquals struct {
	basetypes.StringType

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (t StringTypeWithSemanticEquals) Equal(o attr.Type) bool {
	other, ok := o.(StringTypeWithSemanticEquals)

	if !ok {
		return false
	}

	if t.SemanticEquals != other.SemanticEquals {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t StringTypeWithSemanticEquals) String() string {
	return fmt.Sprintf("StringTypeWithSemanticEquals(%t)", t.SemanticEquals)
}

func (t StringTypeWithSemanticEquals) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := StringValueWithSemanticEquals{
		StringValue:               in,
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}

	return value, diags
}

func (t StringTypeWithSemanticEquals) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t StringTypeWithSemanticEquals) ValueType(ctx context.Context) attr.Value {
	return StringValueWithSemanticEquals{
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}
}

type StringValueWithSemanticEquals struct {
	basetypes.StringValue

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (v StringValueWithSemanticEquals) Equal(o attr.Value) bool {
	other, ok := o.(StringValueWithSemanticEquals)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v StringValueWithSemanticEquals) StringSemanticEquals(ctx context.Context, otherV basetypes.StringValuable) (bool, diag.Diagnostics) {
	return v.SemanticEquals, v.SemanticEqualsDiagnostics
}

func (v StringValueWithSemanticEquals) Type(ctx context.Context) attr.Type {
	return StringTypeWithSemanticEquals{
		SemanticEquals:            v.SemanticEquals,
		SemanticEqualsDiagnostics: v.SemanticEqualsDiagnostics,
	}
}
