package reflect_test

import (
	"context"
	"reflect"
	"testing"

	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPointer_notAPointer(t *testing.T) {
	t.Parallel()

	var s string
	_, err := refl.Pointer(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if expected := ": can't dereference pointer, not a pointer, is a string (string)"; err.Error() != expected {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestPointer_nilPointer(t *testing.T) {
	t.Parallel()

	var s *string
	got, err := refl.Pointer(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if got.Interface() == nil {
		t.Error("Expected \"hello\", got nil")
	}
	if *(got.Interface().(*string)) != "hello" {
		t.Errorf("Expected \"hello\", got %+v", *(got.Interface().(*string)))
	}
}

func TestPointer_simple(t *testing.T) {
	t.Parallel()

	var s string
	got, err := refl.Pointer(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(&s), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if got.Interface() == nil {
		t.Error("Expected \"hello\", got nil")
	}
	if *(got.Interface().(*string)) != "hello" {
		t.Errorf("Expected \"hello\", got %+v", *(got.Interface().(*string)))
	}
}

func TestPointer_pointerPointer(t *testing.T) {
	t.Parallel()

	var s *string
	got, err := refl.Pointer(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(&s), refl.Options{}, tftypes.NewAttributePath())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	if got.Interface() == nil {
		t.Error("Expected \"hello\", got nil")
	}
	if **(got.Interface().(**string)) != "hello" {
		t.Errorf("Expected \"hello\", got %+v", **(got.Interface().(**string)))
	}
}
