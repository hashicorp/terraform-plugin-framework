package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type person struct {
	Name    types.String `tfsdk:"name"`
	Age     types.Int64  `tfsdk:"age"`
	OptedIn types.Bool   `tfsdk:"opted_in"`
}

func TestValueFrom(t *testing.T) {
	t.Parallel()

	personAttrTypes := map[string]attr.Type{
		"name":     types.StringType,
		"age":      types.Int64Type,
		"opted_in": types.BoolType,
	}

	x := person{
		Name:    types.String{Value: "x"},
		Age:     types.Int64{Value: 30},
		OptedIn: types.Bool{Value: true},
	}

	y := person{
		Name:    types.String{Value: "y"},
		Age:     types.Int64{Value: 23},
		OptedIn: types.Bool{Value: false},
	}

	xObj := types.Object{
		AttrTypes: personAttrTypes,
		Attrs: map[string]attr.Value{
			"name":     types.String{Value: "x", Unknown: false, Null: false},
			"age":      types.Int64{Value: 30, Unknown: false, Null: false},
			"opted_in": types.Bool{Value: true, Unknown: false, Null: false},
		},
	}

	yObj := types.Object{
		AttrTypes: personAttrTypes,
		Attrs: map[string]attr.Value{
			"name":     types.String{Value: "y", Unknown: false, Null: false},
			"age":      types.Int64{Value: 23, Unknown: false, Null: false},
			"opted_in": types.Bool{Value: false, Unknown: false, Null: false},
		},
	}

	type testCase struct {
		target        attr.Value
		val           interface{}
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
			val: x,
			target: types.Object{
				AttrTypes: personAttrTypes,
			},
			expected: xObj,
		},
		"list": {
			val: []person{x, y},
			target: types.List{
				ElemType: types.ObjectType{
					AttrTypes: personAttrTypes,
				},
			},
			expected: types.List{
				ElemType: types.ObjectType{
					AttrTypes: personAttrTypes,
				},
				Elems: []attr.Value{xObj, yObj},
			},
		},
		//"incompatible-type": {
		//	val:    0,
		//	target: types.String{},
		//	expectedDiags: diag.Diagnostics{
		//		diag.WithPath(
		//			tftypes.NewAttributePath(),
		//			reflect.DiagIntoIncompatibleType{
		//				Val:        tftypes.NewValue(tftypes.String, ""),
		//				TargetType: goreflect.TypeOf(int64(0)),
		//				Err:        fmt.Errorf("unexpected error was encountered trying to convert from value"),
		//			},
		//		),
		//	},
		//},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := ValueFrom(context.Background(), tc.target.Type(context.Background()), &tc.target, &tc.val)

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
