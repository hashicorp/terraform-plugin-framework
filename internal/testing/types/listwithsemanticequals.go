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
	_ basetypes.ListTypable                    = ListTypeWithSemanticEquals{}
	_ basetypes.ListValuableWithSemanticEquals = ListValueWithSemanticEquals{}
)

// ListTypeWithSemanticEquals is a ListType associated with
// ListValueWithSemanticEquals, which implements semantic equality logic that
// returns the SemanticEquals boolean for testing.
type ListTypeWithSemanticEquals struct {
	basetypes.ListType

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (t ListTypeWithSemanticEquals) Equal(o attr.Type) bool {
	other, ok := o.(ListTypeWithSemanticEquals)

	if !ok {
		return false
	}

	if t.SemanticEquals != other.SemanticEquals {
		return false
	}

	return t.ListType.Equal(other.ListType)
}

func (t ListTypeWithSemanticEquals) String() string {
	return fmt.Sprintf("ListTypeWithSemanticEquals(%t)", t.SemanticEquals)
}

func (t ListTypeWithSemanticEquals) ValueFromList(ctx context.Context, in basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := ListValueWithSemanticEquals{
		ListValue:                 in,
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}

	return value, diags
}

func (t ListTypeWithSemanticEquals) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.ListType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.ListValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromList(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting ListValue to ListValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t ListTypeWithSemanticEquals) ValueType(ctx context.Context) attr.Value {
	return ListValueWithSemanticEquals{
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}
}

type ListValueWithSemanticEquals struct {
	basetypes.ListValue

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (v ListValueWithSemanticEquals) Equal(o attr.Value) bool {
	other, ok := o.(ListValueWithSemanticEquals)

	if !ok {
		return false
	}

	return v.ListValue.Equal(other.ListValue)
}

func (v ListValueWithSemanticEquals) ListSemanticEquals(ctx context.Context, otherV basetypes.ListValuable) (bool, diag.Diagnostics) {
	return v.SemanticEquals, v.SemanticEqualsDiagnostics
}

func (v ListValueWithSemanticEquals) Type(ctx context.Context) attr.Type {
	return ListTypeWithSemanticEquals{
		SemanticEquals:            v.SemanticEquals,
		SemanticEqualsDiagnostics: v.SemanticEqualsDiagnostics,
	}
}
