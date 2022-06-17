package path_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestPathAtListIndex(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		index    int
		expected path.Path
	}{
		"empty": {
			path:     path.EmptyPath(),
			index:    1,
			expected: path.EmptyPath().AtListIndex(1),
		},
		"shallow": {
			path:     path.RootPath("test"),
			index:    1,
			expected: path.RootPath("test").AtListIndex(1),
		},
		"deep": {
			path:     path.RootPath("test1").AtListIndex(0).AtName("test2"),
			index:    1,
			expected: path.RootPath("test1").AtListIndex(0).AtName("test2").AtListIndex(1),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.path.AtListIndex(testCase.index)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestPathAtMapKey(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		key      string
		expected path.Path
	}{
		"empty": {
			path:     path.EmptyPath(),
			key:      "test-key",
			expected: path.EmptyPath().AtMapKey("test-key"),
		},
		"shallow": {
			path:     path.RootPath("test"),
			key:      "test-key",
			expected: path.RootPath("test").AtMapKey("test-key"),
		},
		"deep": {
			path:     path.RootPath("test1").AtListIndex(0).AtName("test2"),
			key:      "test-key",
			expected: path.RootPath("test1").AtListIndex(0).AtName("test2").AtMapKey("test-key"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.path.AtMapKey(testCase.key)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestPathAtName(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		name     string
		expected path.Path
	}{
		"empty": {
			path:     path.EmptyPath(),
			name:     "test",
			expected: path.RootPath("test"),
		},
		"shallow": {
			path:     path.RootPath("test1"),
			name:     "test2",
			expected: path.RootPath("test1").AtName("test2"),
		},
		"deep": {
			path:     path.RootPath("test1").AtListIndex(0),
			name:     "test2",
			expected: path.RootPath("test1").AtListIndex(0).AtName("test2"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.path.AtName(testCase.name)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestPathAtSetValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		value    attr.Value
		expected path.Path
	}{
		"empty": {
			path:     path.EmptyPath(),
			value:    types.String{Value: "test"},
			expected: path.EmptyPath().AtSetValue(types.String{Value: "test"}),
		},
		"shallow": {
			path:     path.RootPath("test"),
			value:    types.String{Value: "test"},
			expected: path.RootPath("test").AtSetValue(types.String{Value: "test"}),
		},
		"deep": {
			path:     path.RootPath("test1").AtListIndex(0).AtName("test2"),
			value:    types.String{Value: "test"},
			expected: path.RootPath("test1").AtListIndex(0).AtName("test2").AtSetValue(types.String{Value: "test"}),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.path.AtSetValue(testCase.value)

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestPathCopy(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		expected path.Path
	}{
		"empty": {
			path:     path.EmptyPath(),
			expected: path.EmptyPath(),
		},
		"shallow": {
			path:     path.RootPath("test"),
			expected: path.RootPath("test"),
		},
		"deep": {
			path:     path.RootPath("test1").AtListIndex(0).AtName("test2"),
			expected: path.RootPath("test1").AtListIndex(0).AtName("test2"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.path.Copy()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestPathEqual(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		other    path.Path
		expected bool
	}{
		"empty-empty": {
			path:     path.EmptyPath(),
			other:    path.EmptyPath(),
			expected: true,
		},
		"different-length": {
			path:     path.RootPath("test1").AtName("test2"),
			other:    path.RootPath("test1"),
			expected: false,
		},
		"different-step-shallow": {
			path:     path.RootPath("test"),
			other:    path.RootPath("not-test"),
			expected: false,
		},
		"different-step-deep": {
			path:     path.RootPath("test1").AtListIndex(0).AtName("test2"),
			other:    path.RootPath("test2").AtListIndex(0).AtName("not-test2"),
			expected: false,
		},
		"equal-shallow": {
			path:     path.RootPath("test"),
			other:    path.RootPath("test"),
			expected: true,
		},
		"equal-deep": {
			path:     path.RootPath("test1").AtListIndex(0).AtName("test2"),
			other:    path.RootPath("test1").AtListIndex(0).AtName("test2"),
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.path.Equal(testCase.other)

			if got != testCase.expected {
				t.Errorf("expected %t, got %t", testCase.expected, got)
			}
		})
	}
}

func TestPathParentPath(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		expected path.Path
	}{
		"empty": {
			path:     path.EmptyPath(),
			expected: path.EmptyPath(),
		},
		"one": {
			path:     path.RootPath("test"),
			expected: path.EmptyPath(),
		},
		"two": {
			path:     path.RootPath("test").AtListIndex(1),
			expected: path.RootPath("test"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.path.ParentPath()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestPathSteps(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		expected path.PathSteps
	}{
		"empty": {
			path:     path.EmptyPath(),
			expected: path.PathSteps{},
		},
		"one": {
			path: path.RootPath("test"),
			expected: path.PathSteps{
				path.PathStepAttributeName("test"),
			},
		},
		"two": {
			path: path.RootPath("test").AtListIndex(1),
			expected: path.PathSteps{
				path.PathStepAttributeName("test"),
				path.PathStepElementKeyInt(1),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.path.Steps()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestPathString(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		path     path.Path
		expected string
	}{
		"empty": {
			path:     path.EmptyPath(),
			expected: "",
		},
		"AttributeName": {
			path:     path.RootPath("test"),
			expected: `test`,
		},
		"AttributeName-AttributeName": {
			path:     path.RootPath("test1").AtName("test2"),
			expected: `test1.test2`,
		},
		"AttributeName-AttributeName-AttributeName": {
			path:     path.RootPath("test1").AtName("test2").AtName("test3"),
			expected: `test1.test2.test3`,
		},
		"AttributeName-ElementKeyInt": {
			path:     path.RootPath("test").AtListIndex(0),
			expected: `test[0]`,
		},
		"AttributeName-ElementKeyInt-AttributeName": {
			path:     path.RootPath("test1").AtListIndex(0).AtName("test2"),
			expected: `test1[0].test2`,
		},
		"AttributeName-ElementKeyInt-ElementKeyInt": {
			path:     path.RootPath("test").AtListIndex(0).AtListIndex(1),
			expected: `test[0][1]`,
		},
		"AttributeName-ElementKeyString": {
			path:     path.RootPath("test").AtMapKey("test-key"),
			expected: `test["test-key"]`,
		},
		"AttributeName-ElementKeyString-AttributeName": {
			path:     path.RootPath("test1").AtMapKey("test-key").AtName("test2"),
			expected: `test1["test-key"].test2`,
		},
		"AttributeName-ElementKeyString-ElementKeyString": {
			path:     path.RootPath("test").AtMapKey("test-key1").AtMapKey("test-key2"),
			expected: `test["test-key1"]["test-key2"]`,
		},
		"AttributeName-ElementKeyValue": {
			path:     path.RootPath("test").AtSetValue(types.String{Value: "test-value"}),
			expected: `test[Value("test-value")]`,
		},
		"AttributeName-ElementKeyValue-AttributeName": {
			path: path.RootPath("test").AtSetValue(types.Object{
				Attrs: map[string]attr.Value{
					"test_attr_1": types.Bool{Value: true},
					"test_attr_2": types.String{Value: "test-value"},
				},
				AttrTypes: map[string]attr.Type{
					"test_attr_1": types.BoolType,
					"test_attr_2": types.StringType,
				},
			}).AtName("test_attr_1"),
			expected: `test[Value({"test_attr_1":true,"test_attr_2":"test-value"})].test_attr_1`,
		},
		"ElementKeyInt": {
			path:     path.EmptyPath().AtListIndex(0),
			expected: `[0]`,
		},
		"ElementKeyString": {
			path:     path.EmptyPath().AtMapKey("test"),
			expected: `["test"]`,
		},
		"ElementKeyValue": {
			path:     path.EmptyPath().AtSetValue(types.String{Value: "test"}),
			expected: `[Value("test")]`,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.path.String()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
