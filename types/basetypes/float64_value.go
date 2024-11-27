// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package basetypes

import (
	"context"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/refinement"
	tfrefinements "github.com/hashicorp/terraform-plugin-go/tftypes/refinement"
)

var (
	_ Float64Valuable                   = Float64Value{}
	_ Float64ValuableWithSemanticEquals = Float64Value{}
	_ attr.ValueWithNotNullRefinement   = Float64Value{}
)

// Float64Valuable extends attr.Value for float64 value types.
// Implement this interface to create a custom Float64 value type.
type Float64Valuable interface {
	attr.Value

	// ToFloat64Value should convert the value type to a Float64.
	ToFloat64Value(ctx context.Context) (Float64Value, diag.Diagnostics)
}

// Float64ValuableWithSemanticEquals extends Float64Valuable with semantic
// equality logic.
type Float64ValuableWithSemanticEquals interface {
	Float64Valuable

	// Float64SemanticEquals should return true if the given value is
	// semantically equal to the current value. This logic is used to prevent
	// Terraform data consistency errors and resource drift where a value change
	// may have inconsequential differences, such as rounding.
	//
	// Only known values are compared with this method as changing a value's
	// state implicitly represents a different value.
	Float64SemanticEquals(context.Context, Float64Valuable) (bool, diag.Diagnostics)
}

// Float64Null creates a Float64 with a null value. Determine whether the value is
// null via the Float64 type IsNull method.
func NewFloat64Null() Float64Value {
	return Float64Value{
		state: attr.ValueStateNull,
	}
}

// Float64Unknown creates a Float64 with an unknown value. Determine whether the
// value is unknown via the Float64 type IsUnknown method.
//
// Setting the deprecated Float64 type Null, Unknown, or Value fields after
// creating a Float64 with this function has no effect.
func NewFloat64Unknown() Float64Value {
	return Float64Value{
		state: attr.ValueStateUnknown,
	}
}

// Float64Value creates a Float64 with a known value. Access the value via the Float64
// type ValueFloat64 method. Passing a value of `NaN` will result in a panic.
//
// Setting the deprecated Float64 type Null, Unknown, or Value fields after
// creating a Float64 with this function has no effect.
func NewFloat64Value(value float64) Float64Value {
	return Float64Value{
		state: attr.ValueStateKnown,
		value: big.NewFloat(value),
	}
}

// NewFloat64PointerValue creates a Float64 with a null value if nil or a known
// value. Access the value via the Float64 type ValueFloat64Pointer method.
// Passing a value of `NaN` will result in a panic.
func NewFloat64PointerValue(value *float64) Float64Value {
	if value == nil {
		return NewFloat64Null()
	}

	return NewFloat64Value(*value)
}

// Float64Value represents a 64-bit floating point value, exposed as a float64.
type Float64Value struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value *big.Float

	// refinements represents the unknown value refinement data associated with this Value.
	// This field is only populated for unknown values.
	refinements refinement.Refinements
}

// Float64SemanticEquals returns true if the given Float64Value is semantically equal to the current Float64Value.
// The underlying value *big.Float can store more precise float values then the Go built-in float64 type. This only
// compares the precision of the value that can be represented as the Go built-in float64, which is 53 bits of precision.
func (f Float64Value) Float64SemanticEquals(ctx context.Context, newValuable Float64Valuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(Float64Value)
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", f)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	return f.ValueFloat64() == newValue.ValueFloat64(), diags
}

// Equal returns true if `other` is a Float64 and has the same value as `f`.
func (f Float64Value) Equal(other attr.Value) bool {
	o, ok := other.(Float64Value)

	if !ok {
		return false
	}

	if f.state != o.state {
		return false
	}

	if len(f.refinements) != len(o.refinements) {
		return false
	}

	if len(f.refinements) > 0 && !f.refinements.Equal(o.refinements) {
		return false
	}

	if f.state != attr.ValueStateKnown {
		return true
	}

	// Not possible to create a known Float64Value with a nil value, but check anyways
	if f.value == nil || o.value == nil {
		return f.value == o.value
	}

	return f.value.Cmp(o.value) == 0
}

