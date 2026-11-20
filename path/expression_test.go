// Copyright IBM Corp. 2021, 2026
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var benchmarkExpressionResult path.Expression

func BenchmarkExpressionBuilders(b *testing.B) {
	expressions := map[string]path.Expression{
		"empty": path.MatchRelative(),
		"root":  path.MatchRoot("root"),
		"deep":  path.MatchRoot("root").AtName("nested").AtListIndex(0).AtName("leaf"),
	}
	value := types.StringValue("element")

	for name, e := range expressions {
		b.Run(name, func(b *testing.B) {
			b.Run("AtAnyListIndex", func(b *testing.B) {
				for b.Loop() {
					benchmarkExpressionResult = e.AtAnyListIndex()
				}
			})
			b.Run("AtAnyMapKey", func(b *testing.B) {
				for b.Loop() {
					benchmarkExpressionResult = e.AtAnyMapKey()
				}
			})
			b.Run("AtAnySetValue", func(b *testing.B) {
				for b.Loop() {
					benchmarkExpressionResult = e.AtAnySetValue()
				}
			})
			b.Run("AtListIndex", func(b *testing.B) {
				for b.Loop() {
					benchmarkExpressionResult = e.AtListIndex(0)
				}
			})
			b.Run("AtMapKey", func(b *testing.B) {
				for b.Loop() {
					benchmarkExpressionResult = e.AtMapKey("key")
				}
			})
			b.Run("AtName", func(b *testing.B) {
				for b.Loop() {
					benchmarkExpressionResult = e.AtName("attribute")
				}
			})
			b.Run("AtParent", func(b *testing.B) {
				for b.Loop() {
					benchmarkExpressionResult = e.AtParent()
				}
			})
			b.Run("AtSetValue", func(b *testing.B) {
				for b.Loop() {
					benchmarkExpressionResult = e.AtSetValue(value)
				}
			})
		})
	}
}

func BenchmarkExpressionChain(b *testing.B) {
	root := path.MatchRoot("root")

	b.Run("one", func(b *testing.B) {
		for b.Loop() {
			benchmarkExpressionResult = root.AtName("one")
		}
	})
	b.Run("two", func(b *testing.B) {
		for b.Loop() {
			benchmarkExpressionResult = root.AtName("one").AtListIndex(0)
		}
	})
	b.Run("four", func(b *testing.B) {
		for b.Loop() {
			benchmarkExpressionResult = root.AtName("one").AtListIndex(0).AtName("two").AtAnyMapKey()
		}
	})
	b.Run("eight", func(b *testing.B) {
		for b.Loop() {
			benchmarkExpressionResult = root.AtName("one").AtListIndex(0).AtName("two").AtAnyMapKey().AtName("three").AtAnyListIndex().AtName("four").AtParent()
		}
	})
}

func BenchmarkExpressionCopy(b *testing.B) {
	for _, depth := range []int{0, 1, 2, 3, 4, 8, 16, 32} {
		expression := path.MatchRelative()
		for range depth {
			expression = expression.AtName("attribute")
		}

		b.Run(fmt.Sprintf("depth-%d", depth), func(b *testing.B) {
			for b.Loop() {
				benchmarkExpressionResult = expression.Copy()
			}
		})
	}
}

func BenchmarkExpressionMerge(b *testing.B) {
	others := map[string]path.Expression{
		"relative-empty": path.MatchRelative(),
		"relative-one":   path.MatchRelative().AtName("one"),
		"relative-four":  path.MatchRelative().AtName("one").AtListIndex(0).AtParent().AtName("two"),
		"root":           path.MatchRoot("other").AtName("nested"),
	}

	for _, depth := range []int{0, 1, 2, 3, 4, 8, 16, 32} {
		expression := path.MatchRelative()
		for range depth {
			expression = expression.AtName("attribute")
		}

		b.Run(fmt.Sprintf("depth-%d", depth), func(b *testing.B) {
			for name, other := range others {
				b.Run(name, func(b *testing.B) {
					for b.Loop() {
						benchmarkExpressionResult = expression.Merge(other)
					}
				})
			}
		})
	}
}

var benchmarkExpressionsResult path.Expressions

