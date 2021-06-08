package reflect

import (
	"context"
	"reflect"
	"testing"
)

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
