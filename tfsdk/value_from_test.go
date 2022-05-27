package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type person struct {
	Name    types.String `tfsdk:"name"`
	Age     types.Int64  `tfsdk:"age"`
	OptedIn types.Bool   `tfsdk:"opted_in"`
	Address types.List   `tfsdk:"address"`
}

func TestValueFrom(t *testing.T) {
	t.Parallel()

	personAttrTypes := map[string]attr.Type{
		"name":     types.StringType,
		"age":      types.Int64Type,
		"opted_in": types.BoolType,
		"address": types.ListType{
			ElemType: types.StringType,
		},
	}

	mrX := person{
		Name:    types.String{Value: "x"},
		Age:     types.Int64{Value: 30},
		OptedIn: types.Bool{Value: true},
		Address: types.List{
			ElemType: types.StringType,
			Elems: []attr.Value{
				types.String{Value: "1"},
				types.String{Value: "Beckford Close"},
				types.String{Value: "Gotham"},
			},
		},
	}

	mrsY := person{
		Name:    types.String{Value: "y"},
		Age:     types.Int64{Value: 23},
		OptedIn: types.Bool{Value: false},
		Address: types.List{
			ElemType: types.StringType,
			Elems: []attr.Value{
				types.String{Value: "2"},
				types.String{Value: "Windmill Close"},
				types.String{Value: "Smallville"},
			},
		},
	}

	expectedMrXObj := types.Object{
		AttrTypes: personAttrTypes,
		Attrs: map[string]attr.Value{
			"name":     types.String{Value: "x", Unknown: false, Null: false},
			"age":      types.Int64{Value: 30, Unknown: false, Null: false},
			"opted_in": types.Bool{Value: true, Unknown: false, Null: false},
			"address": types.List{
				ElemType: types.StringType,
				Elems: []attr.Value{
					types.String{Value: "1"},
					types.String{Value: "Beckford Close"},
					types.String{Value: "Gotham"},
				},
			},
		},
	}

	expectedMrsYObj := types.Object{
		AttrTypes: personAttrTypes,
		Attrs: map[string]attr.Value{
			"name":     types.String{Value: "y", Unknown: false, Null: false},
			"age":      types.Int64{Value: 23, Unknown: false, Null: false},
			"opted_in": types.Bool{Value: false, Unknown: false, Null: false},
			"address": types.List{
				ElemType: types.StringType,
				Elems: []attr.Value{
					types.String{Value: "2"},
					types.String{Value: "Windmill Close"},
					types.String{Value: "Smallville"},
				},
			},
		},
	}

	type testCase struct {
		val           interface{}
		target        attr.Value
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}

	tests := map[string]testCase{
		"primitive": {
			val:      "hello",
			target:   types.String{},
			expected: types.String{Value: "hello", Unknown: false, Null: false},
		},
		"struct": {
			val: mrX,
			target: types.Object{
				AttrTypes: personAttrTypes,
			},
			expected: expectedMrXObj,
		},
		"list": {
			val: []person{mrX, mrsY},
			target: types.List{
				ElemType: types.ObjectType{
					AttrTypes: personAttrTypes,
				},
			},
			expected: types.List{
				ElemType: types.ObjectType{
					AttrTypes: personAttrTypes,
				},
				Elems: []attr.Value{expectedMrXObj, expectedMrsYObj},
			},
		},
		"incompatible-type": {
			val:    0,
			target: types.String{},
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					tftypes.NewAttributePath(),
					diag.NewErrorDiagnostic(
						"Value Conversion Error",
						"An unexpected error was encountered trying to convert the Terraform value. This is always an error in the provider. Please report the following to the provider developer:\n\ncan't unmarshal tftypes.Number into *string, expected string",
					),
				),
			},
		},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := ValueFrom(context.Background(), tc.val, tc.target.Type(context.Background()), &tc.target)

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
