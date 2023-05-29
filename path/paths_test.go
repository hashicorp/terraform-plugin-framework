// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPathsAppend(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		paths    path.Paths
		add      path.Paths
		expected path.Paths
	}{
		"nil-nil": {
			paths:    nil,
			add:      nil,
			expected: nil,
		},
		"nil-nonempty": {
			paths:    nil,
			add:      path.Paths{path.Root("test")},
			expected: path.Paths{path.Root("test")},
		},
		"nonempty-nil": {
			paths:    path.Paths{path.Root("test")},
			add:      nil,
			expected: path.Paths{path.Root("test")},
		},
		"empty-empty": {
			paths:    path.Paths{},
			add:      path.Paths{},
			expected: path.Paths{},
		},
		"empty-nonempty": {
			paths:    path.Paths{},
			add:      path.Paths{path.Root("test")},
			expected: path.Paths{path.Root("test")},
		},
		"nonempty-empty": {
			paths:    path.Paths{path.Root("test")},
			add:      path.Paths{},
			expected: path.Paths{path.Root("test")},
		},
		"nonempty-nonempty": {
			paths: path.Paths{
				path.Root("test1"),
				path.Root("test2"),
			},
			add: path.Paths{
				path.Root("test3"),
				path.Root("test4"),
			},
			expected: path.Paths{
				path.Root("test1"),
				path.Root("test2"),
				path.Root("test3"),
				path.Root("test4"),
			},
		},
		"deduplication": {
			paths: path.Paths{
				path.Root("test1"),
				path.Root("test2"),
			},
			add: path.Paths{
				path.Root("test1"),
				path.Root("test3"),
			},
			expected: path.Paths{
				path.Root("test1"),
				path.Root("test2"),
				path.Root("test3"),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.paths.Append(testCase.add...)

			if diff := cmp.Diff(testCase.paths, testCase.expected); diff != "" {
				t.Errorf("unexpected original difference: %s", diff)
			}

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected result difference: %s", diff)
			}
		})
	}
}

func TestPathsContains(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		paths    path.Paths
		contains path.Path
		expected bool
	}{
		"paths-nil": {
			paths:    nil,
			contains: path.Root("test"),
			expected: false,
		},
		"paths-empty": {
			paths:    path.Paths{},
			contains: path.Root("test"),
			expected: false,
		},
		"contains-empty": {
			paths: path.Paths{
				path.Root("test"),
			},
			contains: path.Empty(),
			expected: false,
		},
		"contains-middle": {
			paths: path.Paths{
				path.Root("test1").AtName("test1_attr"),
				path.Root("test2").AtName("test2_attr"),
				path.Root("test3").AtName("test3_attr"),
			},
			contains: path.Root("test2").AtName("test2_attr"),
			expected: true,
		},
		"contains-end": {
			paths: path.Paths{
				path.Root("test1").AtName("test1_attr"),
				path.Root("test2").AtName("test2_attr"),
				path.Root("test3").AtName("test3_attr"),
			},
			contains: path.Root("test3").AtName("test3_attr"),
			expected: true,
		},
		"AttributeName-different": {
			paths: path.Paths{
				path.Root("test"),
			},
			contains: path.Root("not-test"),
			expected: false,
		},
		"AttributeName-equal": {
			paths: path.Paths{
				path.Root("test"),
			},
			contains: path.Root("test"),
			expected: true,
		},
		"ElementKeyInt-different": {
			paths: path.Paths{
				path.Empty().AtListIndex(0),
			},
			contains: path.Empty().AtListIndex(1),
			expected: false,
		},
		"ElementKeyInt-equal": {
			paths: path.Paths{
				path.Empty().AtListIndex(0),
			},
			contains: path.Empty().AtListIndex(0),
			expected: true,
		},
		"ElementKeyString-different": {
			paths: path.Paths{
				path.Empty().AtMapKey("test"),
			},
			contains: path.Empty().AtMapKey("not-test"),
			expected: false,
		},
		"ElementKeyString-equal": {
			paths: path.Paths{
				path.Empty().AtMapKey("test"),
			},
			contains: path.Empty().AtMapKey("test"),
			expected: true,
		},
		"ElementKeyValue-different": {
			paths: path.Paths{
				path.Empty().AtSetValue(types.StringValue("test")),
			},
			contains: path.Empty().AtSetValue(types.StringValue("not-test")),
			expected: false,
		},
		"ElementKeyValue-equal": {
			paths: path.Paths{
				path.Empty().AtSetValue(types.StringValue("test")),
			},
			contains: path.Empty().AtSetValue(types.StringValue("test")),
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
				path.Root("test"),
			},
			expected: "[test]",
		},
		"one-empty": {
			paths: path.Paths{
				path.Empty(),
			},
			expected: "[]",
		},
		"two": {
			paths: path.Paths{
				path.Root("test1"),
				path.Root("test2"),
			},
			expected: "[test1,test2]",
		},
		"two-empty": {
			paths: path.Paths{
				path.Root("test"),
				path.Empty(),
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
