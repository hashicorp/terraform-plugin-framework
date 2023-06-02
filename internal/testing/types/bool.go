// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.BoolTypable  = BoolType{}
	_ basetypes.BoolValuable = Bool{}
)

// BoolType is a reimplementation of types.BoolType that can be used as a base
// for other extension types in testing.
type BoolType struct{}

func (t BoolType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, t.String())
}

func (t BoolType) Equal(o attr.Type) bool {
	other, ok := o.(BoolType)
	if !ok {
		return false
	}
	return t == other
}

func (t BoolType) String() string {
	return "testtypes.BoolType"
}

func (t BoolType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.Bool
}

func (t BoolType) ValueFromBool(ctx context.Context, in basetypes.BoolValue) (basetypes.BoolValuable, diag.Diagnostics) {
	if in.IsNull() {
		return Bool{
			Bool:      types.BoolNull(),
			CreatedBy: t,
		}, nil
	}

	if in.IsUnknown() {
		return Bool{
			Bool:      types.BoolUnknown(),
			CreatedBy: t,
		}, nil
	}

	return Bool{
		Bool:      types.BoolValue(in.ValueBool()),
		CreatedBy: t,
	}, nil
}

func (t BoolType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	if in.IsNull() {
		return Bool{
			Bool:      types.BoolNull(),
			CreatedBy: t,
		}, nil
	}
	if !in.IsKnown() {
		return Bool{
			Bool:      types.BoolUnknown(),
			CreatedBy: t,
		}, nil
	}
	var b bool
	err := in.As(&b)
	if err != nil {
		return nil, err
	}
	return Bool{Bool: types.BoolValue(b), CreatedBy: t}, nil
}

// ValueType returns the Value type.
func (t BoolType) ValueType(_ context.Context) attr.Value {
	return Bool{}
}

type Bool struct {
	types.Bool

	CreatedBy attr.Type
}

func (b Bool) Type(_ context.Context) attr.Type {
	return b.CreatedBy
}

func (b Bool) Equal(o attr.Value) bool {
	ob, ok := o.(Bool)
	if !ok {
		return false
	}
	return b.Bool.Equal(ob.Bool)
}

func (b Bool) IsNull() bool {
	return b.Bool.IsNull()
}

func (b Bool) IsUnknown() bool {
	return b.Bool.IsUnknown()
}

func (b Bool) String() string {
	return b.Bool.String()
}