func BenchmarkExpressionMergeExpressions(b *testing.B) {
	base := path.MatchRoot("root").AtName("nested")
	for _, count := range []int{0, 1, 2, 4, 8, 16, 32} {
		unique := make(path.Expressions, count)
		duplicates := make(path.Expressions, count)
		for i := range count {
			unique[i] = path.MatchRelative().AtListIndex(i)
			duplicates[i] = path.MatchRelative().AtListIndex(0)
		}

		b.Run(fmt.Sprintf("count-%d", count), func(b *testing.B) {
			b.Run("unique", func(b *testing.B) {
				for b.Loop() {
					benchmarkExpressionsResult = base.MergeExpressions(unique...)
				}
			})
			b.Run("duplicates", func(b *testing.B) {
				for b.Loop() {
					benchmarkExpressionsResult = base.MergeExpressions(duplicates...)
				}
			})
		})
	}
}

func BenchmarkExpressionResolve(b *testing.B) {
	for _, depth := range []int{0, 1, 2, 3, 4, 8, 16, 32} {
		expression := path.MatchRelative()
		for range depth {
			expression = expression.AtName("attribute")
		}

		b.Run(fmt.Sprintf("depth-%d", depth), func(b *testing.B) {
			cases := map[string]path.Expression{
				"no-parent":   expression,
				"parent":      expression.AtName("child").AtParent().AtName("sibling"),
				"beyond-root": path.MatchRelative().AtParent().Merge(expression),
			}
			for name, candidate := range cases {
				b.Run(name, func(b *testing.B) {
					for b.Loop() {
						benchmarkExpressionResult = candidate.Resolve()
					}
				})
			}
		})
	}
}

func TestExpressionBuilders(t *testing.T) {
	t.Parallel()

	value := types.StringValue("element")
	testCases := map[string]struct {
		build func(path.Expression) path.Expression
		step  path.ExpressionStep
	}{
		"any list index": {func(e path.Expression) path.Expression { return e.AtAnyListIndex() }, path.ExpressionStepElementKeyIntAny{}},
		"any map key":    {func(e path.Expression) path.Expression { return e.AtAnyMapKey() }, path.ExpressionStepElementKeyStringAny{}},
		"any set value":  {func(e path.Expression) path.Expression { return e.AtAnySetValue() }, path.ExpressionStepElementKeyValueAny{}},
		"list index":     {func(e path.Expression) path.Expression { return e.AtListIndex(1) }, path.ExpressionStepElementKeyIntExact(1)},
		"map key":        {func(e path.Expression) path.Expression { return e.AtMapKey("key") }, path.ExpressionStepElementKeyStringExact("key")},
		"name":           {func(e path.Expression) path.Expression { return e.AtName("attribute") }, path.ExpressionStepAttributeNameExact("attribute")},
		"parent":         {func(e path.Expression) path.Expression { return e.AtParent() }, path.ExpressionStepParent{}},
		"set value":      {func(e path.Expression) path.Expression { return e.AtSetValue(value) }, path.ExpressionStepElementKeyValueExact{Value: value}},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			for _, root := range []bool{false, true} {
				original := path.MatchRelative().AtName("base")
				if root {
					original = path.MatchRoot("base")
				}

				got := testCase.build(original)
				_ = original.AtName("other")
				wantSteps := path.ExpressionSteps{path.ExpressionStepAttributeNameExact("base"), testCase.step}
				if diff := cmp.Diff(wantSteps, got.Steps()); diff != "" {
					t.Errorf("unexpected steps (root=%t): %s", root, diff)
				}
				if diff := cmp.Diff(path.ExpressionSteps{path.ExpressionStepAttributeNameExact("base")}, original.Steps()); diff != "" {
					t.Errorf("original changed (root=%t): %s", root, diff)
				}

				if !root {
					wantSteps = append(path.ExpressionSteps{path.ExpressionStepAttributeNameExact("prefix")}, wantSteps...)
				}
				merged := path.MatchRoot("prefix").Merge(got)
				if diff := cmp.Diff(wantSteps, merged.Steps()); diff != "" {
					t.Errorf("unexpected merged steps (root=%t): %s", root, diff)
				}
			}
		})
	}
}

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
		"zero": {
			expression: path.Expression{},
			expected:   path.MatchRelative(),
		},
		"empty": {
			expression: path.MatchRelative(),
			expected:   path.MatchRelative(),
		},
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
		"duplicates": {
			expression: path.MatchRoot("root"),
			others: path.Expressions{
				path.MatchRelative().AtName("first"),
				path.MatchRoot("root").AtName("first"),
				path.MatchRelative().AtName("second"),
				path.MatchRelative().AtName("first"),
			},
			expected: path.Expressions{
				path.MatchRoot("root").AtName("first"),
				path.MatchRoot("root").AtName("second"),
			},
		},
	}

	for name, testCase := range testCases {
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
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.expression.String()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
