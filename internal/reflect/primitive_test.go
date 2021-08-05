package reflect_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPrimitive_string(t *testing.T) {
	t.Parallel()

	var s string

	result, diags := refl.Primitive(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), tftypes.NewAttributePath())
	if diagsHasErrors(diags) {
		t.Errorf("Unexpected error: %s", diagsString(diags))
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

	result, diags := refl.Primitive(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), tftypes.NewAttributePath())
	if diagsHasErrors(diags) {
		t.Errorf("Unexpected error: %s", diagsString(diags))
	}
	reflect.ValueOf(&s).Elem().Set(result)
	if s != "hello" {
		t.Errorf("Expected %q, got %q", "hello", s)
	}
}

func TestPrimitive_bool(t *testing.T) {
	t.Parallel()

	var b bool

	result, diags := refl.Primitive(context.Background(), types.BoolType, tftypes.NewValue(tftypes.Bool, true), reflect.ValueOf(b), tftypes.NewAttributePath())
	if diagsHasErrors(diags) {
		t.Errorf("Unexpected error: %s", diagsString(diags))
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

	result, diags := refl.Primitive(context.Background(), types.BoolType, tftypes.NewValue(tftypes.Bool, true), reflect.ValueOf(b), tftypes.NewAttributePath())
	if diagsHasErrors(diags) {
		t.Errorf("Unexpected error: %s", diagsString(diags))
	}
	reflect.ValueOf(&b).Elem().Set(result)
	if b != true {
		t.Errorf("Expected %v, got %v", true, b)
	}
}

func TestFromString(t *testing.T) {
	expectedVal := types.String{
		Value: "mystring",
	}
	actualVal, diags := refl.FromString(context.Background(), types.StringType, "mystring", tftypes.NewAttributePath())
	if diagsHasErrors(diags) {
		t.Fatalf("Unexpected error: %s", diagsString(diags))
	}
	if !expectedVal.Equal(actualVal) {
		t.Fatalf("fail: got %+v, wanted %+v", actualVal, expectedVal)
	}
}

func TestFromString_AttrTypeWithValidate_Error(t *testing.T) {
	_, diags := refl.FromString(context.Background(), testtypes.StringTypeWithValidateError{}, "mystring", tftypes.NewAttributePath())
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}
	if !cmp.Equal(diags[0], testtypes.TestErrorDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagString(testtypes.TestErrorDiagnostic), diagString(diags[0]))
	}
}

func TestFromString_AttrTypeWithValidate_Warning(t *testing.T) {
	expectedVal := types.String{
		Value: "mystring",
	}
	actualVal, diags := refl.FromString(context.Background(), testtypes.StringTypeWithValidateWarning{}, "mystring", tftypes.NewAttributePath())
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}
	if !cmp.Equal(diags[0], testtypes.TestWarningDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagString(testtypes.TestWarningDiagnostic), diagString(diags[0]))
	}
	if !expectedVal.Equal(actualVal) {
		t.Fatalf("unexpected value: got %+v, wanted %+v", actualVal, expectedVal)
	}
}

func TestFromBool(t *testing.T) {
	// the rare exhaustive test
	cases := map[string]struct {
		val          bool
		typ          attr.Type
		expected     attr.Value
		expectedDiag *tfprotov6.Diagnostic
	}{
		"true": {
			val: true,
			typ: types.BoolType,
			expected: types.Bool{
				Value: true,
			},
		},
		"false": {
			val: false,
			typ: types.BoolType,
			expected: types.Bool{
				Value: false,
			},
		},
		"WithValidateWarning": {
			val: true,
			typ: testtypes.BoolTypeWithValidateWarning{},
			expected: types.Bool{
				Value: true,
			},
			expectedDiag: testtypes.TestWarningDiagnostic,
		},
		"WithValidateError": {
			val:          true,
			typ:          testtypes.BoolTypeWithValidateError{},
			expectedDiag: testtypes.TestErrorDiagnostic,
		},
	}

	for _, tc := range cases {
		actualVal, diags := refl.FromBool(context.Background(), tc.typ, tc.val, tftypes.NewAttributePath())
		if tc.expectedDiag == nil && diagsHasErrors(diags) {
			t.Fatalf("Unexpected error: %s", diagsString(diags))
		}
		if tc.expectedDiag != nil {
			if len(diags) == 0 {
				t.Fatalf("Expected diagnostic, got none")
			}

			if !cmp.Equal(tc.expectedDiag, diags[0]) {
				t.Fatalf("Expected diagnostic:\n\n%s\n\nGot diagnostic:\n\n%s\n\n", diagString(tc.expectedDiag), diagString(diags[0]))
			}
		}
		if !diagsHasErrors(diags) && !tc.expected.Equal(actualVal) {
			t.Fatalf("fail: got %+v, wanted %+v", actualVal, tc.expected)
		}
	}
}
