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

func newStringPointer(in string) *string {
	return &in
}

func newStringPointerPointer(in string) **string {
	stringPointer := &in
	return &stringPointer
}

func newInt64Pointer(in int64) *int64 {
	return &in
}

type personAsFWTypes struct {
	Name    types.String `tfsdk:"name"`
	Age     types.Int64  `tfsdk:"age"`
	OptedIn types.Bool   `tfsdk:"opted_in"`
	Address types.Map    `tfsdk:"address"`
	Colours types.List   `tfsdk:"colours"`
}

type personAsGoTypes struct {
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
		"primitive": {
			val:      types.String{Value: "hello"},
			target:   newStringPointer(""),
			expected: newStringPointer("hello"),
		},
		"struct - FW": {
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
			target: &personAsFWTypes{},
			expected: &personAsFWTypes{
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
		"struct - Go": {
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
			target: &personAsGoTypes{},
			expected: &personAsGoTypes{
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

			if diff := cmp.Diff(tc.expected, tc.target); diff != "" {
				t.Fatalf("Unexpected diff in results (-wanted, +got): %s", diff)
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
