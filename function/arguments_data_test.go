// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package function_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
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
		argumentsData function.ArgumentsData
		targets       []any
		expected      []any
		expectedErr   *function.FuncError
	}{
		"no-argument-data": {
			argumentsData: function.NewArgumentsData(nil),
			targets:       []any{new(bool)},
			expected:      []any{new(bool)},
			expectedErr: function.NewFuncError("Invalid Argument Data Usage: When attempting to fetch argument data during the function call, the provider code incorrectly attempted to read argument data. " +
				"This is always an issue in the provider code and should be reported to the provider developers.\n\n" +
				"Function does not have argument data."),
		},
		"invalid-targets-too-few": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
				basetypes.NewBoolNull(),
			}),
			targets:  []any{new(bool)},
			expected: []any{new(bool)},
			expectedErr: function.NewFuncError("Invalid Argument Data Usage: When attempting to fetch argument data during the function call, the provider code incorrectly attempted to read argument data. " +
				"The Get call requires all parameters and the final variadic parameter, if implemented, to be in the targets. " +
				"This is always an error in the provider code and should be reported to the provider developers.\n\n" +
				"Given targets count: 1, expected targets count: 2"),
		},
		"invalid-targets-too-many": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
			}),
			targets:  []any{new(bool), new(bool)},
			expected: []any{new(bool), new(bool)},
			expectedErr: function.NewFuncError("Invalid Argument Data Usage: When attempting to fetch argument data during the function call, the provider code incorrectly attempted to read argument data. " +
				"The Get call requires all parameters and the final variadic parameter, if implemented, to be in the targets. " +
				"This is always an error in the provider code and should be reported to the provider developers.\n\n" +
				"Given targets count: 2, expected targets count: 1"),
		},
		"invalid-target": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
			}),
			targets:  []any{new(basetypes.StringValue)},
			expected: []any{new(basetypes.StringValue)},
			expectedErr: function.NewFuncError("Value Conversion Error: An unexpected error was encountered trying to convert into a Terraform value. " +
				"This is always an error in the provider. Please report the following to the provider developer:\n\n" +
				"Cannot use attr.Value basetypes.StringValue, only basetypes.BoolValue is supported because basetypes.BoolType is the type in the schema"),
		},
		"attr-value": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
				basetypes.NewInt64Unknown(),
				basetypes.NewStringValue("test"),
			}),
			targets: []any{
				new(attr.Value),
				new(attr.Value),
				new(attr.Value),
			},
			expected: []any{
				pointer(attr.Value(basetypes.NewBoolNull())),
				pointer(attr.Value(basetypes.NewInt64Unknown())),
				pointer(attr.Value(basetypes.NewStringValue("test"))),
			},
		},
		"attr-value-variadic": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
				basetypes.NewInt64Unknown(),
				basetypes.NewStringValue("test"),
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("test1"),
						basetypes.NewStringValue("test2"),
					},
				),
			}),
			targets: []any{
				new(attr.Value),
				new(attr.Value),
				new(attr.Value),
				new(attr.Value),
			},
			expected: []any{
				pointer(attr.Value(basetypes.NewBoolNull())),
				pointer(attr.Value(basetypes.NewInt64Unknown())),
				pointer(attr.Value(basetypes.NewStringValue("test"))),
				pointer(attr.Value(
					basetypes.NewTupleValueMust(
						[]attr.Type{
							basetypes.StringType{},
							basetypes.StringType{},
						},
						[]attr.Value{
							basetypes.NewStringValue("test1"),
							basetypes.NewStringValue("test2"),
						},
					),
				)),
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
		"framework-type-variadic": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
				basetypes.NewInt64Unknown(),
				basetypes.NewStringValue("test"),
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("test1"),
						basetypes.NewStringValue("test2"),
					},
				),
			}),
			targets: []any{
				new(basetypes.BoolValue),
				new(basetypes.Int64Value),
				new(basetypes.StringValue),
				new(basetypes.TupleValue),
			},
			expected: []any{
				pointer(basetypes.NewBoolNull()),
				pointer(basetypes.NewInt64Unknown()),
				pointer(basetypes.NewStringValue("test")),
				pointer(
					basetypes.NewTupleValueMust(
						[]attr.Type{
							basetypes.StringType{},
							basetypes.StringType{},
						},
						[]attr.Value{
							basetypes.NewStringValue("test1"),
							basetypes.NewStringValue("test2"),
						},
					),
				),
			},
		},
		"dynamic-framework-type": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewDynamicValue(basetypes.NewStringValue("dynamic_test")),
				basetypes.NewDynamicValue(basetypes.NewListValueMust(
					basetypes.StringType{},
					[]attr.Value{
						basetypes.NewStringValue("hello"),
						basetypes.NewStringValue("dynamic"),
						basetypes.NewStringValue("world"),
					},
				)),
			}),
			targets: []any{
				new(basetypes.DynamicValue),
				new(basetypes.DynamicValue),
			},
			expected: []any{
				pointer(basetypes.NewDynamicValue(basetypes.NewStringValue("dynamic_test"))),
				pointer(basetypes.NewDynamicValue(basetypes.NewListValueMust(
					basetypes.StringType{},
					[]attr.Value{
						basetypes.NewStringValue("hello"),
						basetypes.NewStringValue("dynamic"),
						basetypes.NewStringValue("world"),
					},
				))),
			},
		},
		"dynamic-framework-type-variadic": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.DynamicType{},
						basetypes.DynamicType{},
						basetypes.DynamicType{},
					},
					[]attr.Value{
						basetypes.NewDynamicValue(basetypes.NewStringValue("test1")),
						basetypes.NewDynamicValue(basetypes.NewNumberValue(big.NewFloat(1.23))),
						basetypes.NewDynamicValue(basetypes.NewBoolValue(true)),
					},
				),
			}),
			targets: []any{
				new([]basetypes.DynamicValue),
			},
			expected: []any{
				pointer([]basetypes.DynamicValue{
					basetypes.NewDynamicValue(basetypes.NewStringValue("test1")),
					basetypes.NewDynamicValue(basetypes.NewNumberValue(big.NewFloat(1.23))),
					basetypes.NewDynamicValue(basetypes.NewBoolValue(true)),
				}),
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
		"reflection-variadic": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
				basetypes.NewTupleValueMust(
					[]attr.Type{
						basetypes.StringType{},
						basetypes.StringType{},
					},
					[]attr.Value{
						basetypes.NewStringValue("test1"),
						basetypes.NewStringValue("test2"),
					},
				),
			}),
			targets: []any{
				new(*bool),
				new([]string),
			},
			expected: []any{
				pointer((*bool)(nil)),
				pointer([]string{
					"test1",
					"test2",
				}),
			},
		},
		"reflection-variadic-empty": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
				basetypes.NewTupleValueMust([]attr.Type{}, []attr.Value{}),
			}),
			targets: []any{
				new(*bool),
				new([]string),
			},
			expected: []any{
				pointer((*bool)(nil)),
				pointer([]string{}),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.argumentsData.Get(context.Background(), testCase.targets...)

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
				cmp.Transformer("TupleValue", func(v *basetypes.TupleValue) basetypes.TupleValue {
					return *v
				}),
				cmp.Transformer("DynamicValue", func(v *basetypes.DynamicValue) basetypes.DynamicValue {
					return *v
				}),
			}

			if diff := cmp.Diff(testCase.targets, testCase.expected, options...); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			if diff := cmp.Diff(err, testCase.expectedErr); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}

