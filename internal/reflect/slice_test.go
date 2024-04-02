// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFromSlice(t *testing.T) {
	t.Parallel()

	var s []string

	testCases := map[string]struct {
		typ           attr.Type
		val           reflect.Value
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"null": {
			typ: types.ListType{
				ElemType: types.StringType,
			},
			val:      reflect.ValueOf(s),
			expected: types.ListNull(types.StringType),
		},
		"nullWithValidateError": {
			typ: testtypes.ListTypeWithValidateError{
				ListType: types.ListType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf(s),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
		"nullWithValidateAttributeError": {
			typ: testtypes.ListTypeWithValidateAttributeError{
				ListType: types.ListType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf(s),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
		"nullWithValidateWarning": {
			typ: testtypes.ListTypeWithValidateWarning{
				ListType: types.ListType{
					ElemType: types.StringType,
				},
			},
			val:      reflect.ValueOf(s),
			expected: types.ListNull(types.StringType),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"nullWithValidateAttributeWarning": {
			typ: testtypes.ListTypeWithValidateAttributeWarning{
				ListType: types.ListType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf(s),
			expected: testtypes.ListValueWithValidateAttributeWarning{
				List: types.ListNull(types.StringType),
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"list": {
			typ: types.ListType{ElemType: types.StringType},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expected: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("a"),
				types.StringValue("b"),
				types.StringValue("c"),
			}),
		},
		"listElemWithValidateError": {
			typ: types.ListType{
				ElemType: testtypes.StringTypeWithValidateError{
					StringType: testtypes.StringType{},
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty().AtListIndex(0)),
			},
		},
		"listElemWithValidateAttributeError": {
			typ: types.ListType{
				ElemType: testtypes.StringTypeWithValidateAttributeError{
					StringType: testtypes.StringType{},
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty().AtListIndex(0)),
			},
		},
		"listElemWithValidateWarning": {
			typ: types.ListType{
				ElemType: testtypes.StringTypeWithValidateWarning{
					StringType: testtypes.StringType{},
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expected: types.ListValueMust(testtypes.StringTypeWithValidateWarning{}, []attr.Value{
				testtypes.String{
					InternalString: types.StringValue("a"),
					CreatedBy:      testtypes.StringTypeWithValidateWarning{},
				},
				testtypes.String{
					InternalString: types.StringValue("b"),
					CreatedBy:      testtypes.StringTypeWithValidateWarning{},
				},
				testtypes.String{
					InternalString: types.StringValue("c"),
					CreatedBy:      testtypes.StringTypeWithValidateWarning{},
				},
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty().AtListIndex(0)),
				testtypes.TestWarningDiagnostic(path.Empty().AtListIndex(1)),
				testtypes.TestWarningDiagnostic(path.Empty().AtListIndex(2)),
			},
		},
		"listElemWithValidateAttributeWarning": {
			typ: types.ListType{
				ElemType: testtypes.StringTypeWithValidateAttributeWarning{
					StringType: testtypes.StringType{},
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expected: types.ListValueMust(testtypes.StringTypeWithValidateAttributeWarning{}, []attr.Value{
				testtypes.StringValueWithValidateAttributeWarning{
					InternalString: testtypes.String{
						InternalString: types.StringValue("a"),
						CreatedBy:      testtypes.StringTypeWithValidateAttributeWarning{},
					},
				},
				testtypes.StringValueWithValidateAttributeWarning{
					InternalString: testtypes.String{
						InternalString: types.StringValue("b"),
						CreatedBy:      testtypes.StringTypeWithValidateAttributeWarning{},
					},
				},
				testtypes.StringValueWithValidateAttributeWarning{
					InternalString: testtypes.String{
						InternalString: types.StringValue("c"),
						CreatedBy:      testtypes.StringTypeWithValidateAttributeWarning{},
					},
				},
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty().AtListIndex(0)),
				testtypes.TestWarningDiagnostic(path.Empty().AtListIndex(1)),
				testtypes.TestWarningDiagnostic(path.Empty().AtListIndex(2)),
			},
		},
		"tuple": {
			typ: types.TupleType{
				ElemTypes: []attr.Type{
					types.StringType,
					types.StringType,
					types.StringType,
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expected: types.TupleValueMust(
				[]attr.Type{
					types.StringType,
					types.StringType,
					types.StringType,
				},
				[]attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
					types.StringValue("c"),
				},
			),
		},
		"tupleElemWithValidateError": {
			typ: types.TupleType{
				ElemTypes: []attr.Type{
					testtypes.StringTypeWithValidateError{
						StringType: testtypes.StringType{},
					},
					testtypes.StringTypeWithValidateError{
						StringType: testtypes.StringType{},
					},
					testtypes.StringTypeWithValidateError{
						StringType: testtypes.StringType{},
					},
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty().AtListIndex(0)),
			},
		},
		"tupleElemWithValidateAttributeError": {
			typ: types.TupleType{
				ElemTypes: []attr.Type{
					testtypes.StringTypeWithValidateAttributeError{
						StringType: testtypes.StringType{},
					},
					testtypes.StringTypeWithValidateAttributeError{
						StringType: testtypes.StringType{},
					},
					testtypes.StringTypeWithValidateAttributeError{
						StringType: testtypes.StringType{},
					},
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty().AtListIndex(0)),
			},
		},
		"tupleElemWithValidateWarning": {
			typ: types.TupleType{
				ElemTypes: []attr.Type{
					testtypes.StringTypeWithValidateWarning{
						StringType: testtypes.StringType{},
					},
					testtypes.StringTypeWithValidateWarning{
						StringType: testtypes.StringType{},
					},
					testtypes.StringTypeWithValidateWarning{
						StringType: testtypes.StringType{},
					},
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expected: types.TupleValueMust(
				[]attr.Type{
					testtypes.StringTypeWithValidateWarning{},
					testtypes.StringTypeWithValidateWarning{},
					testtypes.StringTypeWithValidateWarning{},
				},
				[]attr.Value{
					testtypes.String{
						InternalString: types.StringValue("a"),
						CreatedBy:      testtypes.StringTypeWithValidateWarning{},
					},
					testtypes.String{
						InternalString: types.StringValue("b"),
						CreatedBy:      testtypes.StringTypeWithValidateWarning{},
					},
					testtypes.String{
						InternalString: types.StringValue("c"),
						CreatedBy:      testtypes.StringTypeWithValidateWarning{},
					},
				}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty().AtListIndex(0)),
				testtypes.TestWarningDiagnostic(path.Empty().AtListIndex(1)),
				testtypes.TestWarningDiagnostic(path.Empty().AtListIndex(2)),
			},
		},
		"tupleElemWithValidateAttributeWarning": {
			typ: types.TupleType{
				ElemTypes: []attr.Type{
					testtypes.StringTypeWithValidateAttributeWarning{
						StringType: testtypes.StringType{},
					},
					testtypes.StringTypeWithValidateAttributeWarning{
						StringType: testtypes.StringType{},
					},
					testtypes.StringTypeWithValidateAttributeWarning{
						StringType: testtypes.StringType{},
					},
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expected: types.TupleValueMust(
				[]attr.Type{
					testtypes.StringTypeWithValidateAttributeWarning{},
					testtypes.StringTypeWithValidateAttributeWarning{},
					testtypes.StringTypeWithValidateAttributeWarning{},
				},
				[]attr.Value{
					testtypes.StringValueWithValidateAttributeWarning{
						InternalString: testtypes.String{
							InternalString: types.StringValue("a"),
							CreatedBy:      testtypes.StringTypeWithValidateAttributeWarning{},
						},
					},
					testtypes.StringValueWithValidateAttributeWarning{
						InternalString: testtypes.String{
							InternalString: types.StringValue("b"),
							CreatedBy:      testtypes.StringTypeWithValidateAttributeWarning{},
						},
					},
					testtypes.StringValueWithValidateAttributeWarning{
						InternalString: testtypes.String{
							InternalString: types.StringValue("c"),
							CreatedBy:      testtypes.StringTypeWithValidateAttributeWarning{},
						},
					},
				}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty().AtListIndex(0)),
				testtypes.TestWarningDiagnostic(path.Empty().AtListIndex(1)),
				testtypes.TestWarningDiagnostic(path.Empty().AtListIndex(2)),
			},
		},
		"listWithValidateError": {
			typ: testtypes.ListTypeWithValidateError{
				ListType: types.ListType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
		"listWithValidateAttributeError": {
			typ: testtypes.ListTypeWithValidateAttributeError{
				ListType: types.ListType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
		"listWithValidateWarning": {
			typ: testtypes.ListTypeWithValidateWarning{
				ListType: types.ListType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expected: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("a"),
					types.StringValue("b"),
					types.StringValue("c"),
				},
			),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"listWithValidateAttributeWarning": {
			typ: testtypes.ListTypeWithValidateAttributeWarning{
				ListType: types.ListType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf([]string{"a", "b", "c"}),
			expected: testtypes.ListValueWithValidateAttributeWarning{
				List: types.ListValueMust(
					types.StringType,
					[]attr.Value{
						types.StringValue("a"),
						types.StringValue("b"),
						types.StringValue("c"),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := refl.FromSlice(context.Background(), tc.typ, tc.val, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}
