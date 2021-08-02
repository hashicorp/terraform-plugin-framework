package reflect_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
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

func TestFromPointer_AttrTypeWithValidate_Error(t *testing.T) {
	v := "hello, world"
	_, diags := refl.FromPointer(context.Background(), testtypes.StringTypeWithValidateError{}, reflect.ValueOf(&v), tftypes.NewAttributePath())
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}
	if !reflect.DeepEqual(diags[0], testtypes.TestErrorDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagString(testtypes.TestErrorDiagnostic), diagString(diags[0]))
	}
}

func TestFromPointer_AttrTypeWithValidate_Warning(t *testing.T) {
	expectedVal := types.String{
		Value: "hello, world",
	}
	v := "hello, world"
	actualVal, diags := refl.FromPointer(context.Background(), testtypes.StringTypeWithValidateWarning{}, reflect.ValueOf(&v), tftypes.NewAttributePath())
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics, got none")
	}
	if !reflect.DeepEqual(diags[0], testtypes.TestWarningDiagnostic) {
		t.Fatalf("expected diagnostic:\n\n%s\n\ngot diagnostic:\n\n%s\n\n", diagString(testtypes.TestWarningDiagnostic), diagString(diags[0]))
	}
	if !expectedVal.Equal(actualVal) {
		t.Fatalf("unexpected value: got %+v, wanted %+v", actualVal, expectedVal)
	}
}
