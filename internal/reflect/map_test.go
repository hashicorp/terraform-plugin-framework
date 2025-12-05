// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package reflect_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/internal/testing/testtypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestReflectMap_string(t *testing.T) {
	t.Parallel()

	var m map[string]string

	expected := map[string]string{
		"a": "red",
		"b": "blue",
		"c": "green",
	}

	result, diags := refl.Map(context.Background(), types.MapType{
		ElemType: types.StringType,
	}, tftypes.NewValue(tftypes.Map{
		ElementType: tftypes.String,
	}, map[string]tftypes.Value{
		"a": tftypes.NewValue(tftypes.String, "red"),
		"b": tftypes.NewValue(tftypes.String, "blue"),
		"c": tftypes.NewValue(tftypes.String, "green"),
	}), reflect.ValueOf(m), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&m).Elem().Set(result)
	for k, v := range expected {
		if got, ok := m[k]; !ok {
			t.Errorf("Expected %q to be set to %q, wasn't set", k, v)
		} else if got != v {
			t.Errorf("Expected %q to be %q, got %q", k, v, got)
		}
	}
}

func TestFromMap(t *testing.T) {
	t.Parallel()

	var m map[string]string

	testCases := map[string]struct {
		typ           attr.TypeWithElementType
		val           reflect.Value
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"null": {
			typ: types.MapType{
				ElemType: types.StringType,
			},
			val:      reflect.ValueOf(m),
			expected: types.MapNull(types.StringType),
		},
		"nullWithValidateError": {
			typ: testtypes.MapTypeWithValidateError{
				MapType: types.MapType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf(m),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
		"nullWithValidateAttributeError": {
			typ: testtypes.MapTypeWithValidateAttributeError{
				MapType: types.MapType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf(m),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
		"nullWithValidateWarning": {
			typ: testtypes.MapTypeWithValidateWarning{
				MapType: types.MapType{
					ElemType: types.StringType,
				},
			},
			val:      reflect.ValueOf(m),
			expected: types.MapNull(types.StringType),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"nullWithValidateAttributeWarning": {
			typ: testtypes.MapTypeWithValidateAttributeWarning{
				MapType: types.MapType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf(m),
			expected: testtypes.MapValueWithValidateAttributeWarning{
				Map: types.MapNull(types.StringType),
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"map": {
			typ: types.MapType{ElemType: types.StringType},
			val: reflect.ValueOf(map[string]string{"one": "a"}),
			expected: types.MapValueMust(types.StringType, map[string]attr.Value{
				"one": types.StringValue("a"),
			}),
		},
		"mapElemWithValidateError": {
			typ: types.MapType{
				ElemType: testtypes.StringTypeWithValidateError{
					StringType: testtypes.StringType{},
				},
			},
			val: reflect.ValueOf(map[string]string{"one": "a"}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty().AtMapKey("one")),
			},
		},
		"mapElemWithValidateAttributeError": {
			typ: types.MapType{
				ElemType: testtypes.StringTypeWithValidateAttributeError{
					StringType: testtypes.StringType{},
				},
			},
			val: reflect.ValueOf(map[string]string{"one": "a"}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty().AtMapKey("one")),
			},
		},
		"mapElemWithValidateWarning": {
			typ: types.MapType{
				ElemType: testtypes.StringTypeWithValidateWarning{
					StringType: testtypes.StringType{},
				},
			},
			val: reflect.ValueOf(map[string]string{"one": "a"}),
			expected: types.MapValueMust(testtypes.StringTypeWithValidateWarning{}, map[string]attr.Value{
				"one": testtypes.String{
					InternalString: types.StringValue("a"),
					CreatedBy:      testtypes.StringTypeWithValidateWarning{},
				},
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty().AtMapKey("one")),
			},
		},
		"mapElemWithValidateAttributeWarning": {
			typ: types.MapType{
				ElemType: testtypes.StringTypeWithValidateAttributeWarning{
					StringType: testtypes.StringType{},
				},
			},
			val: reflect.ValueOf(map[string]string{"one": "a"}),
			expected: types.MapValueMust(testtypes.StringTypeWithValidateAttributeWarning{}, map[string]attr.Value{
				"one": testtypes.StringValueWithValidateAttributeWarning{
					InternalString: testtypes.String{
						InternalString: types.StringValue("a"),
						CreatedBy:      testtypes.StringTypeWithValidateAttributeWarning{},
					},
				},
			}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty().AtMapKey("one")),
			},
		},
		"mapWithValidateError": {
			typ: testtypes.MapTypeWithValidateError{
				MapType: types.MapType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf(map[string]string{"one": "a"}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
		"mapWithValidateAttributeError": {
			typ: testtypes.MapTypeWithValidateAttributeError{
				MapType: types.MapType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf(map[string]string{"one": "a"}),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
		"mapWithValidateWarning": {
			typ: testtypes.MapTypeWithValidateWarning{
				MapType: types.MapType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf(map[string]string{"one": "a"}),
			expected: types.MapValueMust(
				types.StringType,
				map[string]attr.Value{
					"one": types.StringValue("a"),
				},
			),
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"listWithValidateAttributeWarning": {
			typ: testtypes.MapTypeWithValidateAttributeWarning{
				MapType: types.MapType{
					ElemType: types.StringType,
				},
			},
			val: reflect.ValueOf(map[string]string{"one": "a"}),
			expected: testtypes.MapValueWithValidateAttributeWarning{
				Map: types.MapValueMust(
					types.StringType,
					map[string]attr.Value{
						"one": types.StringValue("a"),
					},
				),
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := refl.FromMap(context.Background(), tc.typ, tc.val, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}
