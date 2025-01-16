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
	tfrefinement "github.com/hashicorp/terraform-plugin-go/tftypes/refinement"
)

var (
	_ NumberValuable                  = NumberValue{}
	_ attr.ValueWithNotNullRefinement = NumberValue{}
)

// NumberValuable extends attr.Value for number value types.
// Implement this interface to create a custom Number value type.
type NumberValuable interface {
	attr.Value

	// ToNumberValue should convert the value type to a Number.
	ToNumberValue(ctx context.Context) (NumberValue, diag.Diagnostics)
}

// NumberValuableWithSemanticEquals extends NumberValuable with semantic
// equality logic.
type NumberValuableWithSemanticEquals interface {
	NumberValuable

	// NumberSemanticEquals should return true if the given value is
	// semantically equal to the current value. This logic is used to prevent
	// Terraform data consistency errors and resource drift where a value change
	// may have inconsequential differences, such as rounding.
	//
	// Only known values are compared with this method as changing a value's
	// state implicitly represents a different value.
	NumberSemanticEquals(context.Context, NumberValuable) (bool, diag.Diagnostics)
}

// NewNumberNull creates a Number with a null value. Determine whether the value is
// null via the Number type IsNull method.
func NewNumberNull() NumberValue {
	return NumberValue{
		state: attr.ValueStateNull,
	}
}

// NewNumberUnknown creates a Number with an unknown value. Determine whether the
// value is unknown via the Number type IsUnknown method.
func NewNumberUnknown() NumberValue {
	return NumberValue{
		state: attr.ValueStateUnknown,
	}
}

// NewNumberValue creates a Number with a known value. Access the value via the Number
// type ValueBigFloat method. If the given value is nil, a null Number is created.
func NewNumberValue(value *big.Float) NumberValue {
	if value == nil {
		return NewNumberNull()
	}

	return NumberValue{
		state: attr.ValueStateKnown,
		value: value,
	}
}

// NumberValue represents a number value, exposed as a *big.Float. Numbers can be
// floats or integers.
type NumberValue struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value *big.Float

	// refinements represents the unknown value refinement data associated with this Value.
	// This field is only populated for unknown values.
	refinements refinement.Refinements
}

// Type returns a NumberType.
func (n NumberValue) Type(_ context.Context) attr.Type {
	return NumberType{}
}

