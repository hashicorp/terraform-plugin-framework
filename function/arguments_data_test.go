// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/function"
	fwreflect "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func TestArgumentsDataEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		argumentsData function.ArgumentsData
		other         function.ArgumentsData
		expected      bool
	}{
		"zero-zero": {
			argumentsData: function.ArgumentsData{},
			other:         function.ArgumentsData{},
			expected:      true,
		},
		"nil-nil": {
			argumentsData: function.NewArgumentsData(nil),
			other:         function.NewArgumentsData(nil),
			expected:      true,
		},
		"empty-empty": {
			argumentsData: function.NewArgumentsData([]attr.Value{}),
			other:         function.NewArgumentsData([]attr.Value{}),
			expected:      true,
		},
		"equal": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewStringValue("test"),
			}),
			other: function.NewArgumentsData([]attr.Value{
				basetypes.NewStringValue("test"),
			}),
			expected: true,
		},
		"different-types": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewStringValue("test"),
			}),
			other: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
			}),
			expected: false,
		},
		"different-values": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewStringValue("test1"),
			}),
			other: function.NewArgumentsData([]attr.Value{
				basetypes.NewStringValue("test2"),
			}),
			expected: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.argumentsData.Equal(testCase.other)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestArgumentsDataGet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		argumentsData      function.ArgumentsData
		targets            []any
		expected           []any
		expectedDiagnotics diag.Diagnostics
	}{
		"no-argument-data": {
			argumentsData: function.NewArgumentsData(nil),
			targets:       []any{new(bool)},
			expected:      []any{new(bool)},
			expectedDiagnotics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Argument Data Usage",
					"When attempting to fetch argument data during the function call, the provider code incorrectly attempted to read argument data. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Function does not have argument data.",
				),
			},
		},
		"invalid-targets-too-few": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
				basetypes.NewBoolNull(),
			}),
			targets:  []any{new(bool)},
			expected: []any{new(bool)},
			expectedDiagnotics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Argument Data Usage",
					"When attempting to fetch argument data during the function call, the provider code incorrectly attempted to read argument data. "+
						"The Get call requires all parameters and the final variadic parameter, if implemented, to be in the targets. "+
						"This is always an error in the provider code and should be reported to the provider developers.\n\n"+
						"Given targets count: 1, expected targets count: 2",
				),
			},
		},
		"invalid-targets-too-many": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
			}),
			targets:  []any{new(bool), new(bool)},
			expected: []any{new(bool), new(bool)},
			expectedDiagnotics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Argument Data Usage",
					"When attempting to fetch argument data during the function call, the provider code incorrectly attempted to read argument data. "+
						"The Get call requires all parameters and the final variadic parameter, if implemented, to be in the targets. "+
						"This is always an error in the provider code and should be reported to the provider developers.\n\n"+
						"Given targets count: 2, expected targets count: 1",
				),
			},
		},
		"invalid-target": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
			}),
			targets:  []any{new(basetypes.StringValue)},
			expected: []any{new(basetypes.StringValue)},
			expectedDiagnotics: diag.Diagnostics{
				diag.WithPath(
					path.Empty(),
					fwreflect.DiagNewAttributeValueIntoWrongType{
						ValType:    reflect.TypeOf(basetypes.BoolValue{}),
						TargetType: reflect.TypeOf(basetypes.StringValue{}),
						SchemaType: basetypes.BoolType{},
					},
				),
			},
		},
		"framework-type": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
				basetypes.NewInt64Unknown(),
				basetypes.NewStringValue("test"),
			}),
			targets: []any{
				new(basetypes.BoolValue),
				new(basetypes.Int64Value),
				new(basetypes.StringValue),
			},
			expected: []any{
				pointer(basetypes.NewBoolNull()),
				pointer(basetypes.NewInt64Unknown()),
				pointer(basetypes.NewStringValue("test")),
			},
		},
		"reflection": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
				basetypes.NewStringValue("test"),
			}),
			targets: []any{
				new(*bool),
				new(string),
			},
			expected: []any{
				pointer((*bool)(nil)),
				pointer("test"),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.argumentsData.Get(context.Background(), testCase.targets...)

			// Prevent awkwardness with comparing pointers in []any
			options := cmp.Options{
				cmp.Transformer("BoolValue", func(v *basetypes.BoolValue) basetypes.BoolValue {
					return *v
				}),
				cmp.Transformer("Int64Value", func(v *basetypes.Int64Value) basetypes.Int64Value {
					return *v
				}),
				cmp.Transformer("StringValue", func(v *basetypes.StringValue) basetypes.StringValue {
					return *v
				}),
			}

			if diff := cmp.Diff(testCase.targets, testCase.expected, options...); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnotics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestArgumentsDataGetArgument(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		argumentsData      function.ArgumentsData
		position           int
		target             any
		expected           any
		expectedDiagnotics diag.Diagnostics
	}{
		"no-argument-data": {
			argumentsData: function.NewArgumentsData(nil),
			position:      0,
			target:        new(bool),
			expected:      new(bool),
			expectedDiagnotics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Argument Data Usage",
					"When attempting to fetch argument data during the function call, the provider code incorrectly attempted to read argument data. "+
						"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
						"Function does not have argument data.",
				),
			},
		},
		"invalid-position": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
			}),
			position: 1,
			target:   new(bool),
			expected: new(bool),
			expectedDiagnotics: diag.Diagnostics{
				diag.NewErrorDiagnostic(
					"Invalid Argument Data Position",
					"When attempting to fetch argument data during the function call, the provider code attempted to read a non-existent argument position. "+
						"Function argument positions are 0-based and any final variadic parameter is represented as one argument position with an ordered list of the parameter data type. "+
						"This is always an error in the provider code and should be reported to the provider developers.\n\n"+
						"Given argument position: 1, last argument position: 0",
				),
			},
		},
		"invalid-target": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
			}),
			position: 0,
			target:   new(basetypes.StringValue),
			expected: new(basetypes.StringValue),
			expectedDiagnotics: diag.Diagnostics{
				diag.WithPath(
					path.Empty(),
					fwreflect.DiagNewAttributeValueIntoWrongType{
						ValType:    reflect.TypeOf(basetypes.BoolValue{}),
						TargetType: reflect.TypeOf(basetypes.StringValue{}),
						SchemaType: basetypes.BoolType{},
					},
				),
			},
		},
		"framework-type": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
			}),
			position: 0,
			target:   new(basetypes.BoolValue),
			expected: pointer(basetypes.NewBoolNull()),
		},
		"reflection": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
			}),
			position: 0,
			target:   new(*bool),
			expected: pointer((*bool)(nil)),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			diags := testCase.argumentsData.GetArgument(context.Background(), testCase.position, testCase.target)

			// Prevent awkwardness with comparing empty interface pointers
			options := cmp.Options{
				cmp.Transformer("BoolValue", func(v *basetypes.BoolValue) basetypes.BoolValue {
					return *v
				}),
				cmp.Transformer("StringValue", func(v *basetypes.StringValue) basetypes.StringValue {
					return *v
				}),
			}

			if diff := cmp.Diff(testCase.target, testCase.expected, options...); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(diags, testCase.expectedDiagnotics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
