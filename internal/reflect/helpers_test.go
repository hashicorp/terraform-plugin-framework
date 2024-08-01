// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

type ExampleStruct struct {
	StrField  string `tfsdk:"str_field"`
	IntField  int    `tfsdk:"int_field"`
	BoolField bool   `tfsdk:"bool_field"`
	IgnoreMe  string `tfsdk:"-"`

	unexported          string //nolint:structcheck,unused
	unexportedAndTagged string `tfsdk:"unexported_and_tagged"`
}

type NestedEmbed struct {
	ListField []string `tfsdk:"list_field"`
	DoubleNestedEmbed
}

type DoubleNestedEmbed struct {
	Map map[string]string `tfsdk:"map_field"`
	ExampleStruct
}

type EmbedWithDuplicates struct {
	StrField1 string `tfsdk:"str_field"`
	StrField2 string `tfsdk:"str_field"`
}

type StructWithInvalidTag struct {
	InvalidField string `tfsdk:"*()-"`
}

func TestGetStructTags(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		in           any
		expectedTags map[string][]int
		expectedErr  error
	}{
		"struct": {
			in: ExampleStruct{},
			expectedTags: map[string][]int{
				"str_field":  {0},
				"int_field":  {1},
				"bool_field": {2},
			},
		},
		"struct-err-duplicate-fields": {
			in: struct {
				StrField string `tfsdk:"str_field"`
				IntField string `tfsdk:"str_field"`
			}{},
			expectedErr: errors.New(`str_field: can't use tfsdk tag "str_field" for both StrField and IntField fields`),
		},
		"embedded-struct-err-duplicate-fields": {
			in: struct {
				EmbedWithDuplicates
			}{},
			expectedErr: errors.New(`error retrieving embedded struct "EmbedWithDuplicates" field tags: str_field: can't use tfsdk tag "str_field" for both StrField1 and StrField2 fields`),
		},
		"embedded-struct-err-duplicate-fields-from-promote": {
			in: struct {
				StrField      string `tfsdk:"str_field"`
				ExampleStruct        // Contains a `tfsdk:"str_field"`
			}{},
			expectedErr: errors.New(`embedded struct "ExampleStruct" promotes a field with a duplicate tfsdk tag "str_field", conflicts with "StrField" tfsdk tag`),
		},
		"struct-err-invalid-field": {
			in:          StructWithInvalidTag{},
			expectedErr: errors.New(`*()-: invalid tfsdk tag, must only use lowercase letters, underscores, and numbers, and must start with a letter`),
		},
		"struct-err-missing-tfsdk-tag": {
			in: struct {
				ExampleField string
			}{},
			expectedErr: errors.New(`: need a struct tag for "tfsdk" on ExampleField`),
		},
		"struct-err-empty-tfsdk-tag": {
			in: struct {
				ExampleField string `tfsdk:""`
			}{},
			expectedErr: errors.New(`: invalid tfsdk tag, must only use lowercase letters, underscores, and numbers, and must start with a letter`),
		},
		"ignore-embedded-struct": {
			in: struct {
				ExampleStruct `tfsdk:"-"`
				Field5        string `tfsdk:"field5"`
			}{},
			expectedTags: map[string][]int{
				"field5": {1},
			},
		},
		"embedded-struct": {
			in: struct {
				ExampleStruct
				Field5 string `tfsdk:"field5"`
			}{},
			expectedTags: map[string][]int{
				"str_field":  {0, 0},
				"int_field":  {0, 1},
				"bool_field": {0, 2},
				"field5":     {1},
			},
		},
		"nested-embedded-struct": {
			in: struct {
				NestedEmbed
				Field5 string `tfsdk:"field5"`
			}{},
			expectedTags: map[string][]int{
				"list_field": {0, 0},
				"map_field":  {0, 1, 0},
				"str_field":  {0, 1, 1, 0},
				"int_field":  {0, 1, 1, 1},
				"bool_field": {0, 1, 1, 2},
				"field5":     {1},
			},
		},
		"embedded-struct-unexported": {
			in: struct {
				ExampleStruct
				Field5 string `tfsdk:"field5"`

				unexported          string //nolint:structcheck,unused
				unexportedAndTagged string `tfsdk:"unexported_and_tagged"`
			}{},
			expectedTags: map[string][]int{
				"str_field":  {0, 0},
				"int_field":  {0, 1},
				"bool_field": {0, 2},
				"field5":     {1},
			},
		},
		"embedded-struct-err-cannot-have-empty-tfsdk-tag": {
			in: struct {
				ExampleStruct `tfsdk:""` // Can't put a tfsdk tag here
			}{},
			expectedErr: errors.New(`: embedded struct field ExampleStruct cannot have tfsdk tag`),
		},
		"embedded-struct-err-cannot-have-tfsdk-tag": {
			in: struct {
				ExampleStruct `tfsdk:"example_field"` // Can't put a tfsdk tag here
			}{},
			expectedErr: errors.New(`example_field: embedded struct field ExampleStruct cannot have tfsdk tag`),
		},
		"embedded-struct-err-invalid": {
			in: struct {
				StructWithInvalidTag // Contains an invalid "tfsdk" tag
			}{},
			expectedErr: errors.New(`error retrieving embedded struct "StructWithInvalidTag" field tags: *()-: invalid tfsdk tag, must only use lowercase letters, underscores, and numbers, and must start with a letter`),
		},
		// NOTE: The following tests are for embedded struct pointers, despite them not being explicitly supported by the framework reflect package.
		// Embedded struct pointers still produce a valid field index, but are later rejected when retrieving them. These tests just ensure that there
		// are no panics when retrieving the field index for an embedded struct pointer field
		"ignore-embedded-struct-ptr": {
			in: struct {
				*ExampleStruct `tfsdk:"-"`
				Field5         string `tfsdk:"field5"`
			}{},
			expectedTags: map[string][]int{
				"field5": {1},
			},
		},
		"embedded-struct-ptr": {
			in: struct {
				*ExampleStruct
				Field5 string `tfsdk:"field5"`
			}{},
			expectedTags: map[string][]int{
				"str_field":  {0, 0},
				"int_field":  {0, 1},
				"bool_field": {0, 2},
				"field5":     {1},
			},
		},
		"embedded-struct-ptr-unexported": {
			in: struct {
				*ExampleStruct
				Field5 string `tfsdk:"field5"`

				unexported          string //nolint:structcheck,unused
				unexportedAndTagged string `tfsdk:"unexported_and_tagged"`
			}{},
			expectedTags: map[string][]int{
				"str_field":  {0, 0},
				"int_field":  {0, 1},
				"bool_field": {0, 2},
				"field5":     {1},
			},
		},
		"embedded-struct-ptr-err-cannot-have-empty-tfsdk-tag": {
			in: struct {
				*ExampleStruct `tfsdk:""` // Can't put a tfsdk tag here
			}{},
			expectedErr: errors.New(`: embedded struct field ExampleStruct cannot have tfsdk tag`),
		},
		"embedded-struct-ptr-err-cannot-have-tfsdk-tag": {
			in: struct {
				*ExampleStruct `tfsdk:"example_field"` // Can't put a tfsdk tag here
			}{},
			expectedErr: errors.New(`example_field: embedded struct field ExampleStruct cannot have tfsdk tag`),
		},
		"embedded-struct-ptr-err-duplicate-fields": {
			in: struct {
				*EmbedWithDuplicates
			}{},
			expectedErr: errors.New(`error retrieving embedded struct "EmbedWithDuplicates" field tags: str_field: can't use tfsdk tag "str_field" for both StrField1 and StrField2 fields`),
		},
		"embedded-struct-ptr-err-duplicate-fields-from-promote": {
			in: struct {
				StrField       string `tfsdk:"str_field"`
				*ExampleStruct        // Contains a `tfsdk:"str_field"`
			}{},
			expectedErr: errors.New(`embedded struct "ExampleStruct" promotes a field with a duplicate tfsdk tag "str_field", conflicts with "StrField" tfsdk tag`),
		},
		"embedded-struct-ptr-err-invalid": {
			in: struct {
				*StructWithInvalidTag // Contains an invalid "tfsdk" tag
			}{},
			expectedErr: errors.New(`error retrieving embedded struct "StructWithInvalidTag" field tags: *()-: invalid tfsdk tag, must only use lowercase letters, underscores, and numbers, and must start with a letter`),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tags, err := getStructTags(context.Background(), reflect.TypeOf(testCase.in), path.Empty())
			if err != nil {
				if testCase.expectedErr == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedErr.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedErr, err)
				}
			}

			if err == nil && testCase.expectedErr != nil {
				t.Fatalf("got no error, expected: %s", testCase.expectedErr)
			}

			if diff := cmp.Diff(tags, testCase.expectedTags); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestTrueReflectValue(t *testing.T) {
	t.Parallel()

	var iface, otherIface interface{}
	var stru struct{}

	// test that when nothing needs unwrapped, we get the right answer
	if got := trueReflectValue(reflect.ValueOf(stru)).Kind(); got != reflect.Struct {
		t.Errorf("Expected %s, got %s", reflect.Struct, got)
	}

	// test that we can unwrap pointers
	if got := trueReflectValue(reflect.ValueOf(&stru)).Kind(); got != reflect.Struct {
		t.Errorf("Expected %s, got %s", reflect.Struct, got)
	}

	// test that we can unwrap interfaces
	iface = stru
	if got := trueReflectValue(reflect.ValueOf(iface)).Kind(); got != reflect.Struct {
		t.Errorf("Expected %s, got %s", reflect.Struct, got)
	}

	// test that we can unwrap pointers inside interfaces, and pointers to
	// interfaces with pointers inside them
	iface = &stru
	if got := trueReflectValue(reflect.ValueOf(iface)).Kind(); got != reflect.Struct {
		t.Errorf("Expected %s, got %s", reflect.Struct, got)
	}
	if got := trueReflectValue(reflect.ValueOf(&iface)).Kind(); got != reflect.Struct {
		t.Errorf("Expected %s, got %s", reflect.Struct, got)
	}

	// test that we can unwrap pointers to interfaces inside other
	// interfaces, and pointers to interfaces inside pointers to
	// interfaces.
	otherIface = &iface
	if got := trueReflectValue(reflect.ValueOf(otherIface)).Kind(); got != reflect.Struct {
		t.Errorf("Expected %s, got %s", reflect.Struct, got)
	}
	if got := trueReflectValue(reflect.ValueOf(&otherIface)).Kind(); got != reflect.Struct {
		t.Errorf("Expected %s, got %s", reflect.Struct, got)
	}
}

