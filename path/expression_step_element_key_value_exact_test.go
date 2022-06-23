package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpressionStepElementKeyValueExactEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyValueExact
		other    path.ExpressionStep
		expected bool
	}{
		"ExpressionStepAttributeNameExact": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "test"}},
			other:    path.ExpressionStepAttributeNameExact("test"),
			expected: false,
		},
		"ExpressionStepElementKeyIntExact": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "test"}},
			other:    path.ExpressionStepElementKeyIntExact(0),
			expected: false,
		},
		"ExpressionStepElementKeyStringExact": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "test"}},
			other:    path.ExpressionStepElementKeyStringExact("test"),
			expected: false,
		},
		"ExpressionStepElementKeyValueExact-different": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "test"}},
			other:    path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "not-test"}},
			expected: false,
		},
		"ExpressionStepElementKeyValueExact-equal": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "test"}},
			other:    path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "test"}},
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

func TestExpressionStepElementKeyValueExactMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyValueExact
		pathStep path.PathStep
		expected bool
	}{
		"StepAttributeName": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "test"}},
			pathStep: path.PathStepAttributeName("test"),
			expected: false,
		},
		"StepElementKeyInt": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "test"}},
			pathStep: path.PathStepElementKeyInt(0),
			expected: false,
		},
		"StepElementKeyString": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "test"}},
			pathStep: path.PathStepElementKeyString("test"),
			expected: false,
		},
		"StepElementKeyValue-different": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "test"}},
			pathStep: path.PathStepElementKeyValue{Value: types.String{Value: "not-test"}},
			expected: false,
		},
		"StepElementKeyValue-equal": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "test"}},
			pathStep: path.PathStepElementKeyValue{Value: types.String{Value: "test"}},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.step.Matches(testCase.pathStep)

			if got != testCase.expected {
				t.Errorf("expected %t, got %t", testCase.expected, got)
			}
		})
	}
}

func TestExpressionStepElementKeyValueExactString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyValueExact
		expected string
	}{
		"bool-value": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.Bool{Value: true}},
			expected: `[Value(true)]`,
		},
		"float64-value": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.Float64{Value: 1.2}},
			expected: `[Value(1.200000)]`,
		},
		"int64-value": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.Int64{Value: 123}},
			expected: `[Value(123)]`,
		},
		"list-value": {
			step: path.ExpressionStepElementKeyValueExact{Value: types.List{
				Elems: []attr.Value{
					types.String{Value: "test-element-1"},
					types.String{Value: "test-element-2"},
				},
				ElemType: types.StringType,
			}},
			expected: `[Value(["test-element-1","test-element-2"])]`,
		},
		"map-value": {
			step: path.ExpressionStepElementKeyValueExact{Value: types.Map{
				Elems: map[string]attr.Value{
					"test-key-1": types.String{Value: "test-value-1"},
					"test-key-2": types.String{Value: "test-value-2"},
				},
				ElemType: types.StringType,
			}},
			expected: `[Value({"test-key-1":"test-value-1","test-key-2":"test-value-2"})]`,
		},
		"object-value": {
			step: path.ExpressionStepElementKeyValueExact{Value: types.Object{
				Attrs: map[string]attr.Value{
					"test_attr_1": types.Bool{Value: true},
					"test_attr_2": types.String{Value: "test-value"},
				},
				AttrTypes: map[string]attr.Type{
					"test_attr_1": types.BoolType,
					"test_attr_2": types.StringType,
				},
			}},
			expected: `[Value({"test_attr_1":true,"test_attr_2":"test-value"})]`,
		},
		"string-null": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Null: true}},
			expected: `[Value(<null>)]`,
		},
		"string-unknown": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Unknown: true}},
			expected: `[Value(<unknown>)]`,
		},
		"string-value": {
			step:     path.ExpressionStepElementKeyValueExact{Value: types.String{Value: "test"}},
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
