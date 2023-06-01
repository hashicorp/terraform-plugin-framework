// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
)

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

func TestGetStructTags_success(t *testing.T) {
	t.Parallel()

	type testStruct struct {
		ExportedAndTagged   string `tfsdk:"exported_and_tagged"`
		unexported          string //nolint:structcheck,unused
		unexportedAndTagged string `tfsdk:"unexported_and_tagged"`
		ExportedAndExcluded string `tfsdk:"-"`
	}

	res, err := getStructTags(context.Background(), reflect.ValueOf(testStruct{}), path.Empty())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if len(res) != 1 {
		t.Errorf("Unexpected result: %v", res)
	}
	if res["exported_and_tagged"] != 0 {
		t.Errorf("Unexpected result: %v", res)
	}
}

func TestGetStructTags_untagged(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		ExportedAndUntagged string
	}
	_, err := getStructTags(context.Background(), reflect.ValueOf(testStruct{}), path.Empty())
	if err == nil {
		t.Error("Expected error, got nil")
	}
	expected := `: need a struct tag for "tfsdk" on ExportedAndUntagged`
	if err.Error() != expected {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestGetStructTags_invalidTag(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		InvalidTag string `tfsdk:"invalidTag"`
	}
	_, err := getStructTags(context.Background(), reflect.ValueOf(testStruct{}), path.Empty())
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	expected := `invalidTag: invalid field name, must only use lowercase letters, underscores, and numbers, and must start with a letter`
	if err.Error() != expected {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestGetStructTags_duplicateTag(t *testing.T) {
	t.Parallel()
	type testStruct struct {
		Field1 string `tfsdk:"my_field"`
		Field2 string `tfsdk:"my_field"`
	}
	_, err := getStructTags(context.Background(), reflect.ValueOf(testStruct{}), path.Empty())
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	expected := `my_field: can't use field name for both Field1 and Field2`
	if err.Error() != expected {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestGetStructTags_notAStruct(t *testing.T) {
	t.Parallel()
	var testStruct string

	_, err := getStructTags(context.Background(), reflect.ValueOf(testStruct), path.Empty())
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
	expected := `: can't get struct tags of string, is not a struct`
	if err.Error() != expected {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
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
