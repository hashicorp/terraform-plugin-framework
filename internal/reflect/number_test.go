// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package reflect_test

import (
	"context"
	"math"
	"math/big"
	"reflect"
	"strconv"
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

var (
	overflowInt, _, _            = big.ParseFloat("9223372036854775808", 10, 53, big.ToPositiveInf)
	overflowUint, _, _           = big.ParseFloat("18446744073709551616", 10, 53, big.ToPositiveInf)
	overflowFloat, _, _          = big.ParseFloat("1e10000", 10, 53, big.ToPositiveInf)
	overflowNegativeFloat, _, _  = big.ParseFloat("-1e10000", 10, 53, big.ToPositiveInf)
	underflowInt, _, _           = big.ParseFloat("-9223372036854775809", 10, 53, big.ToNegativeInf)
	underflowFloat, _, _         = big.ParseFloat("1e-1000", 10, 0, big.ToNegativeInf)
	underflowNegativeFloat, _, _ = big.ParseFloat("-1e-1000", 10, 0, big.ToNegativeInf)
)

func TestNumber_bigFloat(t *testing.T) {
	t.Parallel()

	var f *big.Float

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, 123456), reflect.ValueOf(f), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&f).Elem().Set(result)
	if f == nil {
		t.Error("Expected value, got nil")
		return
	}
	if f.Cmp(big.NewFloat(123456)) != 0 {
		t.Errorf("Expected %v, got %v", big.NewFloat(123456), f)
	}
}

func TestNumber_bigInt(t *testing.T) {
	t.Parallel()

	var n *big.Int

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, 123456), reflect.ValueOf(n), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n == nil {
		t.Error("Expected value, got nil")
		return
	}
	if n.Cmp(big.NewInt(123456)) != 0 {
		t.Errorf("Expected %v, got %v", big.NewInt(123456), n)
	}
}

func TestNumber_bigIntRounded(t *testing.T) {
	t.Parallel()

	var n *big.Int

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, 123456.123), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n == nil {
		t.Error("Expected value, got nil")
		return
	}
	if n.Cmp(big.NewInt(123456)) != 0 {
		t.Errorf("Expected %v, got %v", big.NewInt(123456), n)
	}
}

func TestNumber_bigIntRoundingError(t *testing.T) {
	t.Parallel()

	var n *big.Int
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 123456.123 in *big.Int",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, 123456.123), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_int(t *testing.T) {
	t.Parallel()

	var n int

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, 123), reflect.ValueOf(n), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != 123 {
		t.Errorf("Expected %v, got %v", 123, n)
	}
}

func TestNumber_intOverflow(t *testing.T) {
	t.Parallel()

	var n int

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, overflowInt), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if strconv.IntSize == 64 && n != math.MaxInt64 {
		t.Errorf("Expected %v, got %v", math.MaxInt64, n)
	} else if strconv.IntSize == 32 && n != math.MaxInt32 {
		t.Errorf("Expected %v, got %v", math.MaxInt32, n)
	}
}

func TestNumber_intOverflowError(t *testing.T) {
	t.Parallel()

	var n int
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store "+overflowInt.String()+" in int",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, overflowInt), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_intUnderflow(t *testing.T) {
	t.Parallel()

	var n int

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, underflowInt), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if strconv.IntSize == 64 && n != math.MinInt64 {
		t.Errorf("Expected %v, got %v", math.MinInt64, n)
	} else if strconv.IntSize == 32 && n != math.MinInt32 {
		t.Errorf("Expected %v, got %v", math.MinInt32, n)
	}
}

func TestNumber_intUnderflowError(t *testing.T) {
	t.Parallel()

	var n int
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store "+underflowInt.String()+" in int",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, underflowInt), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_int8(t *testing.T) {
	t.Parallel()

	var n int8

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, 123), reflect.ValueOf(n), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != 123 {
		t.Errorf("Expected %v, got %v", 123, n)
	}
}

func TestNumber_int8Overflow(t *testing.T) {
	t.Parallel()

	var n int8

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxInt8+1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MaxInt8 {
		t.Errorf("Expected %v, got %v", math.MaxInt8, n)
	}
}

