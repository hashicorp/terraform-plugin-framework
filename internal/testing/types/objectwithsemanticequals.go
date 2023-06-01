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
	_ basetypes.ObjectTypable                    = ObjectTypeWithSemanticEquals{}
	_ basetypes.ObjectValuableWithSemanticEquals = ObjectValueWithSemanticEquals{}
)

// ObjectTypeWithSemanticEquals is a ObjectType associated with
// ObjectValueWithSemanticEquals, which implements semantic equality logic that
// returns the SemanticEquals boolean for testing.
type ObjectTypeWithSemanticEquals struct {
	basetypes.ObjectType

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (t ObjectTypeWithSemanticEquals) Equal(o attr.Type) bool {
	other, ok := o.(ObjectTypeWithSemanticEquals)

	if !ok {
		return false
	}

	if t.SemanticEquals != other.SemanticEquals {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ObjectTypeWithSemanticEquals) String() string {
	return fmt.Sprintf("ObjectTypeWithSemanticEquals(%t)", t.SemanticEquals)
}

func (t ObjectTypeWithSemanticEquals) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := ObjectValueWithSemanticEquals{
		ObjectValue:               in,
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}

	return value, diags
}

func (t ObjectTypeWithSemanticEquals) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.ObjectType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.ObjectValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromObject(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting ObjectValue to ObjectValuable: %v", diags)
	}

	return stringValuable, nil
}

func (t ObjectTypeWithSemanticEquals) ValueType(ctx context.Context) attr.Value {
	return ObjectValueWithSemanticEquals{
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}
}

type ObjectValueWithSemanticEquals struct {
	basetypes.ObjectValue

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (v ObjectValueWithSemanticEquals) Equal(o attr.Value) bool {
	other, ok := o.(ObjectValueWithSemanticEquals)

	if !ok {
		return false
	}

	return v.ObjectValue.Equal(other.ObjectValue)
}

func (v ObjectValueWithSemanticEquals) ObjectSemanticEquals(ctx context.Context, otherV basetypes.ObjectValuable) (bool, diag.Diagnostics) {
	return v.SemanticEquals, v.SemanticEqualsDiagnostics
}

func (v ObjectValueWithSemanticEquals) Type(ctx context.Context) attr.Type {
	return ObjectTypeWithSemanticEquals{
		SemanticEquals:            v.SemanticEquals,
		SemanticEqualsDiagnostics: v.SemanticEqualsDiagnostics,
	}
}