// ToTerraformValue returns the data contained in the Float64 as a tftypes.Value.
func (f Float64Value) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	switch f.state {
	case attr.ValueStateKnown:
		if err := tftypes.ValidateValue(tftypes.Number, f.value); err != nil {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.Number, f.value), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(tftypes.Number, nil), nil
	case attr.ValueStateUnknown:
		if len(f.refinements) == 0 {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), nil
		}

		unknownValRefinements := make(tfrefinements.Refinements, 0)
		for _, refn := range f.refinements {
			switch refnVal := refn.(type) {
			case refinement.NotNull:
				unknownValRefinements[tfrefinements.KeyNullness] = tfrefinements.NewNullness(false)
			case refinement.Float64LowerBound:
				lowerBound := big.NewFloat(refnVal.LowerBound())
				unknownValRefinements[tfrefinements.KeyNumberLowerBound] = tfrefinements.NewNumberLowerBound(lowerBound, refnVal.IsInclusive())
			case refinement.Float64UpperBound:
				upperBound := big.NewFloat(refnVal.UpperBound())
				unknownValRefinements[tfrefinements.KeyNumberUpperBound] = tfrefinements.NewNumberUpperBound(upperBound, refnVal.IsInclusive())
			}
		}
		unknownVal := tftypes.NewValue(tftypes.Number, tftypes.UnknownValue)

		return unknownVal.Refine(unknownValRefinements), nil
	default:
		panic(fmt.Sprintf("unhandled Float64 state in ToTerraformValue: %s", f.state))
	}
}

// Type returns a Float64Type.
func (f Float64Value) Type(ctx context.Context) attr.Type {
	return Float64Type{}
}

// IsNull returns true if the Float64 represents a null value.
func (f Float64Value) IsNull() bool {
	return f.state == attr.ValueStateNull
}

// IsUnknown returns true if the Float64 represents a currently unknown value.
func (f Float64Value) IsUnknown() bool {
	return f.state == attr.ValueStateUnknown
}

// String returns a human-readable representation of the Float64 value.
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (f Float64Value) String() string {
	if f.IsUnknown() {
		if len(f.refinements) == 0 {
			return attr.UnknownValueString
		}

		return fmt.Sprintf("<unknown, %s>", f.refinements.String())
	}

	if f.IsNull() {
		return attr.NullValueString
	}

	f64 := f.ValueFloat64()
	return fmt.Sprintf("%f", f64)
}

// ValueFloat64 returns the known float64 value. If Float64 is null or unknown, returns
// 0.0.
func (f Float64Value) ValueFloat64() float64 {
	if f.IsNull() || f.IsUnknown() {
		return float64(0.0)
	}

	f64, _ := f.value.Float64()
	return f64
}

// ValueFloat64Pointer returns a pointer to the known float64 value, nil for a
// null value, or a pointer to 0.0 for an unknown value.
func (f Float64Value) ValueFloat64Pointer() *float64 {
	if f.IsNull() {
		return nil
	}

	if f.IsUnknown() {
		f64 := float64(0.0)
		return &f64
	}

	f64, _ := f.value.Float64()
	return &f64
}

// ToFloat64Value returns Float64.
func (f Float64Value) ToFloat64Value(context.Context) (Float64Value, diag.Diagnostics) {
	return f, nil
}

// RefineAsNotNull will return an unknown Float64Value that includes a value refinement that:
//   - Indicates the float64 value will not be null once it becomes known.
//
// If the provided Float64Value is null or known, then the Float64Value will be returned unchanged.
func (f Float64Value) RefineAsNotNull() Float64Value {
	if !f.IsUnknown() {
		return f
	}

	newRefinements := make(refinement.Refinements, len(f.refinements))
	for i, refn := range f.refinements {
		newRefinements[i] = refn
	}

	newRefinements[refinement.KeyNotNull] = refinement.NewNotNull()

	newUnknownVal := NewFloat64Unknown()
	newUnknownVal.refinements = newRefinements

	return newUnknownVal
}

