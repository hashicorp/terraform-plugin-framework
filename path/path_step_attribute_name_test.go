// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPathStepAttributeNameEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.PathStepAttributeName
		other    path.PathStep
		expected bool
	}{
		"PathStepAttributeName-different": {
			step:     path.PathStepAttributeName("test"),
			other:    path.PathStepAttributeName("not-test"),
			expected: false,
		},
		"PathStepAttributeName-equal": {
			step:     path.PathStepAttributeName("test"),
			other:    path.PathStepAttributeName("test"),
			expected: true,
		},
		"PathStepElementKeyInt": {
			step:     path.PathStepAttributeName("test"),
			other:    path.PathStepElementKeyInt(0),
			expected: false,
		},
		"PathStepElementKeyString": {
			step:     path.PathStepAttributeName("test"),
			other:    path.PathStepElementKeyString("test"),
			expected: false,
		},
		"PathStepElementKeyValue": {
			step:     path.PathStepAttributeName("test"),
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

func TestPathStepAttributeNameExpressionStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.PathStepAttributeName
		expected path.ExpressionStep
	}{
		"basic": {
			step:     path.PathStepAttributeName("test"),
			expected: path.ExpressionStepAttributeNameExact("test"),
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

func TestPathStepAttributeNameString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.PathStepAttributeName
		expected string
	}{
		"basic": {
			step:     path.PathStepAttributeName("test"),
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
