// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestFloat64TypeValidate(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		in       tftypes.Value
		expected diag.Diagnostics
	}{
		"zero-float": {
			in:       tftypes.NewValue(tftypes.Number, big.NewFloat(0.0)),
			expected: nil,
		},
		"negative-integer": {
			in:       tftypes.NewValue(tftypes.Number, big.NewFloat(-123)),
			expected: nil,
		},
		"positive-integer": {
			in:       tftypes.NewValue(tftypes.Number, big.NewFloat(123)),
			expected: nil,
		},
		"positive-float": {
			in:       tftypes.NewValue(tftypes.Number, big.NewFloat(123.45)),
			expected: nil,
		},
		"negative-float": {
			in:       tftypes.NewValue(tftypes.Number, big.NewFloat(123.45)),
			expected: nil,
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/613
		"zero-string-float": {
			in:       tftypes.NewValue(tftypes.Number, testMustParseFloat("0.0")),
			expected: nil,
		},
		"positive-string-float": {
			in:       tftypes.NewValue(tftypes.Number, testMustParseFloat("123.2")),
			expected: nil,
		},
		"negative-string-float": {
			in:       tftypes.NewValue(tftypes.Number, testMustParseFloat("-123.2")),
			expected: nil,
		},
		// Reference: https://pkg.go.dev/math/big#Float.Float64
		// Reference: https://pkg.go.dev/math#pkg-constants
		"SmallestNonzeroFloat64": {
			in:       tftypes.NewValue(tftypes.Number, big.NewFloat(math.SmallestNonzeroFloat64)),
			expected: nil,
		},
		"SmallestNonzeroFloat64-below": {
			in: tftypes.NewValue(tftypes.Number, testMustParseFloat("4.9406564584124654417656879286822137236505980e-325")),
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Float64 Type Validation Error",
					fmt.Sprintf("Value %s cannot be represented as a 64-bit floating point.", testMustParseFloat("4.9406564584124654417656879286822137236505980e-325")),
				),
			},
		},
		// Reference: https://pkg.go.dev/math/big#Float.Float64
		// Reference: https://pkg.go.dev/math#pkg-constants
		"MaxFloat64": {
			in:       tftypes.NewValue(tftypes.Number, big.NewFloat(math.MaxFloat64)),
			expected: nil,
		},
		"MaxFloat64-above": {
			in: tftypes.NewValue(tftypes.Number, testMustParseFloat("1.79769313486231570814527423731704356798070e+309")),
			expected: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Float64 Type Validation Error",
					fmt.Sprintf("Value %s cannot be represented as a 64-bit floating point.", testMustParseFloat("1.79769313486231570814527423731704356798070e+309")),
				),
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := Float64Type{}.Validate(context.Background(), testCase.in, path.Root("test"))

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64TypeValueFromTerraform(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       tftypes.Value
		expectation attr.Value
		expectedErr string
	}
	tests := map[string]testCase{
		"value-int": {
			input:       tftypes.NewValue(tftypes.Number, 123),
			expectation: NewFloat64Value(123.0),
		},
		"value-float": {
			input:       tftypes.NewValue(tftypes.Number, 123.456),
			expectation: NewFloat64Value(123.456),
		},
		"unknown": {
			input:       tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
			expectation: NewFloat64Unknown(),
		},
		"null": {
			input:       tftypes.NewValue(tftypes.Number, nil),
			expectation: NewFloat64Null(),
		},
		"wrongType": {
			input:       tftypes.NewValue(tftypes.String, "oops"),
			expectedErr: "can't unmarshal tftypes.String into *big.Float, expected *big.Float",
		},
		// Reference: https://github.com/hashicorp/terraform-plugin-framework/issues/647
		"zero-string-float": {
			input:       tftypes.NewValue(tftypes.Number, testMustParseFloat("0.0")),
			expectation: NewFloat64Value(0.0),
		},
		"positive-string-float": {
			input:       tftypes.NewValue(tftypes.Number, testMustParseFloat("123.2")),
			expectation: NewFloat64Value(123.2),
		},
		"negative-string-float": {
			input:       tftypes.NewValue(tftypes.Number, testMustParseFloat("-123.2")),
			expectation: NewFloat64Value(-123.2),
		},
		// Reference: https://pkg.go.dev/math/big#Float.Float64
		// Reference: https://pkg.go.dev/math#pkg-constants
		"SmallestNonzeroFloat64": {
			input:       tftypes.NewValue(tftypes.Number, big.NewFloat(math.SmallestNonzeroFloat64)),
			expectation: NewFloat64Value(math.SmallestNonzeroFloat64),
		},
		"SmallestNonzeroFloat64-below": {
			input:       tftypes.NewValue(tftypes.Number, testMustParseFloat("4.9406564584124654417656879286822137236505980e-325")),
			expectedErr: fmt.Sprintf("Value %s cannot be represented as a 64-bit floating point.", testMustParseFloat("4.9406564584124654417656879286822137236505980e-325")),
		},
		// Reference: https://pkg.go.dev/math/big#Float.Float64
		// Reference: https://pkg.go.dev/math#pkg-constants
		"MaxFloat64": {
			input:       tftypes.NewValue(tftypes.Number, big.NewFloat(math.MaxFloat64)),
			expectation: NewFloat64Value(math.MaxFloat64),
		},
		"MaxFloat64-above": {
			input:       tftypes.NewValue(tftypes.Number, testMustParseFloat("1.79769313486231570814527423731704356798070e+309")),
			expectedErr: fmt.Sprintf("Value %s cannot be represented as a 64-bit floating point.", testMustParseFloat("1.79769313486231570814527423731704356798070e+309")),
		},
	}
	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			got, err := Float64Type{}.ValueFromTerraform(ctx, test.input)
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
