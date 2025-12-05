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
	_ basetypes.Float32Typable                    = Float32TypeWithSemanticEquals{}
	_ basetypes.Float32ValuableWithSemanticEquals = Float32ValueWithSemanticEquals{}
)

// Float32TypeWithSemanticEquals is a Float32Type associated with
// Float32ValueWithSemanticEquals, which implements semantic equality logic that
// returns the SemanticEquals boolean for testing.
type Float32TypeWithSemanticEquals struct {
	basetypes.Float32Type

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (t Float32TypeWithSemanticEquals) Equal(o attr.Type) bool {
	other, ok := o.(Float32TypeWithSemanticEquals)

	if !ok {
		return false
	}

	if t.SemanticEquals != other.SemanticEquals {
		return false
	}

	return t.Float32Type.Equal(other.Float32Type)
}

func (t Float32TypeWithSemanticEquals) String() string {
	return fmt.Sprintf("Float32TypeWithSemanticEquals(%t)", t.SemanticEquals)
}

func (t Float32TypeWithSemanticEquals) ValueFromFloat32(ctx context.Context, in basetypes.Float32Value) (basetypes.Float32Valuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := Float32ValueWithSemanticEquals{
		Float32Value:              in,
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}

	return value, diags
}

func (t Float32TypeWithSemanticEquals) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.Float32Type.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.Float32Value)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromFloat32(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting Float32Value to Float32Valuable: %v", diags)
	}

	return stringValuable, nil
}

func (t Float32TypeWithSemanticEquals) ValueType(ctx context.Context) attr.Value {
	return Float32ValueWithSemanticEquals{
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}
}

type Float32ValueWithSemanticEquals struct {
	basetypes.Float32Value

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (v Float32ValueWithSemanticEquals) Equal(o attr.Value) bool {
	other, ok := o.(Float32ValueWithSemanticEquals)

	if !ok {
		return false
	}

	return v.Float32Value.Equal(other.Float32Value)
}

func (v Float32ValueWithSemanticEquals) Float32SemanticEquals(ctx context.Context, otherV basetypes.Float32Valuable) (bool, diag.Diagnostics) {
	return v.SemanticEquals, v.SemanticEqualsDiagnostics
}

func (v Float32ValueWithSemanticEquals) Type(ctx context.Context) attr.Type {
	return Float32TypeWithSemanticEquals{
		SemanticEquals:            v.SemanticEquals,
		SemanticEqualsDiagnostics: v.SemanticEqualsDiagnostics,
	}
}
