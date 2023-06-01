// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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

	gotStr, ok := got.Interface().(string)
	if !ok {
		t.Fatalf("expected type of string, got %T", got.Interface())
	}

	if gotStr != "" {
		t.Errorf("expected \"\", got %v", got.Interface())
	}
}

func TestPointerSafeZeroValue_onePointer(t *testing.T) {
	t.Parallel()

	var s *string
	got := pointerSafeZeroValue(context.Background(), reflect.ValueOf(s))

	gotStr, ok := got.Interface().(*string)
	if !ok {
		t.Fatalf("expected type of *string, got %T", got.Interface())
	}

	if got.Interface() != nil && *(gotStr) != "" {
		t.Errorf("expected \"\", got %v", got.Interface())
	}
}

func TestPointerSafeZeroValue_twoPointers(t *testing.T) {
	t.Parallel()

	var s **string
	got := pointerSafeZeroValue(context.Background(), reflect.ValueOf(s))

	gotStr, ok := got.Interface().(**string)
	if !ok {
		t.Fatalf("expected type of **string, got %T", got.Interface())
	}

	if got.Interface() != nil && **(gotStr) != "" {
		t.Errorf("expected \"\", got %v", got.Interface())
	}
}

func TestPointerSafeZeroValue_threePointers(t *testing.T) {
	t.Parallel()

	var s ***string
	got := pointerSafeZeroValue(context.Background(), reflect.ValueOf(s))

	gotStr, ok := got.Interface().(***string)
	if !ok {
		t.Fatalf("expected type of ***string, got %T", got.Interface())
	}

	if got.Interface() != nil && ***(gotStr) != "" {
		t.Errorf("expected \"\", got %v", got.Interface())
	}
}

func TestPointerSafeZeroValue_tenPointers(t *testing.T) {
	t.Parallel()

	var s **********string
	got := pointerSafeZeroValue(context.Background(), reflect.ValueOf(s))

	gotStr, ok := got.Interface().(**********string)
	if !ok {
		t.Fatalf("expected type of **********string, got %T", got.Interface())
	}

	if got.Interface() != nil && **********(gotStr) != "" {
		t.Errorf("expected \"\", got %v", got.Interface())
	}
}
