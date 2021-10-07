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
			expectation: Bool{Value: true},
		},
		"false": {
			input:       tftypes.NewValue(tftypes.Bool, false),
			expectation: Bool{Value: false},
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Bool, tftypes.UnknownValue),
			expectation: Bool{Unknown: true},
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Bool, nil),
			expectation: Bool{Null: true},
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
		input       Bool
		expectation interface{}
	}
	tests := map[string]testCase{
		"true": {
			input:       Bool{Value: true},
			expectation: true,
		},
		"false": {
			input:       Bool{Value: false},
			expectation: false,
		},
		"unknown": {
			input:       Bool{Unknown: true},
			expectation: tftypes.UnknownValue,
		},
		"null": {
			input:       Bool{Null: true},
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
		input       Bool
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"true-true": {
			input:       Bool{Value: true},
			candidate:   Bool{Value: true},
			expectation: true,
		},
		"true-false": {
			input:       Bool{Value: true},
			candidate:   Bool{Value: false},
			expectation: false,
		},
		"true-unknown": {
			input:       Bool{Value: true},
			candidate:   Bool{Unknown: true},
			expectation: false,
		},
		"true-null": {
			input:       Bool{Value: true},
			candidate:   Bool{Null: true},
			expectation: false,
		},
		"true-wrongType": {
			input:       Bool{Value: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"true-nil": {
			input:       Bool{Value: true},
			candidate:   nil,
			expectation: false,
		},
		"false-true": {
			input:       Bool{Value: false},
			candidate:   Bool{Value: true},
			expectation: false,
		},
		"false-false": {
			input:       Bool{Value: false},
			candidate:   Bool{Value: false},
			expectation: true,
		},
		"false-unknown": {
			input:       Bool{Value: false},
			candidate:   Bool{Unknown: true},
			expectation: false,
		},
		"false-null": {
			input:       Bool{Value: false},
			candidate:   Bool{Null: true},
			expectation: false,
		},
		"false-wrongType": {
			input:       Bool{Value: false},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"false-nil": {
			input:       Bool{Value: false},
			candidate:   nil,
			expectation: false,
		},
		"unknown-true": {
			input:       Bool{Unknown: true},
			candidate:   Bool{Value: true},
			expectation: false,
		},
		"unknown-false": {
			input:       Bool{Unknown: true},
			candidate:   Bool{Value: false},
			expectation: false,
		},
		"unknown-unknown": {
			input:       Bool{Unknown: true},
			candidate:   Bool{Unknown: true},
			expectation: true,
		},
		"unknown-null": {
			input:       Bool{Unknown: true},
			candidate:   Bool{Null: true},
			expectation: false,
		},
		"unknown-wrongType": {
			input:       Bool{Unknown: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"unknown-nil": {
			input:       Bool{Unknown: true},
			candidate:   nil,
			expectation: false,
		},
		"null-true": {
			input:       Bool{Null: true},
			candidate:   Bool{Value: true},
			expectation: false,
		},
		"null-false": {
			input:       Bool{Null: true},
			candidate:   Bool{Value: false},
			expectation: false,
		},
		"null-unknown": {
			input:       Bool{Null: true},
			candidate:   Bool{Unknown: true},
			expectation: false,
		},
		"null-null": {
			input:       Bool{Null: true},
			candidate:   Bool{Null: true},
			expectation: true,
		},
		"null-wrongType": {
			input:       Bool{Null: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"null-nil": {
			input:       Bool{Null: true},
			candidate:   nil,
			expectation: false,
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

func TestBoolMarshalJSON(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Bool
		expectation []byte
	}
	tests := map[string]testCase{
		"unknown produces null": {
			input:       Bool{Unknown: true},
			expectation: []byte("null"),
		},
		"null produces null": {
			input:       Bool{Null: true},
			expectation: []byte("null"),
		},
		"false produces false": {
			input:       Bool{Value: false},
			expectation: []byte("false"),
		},
		"true produces true": {
			input:       Bool{Value: true},
			expectation: []byte("true"),
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := test.input.MarshalJSON()
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %v, got %v", test.expectation, got)
			}
		})
	}
}

func TestBoolUnmarshalJSON(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       []byte
		expectation Bool
	}
	tests := map[string]testCase{
		"null produces null": {
			input:       []byte("null"),
			expectation: Bool{Null: true},
		},
		"false produces false": {
			input:       []byte("false"),
			expectation: Bool{Value: false},
		},
		"true produces true": {
			input:       []byte("true"),
			expectation: Bool{Value: true},
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var got Bool
			err := got.UnmarshalJSON(test.input)
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %v, got %v", test.expectation, got)
			}
		})
	}
}
