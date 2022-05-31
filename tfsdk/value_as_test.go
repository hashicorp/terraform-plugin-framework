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

func newBigFloatPointer(in uint64) *big.Float {
	return new(big.Float)
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
			val:      types.Bool{Value: true},
			target:   newBoolPointer(false),
			expected: newBoolPointer(true),
		},
		"primitive bool pointer pointer": {
			val:      types.Bool{Value: true},
			target:   newBoolPointerPointer(false),
			expected: newBoolPointerPointer(true),
		},
		"primitive float64 pointer": {
			val:      types.Float64{Value: 12.3},
			target:   newFloatPointer(0.0),
			expected: newFloatPointer(12.3),
		},
		"primitive float64 pointer pointer": {
			val:      types.Float64{Value: 12.3},
			target:   newFloatPointerPointer(0.0),
			expected: newFloatPointerPointer(12.3),
		},
		"primitive int64 pointer": {
			val:      types.Int64{Value: 12},
			target:   newInt64Pointer(0),
			expected: newInt64Pointer(12),
		},
		"primitive int64 pointer pointer": {
			val:      types.Int64{Value: 12},
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
			val:      types.Number{Value: new(big.Float).SetUint64(722770156065510359)},
			target:   newBigFloatPointerPointer(0),
			expected: newBigFloatPointerPointer(722770156065510359),
		},
		"primitive string pointer": {
			val:      types.String{Value: "hello"},
			target:   newStringPointer(""),
			expected: newStringPointer("hello"),
		},
		"primitive string pointer pointer": {
			val:      types.String{Value: "hello"},
			target:   newStringPointerPointer(""),
			expected: newStringPointerPointer("hello"),
		},
		"list": {
			val: types.List{
				Elems: []attr.Value{
					types.String{Value: "hello"},
					types.String{Value: "world"},
				},
				ElemType: types.StringType,
			},
			target: &[]string{},
			expected: &[]string{
				"hello",
				"world",
			},
		},
		"map": {
			val: types.Map{
				Elems: map[string]attr.Value{
					"hello":   types.String{Value: "world"},
					"goodbye": types.String{Value: "world"},
				},
				ElemType: types.StringType,
			},
			target: &map[string]string{},
			expected: &map[string]string{
				"hello":   "world",
				"goodbye": "world",
			},
		},
		"object": {
			val: types.Object{
				Attrs: map[string]attr.Value{
					"name":     types.String{Value: "Boris"},
					"age":      types.Int64{Value: 25},
					"opted_in": types.Bool{Value: true},
				},
				AttrTypes: map[string]attr.Type{
					"name":     types.StringType,
					"age":      types.Int64Type,
					"opted_in": types.BoolType,
				},
			},
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
			val: types.Set{
				Elems: []attr.Value{
					types.Bool{Value: true},
					types.Bool{},
				},
				ElemType: types.BoolType,
			},
			target: &[]bool{},
			expected: &[]bool{
				true,
				false,
			},
		},
		"struct framework types": {
			val: types.Object{
				Attrs: map[string]attr.Value{
					"name":     types.String{Value: "Boris"},
					"age":      types.Int64{Value: 25},
					"opted_in": types.Bool{Value: true},
					"address": types.Map{
						Elems: map[string]attr.Value{
							"first_line": types.String{Value: "10 Downing Street"},
							"postcode":   types.String{Value: "SW1A 2AA"},
						},
						ElemType: types.StringType,
					},
					"colours": types.List{
						Elems: []attr.Value{
							types.String{Value: "red"},
							types.String{Value: "green"},
							types.String{Value: "blue"},
						},
						ElemType: types.StringType,
					},
				},
				AttrTypes: map[string]attr.Type{
					"name":     types.StringType,
					"age":      types.Int64Type,
					"opted_in": types.BoolType,
					"address":  types.MapType{ElemType: types.StringType},
					"colours":  types.ListType{ElemType: types.StringType},
				},
			},
			target: &personFrameworkTypes{},
			expected: &personFrameworkTypes{
				Name:    types.String{Value: "Boris"},
				Age:     types.Int64{Value: 25},
				OptedIn: types.Bool{Value: true},
				Address: types.Map{
					Elems: map[string]attr.Value{
						"first_line": types.String{Value: "10 Downing Street"},
						"postcode":   types.String{Value: "SW1A 2AA"},
					},
					ElemType: types.StringType,
				},
				Colours: types.List{
					Elems: []attr.Value{
						types.String{Value: "red"},
						types.String{Value: "green"},
						types.String{Value: "blue"},
					},
					ElemType: types.StringType,
				},
			},
		},
		"struct go types": {
			val: types.Object{
				Attrs: map[string]attr.Value{
					"name":     types.String{Value: "Boris"},
					"age":      types.Int64{Value: 25},
					"opted_in": types.Bool{Value: true},
					"address": types.Map{
						Elems: map[string]attr.Value{
							"first_line": types.String{Value: "10 Downing Street"},
							"postcode":   types.String{Value: "SW1A 2AA"},
						},
						ElemType: types.StringType,
					},
					"colours": types.List{
						Elems: []attr.Value{
							types.String{Value: "red"},
							types.String{Value: "green"},
							types.String{Value: "blue"},
						},
						ElemType: types.StringType,
					},
				},
				AttrTypes: map[string]attr.Type{
					"name":     types.StringType,
					"age":      types.Int64Type,
					"opted_in": types.BoolType,
					"address":  types.MapType{ElemType: types.StringType},
					"colours":  types.ListType{ElemType: types.StringType},
				},
			},
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
			val:    types.String{Value: "hello"},
			target: newInt64Pointer(0),
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					tftypes.NewAttributePath(),
					reflect.DiagIntoIncompatibleType{
						Val:        tftypes.NewValue(tftypes.String, "hello"),
						TargetType: goreflect.TypeOf(int64(0)),
						Err:        fmt.Errorf("can't unmarshal %s into %T, expected *big.Float", tftypes.String, big.NewFloat(0)),
					},
				),
			},
		},
		"different-type": {
			val:    types.String{Value: "hello"},
			target: &testtypes.String{},
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					tftypes.NewAttributePath(),
					reflect.DiagNewAttributeValueIntoWrongType{
						ValType:    goreflect.TypeOf(types.String{Value: "hello"}),
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

			// Cannot use cmp.Diff for comparing big.Float, requires usage of *big.Float.Cmp()
			switch tc.expected.(type) {
			case *big.Float:
				if diff := tc.expected.(*big.Float).Cmp(tc.target.(*big.Float)); diff != 0 {
					t.Fatalf("Unexpected diff in results: %d", diff)
				}
			case **big.Float:
				if diff := (*tc.expected.(**big.Float)).Cmp(*tc.target.(**big.Float)); diff != 0 {
					t.Fatalf("Unexpected diff in results: %d", diff)
				}
			default:
				if diff := cmp.Diff(tc.expected, tc.target); diff != "" {
					t.Fatalf("Unexpected diff in results (-wanted, +got): %s", diff)
				}
			}
		})
	}
}

func TestValueAs_generic(t *testing.T) {
	t.Parallel()

	var target attr.Value
	val := types.String{Value: "hello"}
	diags := ValueAs(context.Background(), val, &target)
	if len(diags) > 0 {
		t.Fatalf("Unexpected diagnostics: %s", diags)
	}
	if !val.Equal(target) {
		t.Errorf("Expected target to be %v, got %v", val, target)
	}
}