func TestNumber_int8OverflowError(t *testing.T) {
	t.Parallel()

	var n int8
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 128 in int8",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxInt8+1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_int8Underflow(t *testing.T) {
	t.Parallel()

	var n int8

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MinInt8-1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MinInt8 {
		t.Errorf("Expected %v, got %v", math.MinInt8, n)
	}
}

func TestNumber_int8UnderflowError(t *testing.T) {
	t.Parallel()

	var n int8
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store -129 in int8",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MinInt8-1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_int16(t *testing.T) {
	t.Parallel()
}

func TestNumber_int16Overflow(t *testing.T) {
	t.Parallel()

	var n int16

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxInt16+1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MaxInt16 {
		t.Errorf("Expected %v, got %v", math.MaxInt16, n)
	}
}

func TestNumber_int16OverflowError(t *testing.T) {
	t.Parallel()

	var n int16
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 32768 in int16",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxInt16+1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_int16Underflow(t *testing.T) {
	t.Parallel()

	var n int16

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MinInt16-1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MinInt16 {
		t.Errorf("Expected %v, got %v", math.MinInt16, n)
	}
}

func TestNumber_int16UnderflowError(t *testing.T) {
	t.Parallel()

	var n int16
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store -32769 in int16",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MinInt16-1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_int32(t *testing.T) {
	t.Parallel()
}

func TestNumber_int32Overflow(t *testing.T) {
	t.Parallel()

	var n int32

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxInt32+1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MaxInt32 {
		t.Errorf("Expected %v, got %v", math.MaxInt32, n)
	}
}

func TestNumber_int32OverflowError(t *testing.T) {
	t.Parallel()

	var n int32
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 2147483648 in int32",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxInt32+1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_int32Underflow(t *testing.T) {
	t.Parallel()

	var n int32

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MinInt32-1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MinInt32 {
		t.Errorf("Expected %v, got %v", math.MinInt32, n)
	}
}

func TestNumber_int32UnderflowError(t *testing.T) {
	t.Parallel()

	var n int32
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store -2147483649 in int32",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MinInt32-1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_int64(t *testing.T) {
	t.Parallel()

	var n int64

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, 123), reflect.ValueOf(n), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != 123 {
		t.Errorf("Expected %v, got %v", 123, n)
	}
}

func TestNumber_int64Overflow(t *testing.T) {
	t.Parallel()

	var n int64

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, overflowInt), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MaxInt64 {
		t.Errorf("Expected %v, got %v", math.MaxInt64, n)
	}
}

func TestNumber_int64OverflowError(t *testing.T) {
	t.Parallel()

	var n int64
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 9.223372037e+18 in int64",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, overflowInt), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_int64Underflow(t *testing.T) {
	t.Parallel()

	var n int64

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, underflowInt), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MinInt64 {
		t.Errorf("Expected %v, got %v", math.MinInt64, n)
	}
}

func TestNumber_int64UnderflowError(t *testing.T) {
	t.Parallel()

	var n int64
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store -9.223372037e+18 in int64",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, underflowInt), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_uint(t *testing.T) {
	t.Parallel()

	var n uint

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, 123), reflect.ValueOf(n), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != 123 {
		t.Errorf("Expected %v, got %v", 123, n)
	}
}

func TestNumber_uintOverflow(t *testing.T) {
	t.Parallel()

	var n uint

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, overflowUint), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if strconv.IntSize == 64 && n != math.MaxUint64 {
		t.Errorf("Expected %v, got %v", uint64(math.MaxUint64), n)
	} else if strconv.IntSize == 32 && n != math.MaxUint32 {
		t.Errorf("Expected %v, got %v", math.MaxUint32, n)
	}
}

func TestNumber_uintOverflowError(t *testing.T) {
	t.Parallel()

	var n uint
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store "+overflowUint.String()+" in uint",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, overflowUint), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_uintUnderflow(t *testing.T) {
	t.Parallel()

	var n uint

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, -1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != 0 {
		t.Errorf("Expected %v, got %v", 0, n)
	}
}

func TestNumber_uintUnderflowError(t *testing.T) {
	t.Parallel()

	var n uint
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store -1 in uint",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, -1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_uint8(t *testing.T) {
	t.Parallel()

	var n uint8

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, 123), reflect.ValueOf(n), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != 123 {
		t.Errorf("Expected %v, got %v", 123, n)
	}
}

