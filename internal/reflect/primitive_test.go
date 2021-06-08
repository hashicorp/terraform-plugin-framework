package reflect_test

import (
	"context"
	"reflect"
	"testing"

	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPrimitive_string(t *testing.T) {
	t.Parallel()

	var s string

	result, err := refl.Primitive(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
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

	result, err := refl.Primitive(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	reflect.ValueOf(&s).Elem().Set(result)
	if s != "hello" {
		t.Errorf("Expected %q, got %q", "hello", s)
	}
}

func TestPrimitive_bool(t *testing.T) {
	t.Parallel()

	var b bool

	result, err := refl.Primitive(context.Background(), types.BoolType, tftypes.NewValue(tftypes.Bool, true), reflect.ValueOf(b), tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
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

	result, err := refl.Primitive(context.Background(), types.BoolType, tftypes.NewValue(tftypes.Bool, true), reflect.ValueOf(b), tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	reflect.ValueOf(&b).Elem().Set(result)
	if b != true {
		t.Errorf("Expected %v, got %v", true, b)
	}
}
