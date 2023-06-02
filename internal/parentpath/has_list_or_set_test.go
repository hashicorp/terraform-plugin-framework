// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package parentpath_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/internal/parentpath"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestHasListOrSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		expected bool
	}{
		"empty": {
			path:     path.Empty(),
			expected: false,
		},
		"AttributeName": {
			path:     path.Root("test"),
			expected: false,
		},
		"AttributeName-AttributeName": {
			path:     path.Root("test").AtName("nested_test"),
			expected: false,
		},
		"AttributeName-AttributeName-ElementKeyInt": {
			path:     path.Root("test").AtName("nested_test").AtListIndex(0),
			expected: true,
		},
		"AttributeName-AttributeName-ElementKeyInt-AttributeName": {
			path:     path.Root("test").AtName("nested_test").AtListIndex(0).AtName("nested_nested_test"),
			expected: true,
		},
		"AttributeName-AttributeName-ElementKeyString": {
			path:     path.Root("test").AtMapKey("testkey"),
			expected: false,
		},
		"AttributeName-AttributeName-ElementKeyValue": {
			path: path.Root("test").AtSetValue(
				types.SetValueMust(
					types.StringType,
					[]attr.Value{types.StringValue("testvalue")},
				),
			),
			expected: true,
		},
		"AttributeName-ElementKeyInt": {
			path:     path.Root("test").AtListIndex(0),
			expected: true,
		},
		"AttributeName-ElementKeyInt-AttributeName": {
			path:     path.Root("test").AtListIndex(0).AtName("nested_test"),
			expected: true,
		},
		"AttributeName-ElementKeyInt-ElementKeyInt": {
			path:     path.Root("test").AtListIndex(0).AtListIndex(0),
			expected: true,
		},
		"AttributeName-ElementKeyInt-ElementKeyString": {
			path:     path.Root("test").AtListIndex(0).AtMapKey("testkey"),
			expected: true,
		},
		"AttributeName-ElementKeyInt-ElementKeyValue": {
			path: path.Root("test").AtSetValue(
				types.SetValueMust(
					types.StringType,
					[]attr.Value{types.StringValue("testvalue")},
				),
			),
			expected: true,
		},
		"AttributeName-ElementKeyString": {
			path:     path.Root("test").AtMapKey("testkey"),
			expected: false,
		},
		"AttributeName-ElementKeyString-AttributeName": {
			path:     path.Root("test").AtMapKey("testkey").AtName("nested_test"),
			expected: false,
		},
		"AttributeName-ElementKeyString-AttributeName-ElementKeyInt": {
			path:     path.Root("test").AtMapKey("testkey").AtName("nested_test").AtListIndex(0),
			expected: true,
		},
		"AttributeName-ElementKeyString-AttributeName-ElementKeyInt-AttributeName": {
			path:     path.Root("test").AtMapKey("testkey").AtName("nested_test").AtListIndex(0).AtName("nested_nested_test"),
			expected: true,
		},
		"AttributeName-ElementKeyString-ElementKeyInt": {
			path:     path.Root("test").AtMapKey("testkey").AtListIndex(0),
			expected: true,
		},
		"AttributeName-ElementKeyString-ElementKeyInt-AttributeName": {
			path:     path.Root("test").AtMapKey("testkey").AtListIndex(0).AtName("nested_test"),
			expected: true,
		},
		"AttributeName-ElementKeyString-ElementKeyString": {
			path:     path.Root("test").AtMapKey("testkey").AtMapKey("nested_testkey"),
			expected: false,
		},
		"AttributeName-ElementKeyString-ElementKeyValue": {
			path: path.Root("test").AtMapKey("testkey").AtSetValue(
				types.SetValueMust(
					types.StringType,
					[]attr.Value{types.StringValue("testvalue")},
				),
			),
			expected: true,
		},
		"AttributeName-ElementKeyValue": {
			path: path.Root("test").AtSetValue(
				types.SetValueMust(
					types.StringType,
					[]attr.Value{types.StringValue("testvalue")},
				),
			),
			expected: true,
		},
		"AttributeName-ElementKeyValue-AttributeName": {
			path: path.Root("test").AtSetValue(
				types.SetValueMust(
					types.StringType,
					[]attr.Value{types.StringValue("testvalue")},
				),
			).AtName("nested_test"),
			expected: true,
		},
		"AttributeName-ElementKeyValue-ElementKeyInt": {
			path: path.Root("test").AtSetValue(
				types.SetValueMust(
					types.StringType,
					[]attr.Value{types.StringValue("testvalue")},
				),
			).AtListIndex(0),
			expected: true,
		},
		"AttributeName-ElementKeyValue-ElementKeyString": {
			path: path.Root("test").AtSetValue(
				types.SetValueMust(
					types.StringType,
					[]attr.Value{types.StringValue("testvalue")},
				),
			).AtMapKey("testkey"),
			expected: true,
		},
		"AttributeName-ElementKeyValue-ElementKeyValue": {
			path: path.Root("test").AtSetValue(
				types.SetValueMust(
					types.SetType{
						ElemType: types.StringType,
					},
					[]attr.Value{
						types.SetValueMust(
							types.StringType,
							[]attr.Value{types.StringValue("testvalue")},
						),
					},
				),
			).AtSetValue(
				types.SetValueMust(
					types.StringType,
					[]attr.Value{types.StringValue("testvalue")},
				),
			),
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := parentpath.HasListOrSet(testCase.path)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
