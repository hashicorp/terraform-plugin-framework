// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpressionStepAttributeNameExactEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepAttributeNameExact
		other    path.ExpressionStep
		expected bool
	}{
		"ExpressionStepAttributeNameExact-different": {
			step:     path.ExpressionStepAttributeNameExact("test"),
			other:    path.ExpressionStepAttributeNameExact("not-test"),
			expected: false,
		},
		"ExpressionStepAttributeNameExact-equal": {
			step:     path.ExpressionStepAttributeNameExact("test"),
			other:    path.ExpressionStepAttributeNameExact("test"),
			expected: true,
		},
		"ExpressionStepElementKeyIntExact": {
			step:     path.ExpressionStepAttributeNameExact("test"),
			other:    path.ExpressionStepElementKeyIntExact(0),
			expected: false,
		},
		"ExpressionStepElementKeyStringExact": {
			step:     path.ExpressionStepAttributeNameExact("test"),
			other:    path.ExpressionStepElementKeyStringExact("test"),
			expected: false,
		},
		"ExpressionStepElementKeyValueExact": {
			step:     path.ExpressionStepAttributeNameExact("test"),
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

func TestExpressionStepAttributeNameExactMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepAttributeNameExact
		pathStep path.PathStep
		expected bool
	}{
		"StepAttributeName-different": {
			step:     path.ExpressionStepAttributeNameExact("test"),
			pathStep: path.PathStepAttributeName("not-test"),
			expected: false,
		},
		"StepAttributeName-equal": {
			step:     path.ExpressionStepAttributeNameExact("test"),
			pathStep: path.PathStepAttributeName("test"),
			expected: true,
		},
		"StepElementKeyInt": {
			step:     path.ExpressionStepAttributeNameExact("test"),
			pathStep: path.PathStepElementKeyInt(0),
			expected: false,
		},
		"StepElementKeyString": {
			step:     path.ExpressionStepAttributeNameExact("test"),
			pathStep: path.PathStepElementKeyString("test"),
			expected: false,
		},
		"StepElementKeyValue": {
			step:     path.ExpressionStepAttributeNameExact("test"),
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

func TestExpressionStepAttributeNameExactString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepAttributeNameExact
		expected string
	}{
		"basic": {
			step:     path.ExpressionStepAttributeNameExact("test"),
			expected: "test",
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
