// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestInto_Slices(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		typ           attr.Type
		value         tftypes.Value
		target        []string
		expected      []string
		expectedDiags diag.Diagnostics
	}{
		"list-to-go-slice": {
			typ: types.ListType{ElemType: types.StringType},
			value: tftypes.NewValue(tftypes.List{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
			target:   make([]string, 0),
			expected: []string{"hello", "world"},
		},
		"set-to-go-slice": {
			typ: types.SetType{ElemType: types.StringType},
			value: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
			target:   make([]string, 0),
			expected: []string{"hello", "world"},
		},
		"tuple-to-go-slice-unsupported": {
			typ: types.TupleType{ElemTypes: []attr.Type{types.StringType, types.StringType}},
			value: tftypes.NewValue(tftypes.Tuple{
				ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "hello"),
				tftypes.NewValue(tftypes.String, "world"),
			}),
			target:   make([]string, 0),
			expected: make([]string, 0),
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					path.Empty(),
					refl.DiagIntoIncompatibleType{
						Val: tftypes.NewValue(tftypes.Tuple{
							ElementTypes: []tftypes.Type{tftypes.String, tftypes.String},
						}, []tftypes.Value{
							tftypes.NewValue(tftypes.String, "hello"),
							tftypes.NewValue(tftypes.String, "world"),
						}),
						TargetType: reflect.TypeOf([]string{}),
						Err:        errors.New("cannot reflect tftypes.Tuple[tftypes.String, tftypes.String] using type information provided by basetypes.TupleType, reflection support is currently not implemented for tuples"),
					},
				),
			},
		},
		"list-to-incompatible-type": {
			typ:      types.ListType{ElemType: types.StringType},
			value:    tftypes.NewValue(tftypes.String, "hello"),
			target:   make([]string, 0),
			expected: make([]string, 0),
			expectedDiags: diag.Diagnostics{
				diag.WithPath(
					path.Empty(),
					refl.DiagIntoIncompatibleType{
						Val:        tftypes.NewValue(tftypes.String, "hello"),
						TargetType: reflect.TypeOf([]string{}),
						Err:        errors.New("can't unmarshal tftypes.String into *[]tftypes.Value expected []tftypes.Value"),
					},
				),
			},
		},
	}
	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := refl.Into(context.Background(), testCase.typ, testCase.value, &testCase.target, refl.Options{}, path.Empty())

			if diff := cmp.Diff(testCase.target, testCase.expected); diff != "" {
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
