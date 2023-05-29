// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPrimitive_string(t *testing.T) {
	t.Parallel()

	var s string

	result, diags := refl.Primitive(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&s).Elem().Set(result)
	if s != "hello" {
		t.Errorf("Expected %q, got %q", "hello", s)
	}
}

func TestPrimitive_stringAlias(t *testing.T) {
	t.Parallel()

	type testString string
	var s testString

	result, diags := refl.Primitive(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&s).Elem().Set(result)
	if s != "hello" {
		t.Errorf("Expected %q, got %q", "hello", s)
	}
}

func TestPrimitive_bool(t *testing.T) {
	t.Parallel()

	var b bool

	result, diags := refl.Primitive(context.Background(), types.BoolType, tftypes.NewValue(tftypes.Bool, true), reflect.ValueOf(b), path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&b).Elem().Set(result)
	if b != true {
		t.Errorf("Expected %v, got %v", true, b)
	}
}

func TestPrimitive_boolAlias(t *testing.T) {
	t.Parallel()

	type testBool bool
	var b testBool

	result, diags := refl.Primitive(context.Background(), types.BoolType, tftypes.NewValue(tftypes.Bool, true), reflect.ValueOf(b), path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&b).Elem().Set(result)
	if b != true {
		t.Errorf("Expected %v, got %v", true, b)
	}
}

func TestFromString(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		val           string
		typ           attr.Type
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"basic": {
			val:      "mystring",
			typ:      types.StringType,
			expected: types.StringValue("mystring"),
		},
		"WithValidateWarning": {
			val: "mystring",
			typ: testtypes.StringTypeWithValidateWarning{},
			expected: testtypes.String{
				InternalString: types.StringValue("mystring"),
				CreatedBy:      testtypes.StringTypeWithValidateWarning{},
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"WithValidateError": {
			val: "mystring",
			typ: testtypes.StringTypeWithValidateError{},
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
	}

	for name, tc := range cases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := refl.FromString(context.Background(), tc.typ, tc.val, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFromBool(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		val           bool
		typ           attr.Type
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"true": {
			val:      true,
			typ:      types.BoolType,
			expected: types.BoolValue(true),
		},
		"false": {
			val:      false,
			typ:      types.BoolType,
			expected: types.BoolValue(false),
		},
		"WithValidateWarning": {
			val: true,
			typ: testtypes.BoolTypeWithValidateWarning{},
			expected: testtypes.Bool{
				Bool:      types.BoolValue(true),
				CreatedBy: testtypes.BoolTypeWithValidateWarning{},
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"WithValidateError": {
			val: true,
			typ: testtypes.BoolTypeWithValidateError{},
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
	}

	for name, tc := range cases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := refl.FromBool(context.Background(), tc.typ, tc.val, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}
