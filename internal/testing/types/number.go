// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.NumberTypable  = NumberType{}
	_ basetypes.NumberValuable = Number{}
)

// NumberType is a reimplementation of types.NumberType that can be used as a base
// for other extension types in testing.
type NumberType struct{}

func (t NumberType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, t.String())
}

func (t NumberType) Equal(o attr.Type) bool {
	other, ok := o.(NumberType)
	if !ok {
		return false
	}
	return t == other
}

func (t NumberType) String() string {
	return "testtypes.NumberType"
}

func (t NumberType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.Number
}

func (t NumberType) ValueFromNumber(ctx context.Context, in basetypes.NumberValue) (basetypes.NumberValuable, diag.Diagnostics) {
	if in.IsNull() {
		return Number{
			Number:    types.NumberNull(),
			CreatedBy: t,
		}, nil
	}

	if in.IsUnknown() {
		return Number{
			Number:    types.NumberUnknown(),
			CreatedBy: t,
		}, nil
	}

	return Number{
		Number:    types.NumberValue(in.ValueBigFloat()),
		CreatedBy: t,
	}, nil
}

func (t NumberType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return Number{
			Number:    types.NumberUnknown(),
			CreatedBy: t,
		}, nil
	}
	if in.IsNull() {
		return Number{
			Number:    types.NumberNull(),
			CreatedBy: t,
		}, nil
	}
	n := big.NewFloat(0)
	err := in.As(&n)
	if err != nil {
		return nil, err
	}
	return Number{
		Number:    types.NumberValue(n),
		CreatedBy: t,
	}, nil
}

// ValueType returns the Value type.
func (t NumberType) ValueType(_ context.Context) attr.Value {
	return Number{}
}

type Number struct {
	types.Number

	CreatedBy attr.Type
}

func (n Number) Type(_ context.Context) attr.Type {
	return n.CreatedBy
}

func (n Number) Equal(o attr.Value) bool {
	on, ok := o.(Number)
	if !ok {
		return false
	}
	return n.Number.Equal(on.Number)
}

func (n Number) IsNull() bool {
	return n.Number.IsNull()
}

func (n Number) IsUnknown() bool {
	return n.Number.IsUnknown()
}
