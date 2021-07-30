package reflect_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/types"

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

func TestFromBool(t *testing.T) {
	// the rare exhaustive test
	cases := []struct {
		val      bool
		expected attr.Value
	}{
		{
			true,
			types.Bool{
				Value: true,
			},
		},
		{
			false,
			types.Bool{
				Value: false,
			},
		},
	}

	for _, tc := range cases {
		actualVal, diags := refl.FromBool(context.Background(), types.BoolType, tc.val, tftypes.NewAttributePath())
		if diagsHasErrors(diags) {
			t.Fatalf("Unexpected error: %s", diagsString(diags))
		}

		if !tc.expected.Equal(actualVal) {
			t.Fatalf("fail: got %+v, wanted %+v", actualVal, tc.expected)
		}
	}
}
