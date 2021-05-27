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

	err := reflectPrimitive(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), tftypes.NewAttributePath())
	if err == nil {
		t.Error("Expected error, didn't get one")
	}
	if expected := ": can't set string"; expected != err.Error() {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestReflectPrimitive_stringPointer(t *testing.T) {
	t.Parallel()

	var s string

	err := reflectPrimitive(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(&s), tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if s != "hello" {
		t.Errorf("Expected %q, got %q", "hello", s)
	}
}
