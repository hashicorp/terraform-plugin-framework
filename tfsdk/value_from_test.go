// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfsdk

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type person struct {
	Name     types.String `tfsdk:"name"`
	Age      types.Int64  `tfsdk:"age"`
	OptedIn  types.Bool   `tfsdk:"opted_in"`
	Address  types.List   `tfsdk:"address"`
	FullName types.Map    `tfsdk:"full_name"`
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
		"full_name": types.MapType{
			ElemType: types.StringType,
		},
	}

	mrX := person{
		Name:    types.StringValue("x"),
		Age:     types.Int64Value(30),
		OptedIn: types.BoolValue(true),
		Address: types.ListValueMust(
			types.StringType,
			[]attr.Value{
				types.StringValue("1"),
				types.StringValue("Beckford Close"),
				types.StringValue("Gotham"),
			},
		),
		FullName: types.MapValueMust(
			types.StringType,
			map[string]attr.Value{
				"first":  types.StringValue("x"),
				"middle": types.StringValue("b"),
				"last":   types.StringValue("c"),
			},
		),
	}

	mrsY := person{
		Name:    types.StringValue("y"),
		Age:     types.Int64Value(23),
		OptedIn: types.BoolValue(false),
		Address: types.ListValueMust(
			types.StringType,
			[]attr.Value{
				types.StringValue("2"),
				types.StringValue("Windmill Close"),
				types.StringValue("Smallville"),
			},
		),
		FullName: types.MapValueMust(
			types.StringType,
			map[string]attr.Value{
				"first":  types.StringValue("y"),
				"middle": types.StringValue("e"),
				"last":   types.StringValue("f"),
			},
		),
	}

	expectedMrXObj := types.ObjectValueMust(
		personAttrTypes,
		map[string]attr.Value{
			"name":     types.StringValue("x"),
			"age":      types.Int64Value(30),
			"opted_in": types.BoolValue(true),
			"address": types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("1"),
					types.StringValue("Beckford Close"),
					types.StringValue("Gotham"),
				},
			),
			"full_name": types.MapValueMust(
				types.StringType,
				map[string]attr.Value{
					"first":  types.StringValue("x"),
					"middle": types.StringValue("b"),
					"last":   types.StringValue("c"),
				},
			),
		},
	)

	expectedMrsYObj := types.ObjectValueMust(
		personAttrTypes,
		map[string]attr.Value{
			"name":     types.StringValue("y"),
			"age":      types.Int64Value(23),
			"opted_in": types.BoolValue(false),
			"address": types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("2"),
					types.StringValue("Windmill Close"),
					types.StringValue("Smallville"),
				},
			),
			"full_name": types.MapValueMust(
				types.StringType,
				map[string]attr.Value{
					"first":  types.StringValue("y"),
					"middle": types.StringValue("e"),
					"last":   types.StringValue("f"),
				},
			),
		},
	)

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
			expected: types.StringValue("hello"),
		},
		"struct": {
			val:      mrX,
			target:   types.ObjectNull(personAttrTypes),
			expected: expectedMrXObj,
		},
		"list": {
			val: []person{mrX, mrsY},
			target: types.ListNull(
				types.ObjectType{
					AttrTypes: personAttrTypes,
				},
			),
			expected: types.ListValueMust(
				types.ObjectType{
					AttrTypes: personAttrTypes,
				},
				[]attr.Value{expectedMrXObj, expectedMrsYObj},
			),
		},
		"map": {
			val: map[string]person{
				"x": mrX,
				"y": mrsY,
			},
			target: types.MapNull(
				types.ObjectType{
					AttrTypes: personAttrTypes,
				},
			),
			expected: types.MapValueMust(
				types.ObjectType{
					AttrTypes: personAttrTypes,
				},
				map[string]attr.Value{
					"x": expectedMrXObj,
					"y": expectedMrsYObj,
				},
			),
		},
		"incompatible-type": {
			val:    0,
			target: types.String{},
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					path.Empty(),
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
