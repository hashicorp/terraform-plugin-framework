// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ basetypes.Int32Typable                    = Int32TypeWithSemanticEquals{}
	_ basetypes.Int32ValuableWithSemanticEquals = Int32ValueWithSemanticEquals{}
)

// Int32TypeWithSemanticEquals is an Int32Type associated with
// Int32ValueWithSemanticEquals, which implements semantic equality logic that
// returns the SemanticEquals boolean for testing.
type Int32TypeWithSemanticEquals struct {
	basetypes.Int32Type

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (t Int32TypeWithSemanticEquals) Equal(o attr.Type) bool {
	other, ok := o.(Int32TypeWithSemanticEquals)

	if !ok {
		return false
	}

	if t.SemanticEquals != other.SemanticEquals {
		return false
	}

	return t.Int32Type.Equal(other.Int32Type)
}

func (t Int32TypeWithSemanticEquals) String() string {
	return fmt.Sprintf("Int32TypeWithSemanticEquals(%t)", t.SemanticEquals)
}

func (t Int32TypeWithSemanticEquals) ValueFromInt32(ctx context.Context, in basetypes.Int32Value) (basetypes.Int32Valuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := Int32ValueWithSemanticEquals{
		Int32Value:                in,
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}

	return value, diags
}

func (t Int32TypeWithSemanticEquals) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.Int32Type.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.Int32Value)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromInt32(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting Int32Value to Int32Valuable: %v", diags)
	}

	return stringValuable, nil
}

func (t Int32TypeWithSemanticEquals) ValueType(ctx context.Context) attr.Value {
	return Int32ValueWithSemanticEquals{
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}
}

type Int32ValueWithSemanticEquals struct {
	basetypes.Int32Value

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (v Int32ValueWithSemanticEquals) Equal(o attr.Value) bool {
	other, ok := o.(Int32ValueWithSemanticEquals)

	if !ok {
		return false
	}

	return v.Int32Value.Equal(other.Int32Value)
}

func (v Int32ValueWithSemanticEquals) Int32SemanticEquals(ctx context.Context, otherV basetypes.Int32Valuable) (bool, diag.Diagnostics) {
	return v.SemanticEquals, v.SemanticEqualsDiagnostics
}

func (v Int32ValueWithSemanticEquals) Type(ctx context.Context) attr.Type {
	return Int32TypeWithSemanticEquals{
		SemanticEquals:            v.SemanticEquals,
		SemanticEqualsDiagnostics: v.SemanticEqualsDiagnostics,
	}
}
