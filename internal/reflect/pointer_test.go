// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	refl "github.com/hashicorp/terraform-plugin-framework/internal/reflect"
	testtypes "github.com/hashicorp/terraform-plugin-framework/internal/testing/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestPointer_notAPointer(t *testing.T) {
	t.Parallel()

	var s string
	expectedDiags := diag.Diagnostics{
		diag.WithPath(path.Empty(), refl.DiagIntoIncompatibleType{
			Val:        tftypes.NewValue(tftypes.String, "hello"),
			TargetType: reflect.TypeOf(s),
			Err:        fmt.Errorf("cannot dereference pointer, not a pointer, is a %s (%s)", reflect.TypeOf(s), reflect.TypeOf(s).Kind()),
		}),
	}

	_, diags := refl.Pointer(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestPointer_nilPointer(t *testing.T) {
	t.Parallel()

	var s *string
	got, diags := refl.Pointer(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(s), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if got.Interface() == nil {
		t.Error("Expected \"hello\", got nil")
	}
	gotStr, ok := got.Interface().(*string)
	if !ok {
		t.Fatalf("Expected type of *string, got %T", got.Interface())
	}
	if *(gotStr) != "hello" {
		t.Errorf("Expected \"hello\", got %+v", *(gotStr))
	}
}

func TestPointer_simple(t *testing.T) {
	t.Parallel()

	var s string
	got, diags := refl.Pointer(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(&s), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if got.Interface() == nil {
		t.Error("Expected \"hello\", got nil")
	}
	gotStr, ok := got.Interface().(*string)
	if !ok {
		t.Fatalf("Expected type of *string, got %T", got.Interface())
	}
	if *(gotStr) != "hello" {
		t.Errorf("Expected \"hello\", got %+v", *(gotStr))
	}
}

func TestPointer_pointerPointer(t *testing.T) {
	t.Parallel()

	var s *string
	got, diags := refl.Pointer(context.Background(), types.StringType, tftypes.NewValue(tftypes.String, "hello"), reflect.ValueOf(&s), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	if got.Interface() == nil {
		t.Error("Expected \"hello\", got nil")
	}
	gotStr, ok := got.Interface().(**string)
	if !ok {
		t.Fatalf("Expected type of **string, got %T", got.Interface())
	}
	if **(gotStr) != "hello" {
		t.Errorf("Expected \"hello\", got %+v", **(gotStr))
	}
}

func TestFromPointer(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		typ           attr.Type
		val           reflect.Value
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"simple": {
			typ:      types.StringType,
			val:      reflect.ValueOf(strPtr("hello, world")),
			expected: types.StringValue("hello, world"),
		},
		"null": {
			typ:      types.StringType,
			val:      reflect.ValueOf(new(*string)),
			expected: types.StringNull(),
		},
		"WithValidateError": {
			typ: testtypes.StringTypeWithValidateError{},
			val: reflect.ValueOf(strPtr("hello, world")),
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
		"WithValidateWarning": {
			typ: testtypes.StringTypeWithValidateWarning{},
			val: reflect.ValueOf(strPtr("hello, world")),
			expected: testtypes.String{
				InternalString: types.StringValue("hello, world"),
				CreatedBy:      testtypes.StringTypeWithValidateWarning{},
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
	}

	for name, tc := range testCases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, diags := refl.FromPointer(context.Background(), tc.typ, tc.val, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if diff := cmp.Diff(got, tc.expected); diff != "" {
				t.Errorf("unexpected result (+wanted, -got): %s", diff)
			}
		})
	}
}

func strPtr(s string) *string {
	return &s
}