func TestNumber_uint8Overflow(t *testing.T) {
	t.Parallel()

	var n uint8

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxUint8+1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MaxUint8 {
		t.Errorf("Expected %v, got %v", math.MaxUint8, n)
	}
}

func TestNumber_uint8OverflowError(t *testing.T) {
	t.Parallel()

	var n uint8
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 256 in uint8",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxUint8+1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_uint8Underflow(t *testing.T) {
	t.Parallel()

	var n uint8

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, -1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != 0 {
		t.Errorf("Expected %v, got %v", 0, n)
	}
}

func TestNumber_uint8UnderflowError(t *testing.T) {
	t.Parallel()

	var n uint8
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store -1 in uint8",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, -1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_uint16(t *testing.T) {
	t.Parallel()
}

func TestNumber_uint16Overflow(t *testing.T) {
	t.Parallel()

	var n uint16

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxUint16+1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MaxUint16 {
		t.Errorf("Expected %v, got %v", math.MaxUint16, n)
	}
}

func TestNumber_uint16OverflowError(t *testing.T) {
	t.Parallel()

	var n uint16
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 65536 in uint16",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxUint16+1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_uint16Underflow(t *testing.T) {
	t.Parallel()

	var n uint16

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, -1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != 0 {
		t.Errorf("Expected %v, got %v", 0, n)
	}
}

func TestNumber_uint16UnderflowError(t *testing.T) {
	t.Parallel()

	var n uint16
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store -1 in uint16",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, -1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_uint32(t *testing.T) {
	t.Parallel()
}

func TestNumber_uint32Overflow(t *testing.T) {
	t.Parallel()

	var n uint32

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxUint32+1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MaxUint32 {
		t.Errorf("Expected %v, got %v", math.MaxUint32, n)
	}
}

func TestNumber_uint32OverflowError(t *testing.T) {
	t.Parallel()

	var n uint32
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 4294967296 in uint32",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxUint32+1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_uint32Underflow(t *testing.T) {
	t.Parallel()

	var n uint32

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, -1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != 0 {
		t.Errorf("Expected %v, got %v", 0, n)
	}
}

func TestNumber_uint32UnderflowError(t *testing.T) {
	t.Parallel()

	var n uint32
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store -1 in uint32",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, -1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_uint64(t *testing.T) {
	t.Parallel()

	var n uint64

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, 123), reflect.ValueOf(n), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != 123 {
		t.Errorf("Expected %v, got %v", 123, n)
	}
}

func TestNumber_uint64Overflow(t *testing.T) {
	t.Parallel()

	var n uint64

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, overflowUint), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MaxUint64 {
		t.Errorf("Expected %v, got %v", uint64(math.MaxUint64), n)
	}
}

func TestNumber_uint64OverflowError(t *testing.T) {
	t.Parallel()

	var n uint64
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 1.844674407e+19 in uint64",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, overflowUint), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_uint64Underflow(t *testing.T) {
	t.Parallel()

	var n uint64

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, -1), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != 0 {
		t.Errorf("Expected %v, got %v", 0, n)
	}
}

func TestNumber_uint64UnderflowError(t *testing.T) {
	t.Parallel()

	var n uint64
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store -1 in uint64",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, -1), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_float32(t *testing.T) {
	t.Parallel()
}

func TestNumber_float32Overflow(t *testing.T) {
	t.Parallel()

	var n float32

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxFloat64), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MaxFloat32 {
		t.Errorf("Expected %v, got %v", math.MaxFloat32, n)
	}
}

func TestNumber_float32OverflowError(t *testing.T) {
	t.Parallel()

	var n float32
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 1.797693135e+308 in float32",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.MaxFloat64), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_float32Underflow(t *testing.T) {
	t.Parallel()

	var n float32

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.SmallestNonzeroFloat64), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.SmallestNonzeroFloat32 {
		t.Errorf("Expected %v, got %v", math.SmallestNonzeroFloat32, n)
	}
}

func TestNumber_float32UnderflowError(t *testing.T) {
	t.Parallel()

	var n float32
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 4.940656458e-324 in float32",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, math.SmallestNonzeroFloat64), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_float64(t *testing.T) {
	t.Parallel()

	var n float64

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, 123), reflect.ValueOf(n), refl.Options{}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != 123 {
		t.Errorf("Expected %v, got %v", 123, n)
	}
}

