// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpressionStepElementKeyStringExactEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyStringExact
		other    path.ExpressionStep
		expected bool
	}{
		"ExpressionStepAttributeNameExact": {
			step:     path.ExpressionStepElementKeyStringExact("test"),
			other:    path.ExpressionStepAttributeNameExact("test"),
			expected: false,
		},
		"ExpressionStepElementKeyIntExact": {
			step:     path.ExpressionStepElementKeyStringExact("test"),
			other:    path.ExpressionStepElementKeyIntExact(1),
			expected: false,
		},
		"ExpressionStepElementKeyStringExact-different": {
			step:     path.ExpressionStepElementKeyStringExact("test"),
			other:    path.ExpressionStepElementKeyStringExact("not-test"),
			expected: false,
		},
		"ExpressionStepElementKeyStringExact-equal": {
			step:     path.ExpressionStepElementKeyStringExact("test"),
			other:    path.ExpressionStepElementKeyStringExact("test"),
			expected: true,
		},
		"ExpressionStepElementKeyValueExact": {
			step:     path.ExpressionStepElementKeyStringExact("test"),
			other:    path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test")},
			expected: false,
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

func TestExpressionStepElementKeyStringExactMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyStringExact
		pathStep path.PathStep
		expected bool
	}{
		"StepAttributeName": {
			step:     path.ExpressionStepElementKeyStringExact("test"),
			pathStep: path.PathStepAttributeName("test"),
			expected: false,
		},
		"StepElementKeyInt": {
			step:     path.ExpressionStepElementKeyStringExact("test"),
			pathStep: path.PathStepElementKeyInt(0),
			expected: false,
		},
		"StepElementKeyString-different": {
			step:     path.ExpressionStepElementKeyStringExact("test"),
			pathStep: path.PathStepElementKeyString("not-test"),
			expected: false,
		},
		"StepElementKeyString-equal": {
			step:     path.ExpressionStepElementKeyStringExact("test"),
			pathStep: path.PathStepElementKeyString("test"),
			expected: true,
		},
		"StepElementKeyValue": {
			step:     path.ExpressionStepElementKeyStringExact("test"),
			pathStep: path.PathStepElementKeyValue{Value: types.StringValue("test")},
			expected: false,
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

func TestExpressionStepElementKeyStringExactString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyStringExact
		expected string
	}{
		"basic": {
			step:     path.ExpressionStepElementKeyStringExact("test"),
			expected: `["test"]`,
		},
		"quotes": {
			step:     path.ExpressionStepElementKeyStringExact(`testing is "fun"`),
			expected: `["testing is \"fun\""]`,
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
