// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPathStepElementKeyIntEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.PathStepElementKeyInt
		other    path.PathStep
		expected bool
	}{
		"PathStepAttributeName": {
			step:     path.PathStepElementKeyInt(0),
			other:    path.PathStepAttributeName("test"),
			expected: false,
		},
		"PathStepElementKeyInt-different": {
			step:     path.PathStepElementKeyInt(0),
			other:    path.PathStepElementKeyInt(1),
			expected: false,
		},
		"PathStepElementKeyInt-equal": {
			step:     path.PathStepElementKeyInt(0),
			other:    path.PathStepElementKeyInt(0),
			expected: true,
		},
		"PathStepElementKeyString": {
			step:     path.PathStepElementKeyInt(0),
			other:    path.PathStepElementKeyString("test"),
			expected: false,
		},
		"PathStepElementKeyValue": {
			step:     path.PathStepElementKeyInt(0),
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

func TestPathStepElementKeyIntExpressionStep(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.PathStepElementKeyInt
		expected path.ExpressionStep
	}{
		"basic": {
			step:     path.PathStepElementKeyInt(1),
			expected: path.ExpressionStepElementKeyIntExact(1),
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

func TestPathStepElementKeyIntString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		step     path.PathStepElementKeyInt
		expected string
	}{
		"basic": {
			step:     path.PathStepElementKeyInt(0),
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