func TestCommaSeparatedString(t *testing.T) {
	t.Parallel()
	type testCase struct {
		input    []string
		expected string
	}
	tests := map[string]testCase{
		"empty": {
			input:    []string{},
			expected: "",
		},
		"oneWord": {
			input:    []string{"red"},
			expected: "red",
		},
		"twoWords": {
			input:    []string{"red", "blue"},
			expected: "red and blue",
		},
		"threeWords": {
			input:    []string{"red", "blue", "green"},
			expected: "red, blue, and green",
		},
		"fourWords": {
			input:    []string{"red", "blue", "green", "purple"},
			expected: "red, blue, green, and purple",
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := commaSeparatedString(test.input)
			if got != test.expected {
				t.Errorf("Expected %q, got %q", test.expected, got)
			}
		})
	}
}

func TestIsValidFieldName(t *testing.T) {
	t.Parallel()
	tests := map[string]bool{
		"":    false,
		"a":   true,
		"1":   false,
		"1a":  false,
		"a1":  true,
		"A":   false,
		"a-b": false,
		"a_b": true,
	}
	for in, expected := range tests {
		in, expected := in, expected
		t.Run(fmt.Sprintf("input=%q", in), func(t *testing.T) {
			t.Parallel()

			result := isValidFieldName(in)
			if result != expected {
				t.Errorf("Expected %v, got %v", expected, result)
			}
		})
	}
}

