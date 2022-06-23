package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

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
