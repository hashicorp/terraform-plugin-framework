// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPathStepElementKeyValueEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.PathStepElementKeyValue
		other    path.PathStep
		expected bool
	}{
		"PathStepAttributeName": {
			step:     path.PathStepElementKeyValue{Value: types.StringValue("test")},
			other:    path.PathStepAttributeName("test"),
			expected: false,
		},
		"PathStepElementKeyInt": {
			step:     path.PathStepElementKeyValue{Value: types.StringValue("test")},
			other:    path.PathStepElementKeyInt(0),
			expected: false,
		},
		"PathStepElementKeyString": {
			step:     path.PathStepElementKeyValue{Value: types.StringValue("test")},
			other:    path.PathStepElementKeyString("test"),
			expected: false,
		},
		"PathStepElementKeyValue-different-type": {
			step:     path.PathStepElementKeyValue{Value: types.BoolValue(true)},
			other:    path.PathStepElementKeyValue{Value: types.StringValue("not-test")},
			expected: false,
		},
		"PathStepElementKeyValue-different-value": {
			step:     path.PathStepElementKeyValue{Value: types.StringValue("test")},
			other:    path.PathStepElementKeyValue{Value: types.StringValue("not-test")},
			expected: false,
		},
		"PathStepElementKeyValue-equal": {
			step:     path.PathStepElementKeyValue{Value: types.StringValue("test")},
			other:    path.PathStepElementKeyValue{Value: types.StringValue("test")},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.step.Equal(testCase.other)

			if got != testCase.expected {
				t.Errorf("expected %t, got %t", testCase.expected, got)
			}
		})
	}
}

func TestPathStepElementKeyValueExpressionStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.PathStepElementKeyValue
		expected path.ExpressionStep
	}{
		"basic": {
			step:     path.PathStepElementKeyValue{Value: types.StringValue("test")},
			expected: path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test")},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.step.ExpressionStep()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestPathStepElementKeyValueString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.PathStepElementKeyValue
		expected string
	}{
		"bool-value": {
			step:     path.PathStepElementKeyValue{Value: types.BoolValue(true)},
			expected: `[Value(true)]`,
		},
		"float64-value": {
			step:     path.PathStepElementKeyValue{Value: types.Float64Value(1.2)},
			expected: `[Value(1.200000)]`,
		},
		"int64-value": {
			step:     path.PathStepElementKeyValue{Value: types.Int64Value(123)},
			expected: `[Value(123)]`,
		},
		"list-value": {
			step: path.PathStepElementKeyValue{Value: types.ListValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("test-element-1"),
					types.StringValue("test-element-2"),
				},
			)},
			expected: `[Value(["test-element-1","test-element-2"])]`,
		},
		"map-value": {
			step: path.PathStepElementKeyValue{Value: types.MapValueMust(
				types.StringType,
				map[string]attr.Value{
					"test-key-1": types.StringValue("test-value-1"),
					"test-key-2": types.StringValue("test-value-2"),
				},
			)},
			expected: `[Value({"test-key-1":"test-value-1","test-key-2":"test-value-2"})]`,
		},
		"object-value": {
			step: path.PathStepElementKeyValue{Value: types.ObjectValueMust(
				map[string]attr.Type{
					"test_attr_1": types.BoolType,
					"test_attr_2": types.StringType,
				},
				map[string]attr.Value{
					"test_attr_1": types.BoolValue(true),
					"test_attr_2": types.StringValue("test-value"),
				},
			)},
			expected: `[Value({"test_attr_1":true,"test_attr_2":"test-value"})]`,
		},
		"string-null": {
			step:     path.PathStepElementKeyValue{Value: types.StringNull()},
			expected: `[Value(<null>)]`,
		},
		"string-unknown": {
			step:     path.PathStepElementKeyValue{Value: types.StringUnknown()},
			expected: `[Value(<unknown>)]`,
		},
		"string-value": {
			step:     path.PathStepElementKeyValue{Value: types.StringValue("test")},
			expected: `[Value("test")]`,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.step.String()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
