package types

import (
	"context"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestInt64ValueFromTerraform(t *testing.T) {
	t.Parallel()

	testInt64ValueFromTerraform(t, true)
}

func testInt64ValueFromTerraform(t *testing.T, direct bool) {
	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectation: Int64{Value: 123},
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			expectation: Int64{Unknown: true},
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Number, nil),
			expectation: Int64{Null: true},
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

			f := Int64Type.ValueFromTerraform
			if direct {
				f = int64ValueFromTerraform
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
			if test.expectation.IsNull() != test.input.IsNull() {
				t.Errorf("Expected null-ness match: expected %t, got %t", test.expectation.IsNull(), test.input.IsNull())
			}
			if test.expectation.IsUnknown() != !test.input.IsKnown() {
				t.Errorf("Expected unknown-ness match: expected %t, got %t", test.expectation.IsUnknown(), !test.input.IsKnown())
			}
		})
	}
}

func TestInt64ToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Int64
		expectation interface{}
	}
	tests := map[string]testCase{
		"value": {
			input:       Int64{Value: 123},
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123)),
		},
		"unknown": {
			input:       Int64{Unknown: true},
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"null": {
			input:       Int64{Null: true},
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

func TestInt64Equal(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Int64
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"value-value-same": {
			input:       Int64{Value: 123},
			candidate:   Int64{Value: 123},
			expectation: true,
		},
		"value-value-diff": {
			input:       Int64{Value: 123},
			candidate:   Int64{Value: 456},
			expectation: false,
		},
		"value-unknown": {
			input:       Int64{Value: 123},
			candidate:   Int64{Unknown: true},
			expectation: false,
		},
		"value-null": {
			input:       Int64{Value: 123},
			candidate:   Int64{Null: true},
			expectation: false,
		},
		"value-wrongType": {
			input:       Int64{Value: 123},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"value-nil": {
			input:       Int64{Value: 123},
			candidate:   nil,
			expectation: false,
		},
		"unknown-value": {
			input:       Int64{Unknown: true},
			candidate:   Int64{Value: 123},
			expectation: false,
		},
		"unknown-unknown": {
			input:       Int64{Unknown: true},
			candidate:   Int64{Unknown: true},
			expectation: true,
		},
		"unknown-null": {
			input:       Int64{Unknown: true},
			candidate:   Int64{Null: true},
			expectation: false,
		},
		"unknown-wrongType": {
			input:       Int64{Unknown: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"unknown-nil": {
			input:       Int64{Unknown: true},
			candidate:   nil,
			expectation: false,
		},
		"null-value": {
			input:       Int64{Null: true},
			candidate:   Int64{Value: 123},
			expectation: false,
		},
		"null-unknown": {
			input:       Int64{Null: true},
			candidate:   Int64{Unknown: true},
			expectation: false,
		},
		"null-null": {
			input:       Int64{Null: true},
			candidate:   Int64{Null: true},
			expectation: true,
		},
		"null-wrongType": {
			input:       Int64{Null: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"null-nil": {
			input:       Int64{Null: true},
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
