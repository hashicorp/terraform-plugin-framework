package types

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestStringValueFromTerraform(t *testing.T) {
	t.Parallel()

	testStringValueFromTerraform(t, true)
}

func testStringValueFromTerraform(t *testing.T, direct bool) {
	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"true": {
			input:       tftypes.NewValue(tftypes.String, "hello"),
			expectation: String{Value: "hello"},
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
			expectation: String{Unknown: true},
		},
		"null": {
			input:       tftypes.NewValue(tftypes.String, nil),
			expectation: String{Null: true},
		},
		"wrongType": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectedErr: "can't unmarshal tftypes.Number into *string, expected string",
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			f := StringType.ValueFromTerraform
			if direct {
				f = stringValueFromTerraform
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

func TestStringToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       String
		expectation interface{}
	}
	tests := map[string]testCase{
		"value": {
			input:       String{Value: "hello"},
			expectation: tftypes.NewValue(tftypes.String, "hello"),
		},
		"unknown": {
			input:       String{Unknown: true},
			expectation: tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
		},
		"null": {
			input:       String{Null: true},
			expectation: tftypes.NewValue(tftypes.String, nil),
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

func TestStringEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       String
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"value-value": {
			input:       String{Value: "hello"},
			candidate:   String{Value: "hello"},
			expectation: true,
		},
		"value-diff": {
			input:       String{Value: "hello"},
			candidate:   String{Value: "world"},
			expectation: false,
		},
		"value-unknown": {
			input:       String{Value: "hello"},
			candidate:   String{Unknown: true},
			expectation: false,
		},
		"value-null": {
			input:       String{Value: "hello"},
			candidate:   String{Null: true},
			expectation: false,
		},
		"value-wrongType": {
			input:       String{Value: "hello"},
			candidate:   Number{Value: big.NewFloat(123)},
			expectation: false,
		},
		"value-nil": {
			input:       String{Value: "hello"},
			candidate:   nil,
			expectation: false,
		},
		"unknown-value": {
			input:       String{Unknown: true},
			candidate:   String{Value: "hello"},
			expectation: false,
		},
		"unknown-unknown": {
			input:       String{Unknown: true},
			candidate:   String{Unknown: true},
			expectation: true,
		},
		"unknown-null": {
			input:       String{Unknown: true},
			candidate:   String{Null: true},
			expectation: false,
		},
		"unknown-wrongType": {
			input:       String{Unknown: true},
			candidate:   Number{Value: big.NewFloat(123)},
			expectation: false,
		},
		"unknown-nil": {
			input:       String{Unknown: true},
			candidate:   nil,
			expectation: false,
		},
		"null-value": {
			input:       String{Null: true},
			candidate:   String{Value: "hello"},
			expectation: false,
		},
		"null-unknown": {
			input:       String{Null: true},
			candidate:   String{Unknown: true},
			expectation: false,
		},
		"null-null": {
			input:       String{Null: true},
			candidate:   String{Null: true},
			expectation: true,
		},
		"null-wrongType": {
			input:       String{Null: true},
			candidate:   Number{Value: big.NewFloat(123)},
			expectation: false,
		},
		"null-nil": {
			input:       String{Null: true},
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
