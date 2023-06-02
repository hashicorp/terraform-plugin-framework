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

func TestPathStepsAppend(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps    path.PathSteps
		add      path.PathSteps
		expected path.PathSteps
	}{
		"empty-empty": {
			steps:    path.PathSteps{},
			add:      path.PathSteps{},
			expected: path.PathSteps{},
		},
		"empty-nonempty": {
			steps:    path.PathSteps{},
			add:      path.PathSteps{path.PathStepAttributeName("test")},
			expected: path.PathSteps{path.PathStepAttributeName("test")},
		},
		"nonempty-empty": {
			steps:    path.PathSteps{path.PathStepAttributeName("test")},
			add:      path.PathSteps{},
			expected: path.PathSteps{path.PathStepAttributeName("test")},
		},
		"nonempty-nonempty": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
			},
			add: path.PathSteps{
				path.PathStepAttributeName("add-test"),
				path.PathStepElementKeyString("add-test-key"),
			},
			expected: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
				path.PathStepAttributeName("add-test"),
				path.PathStepElementKeyString("add-test-key"),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.steps.Append(testCase.add...)

			if diff := cmp.Diff(testCase.steps, testCase.expected); diff != "" {
				t.Errorf("unexpected original difference: %s", diff)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected result difference: %s", diff)
			}
		})
	}
}

func TestPathStepsCopy(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps    path.PathSteps
		expected path.PathSteps
	}{
		"nil": {
			steps:    nil,
			expected: nil,
		},
		"empty": {
			steps:    path.PathSteps{},
			expected: path.PathSteps{},
		},
		"shallow": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			expected: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
		},
		"deep": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepElementKeyInt(0),
				path.PathStepAttributeName("test2"),
			},
			expected: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepElementKeyInt(0),
				path.PathStepAttributeName("test2"),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.steps.Copy()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}

			// Ensure original is not modified
			got.Append(path.PathStepAttributeName("modify-test"))

			if diff := cmp.Diff(got, testCase.expected); diff == "" {
				t.Error("unexpected modification")
			}
		})
	}
}

func TestPathStepsEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps    path.PathSteps
		other    path.PathSteps
		expected bool
	}{
		"nil-nil": {
			steps:    nil,
			other:    nil,
			expected: true,
		},
		"empty-empty": {
			steps:    path.PathSteps{},
			other:    path.PathSteps{},
			expected: true,
		},
		"different-length": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test2"),
			},
			other: path.PathSteps{
				path.PathStepAttributeName("test1"),
			},
			expected: false,
		},
		"PathStepAttributeName-different": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			other: path.PathSteps{
				path.PathStepAttributeName("not-test"),
			},
			expected: false,
		},
		"PathStepAttributeName-equal": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			other: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			expected: true,
		},
		"PathStepAttributeName-PathStepElementKeyInt-different": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
			},
			other: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(1),
			},
			expected: false,
		},
		"PathStepAttributeName-PathStepElementKeyInt-equal": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
			},
			other: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
			},
			expected: true,
		},
		"PathStepAttributeName-PathStepElementKeyString-different": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyString("test-key"),
			},
			other: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyString("not-test-key"),
			},
			expected: false,
		},
		"PathStepAttributeName-PathStepElementKeyString-equal": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyString("test-key"),
			},
			other: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyString("test-key"),
			},
			expected: true,
		},
		"PathStepAttributeName-PathStepElementKeyValue-different": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			other: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyValue{Value: types.StringValue("not-test-value")},
			},
			expected: false,
		},
		"PathStepAttributeName-PathStepElementKeyValue-equal": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			other: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			expected: true,
		},
		"PathStepElementKeyInt-different": {
			steps: path.PathSteps{
				path.PathStepElementKeyInt(0),
			},
			other: path.PathSteps{
				path.PathStepElementKeyInt(1),
			},
			expected: false,
		},
		"PathStepElementKeyInt-equal": {
			steps: path.PathSteps{
				path.PathStepElementKeyInt(0),
			},
			other: path.PathSteps{
				path.PathStepElementKeyInt(0),
			},
			expected: true,
		},
		"PathStepElementKeyString-different": {
			steps: path.PathSteps{
				path.PathStepElementKeyString("test"),
			},
			other: path.PathSteps{
				path.PathStepElementKeyString("not-test"),
			},
			expected: false,
		},
		"PathStepElementKeyString-equal": {
			steps: path.PathSteps{
				path.PathStepElementKeyString("test"),
			},
			other: path.PathSteps{
				path.PathStepElementKeyString("test"),
			},
			expected: true,
		},
		"PathStepElementKeyValue-different": {
			steps: path.PathSteps{
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			other: path.PathSteps{
				path.PathStepElementKeyValue{Value: types.StringValue("not-test-value")},
			},
			expected: false,
		},
		"PathStepElementKeyValue-equal": {
			steps: path.PathSteps{
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			other: path.PathSteps{
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.steps.Equal(testCase.other)

			if got != testCase.expected {
				t.Errorf("expected %t, got %t", testCase.expected, got)
			}
		})
	}
}

