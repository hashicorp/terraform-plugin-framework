package types

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestBoolValueFromTerraform(t *testing.T) {
	t.Parallel()

	testBoolValueFromTerraform(t, true)
}

func testBoolValueFromTerraform(t *testing.T, direct bool) {
	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"true": {
			input:       tftypes.NewValue(tftypes.Bool, true),
			expectation: &Bool{Value: true},
		},
		"false": {
			input:       tftypes.NewValue(tftypes.Bool, false),
			expectation: &Bool{Value: false},
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			expectation: &Bool{Unknown: true},
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Bool, nil),
			expectation: &Bool{Null: true},
		},
		"wrongType": {
			input:       tftypes.NewValue(tftypes.String, "oops"),
			expectedErr: "can't unmarshal tftypes.String into *bool, expected boolean",
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			f := BoolType.ValueFromTerraform
			if direct {
				f = boolValueFromTerraform
			}
			got, err := f(ctx, test.input)
			if err != nil {
				if test.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				if test.expectedErr != err.Error() {
					t.Errorf("Expected error to be %q, got %q", test.expectedErr, err.Error())
					return
				}
				// we have an error, and it matches our
				// expectations, we're good
				return
			}
			if err == nil && test.expectedErr != "" {
				t.Errorf("Expected error to be %q, didn't get an error", test.expectedErr)
				return
			}
			if !got.Equal(test.expectation) {
				t.Errorf("Expected %+v, got %+v", test.expectation, got)
			}
		})
	}
}

func TestBoolToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       *Bool
		expectation interface{}
	}
	tests := map[string]testCase{
		"true": {
			input:       &Bool{Value: true},
			expectation: true,
		},
		"false": {
			input:       &Bool{Value: false},
			expectation: false,
		},
		"unknown": {
			input:       &Bool{Unknown: true},
			expectation: tftypes.UnknownValue,
		},
		"null": {
			input:       &Bool{Null: true},
			expectation: nil,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			got, err := test.input.ToTerraformValue(ctx)
			if err != nil {
				t.Errorf("Unexpected error: %s", err)
				return
			}
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %+v, got %+v", test.expectation, got)
			}
		})
	}
}

func TestBoolEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       *Bool
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"true-true": {
			input:       &Bool{Value: true},
			candidate:   &Bool{Value: true},
			expectation: true,
		},
		"true-false": {
			input:       &Bool{Value: true},
			candidate:   &Bool{Value: false},
			expectation: false,
		},
		"true-unknown": {
			input:       &Bool{Value: true},
			candidate:   &Bool{Unknown: true},
			expectation: false,
		},
		"true-null": {
			input:       &Bool{Value: true},
			candidate:   &Bool{Null: true},
			expectation: false,
		},
		"true-wrongType": {
			input:       &Bool{Value: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"true-nil": {
			input:       &Bool{Value: true},
			candidate:   nil,
			expectation: false,
		},
		"false-true": {
			input:       &Bool{Value: false},
			candidate:   &Bool{Value: true},
			expectation: false,
		},
		"false-false": {
			input:       &Bool{Value: false},
			candidate:   &Bool{Value: false},
			expectation: true,
		},
		"false-unknown": {
			input:       &Bool{Value: false},
			candidate:   &Bool{Unknown: true},
			expectation: false,
		},
		"false-null": {
			input:       &Bool{Value: false},
			candidate:   &Bool{Null: true},
			expectation: false,
		},
		"false-wrongType": {
			input:       &Bool{Value: false},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"false-nil": {
			input:       &Bool{Value: false},
			candidate:   nil,
			expectation: false,
		},
		"unknown-true": {
			input:       &Bool{Unknown: true},
			candidate:   &Bool{Value: true},
			expectation: false,
		},
		"unknown-false": {
			input:       &Bool{Unknown: true},
			candidate:   &Bool{Value: false},
			expectation: false,
		},
		"unknown-unknown": {
			input:       &Bool{Unknown: true},
			candidate:   &Bool{Unknown: true},
			expectation: true,
		},
		"unknown-null": {
			input:       &Bool{Unknown: true},
			candidate:   &Bool{Null: true},
			expectation: false,
		},
		"unknown-wrongType": {
			input:       &Bool{Unknown: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"unknown-nil": {
			input:       &Bool{Unknown: true},
			candidate:   nil,
			expectation: false,
		},
		"null-true": {
			input:       &Bool{Null: true},
			candidate:   &Bool{Value: true},
			expectation: false,
		},
		"null-false": {
			input:       &Bool{Null: true},
			candidate:   &Bool{Value: false},
			expectation: false,
		},
		"null-unknown": {
			input:       &Bool{Null: true},
			candidate:   &Bool{Unknown: true},
			expectation: false,
		},
		"null-null": {
			input:       &Bool{Null: true},
			candidate:   &Bool{Null: true},
			expectation: true,
		},
		"null-wrongType": {
			input:       &Bool{Null: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"null-nil": {
			input:       &Bool{Null: true},
			candidate:   nil,
			expectation: false,
		},
		"nil-true": {
			input:       nil,
			candidate:   &Bool{Value: true},
			expectation: false,
		},
		"nil-false": {
			input:       nil,
			candidate:   &Bool{Value: false},
			expectation: false,
		},
		"nil-unknown": {
			input:       nil,
			candidate:   &Bool{Unknown: true},
			expectation: false,
		},
		"nil-wrongType": {
			input:       nil,
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"nil-nil": {
			input:       nil,
			candidate:   nil,
			expectation: true,
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.Equal(test.candidate)
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %v, got %v", test.expectation, got)
			}
		})
	}
}

func TestBoolSetTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		start       *Bool
		input       tftypes.Value
		expectation *Bool
		expectedErr string
	}
	tests := map[string]testCase{
		"true-true": {
			start:       &Bool{Value: true},
			input:       tftypes.NewValue(tftypes.Bool, true),
			expectation: &Bool{Value: true},
		},
		"true-false": {
			start:       &Bool{Value: true},
			input:       tftypes.NewValue(tftypes.Bool, false),
			expectation: &Bool{Value: false},
		},
		"true-unknown": {
			start:       &Bool{Value: true},
			input:       tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			expectation: &Bool{Unknown: true},
		},
		"true-null": {
			start:       &Bool{Value: true},
			input:       tftypes.NewValue(tftypes.Bool, nil),
			expectation: &Bool{Null: true},
		},
		"true-wrongType": {
			start:       &Bool{Value: true},
			input:       tftypes.NewValue(tftypes.String, "oops"),
			expectedErr: "can't unmarshal tftypes.String into *bool, expected boolean",
		},
		"false-true": {
			start:       &Bool{Value: false},
			input:       tftypes.NewValue(tftypes.Bool, true),
			expectation: &Bool{Value: true},
		},
		"false-false": {
			start:       &Bool{Value: false},
			input:       tftypes.NewValue(tftypes.Bool, false),
			expectation: &Bool{Value: false},
		},
		"false-unknown": {
			start:       &Bool{Value: false},
			input:       tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			expectation: &Bool{Unknown: true},
		},
		"false-null": {
			start:       &Bool{Value: false},
			input:       tftypes.NewValue(tftypes.Bool, nil),
			expectation: &Bool{Null: true},
		},
		"false-wrongType": {
			start:       &Bool{Value: false},
			input:       tftypes.NewValue(tftypes.String, "oops"),
			expectedErr: "can't unmarshal tftypes.String into *bool, expected boolean",
		},
		"unknown-true": {
			start:       &Bool{Unknown: true},
			input:       tftypes.NewValue(tftypes.Bool, true),
			expectation: &Bool{Value: true},
		},
		"unknown-false": {
			start:       &Bool{Unknown: true},
			input:       tftypes.NewValue(tftypes.Bool, false),
			expectation: &Bool{Value: false},
		},
		"unknown-unknown": {
			start:       &Bool{Unknown: true},
			input:       tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			expectation: &Bool{Unknown: true},
		},
		"unknown-null": {
			start:       &Bool{Unknown: true},
			input:       tftypes.NewValue(tftypes.Bool, nil),
			expectation: &Bool{Null: true},
		},
		"unknown-wrongType": {
			start:       &Bool{Unknown: true},
			input:       tftypes.NewValue(tftypes.String, "oops"),
			expectedErr: "can't unmarshal tftypes.String into *bool, expected boolean",
		},
		"null-true": {
			start:       &Bool{Null: true},
			input:       tftypes.NewValue(tftypes.Bool, true),
			expectation: &Bool{Value: true},
		},
		"null-false": {
			start:       &Bool{Null: true},
			input:       tftypes.NewValue(tftypes.Bool, false),
			expectation: &Bool{Value: false},
		},
		"null-unknown": {
			start:       &Bool{Null: true},
			input:       tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			expectation: &Bool{Unknown: true},
		},
		"null-null": {
			start:       &Bool{Null: true},
			input:       tftypes.NewValue(tftypes.Bool, nil),
			expectation: &Bool{Null: true},
		},
		"null-wrongType": {
			start:       &Bool{Null: true},
			input:       tftypes.NewValue(tftypes.String, "oops"),
			expectedErr: "can't unmarshal tftypes.String into *bool, expected boolean",
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			err := test.start.SetTerraformValue(ctx, test.input)
			if err != nil {
				if test.expectedErr == "" {
					t.Errorf("Unexpected error: %s", err)
					return
				}
				if test.expectedErr != err.Error() {
					t.Errorf("Expected error to be %q, got %q", test.expectedErr, err.Error())
					return
				}
				// we have an error, and it matches our
				// expectations, we're good
				return
			}
			if err == nil && test.expectedErr != "" {
				t.Errorf("Expected error to be %q, didn't get an error", test.expectedErr)
				return
			}
			if !test.start.Equal(test.expectation) {
				t.Errorf("Expected %+v, got %+v", test.expectation, test.start)
			}
		})
	}
}
