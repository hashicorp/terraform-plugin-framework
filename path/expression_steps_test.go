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

func TestExpressionStepsAppend(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps    path.ExpressionSteps
		add      path.ExpressionSteps
		expected path.ExpressionSteps
	}{
		"empty-empty": {
			steps:    path.ExpressionSteps{},
			add:      path.ExpressionSteps{},
			expected: path.ExpressionSteps{},
		},
		"empty-nonempty": {
			steps:    path.ExpressionSteps{},
			add:      path.ExpressionSteps{path.ExpressionStepAttributeNameExact("test")},
			expected: path.ExpressionSteps{path.ExpressionStepAttributeNameExact("test")},
		},
		"nonempty-empty": {
			steps:    path.ExpressionSteps{path.ExpressionStepAttributeNameExact("test")},
			add:      path.ExpressionSteps{},
			expected: path.ExpressionSteps{path.ExpressionStepAttributeNameExact("test")},
		},
		"nonempty-nonempty": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
			},
			add: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("add-test"),
				path.ExpressionStepElementKeyStringExact("add-test-key"),
			},
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
				path.ExpressionStepAttributeNameExact("add-test"),
				path.ExpressionStepElementKeyStringExact("add-test-key"),
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

func TestExpressionStepsCopy(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps    path.ExpressionSteps
		expected path.ExpressionSteps
	}{
		"nil": {
			steps:    nil,
			expected: nil,
		},
		"empty": {
			steps:    path.ExpressionSteps{},
			expected: path.ExpressionSteps{},
		},
		"shallow": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
		},
		"deep": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepElementKeyIntExact(0),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepElementKeyIntExact(0),
				path.ExpressionStepAttributeNameExact("test2"),
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
			got.Append(path.ExpressionStepAttributeNameExact("modify-test"))

			if diff := cmp.Diff(got, testCase.expected); diff == "" {
				t.Error("unexpected modification")
			}
		})
	}
}

func TestExpressionStepsEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps    path.ExpressionSteps
		other    path.ExpressionSteps
		expected bool
	}{
		"nil-nil": {
			steps:    nil,
			other:    nil,
			expected: true,
		},
		"empty-empty": {
			steps:    path.ExpressionSteps{},
			other:    path.ExpressionSteps{},
			expected: true,
		},
		"different-length": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			other: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
			},
			expected: false,
		},
		"StepAttributeName-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			other: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("not-test"),
			},
			expected: false,
		},
		"StepAttributeName-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			other: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			expected: true,
		},
		"StepAttributeName-StepElementKeyInt-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
			},
			other: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(1),
			},
			expected: false,
		},
		"StepAttributeName-StepElementKeyInt-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
			},
			other: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
			},
			expected: true,
		},
		"StepAttributeName-StepElementKeyString-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyStringExact("test-key"),
			},
			other: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyStringExact("not-test-key"),
			},
			expected: false,
		},
		"StepAttributeName-StepElementKeyString-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyStringExact("test-key"),
			},
			other: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyStringExact("test-key"),
			},
			expected: true,
		},
		"StepAttributeName-StepElementKeyValue-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test-value")},
			},
			other: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("not-test-value")},
			},
			expected: false,
		},
		"StepAttributeName-StepElementKeyValue-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test-value")},
			},
			other: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test-value")},
			},
			expected: true,
		},
		"StepElementKeyInt-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepElementKeyIntExact(0),
			},
			other: path.ExpressionSteps{
				path.ExpressionStepElementKeyIntExact(1),
			},
			expected: false,
		},
		"StepElementKeyInt-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepElementKeyIntExact(0),
			},
			other: path.ExpressionSteps{
				path.ExpressionStepElementKeyIntExact(0),
			},
			expected: true,
		},
		"StepElementKeyString-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepElementKeyStringExact("test"),
			},
			other: path.ExpressionSteps{
				path.ExpressionStepElementKeyStringExact("not-test"),
			},
			expected: false,
		},
		"StepElementKeyString-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepElementKeyStringExact("test"),
			},
			other: path.ExpressionSteps{
				path.ExpressionStepElementKeyStringExact("test"),
			},
			expected: true,
		},
		"StepElementKeyValue-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test-value")},
			},
			other: path.ExpressionSteps{
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("not-test-value")},
			},
			expected: false,
		},
		"StepElementKeyValue-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test-value")},
			},
			other: path.ExpressionSteps{
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test-value")},
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

func TestExpressionStepsLastStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps             path.ExpressionSteps
		expectedLastStep  path.ExpressionStep
		expectedRemaining path.ExpressionSteps
	}{
		"nil": {
			steps:             nil,
			expectedLastStep:  nil,
			expectedRemaining: nil,
		},
		"empty": {
			steps:             path.ExpressionSteps{},
			expectedLastStep:  nil,
			expectedRemaining: nil,
		},
		"one": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			expectedLastStep:  path.ExpressionStepAttributeNameExact("test"),
			expectedRemaining: nil,
		},
		"two": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
			},
			expectedLastStep: path.ExpressionStepElementKeyIntExact(0),
			expectedRemaining: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
		},
		"three": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
				path.ExpressionStepAttributeNameExact("nested-test"),
			},
			expectedLastStep: path.ExpressionStepAttributeNameExact("nested-test"),
			expectedRemaining: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
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

func TestExpressionStepsMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps     path.ExpressionSteps
		pathSteps path.PathSteps
		expected  bool
	}{
		"empty-empty": {
			steps:     path.ExpressionSteps{},
			pathSteps: path.PathSteps{},
			expected:  false,
		},
		"empty-nonempty": {
			steps: path.ExpressionSteps{},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			expected: false,
		},
		"nonempty-empty": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			pathSteps: path.PathSteps{},
			expected:  false,
		},
		"AttributeNameExact-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("not-test"),
			},
			expected: false,
		},
		"AttributeNameExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			expected: true,
		},
		"AttributeNameExact-AttributeNameExact-different-firststep": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test2"),
				path.PathStepAttributeName("test2"),
			},
			expected: false,
		},
		"AttributeNameExact-AttributeNameExact-different-laststep": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test3"),
			},
			expected: false,
		},
		"AttributeNameExact-AttributeNameExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test2"),
			},
			expected: true,
		},
		"AttributeNameExact-AttributeNameExact-Parent-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepParent{},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test2"),
			},
			expected: false,
		},
		"AttributeNameExact-AttributeNameExact-Parent-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepParent{},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
			},
			expected: true,
		},
		"AttributeNameExact-AttributeNameExact-Parent-AttributeNameExact-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test3"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test2"),
			},
			expected: false,
		},
		"AttributeNameExact-AttributeNameExact-Parent-AttributeNameExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test3"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test3"),
			},
			expected: true,
		},
		"AttributeNameExact-ElementKeyIntAny": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntAny{},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
			},
			expected: true,
		},
		"AttributeNameExact-ElementKeyIntExact-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(1),
			},
			expected: false,
		},
		"AttributeNameExact-ElementKeyIntExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
			},
			expected: true,
		},
		"AttributeNameExact-ElementKeyStringAny": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyStringAny{},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyString("test-key"),
			},
			expected: true,
		},
		"AttributeNameExact-ElementKeyStringExact-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyStringExact("test-key"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyString("not-test-key"),
			},
			expected: false,
		},
		"AttributeNameExact-ElementKeyStringExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyStringExact("test-key"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyString("test-key"),
			},
			expected: true,
		},
		"AttributeNameExact-ElementKeyValueAny": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyValueAny{},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			expected: true,
		},
		"AttributeNameExact-ElementKeyValueExact-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test-value")},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyValue{Value: types.StringValue("not-test-value")},
			},
			expected: false,
		},
		"AttributeNameExact-ElementKeyValueExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test-value")},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			expected: true,
		},
		"AttributeNameExact-Parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepParent{},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
			},
			expected: false,
		},
		"AttributeNameExact-Parent-AttributeNameExact-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
			},
			expected: false,
		},
		"AttributeNameExact-Parent-AttributeNameExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test2"),
			},
			expected: true,
		},
		"Parent-AttributeNameExact": {
			steps: path.ExpressionSteps{
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			expected: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.steps.Matches(testCase.pathSteps)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionStepsMatchesParent(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps     path.ExpressionSteps
		pathSteps path.PathSteps
		expected  bool
	}{
		"empty-empty": {
			steps:     path.ExpressionSteps{},
			pathSteps: path.PathSteps{},
			expected:  false,
		},
		"empty-nonempty": {
			steps: path.ExpressionSteps{},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			expected: false,
		},
		"nonempty-empty": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			pathSteps: path.PathSteps{},
			expected:  true,
		},
		"AttributeNameExact-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("not-test"),
			},
			expected: false,
		},
		"AttributeNameExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			expected: false,
		},
		"AttributeNameExact-parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
			},
			expected: true,
		},
		"AttributeNameExact-AttributeNameExact-different-firststep": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test2"),
				path.PathStepAttributeName("test2"),
			},
			expected: false,
		},
		"AttributeNameExact-AttributeNameExact-different-laststep": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test3"),
			},
			expected: false,
		},
		"AttributeNameExact-AttributeNameExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test2"),
			},
			expected: false,
		},
		"AttributeNameExact-AttributeNameExact-parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepAttributeNameExact("test3"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test2"),
			},
			expected: true,
		},
		"AttributeNameExact-AttributeNameExact-Parent-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepParent{},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test2"),
			},
			expected: false,
		},
		"AttributeNameExact-AttributeNameExact-Parent-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepParent{},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
			},
			expected: false,
		},
		"AttributeNameExact-AttributeNameExact-Parent-parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test3"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
			},
			expected: true,
		},
		"AttributeNameExact-AttributeNameExact-Parent-AttributeNameExact-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test3"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test2"),
			},
			expected: false,
		},
		"AttributeNameExact-AttributeNameExact-Parent-AttributeNameExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test3"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test3"),
			},
			expected: false,
		},
		"AttributeNameExact-AttributeNameExact-Parent-AttributeNameExact-parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test3"),
				path.ExpressionStepAttributeNameExact("test4"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepAttributeName("test3"),
			},
			expected: true,
		},
		"AttributeNameExact-ElementKeyIntAny": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntAny{},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
			},
			expected: false,
		},
		"AttributeNameExact-ElementKeyIntAny-parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepElementKeyIntAny{},
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepElementKeyInt(0),
			},
			expected: true,
		},
		"AttributeNameExact-ElementKeyIntExact-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(1),
			},
			expected: false,
		},
		"AttributeNameExact-ElementKeyIntExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(0),
			},
			expected: false,
		},
		"AttributeNameExact-ElementKeyIntExact-parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepElementKeyIntExact(0),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepElementKeyInt(0),
			},
			expected: true,
		},
		"AttributeNameExact-ElementKeyStringAny": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyStringAny{},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyString("test-key"),
			},
			expected: false,
		},
		"AttributeNameExact-ElementKeyStringAny-parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepElementKeyStringAny{},
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepElementKeyString("test-key"),
			},
			expected: true,
		},
		"AttributeNameExact-ElementKeyStringExact-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyStringExact("test-key"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyString("not-test-key"),
			},
			expected: false,
		},
		"AttributeNameExact-ElementKeyStringExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyStringExact("test-key"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyString("test-key"),
			},
			expected: false,
		},
		"AttributeNameExact-ElementKeyStringExact-parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepElementKeyStringExact("test-key"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepElementKeyString("test-key"),
			},
			expected: true,
		},
		"AttributeNameExact-ElementKeyValueAny": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyValueAny{},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			expected: false,
		},
		"AttributeNameExact-ElementKeyValueAny-parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepElementKeyValueAny{},
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			expected: true,
		},
		"AttributeNameExact-ElementKeyValueExact-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test-value")},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyValue{Value: types.StringValue("not-test-value")},
			},
			expected: false,
		},
		"AttributeNameExact-ElementKeyValueExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test-value")},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			expected: false,
		},
		"AttributeNameExact-ElementKeyValueExact-parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test-value")},
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
				path.PathStepElementKeyValue{Value: types.StringValue("test-value")},
			},
			expected: true,
		},
		"AttributeNameExact-Parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepParent{},
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
			},
			expected: false,
		},
		"AttributeNameExact-Parent-AttributeNameExact-different": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test1"),
			},
			expected: false,
		},
		"AttributeNameExact-Parent-AttributeNameExact-equal": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test2"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test2"),
			},
			expected: false,
		},
		"AttributeNameExact-Parent-AttributeNameExact-parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepAttributeNameExact("test3"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test2"),
			},
			expected: true,
		},
		"Parent-AttributeNameExact": {
			steps: path.ExpressionSteps{
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test"),
			},
			pathSteps: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
			expected: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.steps.MatchesParent(testCase.pathSteps)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionStepsNextStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps             path.ExpressionSteps
		expectedNextStep  path.ExpressionStep
		expectedRemaining path.ExpressionSteps
	}{
		"nil": {
			steps:             nil,
			expectedNextStep:  nil,
			expectedRemaining: nil,
		},
		"empty": {
			steps:             path.ExpressionSteps{},
			expectedNextStep:  nil,
			expectedRemaining: nil,
		},
		"one": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			expectedNextStep:  path.ExpressionStepAttributeNameExact("test"),
			expectedRemaining: nil,
		},
		"two": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
			},
			expectedNextStep: path.ExpressionStepAttributeNameExact("test"),
			expectedRemaining: path.ExpressionSteps{
				path.ExpressionStepElementKeyIntExact(0),
			},
		},
		"three": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
				path.ExpressionStepAttributeNameExact("nested-test"),
			},
			expectedNextStep: path.ExpressionStepAttributeNameExact("test"),
			expectedRemaining: path.ExpressionSteps{
				path.ExpressionStepElementKeyIntExact(0),
				path.ExpressionStepAttributeNameExact("nested-test"),
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

func TestExpressionStepsResolve(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps    path.ExpressionSteps
		expected path.ExpressionSteps
	}{
		"nil": {
			steps:    nil,
			expected: nil,
		},
		"empty": {
			steps:    path.ExpressionSteps{},
			expected: nil,
		},
		"AttributeNameExact": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
		},
		"AttributeNameExact-AttributeNameExact": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
		},
		"AttributeNameExact-AttributeNameExact-Parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepParent{},
			},
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
			},
		},
		"AttributeNameExact-AttributeNameExact-Parent-AttributeNameExact": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test3"),
			},
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test3"),
			},
		},
		"AttributeNameExact-Parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepParent{},
			},
			expected: path.ExpressionSteps{},
		},
		"AttributeNameExact-Parent-AttributeNameExact": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test2"),
			},
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test2"),
			},
		},
		"Parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepParent{},
			},
			expected: nil,
		},
		"Parent-AttributeNameExact": {
			steps: path.ExpressionSteps{
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test"),
			},
			expected: nil,
		},
		"Parent-Parent": {
			steps: path.ExpressionSteps{
				path.ExpressionStepParent{},
				path.ExpressionStepParent{},
			},
			expected: nil,
		},
		"Parent-Parent-AttributeNameExact": {
			steps: path.ExpressionSteps{
				path.ExpressionStepParent{},
				path.ExpressionStepParent{},
				path.ExpressionStepAttributeNameExact("test"),
			},
			expected: nil,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.steps.Resolve()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionStepsString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		steps    path.ExpressionSteps
		expected string
	}{
		"nil": {
			steps:    nil,
			expected: ``,
		},
		"empty": {
			steps:    path.ExpressionSteps{},
			expected: ``,
		},
		"AttributeName": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
			expected: `test`,
		},
		"AttributeName-AttributeName": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			expected: `test1.test2`,
		},
		"AttributeName-AttributeName-AttributeName": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepAttributeNameExact("test2"),
				path.ExpressionStepAttributeNameExact("test3"),
			},
			expected: `test1.test2.test3`,
		},
		"AttributeName-ElementKeyInt": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
			},
			expected: `test[0]`,
		},
		"AttributeName-ElementKeyInt-AttributeName": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepElementKeyIntExact(0),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			expected: `test1[0].test2`,
		},
		"AttributeName-ElementKeyInt-ElementKeyInt": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(0),
				path.ExpressionStepElementKeyIntExact(1),
			},
			expected: `test[0][1]`,
		},
		"AttributeName-ElementKeyString": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyStringExact("test-key"),
			},
			expected: `test["test-key"]`,
		},
		"AttributeName-ElementKeyString-AttributeName": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test1"),
				path.ExpressionStepElementKeyStringExact("test-key"),
				path.ExpressionStepAttributeNameExact("test2"),
			},
			expected: `test1["test-key"].test2`,
		},
		"AttributeName-ElementKeyString-ElementKeyString": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyStringExact("test-key1"),
				path.ExpressionStepElementKeyStringExact("test-key2"),
			},
			expected: `test["test-key1"]["test-key2"]`,
		},
		"AttributeName-ElementKeyValue": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test-value")},
			},
			expected: `test[Value("test-value")]`,
		},
		"AttributeName-ElementKeyValue-AttributeName": {
			steps: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyValueExact{Value: types.ObjectValueMust(
					map[string]attr.Type{
						"test_attr_1": types.BoolType,
						"test_attr_2": types.StringType,
					},
					map[string]attr.Value{
						"test_attr_1": types.BoolValue(true),
						"test_attr_2": types.StringValue("test-value"),
					},
				)},
				path.ExpressionStepAttributeNameExact("test_attr_1"),
			},
			expected: `test[Value({"test_attr_1":true,"test_attr_2":"test-value"})].test_attr_1`,
		},
		"ElementKeyInt": {
			steps: path.ExpressionSteps{
				path.ExpressionStepElementKeyIntExact(0),
			},
			expected: `[0]`,
		},
		"ElementKeyString": {
			steps: path.ExpressionSteps{
				path.ExpressionStepElementKeyStringExact("test"),
			},
			expected: `["test"]`,
		},
		"ElementKeyValue": {
			steps: path.ExpressionSteps{
				path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test")},
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