// ToTerraformValue returns the data contained in the Number as a tftypes.Value.
func (n NumberValue) ToTerraformValue(_ context.Context) (tftypes.Value, error) {
	switch n.state {
	case attr.ValueStateKnown:
		if n.value == nil {
			return tftypes.NewValue(tftypes.Number, nil), nil
		}

		if err := tftypes.ValidateValue(tftypes.Number, n.value); err != nil {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(tftypes.Number, n.value), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(tftypes.Number, nil), nil
	case attr.ValueStateUnknown:
		if len(n.refinements) == 0 {
			return tftypes.NewValue(tftypes.Number, tftypes.UnknownValue), nil
		}

		unknownValRefinements := make(tfrefinement.Refinements, 0)
		for _, refn := range n.refinements {
			switch refnVal := refn.(type) {
			case refinement.NotNull:
				unknownValRefinements[tfrefinement.KeyNullness] = tfrefinement.NewNullness(false)
			case refinement.NumberLowerBound:
				unknownValRefinements[tfrefinement.KeyNumberLowerBound] = tfrefinement.NewNumberLowerBound(refnVal.LowerBound(), refnVal.IsInclusive())
			case refinement.NumberUpperBound:
				unknownValRefinements[tfrefinement.KeyNumberUpperBound] = tfrefinement.NewNumberUpperBound(refnVal.UpperBound(), refnVal.IsInclusive())
			}
		}
		unknownVal := tftypes.NewValue(tftypes.Number, tftypes.UnknownValue)

		return unknownVal.Refine(unknownValRefinements), nil
	default:
		panic(fmt.Sprintf("unhandled Number state in ToTerraformValue: %s", n.state))
	}
}

// Equal returns true if `other` is a Number and has the same value as `n`.
func (n NumberValue) Equal(other attr.Value) bool {
	o, ok := other.(NumberValue)

	if !ok {
		return false
	}

	if n.state != o.state {
		return false
	}

	if len(n.refinements) != len(o.refinements) {
		return false
	}

	if len(n.refinements) > 0 && !n.refinements.Equal(o.refinements) {
		return false
	}

	if n.state != attr.ValueStateKnown {
		return true
	}

	return n.value.Cmp(o.value) == 0
}

// IsNull returns true if the Number represents a null value.
func (n NumberValue) IsNull() bool {
	return n.state == attr.ValueStateNull
}

// IsUnknown returns true if the Number represents a currently unknown value.
func (n NumberValue) IsUnknown() bool {
	return n.state == attr.ValueStateUnknown
}

// String returns a human-readable representation of the Number value.
// The string returned here is not protected by any compatibility guarantees,
// and is intended for logging and error reporting.
func (n NumberValue) String() string {
	if n.IsUnknown() {
		if len(n.refinements) == 0 {
			return attr.UnknownValueString
		}

		return fmt.Sprintf("<unknown, %s>", n.refinements.String())
	}

	if n.IsNull() {
		return attr.NullValueString
	}

	return n.value.String()
}

// ValueBigFloat returns the known *big.Float value. If Number is null or unknown, returns
// 0.0.
func (n NumberValue) ValueBigFloat() *big.Float {
	return n.value
}

// ToNumberValue returns Number.
func (n NumberValue) ToNumberValue(context.Context) (NumberValue, diag.Diagnostics) {
	return n, nil
}

// RefineAsNotNull will return an unknown NumberValue that includes a value refinement that:
//   - Indicates the number value will not be null once it becomes known.
//
// If the provided NumberValue is null or known, then the NumberValue will be returned unchanged.
func (n NumberValue) RefineAsNotNull() NumberValue {
	if !n.IsUnknown() {
		return n
	}

	newRefinements := make(refinement.Refinements, len(n.refinements))
	for i, refn := range n.refinements {
		newRefinements[i] = refn
	}

	newRefinements[refinement.KeyNotNull] = refinement.NewNotNull()

	newUnknownVal := NewNumberUnknown()
	newUnknownVal.refinements = newRefinements

	return newUnknownVal
}

// RefineWithLowerBound will return an unknown NumberValue that includes a value refinement that:
//   - Indicates the number value will not be null once it becomes known.
//   - Indicates the number value will not be less than the number provided (lowerBound) once it becomes known.
//
// If the provided NumberValue is null or known, then the NumberValue will be returned unchanged.
func (n NumberValue) RefineWithLowerBound(lowerBound *big.Float, inclusive bool) NumberValue {
	if !n.IsUnknown() {
		return n
	}

	newRefinements := make(refinement.Refinements, len(n.refinements))
	for i, refn := range n.refinements {
		newRefinements[i] = refn
	}

	newRefinements[refinement.KeyNotNull] = refinement.NewNotNull()
	newRefinements[refinement.KeyNumberLowerBound] = refinement.NewNumberLowerBound(lowerBound, inclusive)

	newUnknownVal := NewNumberUnknown()
	newUnknownVal.refinements = newRefinements

	return newUnknownVal
}

// RefineWithUpperBound will return an unknown NumberValue that includes a value refinement that:
//   - Indicates the number value will not be null once it becomes known.
//   - Indicates the number value will not be greater than the number provided (upperBound) once it becomes known.
//
// If the provided NumberValue is null or known, then the NumberValue will be returned unchanged.
func (n NumberValue) RefineWithUpperBound(upperBound *big.Float, inclusive bool) NumberValue {
	if !n.IsUnknown() {
		return n
	}

	newRefinements := make(refinement.Refinements, len(n.refinements))
	for i, refn := range n.refinements {
		newRefinements[i] = refn
	}

	newRefinements[refinement.KeyNotNull] = refinement.NewNotNull()
	newRefinements[refinement.KeyNumberUpperBound] = refinement.NewNumberUpperBound(upperBound, inclusive)

	newUnknownVal := NewNumberUnknown()
	newUnknownVal.refinements = newRefinements

	return newUnknownVal
}

// NotNullRefinement returns value refinement data and a boolean indicating if a NotNull refinement
// exists on the given NumberValue. If an NumberValue contains a NotNull refinement, this indicates that
// the number value is unknown, but the eventual known value will not be null.
//
// A NotNull value refinement can be added to an unknown value via the `RefineAsNotNull` method.
func (n NumberValue) NotNullRefinement() (*refinement.NotNull, bool) {
	if !n.IsUnknown() {
		return nil, false
	}

	refn, ok := n.refinements[refinement.KeyNotNull]
	if !ok {
		return nil, false
	}

	notNullRefn, ok := refn.(refinement.NotNull)
	if !ok {
		return nil, false
	}

	return &notNullRefn, true
}

// LowerBoundRefinement returns value refinement data and a boolean indicating if a NumberLowerBound refinement
// exists on the given NumberValue. If an NumberValue contains a NumberLowerBound refinement, this indicates that
// the number value is unknown, but the eventual known value will not be less than the specified number value
// (either inclusive or exclusive) once it becomes known. The returned boolean should be checked before accessing
// refinement data.
//
// An NumberLowerBound value refinement can be added to an unknown value via the `RefineWithLowerBound` method.
func (n NumberValue) LowerBoundRefinement() (*refinement.NumberLowerBound, bool) {
	if !n.IsUnknown() {
		return nil, false
	}

	refn, ok := n.refinements[refinement.KeyNumberLowerBound]
	if !ok {
		return nil, false
	}

	lowerBoundRefn, ok := refn.(refinement.NumberLowerBound)
	if !ok {
		return nil, false
	}

	return &lowerBoundRefn, true
}

// UpperBoundRefinement returns value refinement data and a boolean indicating if a NumberUpperBound refinement
// exists on the given NumberValue. If an NumberValue contains a NumberUpperBound refinement, this indicates that
// the number value is unknown, but the eventual known value will not be greater than the specified number value
// (either inclusive or exclusive) once it becomes known. The returned boolean should be checked before accessing
// refinement data.
//
// A NumberUpperBound value refinement can be added to an unknown value via the `RefineWithUpperBound` method.
func (n NumberValue) UpperBoundRefinement() (*refinement.NumberUpperBound, bool) {
	if !n.IsUnknown() {
		return nil, false
	}

	refn, ok := n.refinements[refinement.KeyNumberUpperBound]
	if !ok {
		return nil, false
	}

	upperBoundRefn, ok := refn.(refinement.NumberUpperBound)
	if !ok {
		return nil, false
	}

	return &upperBoundRefn, true
}
