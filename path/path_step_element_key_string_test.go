// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPathStepElementKeyStringEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.PathStepElementKeyString
		other    path.PathStep
		expected bool
	}{
		"PathStepAttributeName": {
			step:     path.PathStepElementKeyString("test"),
			other:    path.PathStepAttributeName("test"),
			expected: false,
		},
		"PathStepElementKeyInt": {
			step:     path.PathStepElementKeyString("test"),
			other:    path.PathStepElementKeyInt(0),
			expected: false,
		},
		"PathStepElementKeyString-different": {
			step:     path.PathStepElementKeyString("test"),
			other:    path.PathStepElementKeyString("not-test"),
			expected: false,
		},
		"PathStepElementKeyString-equal": {
			step:     path.PathStepElementKeyString("test"),
			other:    path.PathStepElementKeyString("test"),
			expected: true,
		},
		"PathStepElementKeyValue": {
			step:     path.PathStepElementKeyString("test"),
			other:    path.PathStepElementKeyValue{Value: types.StringValue("test")},
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

func TestPathStepElementKeyStringExpressionStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.PathStepElementKeyString
		expected path.ExpressionStep
	}{
		"basic": {
			step:     path.PathStepElementKeyString("test"),
			expected: path.ExpressionStepElementKeyStringExact("test"),
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

func TestPathStepElementKeyStringString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.PathStepElementKeyString
		expected string
	}{
		"basic": {
			step:     path.PathStepElementKeyString("test"),
			expected: `["test"]`,
		},
		"quotes": {
			step:     path.PathStepElementKeyString(`testing is "fun"`),
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
