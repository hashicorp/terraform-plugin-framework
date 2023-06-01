// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfsdk

import (
	"context"
	"fmt"
	"math/big"
	goreflect "reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func newBoolPointer(in bool) *bool {
	return &in
}

func newBoolPointerPointer(in bool) **bool {
	boolPointer := &in
	return &boolPointer
}

func newFloatPointer(in float64) *float64 {
	return &in
}

func newFloatPointerPointer(in float64) **float64 {
	floatPointer := &in
	return &floatPointer
}

func newInt64Pointer(in int64) *int64 {
	return &in
}

func newInt64PointerPointer(in int64) **int64 {
	intPointer := &in
	return &intPointer
}

func newBigFloatPointerPointer(in uint64) **big.Float {
	bf := new(big.Float).SetUint64(in)
	return &bf
}

func newStringPointer(in string) *string {
	return &in
}

func newStringPointerPointer(in string) **string {
	stringPointer := &in
	return &stringPointer
}

type personFrameworkTypes struct {
	Name    types.String `tfsdk:"name"`
	Age     types.Int64  `tfsdk:"age"`
	OptedIn types.Bool   `tfsdk:"opted_in"`
	Address types.Map    `tfsdk:"address"`
	Colours types.List   `tfsdk:"colours"`
}

type personGoTypes struct {
	Name    string            `tfsdk:"name"`
	Age     int64             `tfsdk:"age"`
	OptedIn bool              `tfsdk:"opted_in"`
	Address map[string]string `tfsdk:"address"`
	Colours []string          `tfsdk:"colours"`
}

