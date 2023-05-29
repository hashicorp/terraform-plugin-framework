// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpressionsAppend(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expressions path.Expressions
		add         path.Expressions
		expected    path.Expressions
	}{
		"nil-nil": {
			expressions: nil,
			add:         nil,
			expected:    nil,
		},
		"nil-nonempty": {
			expressions: nil,
			add:         path.Expressions{path.MatchRoot("test")},
			expected:    path.Expressions{path.MatchRoot("test")},
		},
		"nonempty-nil": {
			expressions: path.Expressions{path.MatchRoot("test")},
			add:         nil,
			expected:    path.Expressions{path.MatchRoot("test")},
		},
		"empty-empty": {
			expressions: path.Expressions{},
			add:         path.Expressions{},
			expected:    path.Expressions{},
		},
		"empty-nonempty": {
			expressions: path.Expressions{},
			add:         path.Expressions{path.MatchRoot("test")},
			expected:    path.Expressions{path.MatchRoot("test")},
		},
		"nonempty-empty": {
			expressions: path.Expressions{path.MatchRoot("test")},
			add:         path.Expressions{},
			expected:    path.Expressions{path.MatchRoot("test")},
		},
		"nonempty-nonempty": {
			expressions: path.Expressions{
				path.MatchRoot("test1"),
				path.MatchRoot("test2"),
			},
			add: path.Expressions{
				path.MatchRoot("test3"),
				path.MatchRoot("test4"),
			},
			expected: path.Expressions{
				path.MatchRoot("test1"),
				path.MatchRoot("test2"),
				path.MatchRoot("test3"),
				path.MatchRoot("test4"),
			},
		},
		"deduplication": {
			expressions: path.Expressions{
				path.MatchRoot("test1"),
				path.MatchRoot("test2"),
			},
			add: path.Expressions{
				path.MatchRoot("test1"),
				path.MatchRoot("test3"),
			},
			expected: path.Expressions{
				path.MatchRoot("test1"),
				path.MatchRoot("test2"),
				path.MatchRoot("test3"),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expressions.Append(testCase.add...)

			if diff := cmp.Diff(testCase.expressions, testCase.expected); diff != "" {
				t.Errorf("unexpected original difference: %s", diff)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected result difference: %s", diff)
			}
		})
	}
}

func TestExpressionsContains(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expressions path.Expressions
		contains    path.Expression
		expected    bool
	}{
		"paths-nil": {
			expressions: nil,
			contains:    path.MatchRoot("test"),
			expected:    false,
		},
		"paths-empty": {
			expressions: path.Expressions{},
			contains:    path.MatchRoot("test"),
			expected:    false,
		},
		"contains-empty": {
			expressions: path.Expressions{
				path.MatchRoot("test"),
			},
			contains: path.MatchRelative(),
			expected: false,
		},
		"contains-middle": {
			expressions: path.Expressions{
				path.MatchRoot("test1").AtName("test1_attr"),
				path.MatchRoot("test2").AtName("test2_attr"),
				path.MatchRoot("test3").AtName("test3_attr"),
			},
			contains: path.MatchRoot("test2").AtName("test2_attr"),
			expected: true,
		},
		"contains-end": {
			expressions: path.Expressions{
				path.MatchRoot("test1").AtName("test1_attr"),
				path.MatchRoot("test2").AtName("test2_attr"),
				path.MatchRoot("test3").AtName("test3_attr"),
			},
			contains: path.MatchRoot("test3").AtName("test3_attr"),
			expected: true,
		},
		"relative-paths-different": {
			expressions: path.Expressions{
				path.MatchRoot("test_parent").AtName("test_child"),
			},
			contains: path.MatchRoot("test_parent").AtName("test_child").AtParent().AtName("test_child"),
			expected: false, // Contains intentionally does not Resolve()
		},
		"AttributeName-different": {
			expressions: path.Expressions{
				path.MatchRoot("test"),
			},
			contains: path.MatchRoot("not-test"),
			expected: false,
		},
		"AttributeName-equal": {
			expressions: path.Expressions{
				path.MatchRoot("test"),
			},
			contains: path.MatchRoot("test"),
			expected: true,
		},
		"ElementKeyInt-different": {
			expressions: path.Expressions{
				path.MatchRelative().AtListIndex(0),
			},
			contains: path.MatchRelative().AtListIndex(1),
			expected: false,
		},
		"ElementKeyInt-equal": {
			expressions: path.Expressions{
				path.MatchRelative().AtListIndex(0),
			},
			contains: path.MatchRelative().AtListIndex(0),
			expected: true,
		},
		"ElementKeyString-different": {
			expressions: path.Expressions{
				path.MatchRelative().AtMapKey("test"),
			},
			contains: path.MatchRelative().AtMapKey("not-test"),
			expected: false,
		},
		"ElementKeyString-equal": {
			expressions: path.Expressions{
				path.MatchRelative().AtMapKey("test"),
			},
			contains: path.MatchRelative().AtMapKey("test"),
			expected: true,
		},
		"ElementKeyValue-different": {
			expressions: path.Expressions{
				path.MatchRelative().AtSetValue(types.StringValue("test")),
			},
			contains: path.MatchRelative().AtSetValue(types.StringValue("not-test")),
			expected: false,
		},
		"ElementKeyValue-equal": {
			expressions: path.Expressions{
				path.MatchRelative().AtSetValue(types.StringValue("test")),
			},
			contains: path.MatchRelative().AtSetValue(types.StringValue("test")),
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expressions.Contains(testCase.contains)

			if got != testCase.expected {
				t.Errorf("expected %t, got %t", testCase.expected, got)
			}
		})
	}
}

func TestExpressionsMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expressions path.Expressions
		match       path.Path
		expected    bool
	}{
		"expressions-nil": {
			expressions: nil,
			match:       path.Root("test"),
			expected:    false,
		},
		"expressions-empty": {
			expressions: path.Expressions{},
			match:       path.Root("test"),
			expected:    false,
		},
		"match-empty": {
			expressions: path.Expressions{
				path.MatchRoot("test"),
			},
			match:    path.Empty(),
			expected: false,
		},
		"match-middle": {
			expressions: path.Expressions{
				path.MatchRoot("test1").AtName("test1_attr"),
				path.MatchRoot("test2").AtName("test2_attr"),
				path.MatchRoot("test3").AtName("test3_attr"),
			},
			match:    path.Root("test2").AtName("test2_attr"),
			expected: true,
		},
		"match-end": {
			expressions: path.Expressions{
				path.MatchRoot("test1").AtName("test1_attr"),
				path.MatchRoot("test2").AtName("test2_attr"),
				path.MatchRoot("test3").AtName("test3_attr"),
			},
			match:    path.Root("test3").AtName("test3_attr"),
			expected: true,
		},
		"AttributeName-different": {
			expressions: path.Expressions{
				path.MatchRoot("test"),
			},
			match:    path.Root("not-test"),
			expected: false,
		},
		"AttributeName-equal": {
			expressions: path.Expressions{
				path.MatchRoot("test"),
			},
			match:    path.Root("test"),
			expected: true,
		},
		"ElementKeyInt-any": {
			expressions: path.Expressions{
				path.MatchRoot("test").AtAnyListIndex(),
			},
			match:    path.Root("test").AtListIndex(1),
			expected: true,
		},
		"ElementKeyInt-different": {
			expressions: path.Expressions{
				path.MatchRoot("test").AtListIndex(0),
			},
			match:    path.Root("test").AtListIndex(1),
			expected: false,
		},
		"ElementKeyInt-equal": {
			expressions: path.Expressions{
				path.MatchRoot("test").AtListIndex(0),
			},
			match:    path.Root("test").AtListIndex(0),
			expected: true,
		},
		"ElementKeyString-any": {
			expressions: path.Expressions{
				path.MatchRoot("test").AtAnyMapKey(),
			},
			match:    path.Root("test").AtMapKey("test"),
			expected: true,
		},
		"ElementKeyString-different": {
			expressions: path.Expressions{
				path.MatchRoot("test").AtMapKey("test"),
			},
			match:    path.Root("test").AtMapKey("not-test"),
			expected: false,
		},
		"ElementKeyString-equal": {
			expressions: path.Expressions{
				path.MatchRoot("test").AtMapKey("test"),
			},
			match:    path.Root("test").AtMapKey("test"),
			expected: true,
		},
		"ElementKeyValue-any": {
			expressions: path.Expressions{
				path.MatchRoot("test").AtAnySetValue(),
			},
			match:    path.Root("test").AtSetValue(types.StringValue("test")),
			expected: true,
		},
		"ElementKeyValue-different": {
			expressions: path.Expressions{
				path.MatchRoot("test").AtSetValue(types.StringValue("test")),
			},
			match:    path.Root("test").AtSetValue(types.StringValue("not-test")),
			expected: false,
		},
		"ElementKeyValue-equal": {
			expressions: path.Expressions{
				path.MatchRoot("test").AtSetValue(types.StringValue("test")),
			},
			match:    path.Root("test").AtSetValue(types.StringValue("test")),
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expressions.Matches(testCase.match)

			if got != testCase.expected {
				t.Errorf("expected %t, got %t", testCase.expected, got)
			}
		})
	}
}

func TestExpressionsString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expressions path.Expressions
		expected    string
	}{
		"nil": {
			expressions: nil,
			expected:    "[]",
		},
		"empty": {
			expressions: path.Expressions{},
			expected:    "[]",
		},
		"one": {
			expressions: path.Expressions{
				path.MatchRoot("test"),
			},
			expected: "[test]",
		},
		"one-empty": {
			expressions: path.Expressions{
				path.Expression{},
			},
			expected: "[]",
		},
		"two": {
			expressions: path.Expressions{
				path.MatchRoot("test1"),
				path.MatchRoot("test2"),
			},
			expected: "[test1,test2]",
		},
		"two-empty": {
			expressions: path.Expressions{
				path.MatchRoot("test"),
				path.Expression{},
			},
			expected: "[test]",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expressions.String()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
