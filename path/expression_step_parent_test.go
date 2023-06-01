// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpressionStepParentEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepParent
		other    path.ExpressionStep
		expected bool
	}{
		"ExpressionStepAttributeNameExact": {
			step:     path.ExpressionStepParent{},
			other:    path.ExpressionStepAttributeNameExact("test"),
			expected: false,
		},
		"ExpressionStepElementKeyIntExact": {
			step:     path.ExpressionStepParent{},
			other:    path.ExpressionStepElementKeyIntExact(0),
			expected: false,
		},
		"ExpressionStepElementKeyStringExact": {
			step:     path.ExpressionStepParent{},
			other:    path.ExpressionStepElementKeyStringExact("test"),
			expected: false,
		},
		"ExpressionStepElementKeyValueExact": {
			step:     path.ExpressionStepParent{},
			other:    path.ExpressionStepElementKeyValueExact{Value: types.StringValue("test")},
			expected: false,
		},
		"StepParent": {
			step:     path.ExpressionStepParent{},
			other:    path.ExpressionStepParent{},
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

func TestExpressionStepParentMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepParent
		pathStep path.PathStep
		expected bool
	}{
		"StepAttributeName": {
			step:     path.ExpressionStepParent{},
			pathStep: path.PathStepAttributeName("test"),
			expected: false,
		},
		"StepElementKeyInt": {
			step:     path.ExpressionStepParent{},
			pathStep: path.PathStepElementKeyInt(0),
			expected: false,
		},
		"StepElementKeyString": {
			step:     path.ExpressionStepParent{},
			pathStep: path.PathStepElementKeyString("test"),
			expected: false,
		},
		"StepElementKeyValue": {
			step:     path.ExpressionStepParent{},
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

func TestExpressionStepParentString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.ExpressionStepParent
		expected string
	}{
		"basic": {
			step:     path.ExpressionStepParent{},
			expected: "<",
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
