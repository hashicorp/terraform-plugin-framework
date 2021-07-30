package reflect_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPointer_notAPointer(t *testing.T) {
	t.Parallel()

	var s string
	_, diags := refl.Pointer(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if expected := "cannot dereference pointer, not a pointer, is a string (string)"; !diagsContainsDetail(diags, expected) {
		t.Errorf("Expected error to be %q, got %s", expected, diagsString(diags))
	}
}

func TestPointer_nilPointer(t *testing.T) {
	t.Parallel()

	var s *string
	got, diags := refl.Pointer(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), refl.Options{}, tftypes.NewAttributePath())
	if diagsHasErrors(diags) {
		t.Errorf("Unexpected error: %s", diagsString(diags))
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
	got, diags := refl.Pointer(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(&s), refl.Options{}, tftypes.NewAttributePath())
	if diagsHasErrors(diags) {
		t.Errorf("Unexpected error: %s", diagsString(diags))
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
	got, diags := refl.Pointer(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(&s), refl.Options{}, tftypes.NewAttributePath())
	if diagsHasErrors(diags) {
		t.Errorf("Unexpected error: %s", diagsString(diags))
	}
	if got.Interface() == nil {
		t.Error("Expected \"hello\", got nil")
	}
	if **(got.Interface().(**string)) != "hello" {
		t.Errorf("Expected \"hello\", got %+v", **(got.Interface().(**string)))
	}
}

func TestFromPointer_simple(t *testing.T) {
	t.Parallel()

	v := "hello, world"
	got, diags := refl.FromPointer(context.Background(), types.StringType, reflect.ValueOf(&v), tftypes.NewAttributePath())
	if diagsHasErrors(diags) {
		t.Errorf("unexpected error: %s", diagsString(diags))
	}
	expected := types.String{
		Value: "hello, world",
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestFromPointer_null(t *testing.T) {
	t.Parallel()

	var v *string
	got, diags := refl.FromPointer(context.Background(), types.StringType, reflect.ValueOf(v), tftypes.NewAttributePath())
	if diagsHasErrors(diags) {
		t.Errorf("unexpected error: %s", diagsString(diags))
	}
	expected := types.String{
		Null: true,
	}
	if diff := cmp.Diff(expected, got); diff != "" {
		t.Errorf("Unexpected diff (+wanted, -got): %s", diff)
	}
}