func TestCanBeNil_struct(t *testing.T) {
	t.Parallel()

	var stru struct{}

	got := canBeNil(reflect.ValueOf(stru))
	if got {
		t.Error("Expected structs to not be nillable, but canBeNil said they were")
	}
}

func TestCanBeNil_structPointer(t *testing.T) {
	t.Parallel()

	var stru struct{}
	struPtr := &stru

	got := canBeNil(reflect.ValueOf(struPtr))
	if !got {
		t.Error("Expected pointers to structs to be nillable, but canBeNil said they weren't")
	}
}

func TestCanBeNil_slice(t *testing.T) {
	t.Parallel()

	slice := []string{}
	got := canBeNil(reflect.ValueOf(slice))
	if !got {
		t.Errorf("Expected slices to be nillable, but canBeNil said they weren't")
	}
}

func TestCanBeNil_map(t *testing.T) {
	t.Parallel()

	m := map[string]string{}
	got := canBeNil(reflect.ValueOf(m))
	if !got {
		t.Errorf("Expected maps to be nillable, but canBeNil said they weren't")
	}
}

func TestCanBeNil_interface(t *testing.T) {
	t.Parallel()

	type myStruct struct {
		Value interface{}
	}

	var s myStruct
	got := canBeNil(reflect.ValueOf(s).FieldByName("Value"))
	if !got {
		t.Errorf("Expected interfaces to be nillable, but canBeNil said they weren't")
	}
}
