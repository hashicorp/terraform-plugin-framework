// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestExpressionAtAnyListIndex(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		expected   path.Expression
	}{
		"shallow": {
			expression: path.MatchRoot("test"),
			expected:   path.MatchRoot("test").AtAnyListIndex(),
		},
		"deep": {
			expression: path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			expected:   path.MatchRoot("test1").AtListIndex(0).AtName("test2").AtAnyListIndex(),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.AtAnyListIndex()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionAtAnyMapKey(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		expected   path.Expression
	}{
		"shallow": {
			expression: path.MatchRoot("test"),
			expected:   path.MatchRoot("test").AtAnyMapKey(),
		},
		"deep": {
			expression: path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			expected:   path.MatchRoot("test1").AtListIndex(0).AtName("test2").AtAnyMapKey(),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.AtAnyMapKey()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionAtAnySetValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		expected   path.Expression
	}{
		"shallow": {
			expression: path.MatchRoot("test"),
			expected:   path.MatchRoot("test").AtAnySetValue(),
		},
		"deep": {
			expression: path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			expected:   path.MatchRoot("test1").AtListIndex(0).AtName("test2").AtAnySetValue(),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.AtAnySetValue()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionAtListIndex(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		index      int
		expected   path.Expression
	}{
		"shallow": {
			expression: path.MatchRoot("test"),
			index:      1,
			expected:   path.MatchRoot("test").AtListIndex(1),
		},
		"deep": {
			expression: path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			index:      1,
			expected:   path.MatchRoot("test1").AtListIndex(0).AtName("test2").AtListIndex(1),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.AtListIndex(testCase.index)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionAtMapKey(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		key        string
		expected   path.Expression
	}{
		"shallow": {
			expression: path.MatchRoot("test"),
			key:        "test-key",
			expected:   path.MatchRoot("test").AtMapKey("test-key"),
		},
		"deep": {
			expression: path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			key:        "test-key",
			expected:   path.MatchRoot("test1").AtListIndex(0).AtName("test2").AtMapKey("test-key"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.AtMapKey(testCase.key)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionAtName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		name       string
		expected   path.Expression
	}{
		"shallow": {
			expression: path.MatchRoot("test1"),
			name:       "test2",
			expected:   path.MatchRoot("test1").AtName("test2"),
		},
		"deep": {
			expression: path.MatchRoot("test1").AtListIndex(0),
			name:       "test2",
			expected:   path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.AtName(testCase.name)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionAtParent(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		expected   path.Expression
	}{
		"shallow": {
			expression: path.MatchRoot("test"),
			expected:   path.MatchRoot("test").AtParent(),
		},
		"deep": {
			expression: path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			expected:   path.MatchRoot("test1").AtListIndex(0).AtName("test2").AtParent(),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.AtParent()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionAtSetValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		value      attr.Value
		expected   path.Expression
	}{
		"shallow": {
			expression: path.MatchRoot("test"),
			value:      types.StringValue("test"),
			expected:   path.MatchRoot("test").AtSetValue(types.StringValue("test")),
		},
		"deep": {
			expression: path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			value:      types.StringValue("test"),
			expected:   path.MatchRoot("test1").AtListIndex(0).AtName("test2").AtSetValue(types.StringValue("test")),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.AtSetValue(testCase.value)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionCopy(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		expected   path.Expression
	}{
		"shallow": {
			expression: path.MatchRoot("test"),
			expected:   path.MatchRoot("test"),
		},
		"deep": {
			expression: path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			expected:   path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.Copy()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		other      path.Expression
		expected   bool
	}{
		"different-length": {
			expression: path.MatchRoot("test1").AtName("test2"),
			other:      path.MatchRoot("test1"),
			expected:   false,
		},
		"different-step-shallow": {
			expression: path.MatchRoot("test"),
			other:      path.MatchRoot("not-test"),
			expected:   false,
		},
		"different-step-deep": {
			expression: path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			other:      path.MatchRoot("test2").AtListIndex(0).AtName("not-test2"),
			expected:   false,
		},
		"equal-shallow": {
			expression: path.MatchRoot("test"),
			other:      path.MatchRoot("test"),
			expected:   true,
		},
		"equal-deep": {
			expression: path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			other:      path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			expected:   true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.Equal(testCase.other)

			if got != testCase.expected {
				t.Errorf("expected %t, got %t", testCase.expected, got)
			}
		})
	}
}

func TestExpressionMatches(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		path       path.Path
		expected   bool
	}{
		"empty-empty": {
			expression: path.Expression{},
			path:       path.Empty(),
			expected:   false,
		},
		"empty-nonempty": {
			expression: path.Expression{},
			path:       path.Root("test"),
			expected:   false,
		},
		"nonempty-empty": {
			expression: path.MatchRoot("test"),
			path:       path.Empty(),
			expected:   false,
		},
		"AttributeNameExact-different": {
			expression: path.MatchRoot("test"),
			path:       path.Root("not-test"),
			expected:   false,
		},
		"AttributeNameExact-equal": {
			expression: path.MatchRoot("test"),
			path:       path.Root("test"),
			expected:   true,
		},
		"AttributeNameExact-AttributeNameExact-different-firststep": {
			expression: path.MatchRoot("test1").AtName("test2"),
			path:       path.Root("test2").AtName("test2"),
			expected:   false,
		},
		"AttributeNameExact-AttributeNameExact-different-laststep": {
			expression: path.MatchRoot("test1").AtName("test2"),
			path:       path.Root("test1").AtName("test3"),
			expected:   false,
		},
		"AttributeNameExact-AttributeNameExact-equal": {
			expression: path.MatchRoot("test1").AtName("test2"),
			path:       path.Root("test1").AtName("test2"),
			expected:   true,
		},
		"AttributeNameExact-AttributeNameExact-Parent-different": {
			expression: path.MatchRoot("test1").AtName("test2").AtParent(),
			path:       path.Root("test2"),
			expected:   false,
		},
		"AttributeNameExact-AttributeNameExact-Parent-equal": {
			expression: path.MatchRoot("test1").AtName("test2").AtParent(),
			path:       path.Root("test1"),
			expected:   true,
		},
		"AttributeNameExact-AttributeNameExact-Parent-AttributeNameExact-different": {
			expression: path.MatchRoot("test1").AtName("test2").AtParent().AtName("test3"),
			path:       path.Root("test1").AtName("test2"),
			expected:   false,
		},
		"AttributeNameExact-AttributeNameExact-Parent-AttributeNameExact-equal": {
			expression: path.MatchRoot("test1").AtName("test2").AtParent().AtName("test3"),
			path:       path.Root("test1").AtName("test3"),
			expected:   true,
		},
		"AttributeNameExact-ElementKeyIntAny": {
			expression: path.MatchRoot("test").AtAnyListIndex(),
			path:       path.Root("test").AtListIndex(0),
			expected:   true,
		},
		"AttributeNameExact-ElementKeyIntExact-different": {
			expression: path.MatchRoot("test").AtListIndex(0),
			path:       path.Root("test").AtListIndex(1),
			expected:   false,
		},
		"AttributeNameExact-ElementKeyIntExact-equal": {
			expression: path.MatchRoot("test").AtListIndex(0),
			path:       path.Root("test").AtListIndex(0),
			expected:   true,
		},
		"AttributeNameExact-ElementKeyStringAny": {
			expression: path.MatchRoot("test").AtAnyMapKey(),
			path:       path.Root("test").AtMapKey("test-key"),
			expected:   true,
		},
		"AttributeNameExact-ElementKeyStringExact-different": {
			expression: path.MatchRoot("test").AtMapKey("test-key"),
			path:       path.Root("test").AtMapKey("not-test-key"),
			expected:   false,
		},
		"AttributeNameExact-ElementKeyStringExact-equal": {
			expression: path.MatchRoot("test").AtMapKey("test-key"),
			path:       path.Root("test").AtMapKey("test-key"),
			expected:   true,
		},
		"AttributeNameExact-ElementKeyValueAny": {
			expression: path.MatchRoot("test").AtAnySetValue(),
			path:       path.Root("test").AtSetValue(types.StringValue("test-value")),
			expected:   true,
		},
		"AttributeNameExact-ElementKeyValueExact-different": {
			expression: path.MatchRoot("test").AtSetValue(types.StringValue("test-value")),
			path:       path.Root("test").AtSetValue(types.StringValue("not-test-value")),
			expected:   false,
		},
		"AttributeNameExact-ElementKeyValueExact-equal": {
			expression: path.MatchRoot("test").AtSetValue(types.StringValue("test-value")),
			path:       path.Root("test").AtSetValue(types.StringValue("test-value")),
			expected:   true,
		},
		"AttributeNameExact-Parent-AttributeNameExact-different": {
			expression: path.MatchRoot("test1").AtParent().AtName("test2"),
			path:       path.Root("test1"),
			expected:   false,
		},
		"AttributeNameExact-Parent-AttributeNameExact-equal": {
			expression: path.MatchRoot("test1").AtParent().AtName("test2"),
			path:       path.Root("test2"),
			expected:   true,
		},
		"Parent-AttributeNameExact": {
			expression: path.MatchRelative().AtParent().AtName("test"),
			path:       path.Root("test"),
			expected:   false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.Matches(testCase.path)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionMatchesParent(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		path       path.Path
		expected   bool
	}{
		"empty-empty": {
			expression: path.Expression{},
			path:       path.Empty(),
			expected:   false,
		},
		"empty-nonempty": {
			expression: path.Expression{},
			path:       path.Root("test"),
			expected:   false,
		},
		"nonempty-empty": {
			expression: path.MatchRoot("test"),
			path:       path.Empty(),
			expected:   true,
		},
		"AttributeNameExact-different": {
			expression: path.MatchRoot("test"),
			path:       path.Root("not-test"),
			expected:   false,
		},
		"AttributeNameExact-equal": {
			expression: path.MatchRoot("test"),
			path:       path.Root("test"),
			expected:   false,
		},
		"AttributeNameExact-parent": {
			expression: path.MatchRoot("test1").AtName("test2"),
			path:       path.Root("test1"),
			expected:   true,
		},
		"AttributeNameExact-AttributeNameExact-different-firststep": {
			expression: path.MatchRoot("test1").AtName("test2"),
			path:       path.Root("test2").AtName("test2"),
			expected:   false,
		},
		"AttributeNameExact-AttributeNameExact-different-laststep": {
			expression: path.MatchRoot("test1").AtName("test2"),
			path:       path.Root("test1").AtName("test3"),
			expected:   false,
		},
		"AttributeNameExact-AttributeNameExact-equal": {
			expression: path.MatchRoot("test1").AtName("test2"),
			path:       path.Root("test1").AtName("test2"),
			expected:   false,
		},
		"AttributeNameExact-AttributeNameExact-parent": {
			expression: path.MatchRoot("test1").AtName("test2").AtName("test3"),
			path:       path.Root("test1").AtName("test2"),
			expected:   true,
		},
		"AttributeNameExact-AttributeNameExact-Parent-different": {
			expression: path.MatchRoot("test1").AtName("test2").AtParent(),
			path:       path.Root("test2"),
			expected:   false,
		},
		"AttributeNameExact-AttributeNameExact-Parent-equal": {
			expression: path.MatchRoot("test1").AtName("test2").AtParent(),
			path:       path.Root("test1"),
			expected:   false,
		},
		"AttributeNameExact-AttributeNameExact-Parent-parent": {
			expression: path.MatchRoot("test1").AtName("test2").AtParent().AtName("test3"),
			path:       path.Root("test1"),
			expected:   true,
		},
		"AttributeNameExact-AttributeNameExact-Parent-AttributeNameExact-different": {
			expression: path.MatchRoot("test1").AtName("test2").AtParent().AtName("test3"),
			path:       path.Root("test1").AtName("test2"),
			expected:   false,
		},
		"AttributeNameExact-AttributeNameExact-Parent-AttributeNameExact-equal": {
			expression: path.MatchRoot("test1").AtName("test2").AtParent().AtName("test3"),
			path:       path.Root("test1").AtName("test3"),
			expected:   false,
		},
		"AttributeNameExact-AttributeNameExact-Parent-AttributeNameExact-parent": {
			expression: path.MatchRoot("test1").AtName("test2").AtParent().AtName("test3").AtName("test4"),
			path:       path.Root("test1").AtName("test3"),
			expected:   true,
		},
		"AttributeNameExact-ElementKeyIntAny": {
			expression: path.MatchRoot("test").AtAnyListIndex(),
			path:       path.Root("test").AtListIndex(0),
			expected:   false,
		},
		"AttributeNameExact-ElementKeyIntAny-parent": {
			expression: path.MatchRoot("test1").AtAnyListIndex().AtName("test2"),
			path:       path.Root("test1").AtListIndex(0),
			expected:   true,
		},
		"AttributeNameExact-ElementKeyIntExact-different": {
			expression: path.MatchRoot("test").AtListIndex(0),
			path:       path.Root("test").AtListIndex(1),
			expected:   false,
		},
		"AttributeNameExact-ElementKeyIntExact-equal": {
			expression: path.MatchRoot("test").AtListIndex(0),
			path:       path.Root("test").AtListIndex(0),
			expected:   false,
		},
		"AttributeNameExact-ElementKeyIntExact-parent": {
			expression: path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			path:       path.Root("test1").AtListIndex(0),
			expected:   true,
		},
		"AttributeNameExact-ElementKeyStringAny": {
			expression: path.MatchRoot("test").AtAnyMapKey(),
			path:       path.Root("test").AtMapKey("test-key"),
			expected:   false,
		},
		"AttributeNameExact-ElementKeyStringAny-parent": {
			expression: path.MatchRoot("test1").AtAnyMapKey().AtName("test2"),
			path:       path.Root("test1").AtMapKey("test-key"),
			expected:   true,
		},
		"AttributeNameExact-ElementKeyStringExact-different": {
			expression: path.MatchRoot("test").AtMapKey("test-key"),
			path:       path.Root("test").AtMapKey("not-test-key"),
			expected:   false,
		},
		"AttributeNameExact-ElementKeyStringExact-equal": {
			expression: path.MatchRoot("test").AtMapKey("test-key"),
			path:       path.Root("test").AtMapKey("test-key"),
			expected:   false,
		},
		"AttributeNameExact-ElementKeyStringExact-parent": {
			expression: path.MatchRoot("test1").AtMapKey("test-key").AtName("test2"),
			path:       path.Root("test1").AtMapKey("test-key"),
			expected:   true,
		},
		"AttributeNameExact-ElementKeyValueAny": {
			expression: path.MatchRoot("test").AtAnySetValue(),
			path:       path.Root("test").AtSetValue(types.StringValue("test-value")),
			expected:   false,
		},
		"AttributeNameExact-ElementKeyValueAny-parent": {
			expression: path.MatchRoot("test1").AtAnySetValue().AtName("test2"),
			path:       path.Root("test1").AtSetValue(types.StringValue("test-value")),
			expected:   true,
		},
		"AttributeNameExact-ElementKeyValueExact-different": {
			expression: path.MatchRoot("test").AtSetValue(types.StringValue("test-value")),
			path:       path.Root("test").AtSetValue(types.StringValue("not-test-value")),
			expected:   false,
		},
		"AttributeNameExact-ElementKeyValueExact-equal": {
			expression: path.MatchRoot("test").AtSetValue(types.StringValue("test-value")),
			path:       path.Root("test").AtSetValue(types.StringValue("test-value")),
			expected:   false,
		},
		"AttributeNameExact-ElementKeyValueExact-parent": {
			expression: path.MatchRoot("test1").AtSetValue(types.StringValue("test-value")).AtName("test2"),
			path:       path.Root("test1").AtSetValue(types.StringValue("test-value")),
			expected:   true,
		},
		"AttributeNameExact-Parent": {
			expression: path.MatchRoot("test1").AtParent(),
			path:       path.Root("test1"),
			expected:   false,
		},
		"AttributeNameExact-Parent-AttributeNameExact-different": {
			expression: path.MatchRoot("test1").AtParent().AtName("test2"),
			path:       path.Root("test1"),
			expected:   false,
		},
		"AttributeNameExact-Parent-AttributeNameExact-equal": {
			expression: path.MatchRoot("test1").AtParent().AtName("test2"),
			path:       path.Root("test2"),
			expected:   false,
		},
		"AttributeNameExact-Parent-AttributeNameExact-parent": {
			expression: path.MatchRoot("test1").AtParent().AtName("test2").AtName("test3"),
			path:       path.Root("test2"),
			expected:   true,
		},
		"Parent-AttributeNameExact": {
			expression: path.MatchRelative().AtParent().AtName("test"),
			path:       path.Root("test"),
			expected:   false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.MatchesParent(testCase.path)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionMerge(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		other      path.Expression
		expected   path.Expression
	}{
		"Relative-further": {
			expression: path.MatchRoot("test1"),
			other:      path.MatchRelative().AtName("test2"),
			expected:   path.MatchRoot("test1").AtName("test2"),
		},
		"Relative-Parent-root-level": {
			expression: path.MatchRoot("test1"),
			other:      path.MatchRelative().AtParent().AtName("test2"),
			expected:   path.MatchRoot("test1").AtParent().AtName("test2"),
		},
		"Relative-Parent-nested-level": {
			expression: path.MatchRoot("test_parent").AtListIndex(1).AtName("test_child1"),
			other:      path.MatchRelative().AtParent().AtName("test_child2"),
			expected:   path.MatchRoot("test_parent").AtListIndex(1).AtName("test_child1").AtParent().AtName("test_child2"),
		},
		"Root": {
			expression: path.MatchRoot("test1"),
			other:      path.MatchRoot("test2"),
			expected:   path.MatchRoot("test2"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.Merge(testCase.other)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected result difference: %s", diff)
			}
		})
	}
}

func TestExpressionMergeExpressions(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		others     path.Expressions
		expected   path.Expressions
	}{
		"nil": {
			expression: path.MatchRoot("test"),
			others:     nil,
			expected: path.Expressions{
				path.MatchRoot("test"),
			},
		},
		"empty": {
			expression: path.MatchRoot("test"),
			others:     path.Expressions{},
			expected: path.Expressions{
				path.MatchRoot("test"),
			},
		},
		"Relative-further": {
			expression: path.MatchRoot("test1"),
			others: path.Expressions{
				path.MatchRelative().AtName("test2"),
			},
			expected: path.Expressions{
				path.MatchRoot("test1").AtName("test2"),
			},
		},
		"Relative-Parent-root-level": {
			expression: path.MatchRoot("test1"),
			others: path.Expressions{
				path.MatchRelative().AtParent().AtName("test2"),
			},
			expected: path.Expressions{
				path.MatchRoot("test1").AtParent().AtName("test2"),
			},
		},
		"Relative-Parent-nested-level": {
			expression: path.MatchRoot("test_parent").AtListIndex(1).AtName("test_child1"),
			others: path.Expressions{
				path.MatchRelative().AtParent().AtName("test_child2"),
			},
			expected: path.Expressions{
				path.MatchRoot("test_parent").AtListIndex(1).AtName("test_child1").AtParent().AtName("test_child2"),
			},
		},
		"Root": {
			expression: path.MatchRoot("test1"),
			others: path.Expressions{
				path.MatchRoot("test2"),
			},
			expected: path.Expressions{
				path.MatchRoot("test2"),
			},
		},
		"multiple": {
			expression: path.MatchRoot("test_parent").AtListIndex(1).AtName("test_child1"),
			others: path.Expressions{
				path.MatchRelative().AtParent().AtName("test2"),
				path.MatchRoot("test3"),
			},
			expected: path.Expressions{
				path.MatchRoot("test_parent").AtListIndex(1).AtName("test_child1").AtParent().AtName("test2"),
				path.MatchRoot("test3"),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.MergeExpressions(testCase.others...)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected result difference: %s", diff)
			}
		})
	}
}

func TestExpressionResolve(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		expected   path.Expression
	}{
		// Refer to TestExpressionStepsResolve for more exhaustive unit
		// testing of the underlying step resolving functionality.
		"AttributeNameExact": {
			expression: path.MatchRoot("test1"),
			expected:   path.MatchRoot("test1"),
		},
		"AttributeNameExact-AttributeNameExact": {
			expression: path.MatchRoot("test1").AtName("test2"),
			expected:   path.MatchRoot("test1").AtName("test2"),
		},
		"AttributeNameExact-Parent-AttributeNameExact": {
			expression: path.MatchRoot("test1").AtParent().AtName("test2"),
			expected:   path.MatchRoot("test2"),
		},
		"AttributeNameExact-ElementKeyIntExact-AttributeNameExact-Parent-AttributeNameExact": {
			expression: path.MatchRoot("test_parent").AtListIndex(1).AtName("test_child1").AtParent().AtName("test_child2"),
			expected:   path.MatchRoot("test_parent").AtListIndex(1).AtName("test_child2"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.Resolve()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionSteps(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		expected   path.ExpressionSteps
	}{
		"one": {
			expression: path.MatchRoot("test"),
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
			},
		},
		"two": {
			expression: path.MatchRoot("test").AtListIndex(1),
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntExact(1),
			},
		},
		"any": {
			expression: path.MatchRoot("test").AtAnyListIndex(),
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepElementKeyIntAny{},
			},
		},
		"parent": {
			expression: path.MatchRoot("test").AtParent(),
			expected: path.ExpressionSteps{
				path.ExpressionStepAttributeNameExact("test"),
				path.ExpressionStepParent{},
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.Steps()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestExpressionString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		expression path.Expression
		expected   string
	}{
		"AttributeNameExact": {
			expression: path.MatchRoot("test"),
			expected:   `test`,
		},
		"AttributeNameExact-AttributeNameExact": {
			expression: path.MatchRoot("test1").AtName("test2"),
			expected:   `test1.test2`,
		},
		"AttributeNameExact-AttributeNameExact-AttributeNameExact": {
			expression: path.MatchRoot("test1").AtName("test2").AtName("test3"),
			expected:   `test1.test2.test3`,
		},
		"AttributeNameExact-ElementKeyIntAny": {
			expression: path.MatchRoot("test").AtAnyListIndex(),
			expected:   `test[*]`,
		},
		"AttributeNameExact-ElementKeyIntExact": {
			expression: path.MatchRoot("test").AtListIndex(0),
			expected:   `test[0]`,
		},
		"AttributeNameExact-ElementKeyIntExact-AttributeNameExact": {
			expression: path.MatchRoot("test1").AtListIndex(0).AtName("test2"),
			expected:   `test1[0].test2`,
		},
		"AttributeNameExact-ElementKeyIntExact-ElementKeyIntExact": {
			expression: path.MatchRoot("test").AtListIndex(0).AtListIndex(1),
			expected:   `test[0][1]`,
		},
		"AttributeNameExact-ElementKeyStringAny": {
			expression: path.MatchRoot("test").AtAnyMapKey(),
			expected:   `test["*"]`,
		},
		"AttributeNameExact-ElementKeyStringExact": {
			expression: path.MatchRoot("test").AtMapKey("test-key"),
			expected:   `test["test-key"]`,
		},
		"AttributeNameExact-ElementKeyStringExact-AttributeNameExact": {
			expression: path.MatchRoot("test1").AtMapKey("test-key").AtName("test2"),
			expected:   `test1["test-key"].test2`,
		},
		"AttributeNameExact-ElementKeyStringExact-ElementKeyStringExact": {
			expression: path.MatchRoot("test").AtMapKey("test-key1").AtMapKey("test-key2"),
			expected:   `test["test-key1"]["test-key2"]`,
		},
		"AttributeNameExact-ElementKeyValueAny": {
			expression: path.MatchRoot("test").AtAnySetValue(),
			expected:   `test[Value(*)]`,
		},
		"AttributeNameExact-ElementKeyValueExact": {
			expression: path.MatchRoot("test").AtSetValue(types.StringValue("test-value")),
			expected:   `test[Value("test-value")]`,
		},
		"AttributeNameExact-ElementKeyValue-AttributeNameExact": {
			expression: path.MatchRoot("test").AtSetValue(types.ObjectValueMust(
				map[string]attr.Type{
					"test_attr_1": types.BoolType,
					"test_attr_2": types.StringType,
				},
				map[string]attr.Value{
					"test_attr_1": types.BoolValue(true),
					"test_attr_2": types.StringValue("test-value"),
				},
			)).AtName("test_attr_1"),
			expected: `test[Value({"test_attr_1":true,"test_attr_2":"test-value"})].test_attr_1`,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.String()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