func TestValueAs(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val           attr.Value
		target        interface{}
		expected      interface{}
		expectedDiags diag.Diagnostics
	}

	tests := map[string]testCase{
		"primitive bool pointer": {
			val:      types.BoolValue(true),
			target:   newBoolPointer(false),
			expected: newBoolPointer(true),
		},
		"primitive bool pointer pointer": {
			val:      types.BoolValue(true),
			target:   newBoolPointerPointer(false),
			expected: newBoolPointerPointer(true),
		},
		"primitive float64 pointer": {
			val:      types.Float64Value(12.3),
			target:   newFloatPointer(0.0),
			expected: newFloatPointer(12.3),
		},
		"primitive float64 pointer pointer": {
			val:      types.Float64Value(12.3),
			target:   newFloatPointerPointer(0.0),
			expected: newFloatPointerPointer(12.3),
		},
		"primitive int64 pointer": {
			val:      types.Int64Value(12),
			target:   newInt64Pointer(0),
			expected: newInt64Pointer(12),
		},
		"primitive int64 pointer pointer": {
			val:      types.Int64Value(12),
			target:   newInt64PointerPointer(0),
			expected: newInt64PointerPointer(12),
		},
		// Following test fails as target.Type() is big.Float not *bigFloat.
		// See https://github.com/hashicorp/terraform-plugin-framework/blob/main/internal/reflect/into.go#L144
		// The switch on target.Kind() then identifies big.Float as reflect.Struct and the reflection fails
		// with "cannot reflect tftypes.Number into a struct, must be an object".
		// See https://github.com/hashicorp/terraform-plugin-framework/blob/main/internal/reflect/into.go#L148
		// "primitive number pointer": {
		// 	val:      types.Number{Value: new(big.Float).SetUint64(722770156065510359)},
		// 	target:   newBigFloatPointer(0),
		// 	expected: newBigFloatPointer(722770156065510359),
		// },
		"primitive number pointer pointer": {
			val:      types.NumberValue(new(big.Float).SetUint64(722770156065510359)),
			target:   newBigFloatPointerPointer(0),
			expected: newBigFloatPointerPointer(722770156065510359),
		},
		"primitive string pointer": {
			val:      types.StringValue("hello"),
			target:   newStringPointer(""),
			expected: newStringPointer("hello"),
		},
		"primitive string pointer pointer": {
			val:      types.StringValue("hello"),
			target:   newStringPointerPointer(""),
			expected: newStringPointerPointer("hello"),
		},
		"list": {
			val: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("hello"),
					types.StringValue("world"),
				},
			),
			target: &[]string{},
			expected: &[]string{
				"hello",
				"world",
			},
		},
		"map": {
			val: types.MapValueMust(
				types.StringType,
				map[string]attr.Value{
					"hello":   types.StringValue("world"),
					"goodbye": types.StringValue("world"),
				},
			),
			target: &map[string]string{},
			expected: &map[string]string{
				"hello":   "world",
				"goodbye": "world",
			},
		},
		"object": {
			val: types.ObjectValueMust(
				map[string]attr.Type{
					"name":     types.StringType,
					"age":      types.Int64Type,
					"opted_in": types.BoolType,
				},
				map[string]attr.Value{
					"name":     types.StringValue("Boris"),
					"age":      types.Int64Value(25),
					"opted_in": types.BoolValue(true),
				},
			),
			target: &struct {
				Name    string `tfsdk:"name"`
				Age     int64  `tfsdk:"age"`
				OptedIn bool   `tfsdk:"opted_in"`
			}{},
			expected: &struct {
				Name    string `tfsdk:"name"`
				Age     int64  `tfsdk:"age"`
				OptedIn bool   `tfsdk:"opted_in"`
			}{
				Name:    "Boris",
				Age:     25,
				OptedIn: true,
			},
		},
		"set": {
			val: types.SetValueMust(
				types.BoolType,
				[]attr.Value{
					types.BoolValue(true),
					types.BoolValue(false),
				},
			),
			target: &[]bool{},
			expected: &[]bool{
				true,
				false,
			},
		},
		"struct framework types": {
			val: types.ObjectValueMust(
				map[string]attr.Type{
					"name":     types.StringType,
					"age":      types.Int64Type,
					"opted_in": types.BoolType,
					"address":  types.MapType{ElemType: types.StringType},
					"colours":  types.ListType{ElemType: types.StringType},
				},
				map[string]attr.Value{
					"name":     types.StringValue("Boris"),
					"age":      types.Int64Value(25),
					"opted_in": types.BoolValue(true),
					"address": types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"first_line": types.StringValue("10 Downing Street"),
							"postcode":   types.StringValue("SW1A 2AA"),
						},
					),
					"colours": types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("red"),
							types.StringValue("green"),
							types.StringValue("blue"),
						},
					),
				},
			),
			target: &personFrameworkTypes{},
			expected: &personFrameworkTypes{
				Name:    types.StringValue("Boris"),
				Age:     types.Int64Value(25),
				OptedIn: types.BoolValue(true),
				Address: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"first_line": types.StringValue("10 Downing Street"),
						"postcode":   types.StringValue("SW1A 2AA"),
					},
				),
				Colours: types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("red"),
						types.StringValue("green"),
						types.StringValue("blue"),
					},
				),
			},
		},
		"struct go types": {
			val: types.ObjectValueMust(
				map[string]attr.Type{
					"name":     types.StringType,
					"age":      types.Int64Type,
					"opted_in": types.BoolType,
					"address":  types.MapType{ElemType: types.StringType},
					"colours":  types.ListType{ElemType: types.StringType},
				},
				map[string]attr.Value{
					"name":     types.StringValue("Boris"),
					"age":      types.Int64Value(25),
					"opted_in": types.BoolValue(true),
					"address": types.MapValueMust(
						types.StringType,
						map[string]attr.Value{
							"first_line": types.StringValue("10 Downing Street"),
							"postcode":   types.StringValue("SW1A 2AA"),
						},
					),
					"colours": types.ListValueMust(
						types.StringType,
						[]attr.Value{
							types.StringValue("red"),
							types.StringValue("green"),
							types.StringValue("blue"),
						},
					),
				},
			),
			target: &personGoTypes{},
			expected: &personGoTypes{
				Name:    "Boris",
				Age:     25,
				OptedIn: true,
				Address: map[string]string{
					"first_line": "10 Downing Street",
					"postcode":   "SW1A 2AA",
				},
				Colours: []string{"red", "green", "blue"},
			},
		},
		"incompatible-type": {
			val:    types.StringValue("hello"),
			target: newInt64Pointer(0),
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					path.Empty(),
					reflect.DiagIntoIncompatibleType{
						Val:        tftypes.NewValue(tftypes.String, "hello"),
						TargetType: goreflect.TypeOf(int64(0)),
						Err:        fmt.Errorf("can't unmarshal %s into %T, expected *big.Float", tftypes.String, big.NewFloat(0)),
					},
				),
			},
		},
		"different-type": {
			val:    types.StringValue("hello"),
			target: &testtypes.String{},
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					path.Empty(),
					reflect.DiagNewAttributeValueIntoWrongType{
						ValType:    goreflect.TypeOf(types.StringValue("hello")),
						TargetType: goreflect.TypeOf(testtypes.String{}),
						SchemaType: types.StringType,
					},
				),
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := ValueAs(context.Background(), tc.val, tc.target)

			if diff := cmp.Diff(tc.expectedDiags, diags); diff != "" {
				t.Fatalf("Unexpected diff in diagnostics (-wanted, +got): %s", diff)
			}

			if diags.HasError() {
				return
			}

			// cmp.Comparer is used to generate a type-specific comparison for big.Float as cmp.Diff
			// cannot be used owing to the unexported (*big.Float).prec field.
			opt := cmp.Comparer(func(expected **big.Float, target **big.Float) bool {
				//nolint:forcetypeassert // Type assertion is guaranteed by the above `cmp.Comparer` function
				if diff := (*tc.expected.(**big.Float)).Cmp(*tc.target.(**big.Float)); diff != 0 {
					return false
				}

				return true
			})

			if diff := cmp.Diff(tc.expected, tc.target, opt); diff != "" {
				t.Fatalf("Unexpected diff in results (-wanted, +got): %s", diff)
			}
		})
	}
}

func TestValueAs_generic(t *testing.T) {
	t.Parallel()

	var target attr.Value
	val := types.StringValue("hello")
	diags := ValueAs(context.Background(), val, &target)
	if len(diags) > 0 {
		t.Fatalf("Unexpected diagnostics: %s", diags)
	}
	if !val.Equal(target) {
		t.Errorf("Expected target to be %v, got %v", val, target)
	}
}
