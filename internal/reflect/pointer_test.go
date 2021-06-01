package reflect

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestReflectPointer_notAPointer(t *testing.T) {
	t.Parallel()

	var s string
	_, err := reflectPointer(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), Options{}, tftypes.NewAttributePath())
	if expected := ": can't dereference pointer, not a pointer, is a string (string)"; err.Error() != expected {
		t.Errorf("Expected error to be %q, got %q", expected, err.Error())
	}
}

func TestReflectPointer_nilPointer(t *testing.T) {
	t.Parallel()

	var s *string
	got, err := reflectPointer(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), Options{}, tftypes.NewAttributePath())
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

func TestReflectPointer_simple(t *testing.T) {
	t.Parallel()

	var s string
	got, err := reflectPointer(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(&s), Options{}, tftypes.NewAttributePath())
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

func TestReflectPointer_pointerPointer(t *testing.T) {
	t.Parallel()

	var s *string
	got, err := reflectPointer(context.Background(), tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(&s), Options{}, tftypes.NewAttributePath())
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

func TestPointerSafeZeroValue_zeroPointers(t *testing.T) {
	t.Parallel()

	var s string
	got := pointerSafeZeroValue(context.Background(), reflect.ValueOf(s))
	if got.Interface().(string) != "" {
		t.Errorf("expected \"\", got %v", got.Interface())
	}
}

func TestPointerSafeZeroValue_onePointer(t *testing.T) {
	t.Parallel()

	var s *string
	got := pointerSafeZeroValue(context.Background(), reflect.ValueOf(s))
	if got.Interface() != nil && *(got.Interface().(*string)) != "" {
		t.Errorf("expected \"\", got %v", got.Interface())
	}
}

func TestPointerSafeZeroValue_twoPointers(t *testing.T) {
	t.Parallel()

	var s **string
	got := pointerSafeZeroValue(context.Background(), reflect.ValueOf(s))
	if got.Interface() != nil && **(got.Interface().(**string)) != "" {
		t.Errorf("expected \"\", got %v", got.Interface())
	}
}

func TestPointerSafeZeroValue_threePointers(t *testing.T) {
	t.Parallel()

	var s ***string
	got := pointerSafeZeroValue(context.Background(), reflect.ValueOf(s))
	if got.Interface() != nil && ***(got.Interface().(***string)) != "" {
		t.Errorf("expected \"\", got %v", got.Interface())
	}
}

func TestPointerSafeZeroValue_tenPointers(t *testing.T) {
	t.Parallel()

	var s **********string
	got := pointerSafeZeroValue(context.Background(), reflect.ValueOf(s))
	if got.Interface() != nil && **********(got.Interface().(**********string)) != "" {
		t.Errorf("expected \"\", got %v", got.Interface())
	}
}
