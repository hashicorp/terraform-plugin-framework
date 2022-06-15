package types

import (
	"context"
	"math"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestFloat64ValueFromTerraform(t *testing.T) {
	t.Parallel()

	testFloat64ValueFromTerraform(t, true)
}

func testFloat64ValueFromTerraform(t *testing.T, direct bool) {
	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value-int": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectation: Float64{Value: 123.0},
		},
		"value-float": {
			input:       tftypes.NewValue(tftypes.Number, 123.456),
			expectation: Float64{Value: 123.456},
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			expectation: Float64{Unknown: true},
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Number, nil),
			expectation: Float64{Null: true},
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

			f := Float64Type.ValueFromTerraform
			if direct {
				f = float64ValueFromTerraform
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
			if diff := cmp.Diff(got, test.expectation); diff != "" {
				t.Errorf("Unexpected response (+wanted, -got): %s", diff)
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

func TestFloat64ToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float64
		expectation interface{}
	}
	tests := map[string]testCase{
		"value-int": {
			input:       Float64{Value: 123},
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.0)),
		},
		"value-float": {
			input:       Float64{Value: 123.456},
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.456)),
		},
		"unknown": {
			input:       Float64{Unknown: true},
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"null": {
			input:       Float64{Null: true},
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

func TestFloat64Equal(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float64
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"value-value-same": {
			input:       Float64{Value: 123},
			candidate:   Float64{Value: 123},
			expectation: true,
		},
		"value-value-diff": {
			input:       Float64{Value: 123},
			candidate:   Float64{Value: 456},
			expectation: false,
		},
		"value-unknown": {
			input:       Float64{Value: 123},
			candidate:   Float64{Unknown: true},
			expectation: false,
		},
		"value-null": {
			input:       Float64{Value: 123},
			candidate:   Float64{Null: true},
			expectation: false,
		},
		"value-wrongType": {
			input:       Float64{Value: 123},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"value-nil": {
			input:       Float64{Value: 123},
			candidate:   nil,
			expectation: false,
		},
		"unknown-value": {
			input:       Float64{Unknown: true},
			candidate:   Float64{Value: 123},
			expectation: false,
		},
		"unknown-unknown": {
			input:       Float64{Unknown: true},
			candidate:   Float64{Unknown: true},
			expectation: true,
		},
		"unknown-null": {
			input:       Float64{Unknown: true},
			candidate:   Float64{Null: true},
			expectation: false,
		},
		"unknown-wrongType": {
			input:       Float64{Unknown: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"unknown-nil": {
			input:       Float64{Unknown: true},
			candidate:   nil,
			expectation: false,
		},
		"null-value": {
			input:       Float64{Null: true},
			candidate:   Float64{Value: 123},
			expectation: false,
		},
		"null-unknown": {
			input:       Float64{Null: true},
			candidate:   Float64{Unknown: true},
			expectation: false,
		},
		"null-null": {
			input:       Float64{Null: true},
			candidate:   Float64{Null: true},
			expectation: true,
		},
		"null-wrongType": {
			input:       Float64{Null: true},
			candidate:   &String{Value: "oops"},
			expectation: false,
		},
		"null-nil": {
			input:       Float64{Null: true},
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

func TestFloat64String(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float64
		expectation string
	}
	tests := map[string]testCase{
		"less-than-one": {
			input:       Float64{Value: 0.12340984302980000},
			expectation: "0.123410",
		},
		"more-than-one": {
			input:       Float64{Value: 92387938173219.327663},
			expectation: "92387938173219.328125",
		},
		"negative-more-than-one": {
			input:       Float64{Value: -0.12340984302980000},
			expectation: "-0.123410",
		},
		"negative-less-than-one": {
			input:       Float64{Value: -92387938173219.327663},
			expectation: "-92387938173219.328125",
		},
		"min-float64": {
			input:       Float64{Value: math.SmallestNonzeroFloat64},
			expectation: "0.000000",
		},
		"max-float64": {
			input:       Float64{Value: math.MaxFloat64},
			expectation: "179769313486231570814527423731704356798070567525844996598917476803157260780028538760589558632766878171540458953514382464234321326889464182768467546703537516986049910576551282076245490090389328944075868508455133942304583236903222948165808559332123348274797826204144723168738177180919299881250404026184124858368.000000",
		},
		"unknown": {
			input:       Float64{Unknown: true},
			expectation: "<unknown>",
		},
		"null": {
			input:       Float64{Null: true},
			expectation: "<null>",
		},
		"default-0": {
			input:       Float64{},
			expectation: "0.000000",
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.input.String()
			if !cmp.Equal(got, test.expectation) {
				t.Errorf("Expected %q, got %q", test.expectation, got)
			}
		})
	}
}