// RefineWithLowerBound will return an unknown Float64Value that includes a value refinement that:
//   - Indicates the float64 value will not be null once it becomes known.
//   - Indicates the float64 value will not be less than the float64 provided (lowerBound) once it becomes known.
//
// If the provided Float64Value is null or known, then the Float64Value will be returned unchanged.
func (f Float64Value) RefineWithLowerBound(lowerBound float64, inclusive bool) Float64Value {
	if !f.IsUnknown() {
		return f
	}

	newRefinements := make(refinement.Refinements, len(f.refinements))
	for i, refn := range f.refinements {
		newRefinements[i] = refn
	}

	newRefinements[refinement.KeyNotNull] = refinement.NewNotNull()
	newRefinements[refinement.KeyNumberLowerBound] = refinement.NewFloat64LowerBound(lowerBound, inclusive)

	newUnknownVal := NewFloat64Unknown()
	newUnknownVal.refinements = newRefinements

	return newUnknownVal
}

// RefineWithUpperBound will return an unknown Float64Value that includes a value refinement that:
//   - Indicates the float64 value will not be null once it becomes known.
//   - Indicates the float64 value will not be greater than the float64 provided (upperBound) once it becomes known.
//
// If the provided Float64Value is null or known, then the Float64Value will be returned unchanged.
func (f Float64Value) RefineWithUpperBound(upperBound float64, inclusive bool) Float64Value {
	if !f.IsUnknown() {
		return f
	}

	newRefinements := make(refinement.Refinements, len(f.refinements))
	for i, refn := range f.refinements {
		newRefinements[i] = refn
	}

	newRefinements[refinement.KeyNotNull] = refinement.NewNotNull()
	newRefinements[refinement.KeyNumberUpperBound] = refinement.NewFloat64UpperBound(upperBound, inclusive)

	newUnknownVal := NewFloat64Unknown()
	newUnknownVal.refinements = newRefinements

	return newUnknownVal
}

// NotNullRefinement returns value refinement data and a boolean indicating if a NotNull refinement
// exists on the given Float64Value. If an Float64Value contains a NotNull refinement, this indicates that
// the float64 value is unknown, but the eventual known value will not be null.
//
// A NotNull value refinement can be added to an unknown value via the `RefineAsNotNull` method.
func (f Float64Value) NotNullRefinement() (*refinement.NotNull, bool) {
	if !f.IsUnknown() {
		return nil, false
	}

	refn, ok := f.refinements[refinement.KeyNotNull]
	if !ok {
		return nil, false
	}

	notNullRefn, ok := refn.(refinement.NotNull)
	if !ok {
		return nil, false
	}

	return &notNullRefn, true
}

// LowerBoundRefinement returns value refinement data and a boolean indicating if a Float64LowerBound refinement
// exists on the given Float64Value. If an Float64Value contains a Float64LowerBound refinement, this indicates that
// the float64 value is unknown, but the eventual known value will not be less than the specified float64 value
// (either inclusive or exclusive) once it becomes known. The returned boolean should be checked before accessing
// refinement data.
//
// An Float64LowerBound value refinement can be added to an unknown value via the `RefineWithLowerBound` method.
func (f Float64Value) LowerBoundRefinement() (*refinement.Float64LowerBound, bool) {
	if !f.IsUnknown() {
		return nil, false
	}

	refn, ok := f.refinements[refinement.KeyNumberLowerBound]
	if !ok {
		return nil, false
	}

	lowerBoundRefn, ok := refn.(refinement.Float64LowerBound)
	if !ok {
		return nil, false
	}

	return &lowerBoundRefn, true
}

// UpperBoundRefinement returns value refinement data and a boolean indicating if a Float64UpperBound refinement
// exists on the given Float64Value. If an Float64Value contains a Float64UpperBound refinement, this indicates that
// the float64 value is unknown, but the eventual known value will not be greater than the specified float64 value
// (either inclusive or exclusive) once it becomes known. The returned boolean should be checked before accessing
// refinement data.
//
// A Float64UpperBound value refinement can be added to an unknown value via the `RefineWithUpperBound` method.
func (f Float64Value) UpperBoundRefinement() (*refinement.Float64UpperBound, bool) {
	if !f.IsUnknown() {
		return nil, false
	}

	refn, ok := f.refinements[refinement.KeyNumberUpperBound]
	if !ok {
		return nil, false
	}

	upperBoundRefn, ok := refn.(refinement.Float64UpperBound)
	if !ok {
		return nil, false
	}

	return &upperBoundRefn, true
}
