package types

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func numberComparer(i, j *big.Float) bool {
	return (i == nil && j == nil) || (i != nil && j != nil && i.Cmp(j) == 0)
}

func TestNumberValueFromTerraform(t *testing.T) {
	t.Parallel()

	testNumberValueFromTerraform(t, true)
}

func testNumberValueFromTerraform(t *testing.T, direct bool) {
	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectation: Number{Value: big.NewFloat(123)},
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			expectation: Number{Unknown: true},
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Number, nil),
			expectation: Number{Null: true},
		},
		"wrongType": {
			input:       tftypes.NewValue(tftypes.String, "oops"),
			expectedErr: "can't unmarshal tftypes.String into *big.Float, expected *big.Float",
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			f := NumberType.ValueFromTerraform
			if direct {
				f = numberValueFromTerraform
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

func TestNumberToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Number
		expectation tftypes.Value
	}
	tests := map[string]testCase{
		"value": {
			input:       Number{Value: big.NewFloat(123)},
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123)),
		},
		"value-nil": {
			input:       Number{Value: nil},
			expectation: tftypes.NewValue(tftypes.Number, nil),
		},
		"unknown": {
			input:       Number{Unknown: true},
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"null": {
			input:       Number{Null: true},
			expectation: tftypes.NewValue(tftypes.Number, nil),
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
			if !cmp.Equal(got, test.expectation, cmp.Comparer(numberComparer)) {
				t.Errorf("Expected %+v, got %+v", test.expectation, got)
			}
		})
	}
}

func TestNumberEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Number
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"value-value-same": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   Number{Value: big.NewFloat(123)},
			expectation: true,
		},
		"value-value-diff": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   Number{Value: big.NewFloat(456)},
			expectation: false,
		},
		"value-unknown": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   Number{Unknown: true},
			expectation: false,
		},
		"value-null": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   Number{Null: true},
			expectation: false,
		},
		"value-wrongType": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"value-nil": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   nil,
			expectation: false,
		},
		"value-nilValue": {
			input:       Number{Value: big.NewFloat(123)},
			candidate:   Number{Value: nil},
			expectation: false,
		},
		"unknown-value": {
			input:       Number{Unknown: true},
			candidate:   Number{Value: big.NewFloat(123)},
			expectation: false,
		},
		"unknown-unknown": {
			input:       Number{Unknown: true},
			candidate:   Number{Unknown: true},
			expectation: true,
		},
		"unknown-null": {
			input:       Number{Unknown: true},
			candidate:   Number{Null: true},
			expectation: false,
		},
		"unknown-wrongType": {
			input:       Number{Unknown: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"unknown-nil": {
			input:       Number{Unknown: true},
			candidate:   nil,
			expectation: false,
		},
		"unknown-nilValue": {
			input:       Number{Unknown: true},
			candidate:   Number{Value: nil},
			expectation: false,
		},
		"null-value": {
			input:       Number{Null: true},
			candidate:   Number{Value: big.NewFloat(123)},
			expectation: false,
		},
		"null-unknown": {
			input:       Number{Null: true},
			candidate:   Number{Unknown: true},
			expectation: false,
		},
		"null-null": {
			input:       Number{Null: true},
			candidate:   Number{Null: true},
			expectation: true,
		},
		"null-wrongType": {
			input:       Number{Null: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"null-nil": {
			input:       Number{Null: true},
			candidate:   nil,
			expectation: false,
		},
		"null-nilValue": {
			input:       Number{Null: true},
			candidate:   Number{Value: nil},
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
