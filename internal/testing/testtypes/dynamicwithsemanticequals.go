// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testtypes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.DynamicTypable                    = DynamicTypeWithSemanticEquals{}
	_ basetypes.DynamicValuableWithSemanticEquals = DynamicValueWithSemanticEquals{}
)

// DynamicTypeWithSemanticEquals is a DynamicType associated with
// DynamicValueWithSemanticEquals, which implements semantic equality logic that
// returns the SemanticEquals boolean for testing.
type DynamicTypeWithSemanticEquals struct {
	basetypes.DynamicType

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (t DynamicTypeWithSemanticEquals) Equal(o attr.Type) bool {
	other, ok := o.(DynamicTypeWithSemanticEquals)

	if !ok {
		return false
	}

	if t.SemanticEquals != other.SemanticEquals {
		return false
	}

	return t.DynamicType.Equal(other.DynamicType)
}

func (t DynamicTypeWithSemanticEquals) String() string {
	return fmt.Sprintf("DynamicTypeWithSemanticEquals(%t)", t.SemanticEquals)
}

func (t DynamicTypeWithSemanticEquals) ValueFromDynamic(ctx context.Context, in basetypes.DynamicValue) (basetypes.DynamicValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	value := DynamicValueWithSemanticEquals{
		DynamicValue:              in,
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}

	return value, diags
}

func (t DynamicTypeWithSemanticEquals) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.DynamicType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	dynamicValue, ok := attrValue.(basetypes.DynamicValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	dynamicValuable, diags := t.ValueFromDynamic(ctx, dynamicValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting DynamicValue to DynamicValuable: %v", diags)
	}

	return dynamicValuable, nil
}

func (t DynamicTypeWithSemanticEquals) ValueType(ctx context.Context) attr.Value {
	return DynamicValueWithSemanticEquals{
		SemanticEquals:            t.SemanticEquals,
		SemanticEqualsDiagnostics: t.SemanticEqualsDiagnostics,
	}
}

type DynamicValueWithSemanticEquals struct {
	basetypes.DynamicValue

	SemanticEquals            bool
	SemanticEqualsDiagnostics diag.Diagnostics
}

func (v DynamicValueWithSemanticEquals) Equal(o attr.Value) bool {
	other, ok := o.(DynamicValueWithSemanticEquals)

	if !ok {
		return false
	}

	return v.DynamicValue.Equal(other.DynamicValue)
}

func (v DynamicValueWithSemanticEquals) DynamicSemanticEquals(ctx context.Context, otherV basetypes.DynamicValuable) (bool, diag.Diagnostics) {
	return v.SemanticEquals, v.SemanticEqualsDiagnostics
}

func (v DynamicValueWithSemanticEquals) Type(ctx context.Context) attr.Type {
	return DynamicTypeWithSemanticEquals{
		SemanticEquals:            v.SemanticEquals,
		SemanticEqualsDiagnostics: v.SemanticEqualsDiagnostics,
	}
}
