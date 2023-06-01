// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpressionStepElementKeyIntExactEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyIntExact
		other    path.ExpressionStep
		expected bool
	}{
		"ExpressionStepAttributeNameExact": {
			step:     path.ExpressionStepElementKeyIntExact(0),
			other:    path.ExpressionStepAttributeNameExact("test"),
			expected: false,
		},
		"ExpressionStepElementKeyIntAny": {
			step:     path.ExpressionStepElementKeyIntExact(0),
			other:    path.ExpressionStepElementKeyIntAny{},
			expected: false,
		},
		"ExpressionStepElementKeyIntExact-different": {
			step:     path.ExpressionStepElementKeyIntExact(0),
			other:    path.ExpressionStepElementKeyIntExact(1),
			expected: false,
		},
		"ExpressionStepElementKeyIntExact-equal": {
			step:     path.ExpressionStepElementKeyIntExact(0),
			other:    path.ExpressionStepElementKeyIntExact(0),
			expected: true,
		},
		"ExpressionStepElementKeyStringExact": {
			step:     path.ExpressionStepElementKeyIntExact(0),
			other:    path.ExpressionStepElementKeyStringExact("test"),
			expected: false,
		},
		"ExpressionStepElementKeyValueExact": {
			step:     path.ExpressionStepElementKeyIntExact(0),
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

func TestExpressionStepElementKeyIntExactMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyIntExact
		pathStep path.PathStep
		expected bool
	}{
		"StepAttributeName": {
			step:     path.ExpressionStepElementKeyIntExact(0),
			pathStep: path.PathStepAttributeName("test"),
			expected: false,
		},
		"StepElementKeyInt-different": {
			step:     path.ExpressionStepElementKeyIntExact(0),
			pathStep: path.PathStepElementKeyInt(1),
			expected: false,
		},
		"StepElementKeyInt-equal": {
			step:     path.ExpressionStepElementKeyIntExact(0),
			pathStep: path.PathStepElementKeyInt(0),
			expected: true,
		},
		"StepElementKeyString": {
			step:     path.ExpressionStepElementKeyIntExact(0),
			pathStep: path.PathStepElementKeyString("test"),
			expected: false,
		},
		"StepElementKeyValue": {
			step:     path.ExpressionStepElementKeyIntExact(0),
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

func TestExpressionStepElementKeyIntExactString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyIntExact
		expected string
	}{
		"basic": {
			step:     path.ExpressionStepElementKeyIntExact(0),
			expected: "[0]",
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
