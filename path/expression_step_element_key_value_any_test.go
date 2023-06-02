// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpressionStepElementKeyValueAnyEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyValueAny
		other    path.ExpressionStep
		expected bool
	}{
		"ExpressionStepAttributeNameExact": {
			step:     path.ExpressionStepElementKeyValueAny{},
			other:    path.ExpressionStepAttributeNameExact("test"),
			expected: false,
		},
		"ExpressionStepElementKeyIntExact": {
			step:     path.ExpressionStepElementKeyValueAny{},
			other:    path.ExpressionStepElementKeyIntExact(0),
			expected: false,
		},
		"ExpressionStepElementKeyStringExact": {
			step:     path.ExpressionStepElementKeyValueAny{},
			other:    path.ExpressionStepElementKeyStringExact("test"),
			expected: false,
		},
		"ExpressionStepElementKeyValueAny": {
			step:     path.ExpressionStepElementKeyValueAny{},
			other:    path.ExpressionStepElementKeyValueAny{},
			expected: true,
		},
		"ExpressionStepElementKeyValueExact": {
			step:     path.ExpressionStepElementKeyValueAny{},
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

func TestExpressionStepElementKeyValueAnyMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyValueAny
		pathStep path.PathStep
		expected bool
	}{
		"StepAttributeName": {
			step:     path.ExpressionStepElementKeyValueAny{},
			pathStep: path.PathStepAttributeName("test"),
			expected: false,
		},
		"StepElementKeyInt": {
			step:     path.ExpressionStepElementKeyValueAny{},
			pathStep: path.PathStepElementKeyInt(0),
			expected: false,
		},
		"StepElementKeyString": {
			step:     path.ExpressionStepElementKeyValueAny{},
			pathStep: path.PathStepElementKeyString("test"),
			expected: false,
		},
		"StepElementKeyValue": {
			step:     path.ExpressionStepElementKeyValueAny{},
			pathStep: path.PathStepElementKeyValue{Value: types.StringValue("test")},
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

func TestExpressionStepElementKeyValueAnyString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyValueAny
		expected string
	}{
		"basic": {
			step:     path.ExpressionStepElementKeyValueAny{},
			expected: "[Value(*)]",
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