func TestPathStepsExpressionSteps(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps    path.PathSteps
		expected path.ExpressionSteps
	}{
		"nil": {
			steps:    nil,
			expected: path.ExpressionSteps{},
		},
		"empty": {
			steps:    path.PathSteps{},
			expected: path.ExpressionSteps{},
		},
		"one": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
		},
		"two": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(1),
			},
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(1),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.steps.ExpressionSteps()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestPathStepsLastStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps             path.PathSteps
		expectedLastStep  path.PathStep
		expectedRemaining path.PathSteps
	}{
		"nil": {
			steps:             nil,
			expectedLastStep:  nil,
			expectedRemaining: nil,
		},
		"empty": {
			steps:             path.PathSteps{},
			expectedLastStep:  nil,
			expectedRemaining: nil,
		},
		"one": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			expectedLastStep:  path.PathStepAttributeName("test"),
			expectedRemaining: nil,
		},
		"two": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
			},
			expectedLastStep: path.PathStepElementKeyInt(0),
			expectedRemaining: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
		},
		"three": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
				path.PathStepAttributeName("nested-test"),
			},
			expectedLastStep: path.PathStepAttributeName("nested-test"),
			expectedRemaining: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotLastStep, gotRemaining := testCase.steps.LastStep()

			if diff := cmp.Diff(gotLastStep, testCase.expectedLastStep); diff != "" {
				t.Errorf("unexpected last step difference: %s", diff)
			}

			if diff := cmp.Diff(gotRemaining, testCase.expectedRemaining); diff != "" {
				t.Errorf("unexpected remaining difference: %s", diff)
			}
		})
	}
}

func TestPathStepsNextStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps             path.PathSteps
		expectedNextStep  path.PathStep
		expectedRemaining path.PathSteps
	}{
		"nil": {
			steps:             nil,
			expectedNextStep:  nil,
			expectedRemaining: nil,
		},
		"empty": {
			steps:             path.PathSteps{},
			expectedNextStep:  nil,
			expectedRemaining: nil,
		},
		"one": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			expectedNextStep:  path.PathStepAttributeName("test"),
			expectedRemaining: nil,
		},
		"two": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
			},
			expectedNextStep: path.PathStepAttributeName("test"),
			expectedRemaining: path.PathSteps{
				path.PathStepElementKeyInt(0),
			},
		},
		"three": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
				path.PathStepAttributeName("nested-test"),
			},
			expectedNextStep: path.PathStepAttributeName("test"),
			expectedRemaining: path.PathSteps{
				path.PathStepElementKeyInt(0),
				path.PathStepAttributeName("nested-test"),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			gotNextStep, gotRemaining := testCase.steps.NextStep()

			if diff := cmp.Diff(gotNextStep, testCase.expectedNextStep); diff != "" {
				t.Errorf("unexpected next step difference: %s", diff)
			}

			if diff := cmp.Diff(gotRemaining, testCase.expectedRemaining); diff != "" {
				t.Errorf("unexpected remaining difference: %s", diff)
			}
		})
	}
}

func TestPathStepsString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps    path.PathSteps
		expected string
	}{
		"nil": {
			steps:    nil,
			expected: ``,
		},
		"empty": {
			steps:    path.PathSteps{},
			expected: ``,
		},
		"AttributeName": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			expected: `test`,
		},
		"AttributeName-AttributeName": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test2"),
			},
			expected: `test1.test2`,
		},
		"AttributeName-AttributeName-AttributeName": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test2"),
				path.PathStepAttributeName("test3"),
			},
			expected: `test1.test2.test3`,
		},
		"AttributeName-ElementKeyInt": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
			},
			expected: `test[0]`,
		},
		"AttributeName-ElementKeyInt-AttributeName": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepElementKeyInt(0),
				path.PathStepAttributeName("test2"),
			},
			expected: `test1[0].test2`,
		},
		"AttributeName-ElementKeyInt-ElementKeyInt": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
				path.PathStepElementKeyInt(1),
			},
			expected: `test[0][1]`,
		},
		"AttributeName-ElementKeyString": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyString("test-key"),
			},
			expected: `test["test-key"]`,
		},
		"AttributeName-ElementKeyString-AttributeName": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepElementKeyString("test-key"),
				path.PathStepAttributeName("test2"),
			},
			expected: `test1["test-key"].test2`,
		},
		"AttributeName-ElementKeyString-ElementKeyString": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyString("test-key1"),
				path.PathStepElementKeyString("test-key2"),
			},
			expected: `test["test-key1"]["test-key2"]`,
		},
		"AttributeName-ElementKeyValue": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			expected: `test[Value("test-value")]`,
		},
		"AttributeName-ElementKeyValue-AttributeName": {
			steps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyValue{Value: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr_1": types.BoolType,
						"test_attr_2": types.StringType,
					},
					map[string]attr.Value{
						"test_attr_1": types.BoolValue(true),
						"test_attr_2": types.StringValue("test-value"),
					},
				)},
				path.PathStepAttributeName("test_attr_1"),
			},
			expected: `test[Value({"test_attr_1":true,"test_attr_2":"test-value"})].test_attr_1`,
		},
		"ElementKeyInt": {
			steps: path.PathSteps{
				path.PathStepElementKeyInt(0),
			},
			expected: `[0]`,
		},
		"ElementKeyString": {
			steps: path.PathSteps{
				path.PathStepElementKeyString("test"),
			},
			expected: `["test"]`,
		},
		"ElementKeyValue": {
			steps: path.PathSteps{
				path.PathStepElementKeyValue{Value: types.StringValue("test")},
			},
			expected: `[Value("test")]`,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.steps.String()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
