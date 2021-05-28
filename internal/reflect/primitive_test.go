package reflect

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestReflectPrimitive_string(t *testing.T) {
	t.Parallel()

	var s string

	result, err := reflectPrimitive(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	reflect.ValueOf(&s).Elem().Set(result)
	if s != "hello" {
		t.Errorf("Expected %q, got %q", "hello", s)
	}
}

func TestReflectPrimitive_bool(t *testing.T) {
	t.Parallel()

	var b bool

	result, err := reflectPrimitive(context.Background(), tftypes.NewValue(tftypes.Bool, true), reflect.ValueOf(b), tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	reflect.ValueOf(&b).Elem().Set(result)
	if b != true {
		t.Errorf("Expected %v, got %v", true, b)
	}
}
