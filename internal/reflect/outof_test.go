// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package reflect_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFromValue_go_types(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		typ           attr.Type
		value         any
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"nil-go-slice-to-list-value": {
			typ:      types.ListType{ElemType: types.StringType},
			value:    new([]string),
			expected: types.ListNull(types.StringType),
		},
		"nil-go-slice-to-set-value": {
			typ:      types.SetType{ElemType: types.StringType},
			value:    new([]string),
			expected: types.SetNull(types.StringType),
		},
		"nil-go-slice-to-tuple-value": {
			typ:      types.TupleType{ElemTypes: []attr.Type{types.StringType, types.StringType}},
			value:    new([]string),
			expected: types.TupleNull([]attr.Type{types.StringType, types.StringType}),
		},
		"go-slice-to-list-value": {
			typ:   types.ListType{ElemType: types.StringType},
			value: []string{"hello", "world"},
			expected: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("hello"),
					types.StringValue("world"),
				},
			),
		},
		"go-slice-to-set-value": {
			typ:   types.SetType{ElemType: types.StringType},
			value: []string{"hello", "world"},
			expected: types.SetValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("hello"),
					types.StringValue("world"),
				},
			),
		},
		"go-slice-to-tuple-value": {
			typ:   types.TupleType{ElemTypes: []attr.Type{types.StringType, types.StringType}},
			value: []string{"hello", "world"},
			expected: types.TupleValueMust(
				[]attr.Type{
					types.StringType,
					types.StringType,
				},
				[]attr.Value{
					types.StringValue("hello"),
					types.StringValue("world"),
				},
			),
		},
		"go-slice-to-tuple-value-empty": {
			typ:      types.TupleType{ElemTypes: []attr.Type{}},
			value:    []any{},
			expected: types.TupleValueMust([]attr.Type{}, []attr.Value{}),
		},
		"go-slice-to-tuple-value-one-element": {
			typ:   types.TupleType{ElemTypes: []attr.Type{types.BoolType}},
			value: []bool{true},
			expected: types.TupleValueMust(
				[]attr.Type{
					types.BoolType,
				},
				[]attr.Value{
					types.BoolValue(true),
				},
			),
		},
		"go-slice-to-tuple-value-unsupported-no-element-types-with-values": {
			typ:   types.TupleType{ElemTypes: []attr.Type{}},
			value: []string{"hello", "world"},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from slice value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"cannot use type []string as schema type basetypes.TupleType; tuple type contained no element types but received values",
				),
			},
		},
		"go-slice-to-tuple-value-unsupported-multiple-element-types": {
			typ:   types.TupleType{ElemTypes: []attr.Type{types.StringType, types.BoolType}},
			value: []any{"hello", true},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from slice value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"cannot use type []interface {} as schema type basetypes.TupleType; reflection support for tuples is limited to multiple elements of the same element type. Expected all element types to be basetypes.StringType",
				),
			},
		},
		"go-slice-incompatible-type": {
			typ:   types.StringType,
			value: []string{"hello", "world"},
			expectedDiags: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Empty(),
					"Value Conversion Error",
					"An unexpected error was encountered trying to convert from slice value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
						"cannot use type []string as schema type basetypes.StringType; basetypes.StringType must be an attr.TypeWithElementType or attr.TypeWithElementTypes",
				),
			},
		},
		"go-struct-with-dynamic-values": {
			typ: types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"dynamic_attr": types.DynamicType,
					"dynamic_list": types.DynamicType,
				},
			},
			value: struct {
				DynamicAttr types.Dynamic `tfsdk:"dynamic_attr"`
				DynamicList types.Dynamic `tfsdk:"dynamic_list"`
			}{
				DynamicAttr: types.DynamicValue(types.StringValue("hello world")),
				DynamicList: types.DynamicValue(types.ListValueMust(
					types.BoolType,
					[]attr.Value{
						types.BoolValue(true),
						types.BoolValue(false),
					},
				)),
			},
			expected: types.ObjectValueMust(
				map[string]attr.Type{
					"dynamic_attr": types.DynamicType,
					"dynamic_list": types.DynamicType,
				},
				map[string]attr.Value{
					"dynamic_attr": types.DynamicValue(types.StringValue("hello world")),
					"dynamic_list": types.DynamicValue(types.ListValueMust(
						types.BoolType,
						[]attr.Value{
							types.BoolValue(true),
							types.BoolValue(false),
						},
					)),
				},
			),
		},
	}
	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := refl.FromValue(context.Background(), testCase.typ, testCase.value, path.Empty())

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiags); diff != "" {
				for _, d := range diags {
					t.Logf("%s: %s\n%s\n", d.Severity(), d.Summary(), d.Detail())
				}
				t.Errorf("unexpected diagnostics: %s", diff)
			}
		})
	}
}