func TestNumber_float64Overflow(t *testing.T) {
	t.Parallel()

	var n float64

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, overflowFloat), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.MaxFloat64 {
		t.Errorf("Expected %v, got %v", math.MaxFloat64, n)
	}
}

func TestNumber_float64OverflowError(t *testing.T) {
	t.Parallel()

	var n float64
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 1e+10000 in float64",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, overflowFloat), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_float64OverflowNegative(t *testing.T) {
	t.Parallel()

	var n float64

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, overflowNegativeFloat), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != -math.MaxFloat64 {
		t.Errorf("Expected %v, got %v", -math.MaxFloat64, n)
	}
}

func TestNumber_float64OverflowNegativeError(t *testing.T) {
	t.Parallel()

	var n float64
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store -1e+10000 in float64",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, overflowNegativeFloat), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_float64Underflow(t *testing.T) {
	t.Parallel()

	var n float64

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, underflowFloat), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != math.SmallestNonzeroFloat64 {
		t.Errorf("Expected %v, got %v", math.SmallestNonzeroFloat64, n)
	}
}

func TestNumber_float64UnderflowError(t *testing.T) {
	t.Parallel()

	var n float64
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store 1e-1000 in float64",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, underflowFloat), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestNumber_float64UnderflowNegative(t *testing.T) {
	t.Parallel()

	var n float64

	result, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, underflowNegativeFloat), reflect.ValueOf(n), refl.Options{
		AllowRoundingNumbers: true,
	}, path.Empty())
	if diags.HasError() {
		t.Errorf("Unexpected error: %v", diags)
	}
	reflect.ValueOf(&n).Elem().Set(result)
	if n != -math.SmallestNonzeroFloat64 {
		t.Errorf("Expected %v, got %v", -math.SmallestNonzeroFloat64, n)
	}
}

func TestNumber_float64UnderflowNegativeError(t *testing.T) {
	t.Parallel()

	var n float64
	expectedDiags := diag.Diagnostics{
		diag.NewAttributeErrorDiagnostic(
			path.Empty(),
			"Value Conversion Error",
			"An unexpected error was encountered trying to convert to number. This is always an error in the provider. Please report the following to the provider developer:\n\ncannot store -1e-1000 in float64",
		),
	}

	_, diags := refl.Number(context.Background(), types.NumberType, tftypes.NewValue(tftypes.Number, underflowNegativeFloat), reflect.ValueOf(n), refl.Options{}, path.Empty())

	if diff := cmp.Diff(diags, expectedDiags); diff != "" {
		t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
	}
}

func TestFromInt(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		val           int64
		typ           attr.Type
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"0": {
			val:      0,
			typ:      types.NumberType,
			expected: types.NumberValue(big.NewFloat(0)),
		},
		"1": {
			val:      1,
			typ:      types.NumberType,
			expected: types.NumberValue(big.NewFloat(1)),
		},
		"WithValidateWarning": {
			val: 1,
			typ: testtypes.NumberTypeWithValidateWarning{},
			expected: testtypes.Number{
				Number:    types.NumberValue(big.NewFloat(1)),
				CreatedBy: testtypes.NumberType{},
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"WithValidateError": {
			val: 1,
			typ: testtypes.NumberTypeWithValidateError{},
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
	}

	for name, tc := range cases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actualVal, diags := refl.FromInt(context.Background(), tc.typ, tc.val, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if !diags.HasError() && !tc.expected.Equal(actualVal) {
				t.Fatalf("fail: got %+v, wanted %+v", actualVal, tc.expected)
			}
		})
	}
}

func TestFromUint(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		val           uint64
		typ           attr.Type
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"0": {
			val:      0,
			typ:      types.NumberType,
			expected: types.NumberValue(big.NewFloat(0)),
		},
		"1": {
			val:      1,
			typ:      types.NumberType,
			expected: types.NumberValue(big.NewFloat(1)),
		},
		"WithValidateWarning": {
			val: 1,
			typ: testtypes.NumberTypeWithValidateWarning{},
			expected: testtypes.Number{
				Number:    types.NumberValue(big.NewFloat(1)),
				CreatedBy: testtypes.NumberType{},
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"WithValidateError": {
			val: 1,
			typ: testtypes.NumberTypeWithValidateError{},
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
	}

	for name, tc := range cases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actualVal, diags := refl.FromUint(context.Background(), tc.typ, tc.val, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if !diags.HasError() && !tc.expected.Equal(actualVal) {
				t.Fatalf("fail: got %+v, wanted %+v", actualVal, tc.expected)
			}
		})
	}
}

