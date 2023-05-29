// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpressionStepElementKeyIntAnyEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyIntAny
		other    path.ExpressionStep
		expected bool
	}{
		"ExpressionStepAttributeNameExact": {
			step:     path.ExpressionStepElementKeyIntAny{},
			other:    path.ExpressionStepAttributeNameExact("test"),
			expected: false,
		},
		"ExpressionStepElementKeyIntAny": {
			step:     path.ExpressionStepElementKeyIntAny{},
			other:    path.ExpressionStepElementKeyIntAny{},
			expected: true,
		},
		"ExpressionStepElementKeyIntExact": {
			step:     path.ExpressionStepElementKeyIntAny{},
			other:    path.ExpressionStepElementKeyIntExact(0),
			expected: false,
		},
		"ExpressionStepElementKeyStringExact": {
			step:     path.ExpressionStepElementKeyIntAny{},
			other:    path.ExpressionStepElementKeyStringExact("test"),
			expected: false,
		},
		"ExpressionStepElementKeyValueExact": {
			step:     path.ExpressionStepElementKeyIntAny{},
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

func TestExpressionStepElementKeyIntAnyMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyIntAny
		pathStep path.PathStep
		expected bool
	}{
		"StepAttributeName": {
			step:     path.ExpressionStepElementKeyIntAny{},
			pathStep: path.PathStepAttributeName("test"),
			expected: false,
		},
		"StepElementKeyInt": {
			step:     path.ExpressionStepElementKeyIntAny{},
			pathStep: path.PathStepElementKeyInt(0),
			expected: true,
		},
		"StepElementKeyString": {
			step:     path.ExpressionStepElementKeyIntAny{},
			pathStep: path.PathStepElementKeyString("test"),
			expected: false,
		},
		"StepElementKeyValue": {
			step:     path.ExpressionStepElementKeyIntAny{},
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

func TestExpressionStepElementKeyIntAnyString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepElementKeyIntAny
		expected string
	}{
		"basic": {
			step:     path.ExpressionStepElementKeyIntAny{},
			expected: "[*]",
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
