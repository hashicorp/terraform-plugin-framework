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

// testMustParseFloat parses a string into a *big.Float similar to cty and
// tftypes logic or panics on any error.
//
// Reference: https://github.com/hashicorp/go-cty/blob/85980079f637862fa8e43ddc82dd74315e2f4c85/cty/value_init.go#L49
// Reference: https://github.com/hashicorp/terraform-plugin-go/blob/c593d2e0da8d2258b2a22af867c39842a0cb89f7/tftypes/value_msgpack.go#L108
func testMustParseFloat(s string) *big.Float {
	f, _, err := big.ParseFloat(s, 10, 512, big.ToNearestEven)

	if err != nil {
		panic(err)
	}

	return f
}

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

func TestFloat64ValueToTerraformValue(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float64Value
		expectation interface{}
	}
	tests := map[string]testCase{
		"known-int": {
			input:       NewFloat64Value(123),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.0)),
		},
		"known-float": {
			input:       NewFloat64Value(123.456),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.456)),
		},
		"unknown": {
			input:       NewFloat64Unknown(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"null": {
			input:       NewFloat64Null(),
			expectation: tftypes.NewValue(tftypes.Number, nil),
		},
		"deprecated-value-int": {
			input:       NewFloat64Value(123),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.0)),
		},
		"deprecated-value-float": {
			input:       NewFloat64Value(123.456),
			expectation: tftypes.NewValue(tftypes.Number, big.NewFloat(123.456)),
		},
		"deprecated-unknown": {
			input:       NewFloat64Unknown(),
			expectation: tftypes.NewValue(tftypes.Number, tftypes.UnknownValue),
		},
		"deprecated-null": {
			input:       NewFloat64Null(),
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

func TestFloat64ValueEqual(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float64Value
		candidate   attr.Value
		expectation bool
	}
	tests := map[string]testCase{
		"known-known-same": {
			input:       NewFloat64Value(123),
			candidate:   NewFloat64Value(123),
			expectation: true,
		},
		"known-known-diff": {
			input:       NewFloat64Value(123),
			candidate:   NewFloat64Value(456),
			expectation: false,
		},
		"known-unknown": {
			input:       NewFloat64Value(123),
			candidate:   NewFloat64Unknown(),
			expectation: false,
		},
		"known-null": {
			input:       NewFloat64Value(123),
			candidate:   NewFloat64Null(),
			expectation: false,
		},
		"unknown-value": {
			input:       NewFloat64Unknown(),
			candidate:   NewFloat64Value(123),
			expectation: false,
		},
		"unknown-unknown": {
			input:       NewFloat64Unknown(),
			candidate:   NewFloat64Unknown(),
			expectation: true,
		},
		"unknown-null": {
			input:       NewFloat64Unknown(),
			candidate:   NewFloat64Null(),
			expectation: false,
		},
		"null-known": {
			input:       NewFloat64Null(),
			candidate:   NewFloat64Value(123),
			expectation: false,
		},
		"null-unknown": {
			input:       NewFloat64Null(),
			candidate:   NewFloat64Unknown(),
			expectation: false,
		},
		"null-null": {
			input:       NewFloat64Null(),
			candidate:   NewFloat64Null(),
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

func TestFloat64ValueIsNull(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float64Value
		expected bool
	}{
		"known": {
			input:    NewFloat64Value(2.4),
			expected: false,
		},
		"null": {
			input:    NewFloat64Null(),
			expected: true,
		},
		"unknown": {
			input:    NewFloat64Unknown(),
			expected: false,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.IsNull()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64ValueIsUnknown(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float64Value
		expected bool
	}{
		"known": {
			input:    NewFloat64Value(2.4),
			expected: false,
		},
		"null": {
			input:    NewFloat64Null(),
			expected: false,
		},
		"unknown": {
			input:    NewFloat64Unknown(),
			expected: true,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.IsUnknown()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestFloat64ValueString(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input       Float64Value
		expectation string
	}
	tests := map[string]testCase{
		"less-than-one": {
			input:       NewFloat64Value(0.12340984302980000),
			expectation: "0.123410",
		},
		"more-than-one": {
			input:       NewFloat64Value(92387938173219.327663),
			expectation: "92387938173219.328125",
		},
		"negative-more-than-one": {
			input:       NewFloat64Value(-0.12340984302980000),
			expectation: "-0.123410",
		},
		"negative-less-than-one": {
			input:       NewFloat64Value(-92387938173219.327663),
			expectation: "-92387938173219.328125",
		},
		"min-float64": {
			input:       NewFloat64Value(math.SmallestNonzeroFloat64),
			expectation: "0.000000",
		},
		"max-float64": {
			input:       NewFloat64Value(math.MaxFloat64),
			expectation: "179769313486231570814527423731704356798070567525844996598917476803157260780028538760589558632766878171540458953514382464234321326889464182768467546703537516986049910576551282076245490090389328944075868508455133942304583236903222948165808559332123348274797826204144723168738177180919299881250404026184124858368.000000",
		},
		"unknown": {
			input:       NewFloat64Unknown(),
			expectation: "<unknown>",
		},
		"null": {
			input:       NewFloat64Null(),
			expectation: "<null>",
		},
		"zero-value": {
			input:       Float64Value{},
			expectation: "<null>",
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

func TestFloat64ValueValueFloat64(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input    Float64Value
		expected float64
	}{
		"known": {
			input:    NewFloat64Value(2.4),
			expected: 2.4,
		},
		"null": {
			input:    NewFloat64Null(),
			expected: 0.0,
		},
		"unknown": {
			input:    NewFloat64Unknown(),
			expected: 0.0,
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.input.ValueFloat64()

			if diff := cmp.Diff(got, testCase.expected); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}