func TestFromFloat(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		val           float64
		typ           attr.Type
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"0": {
			val:      0,
			typ:      types.NumberType,
			expected: types.NumberValue(big.NewFloat(0)),
		},
		"1": {
			val:      1,
			typ:      types.NumberType,
			expected: types.NumberValue(big.NewFloat(1)),
		},
		"1.234": {
			val:      1.234,
			typ:      types.NumberType,
			expected: types.NumberValue(big.NewFloat(1.234)),
		},
		"WithValidateWarning": {
			val: 1,
			typ: testtypes.NumberTypeWithValidateWarning{},
			expected: testtypes.Number{
				Number:    types.NumberValue(big.NewFloat(1)),
				CreatedBy: testtypes.NumberType{},
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"WithValidateError": {
			val: 1,
			typ: testtypes.NumberTypeWithValidateError{},
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
	}

	for name, tc := range cases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actualVal, diags := refl.FromFloat(context.Background(), tc.typ, tc.val, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if !diags.HasError() && !tc.expected.Equal(actualVal) {
				t.Fatalf("fail: got %+v, wanted %+v", actualVal, tc.expected)
			}
		})
	}
}

func TestFromBigFloat(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		val           *big.Float
		typ           attr.Type
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"0": {
			val:      big.NewFloat(0),
			typ:      types.NumberType,
			expected: types.NumberValue(big.NewFloat(0)),
		},
		"1": {
			val:      big.NewFloat(1),
			typ:      types.NumberType,
			expected: types.NumberValue(big.NewFloat(1)),
		},
		"1.234": {
			val:      big.NewFloat(1.234),
			typ:      types.NumberType,
			expected: types.NumberValue(big.NewFloat(1.234)),
		},
		"WithValidateWarning": {
			val: big.NewFloat(1),
			typ: testtypes.NumberTypeWithValidateWarning{},
			expected: testtypes.Number{
				Number:    types.NumberValue(big.NewFloat(1)),
				CreatedBy: testtypes.NumberType{},
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"WithValidateError": {
			val: big.NewFloat(1),
			typ: testtypes.NumberTypeWithValidateError{},
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
	}

	for name, tc := range cases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actualVal, diags := refl.FromBigFloat(context.Background(), tc.typ, tc.val, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if !diags.HasError() && !tc.expected.Equal(actualVal) {
				t.Fatalf("fail: got %+v, wanted %+v", actualVal, tc.expected)
			}
		})
	}
}

func TestFromBigInt(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		val           *big.Int
		typ           attr.Type
		expected      attr.Value
		expectedDiags diag.Diagnostics
	}{
		"0": {
			val:      big.NewInt(0),
			typ:      types.NumberType,
			expected: types.NumberValue(big.NewFloat(0)),
		},
		"1": {
			val:      big.NewInt(1),
			typ:      types.NumberType,
			expected: types.NumberValue(big.NewFloat(1)),
		},
		"WithValidateWarning": {
			val: big.NewInt(1),
			typ: testtypes.NumberTypeWithValidateWarning{},
			expected: testtypes.Number{
				Number:    types.NumberValue(big.NewFloat(1)),
				CreatedBy: testtypes.NumberTypeWithValidateWarning{},
			},
			expectedDiags: diag.Diagnostics{
				testtypes.TestWarningDiagnostic(path.Empty()),
			},
		},
		"WithValidateError": {
			val: big.NewInt(1),
			typ: testtypes.NumberTypeWithValidateError{},
			expectedDiags: diag.Diagnostics{
				testtypes.TestErrorDiagnostic(path.Empty()),
			},
		},
	}

	for name, tc := range cases {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actualVal, diags := refl.FromBigInt(context.Background(), tc.typ, tc.val, path.Empty())

			if diff := cmp.Diff(diags, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics (+wanted, -got): %s", diff)
			}

			if !diags.HasError() && !tc.expected.Equal(actualVal) {
				t.Fatalf("fail: got %+v, wanted %+v", actualVal, tc.expected)
			}
		})
	}
}
