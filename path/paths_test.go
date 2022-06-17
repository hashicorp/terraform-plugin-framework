package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPathsContains(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		paths    path.Paths
		contains path.Path
		expected bool
	}{
		"paths-nil": {
			paths:    nil,
			contains: path.RootPath("test"),
			expected: false,
		},
		"paths-empty": {
			paths:    path.Paths{},
			contains: path.RootPath("test"),
			expected: false,
		},
		"contains-empty": {
			paths: path.Paths{
				path.RootPath("test"),
			},
			contains: path.EmptyPath(),
			expected: false,
		},
		"contains-middle": {
			paths: path.Paths{
				path.RootPath("test1").AtName("test1_attr"),
				path.RootPath("test2").AtName("test2_attr"),
				path.RootPath("test3").AtName("test3_attr"),
			},
			contains: path.RootPath("test2").AtName("test2_attr"),
			expected: true,
		},
		"contains-end": {
			paths: path.Paths{
				path.RootPath("test1").AtName("test1_attr"),
				path.RootPath("test2").AtName("test2_attr"),
				path.RootPath("test3").AtName("test3_attr"),
			},
			contains: path.RootPath("test3").AtName("test3_attr"),
			expected: true,
		},
		"AttributeName-different": {
			paths: path.Paths{
				path.RootPath("test"),
			},
			contains: path.RootPath("not-test"),
			expected: false,
		},
		"AttributeName-equal": {
			paths: path.Paths{
				path.RootPath("test"),
			},
			contains: path.RootPath("test"),
			expected: true,
		},
		"ElementKeyInt-different": {
			paths: path.Paths{
				path.EmptyPath().AtListIndex(0),
			},
			contains: path.EmptyPath().AtListIndex(1),
			expected: false,
		},
		"ElementKeyInt-equal": {
			paths: path.Paths{
				path.EmptyPath().AtListIndex(0),
			},
			contains: path.EmptyPath().AtListIndex(0),
			expected: true,
		},
		"ElementKeyString-different": {
			paths: path.Paths{
				path.EmptyPath().AtMapKey("test"),
			},
			contains: path.EmptyPath().AtMapKey("not-test"),
			expected: false,
		},
		"ElementKeyString-equal": {
			paths: path.Paths{
				path.EmptyPath().AtMapKey("test"),
			},
			contains: path.EmptyPath().AtMapKey("test"),
			expected: true,
		},
		"ElementKeyValue-different": {
			paths: path.Paths{
				path.EmptyPath().AtSetValue(types.String{Value: "test"}),
			},
			contains: path.EmptyPath().AtSetValue(types.String{Value: "not-test"}),
			expected: false,
		},
		"ElementKeyValue-equal": {
			paths: path.Paths{
				path.EmptyPath().AtSetValue(types.String{Value: "test"}),
			},
			contains: path.EmptyPath().AtSetValue(types.String{Value: "test"}),
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.paths.Contains(testCase.contains)

			if got != testCase.expected {
				t.Errorf("expected %t, got %t", testCase.expected, got)
			}
		})
	}
}

func TestPathsString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		paths    path.Paths
		expected string
	}{
		"nil": {
			paths:    nil,
			expected: "[]",
		},
		"empty": {
			paths:    path.Paths{},
			expected: "[]",
		},
		"one": {
			paths: path.Paths{
				path.RootPath("test"),
			},
			expected: "[test]",
		},
		"one-empty": {
			paths: path.Paths{
				path.EmptyPath(),
			},
			expected: "[]",
		},
		"two": {
			paths: path.Paths{
				path.RootPath("test1"),
				path.RootPath("test2"),
			},
			expected: "[test1,test2]",
		},
		"two-empty": {
			paths: path.Paths{
				path.RootPath("test"),
				path.EmptyPath(),
			},
			expected: "[test]",
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.paths.String()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