func TestArgumentsDataGetArgument(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		argumentsData function.ArgumentsData
		position      int
		target        any
		expected      any
		expectedErr   *function.FuncError
	}{
		"no-argument-data": {
			argumentsData: function.NewArgumentsData(nil),
			position:      0,
			target:        new(bool),
			expected:      new(bool),
			expectedErr: function.NewArgumentFuncError(int64(0), "Invalid Argument Data Usage: When attempting to fetch argument data during the function call, the provider code incorrectly attempted to read argument data. "+
				"This is always an issue in the provider code and should be reported to the provider developers.\n\n"+
				"Function does not have argument data."),
		},
		"invalid-position": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
			}),
			position: 1,
			target:   new(bool),
			expected: new(bool),
			expectedErr: function.NewArgumentFuncError(int64(1), "Invalid Argument Data Position: When attempting to fetch argument data during the function call, the provider code attempted to read a non-existent argument position. "+
				"Function argument positions are 0-based and any final variadic parameter is represented as one argument position with a tuple where each element "+
				"type matches the parameter data type. This is always an error in the provider code and should be reported to the provider developers.\n\n"+
				"Given argument position: 1, last argument position: 0",
			),
		},
		"invalid-target": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolValue(true),
			}),
			position: 0,
			target:   new(basetypes.StringValue),
			expected: new(basetypes.StringValue),
			expectedErr: function.NewFuncError("Value Conversion Error: An unexpected error was encountered trying to convert into a Terraform value. " +
				"This is always an error in the provider. Please report the following to the provider developer:\n\n" +
				"Cannot use attr.Value basetypes.StringValue, only basetypes.BoolValue is supported because basetypes.BoolType is the type in the schema"),
		},
		"attr-value": {
			argumentsData: function.NewArgumentsData([]attr.Value{
				basetypes.NewBoolNull(),
			}),
			position: 0,
			target:   new(attr.Value),
			expected: pointer(attr.Value(basetypes.NewBoolNull())),
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := testCase.argumentsData.GetArgument(context.Background(), testCase.position, testCase.target)

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

			if diff := cmp.Diff(err, testCase.expectedErr); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
